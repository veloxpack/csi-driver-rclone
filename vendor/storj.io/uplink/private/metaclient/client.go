// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package metaclient

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"

	"storj.io/common/errs2"
	"storj.io/common/macaroon"
	"storj.io/common/pb"
	"storj.io/common/rpc"
	"storj.io/common/rpc/rpcstatus"
	"storj.io/common/signing"
	"storj.io/common/storj"
	"storj.io/uplink/internal"
	"storj.io/uplink/private/eestream"
	"storj.io/uplink/private/metaclient/cursed"
)

var (
	mon = monkit.Package()

	// Error is the errs class of standard metainfo errors.
	Error = errs.Class("metaclient")

	zstdDecoder = func() *zstd.Decoder {
		decoder, err := zstd.NewReader(nil,
			zstd.WithDecoderMaxMemory(64<<20),
		)
		if err != nil {
			panic(err)
		}
		return decoder
	}()
)

// Client creates a grpcClient.
type Client struct {
	mu        sync.Mutex
	conn      *rpc.Conn
	client    pb.DRPCMetainfoClient
	apiKeyRaw []byte

	opts DialNodeURLOpts
}

// DialNodeURLOpts contains options for DialNodeURLWithOpts.
type DialNodeURLOpts struct {
	UserAgent       string
	SatelliteSigner signing.Signer
}

// DialNodeURLWithOpts is like DialNodeURL but takes options.
func DialNodeURLWithOpts(ctx context.Context, dialer rpc.Dialer, nodeURL string, apiKey *macaroon.APIKey, opts DialNodeURLOpts) (*Client, error) {
	url, err := storj.ParseNodeURL(nodeURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	if url.ID.IsZero() {
		return nil, Error.New("node ID is required in node URL %q", nodeURL)
	}

	conn, err := dialer.DialNode(ctx, url, rpc.DialOptions{ForceTCPFastOpenMultidialSupport: true})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &Client{
		conn:      conn,
		client:    pb.NewDRPCMetainfoClient(conn),
		apiKeyRaw: apiKey.SerializeRaw(),

		opts: opts,
	}, nil
}

// DialNodeURL dials to metainfo endpoint with the specified api key.
func DialNodeURL(ctx context.Context, dialer rpc.Dialer, nodeURL string, apiKey *macaroon.APIKey, userAgent string) (*Client, error) {
	return DialNodeURLWithOpts(ctx, dialer, nodeURL, apiKey, DialNodeURLOpts{
		UserAgent: userAgent,
	})
}

// Close closes the dialed connection.
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.conn != nil {
		err := client.conn.Close()
		client.conn = nil
		return Error.Wrap(err)
	}

	return nil
}

func (client *Client) header() *pb.RequestHeader {
	return &pb.RequestHeader{
		ApiKey:    client.apiKeyRaw,
		UserAgent: []byte(client.opts.UserAgent),
	}
}

// GetProjectInfo gets the ProjectInfo for the api key associated with the metainfo client.
func (client *Client) GetProjectInfo(ctx context.Context) (response *pb.ProjectInfoResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.ProjectInfo(ctx, &pb.ProjectInfoRequest{
			Header: client.header(),
		})
		return err
	})
	return response, err
}

// AccountLicensesRequest contains parameters for AccountLicenses method.
type AccountLicensesRequest struct {
	Type       string
	BucketName string
}

func (params *AccountLicensesRequest) toRequest(header *pb.RequestHeader) *pb.AccountLicensesRequest {
	return &pb.AccountLicensesRequest{
		Header:     header,
		Type:       params.Type,
		BucketName: params.BucketName,
	}
}

// Licenses represents license information.
type Licenses struct {
	Type      string
	ExpiresAt time.Time
	Key       []byte
}

// AccountLicenses retrieves license information.
func (client *Client) AccountLicenses(ctx context.Context, params AccountLicensesRequest) (licenses []Licenses, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.AccountLicensesResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.AccountLicenses(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	for _, l := range response.Licenses {
		licenses = append(licenses, Licenses{
			Type:      l.Type,
			ExpiresAt: l.ExpiresAt,
			Key:       l.Key,
		})
	}

	return licenses, nil
}

// BeginObjectParams parameters for BeginObject method.
type BeginObjectParams struct {
	Bucket               []byte
	EncryptedObjectKey   []byte
	Redundancy           storj.RedundancyScheme
	EncryptionParameters storj.EncryptionParameters
	ExpiresAt            time.Time

	EncryptedUserData

	Retention Retention
	LegalHold bool

	IfNoneMatch []string
}

func (params *BeginObjectParams) toRequest(header *pb.RequestHeader) *pb.ObjectBeginRequest {
	req := &pb.ObjectBeginRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		ExpiresAt:          params.ExpiresAt,
		RedundancyScheme: &pb.RedundancyScheme{
			Type:             pb.RedundancyScheme_SchemeType(params.Redundancy.Algorithm),
			ErasureShareSize: params.Redundancy.ShareSize,
			MinReq:           int32(params.Redundancy.RequiredShares),
			RepairThreshold:  int32(params.Redundancy.RepairShares),
			SuccessThreshold: int32(params.Redundancy.OptimalShares),
			Total:            int32(params.Redundancy.TotalShares),
		},
		EncryptionParameters: &pb.EncryptionParameters{
			CipherSuite: pb.CipherSuite(params.EncryptionParameters.CipherSuite),
			BlockSize:   int64(params.EncryptionParameters.BlockSize),
		},

		EncryptedMetadata:             params.EncryptedMetadata,
		EncryptedMetadataEncryptedKey: params.EncryptedMetadataEncryptedKey,
		EncryptedMetadataNonce:        params.EncryptedMetadataNonce,
		EncryptedEtag:                 params.EncryptedETag,
		LegalHold:                     params.LegalHold,
	}

	if params.Retention != (Retention{}) {
		req.Retention = &pb.Retention{
			Mode:        pb.Retention_Mode(params.Retention.Mode),
			RetainUntil: params.Retention.RetainUntil,
		}
	}

	return req
}

// BatchItem returns single item for batch request.
func (params *BeginObjectParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectBegin{
			ObjectBegin: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *BeginObjectParams) IsRetriable() bool { return false }

// BeginObjectResponse response for BeginObject request.
type BeginObjectResponse struct {
	StreamID storj.StreamID
}

func newBeginObjectResponse(response *pb.ObjectBeginResponse) BeginObjectResponse {
	return BeginObjectResponse{
		StreamID: response.StreamId,
	}
}

// BeginObject begins object creation.
func (client *Client) BeginObject(ctx context.Context, params BeginObjectParams) (_ BeginObjectResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.ObjectBeginResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.BeginObject(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return BeginObjectResponse{}, Error.Wrap(err)
	}

	return newBeginObjectResponse(response), nil
}

// CommitObjectParams parameters for CommitObject method.
type CommitObjectParams struct {
	StreamID storj.StreamID

	EncryptedUserData

	IfNoneMatch []string
}

func (params *CommitObjectParams) toRequest(header *pb.RequestHeader) *pb.ObjectCommitRequest {
	return &pb.ObjectCommitRequest{
		Header:                        header,
		StreamId:                      params.StreamID,
		EncryptedMetadataNonce:        params.EncryptedMetadataNonce,
		EncryptedMetadata:             params.EncryptedMetadata,
		EncryptedMetadataEncryptedKey: params.EncryptedMetadataEncryptedKey,
		EncryptedEtag:                 params.EncryptedETag,
		IfNoneMatch:                   params.IfNoneMatch,
	}
}

// BatchItem returns single item for batch request.
func (params *CommitObjectParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectCommit{
			ObjectCommit: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *CommitObjectParams) IsRetriable() bool { return false }

// CommitObjectResponse response for CommitObject request.
type CommitObjectResponse struct {
	Object RawObjectItem
}

// CommitObject commits a created object.
// TODO remove when all code will be adjusted.
func (client *Client) CommitObject(ctx context.Context, params CommitObjectParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	return WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.CommitObject(ctx, params.toRequest(client.header()))
		return err
	})
}

// CommitObjectWithResponse commits a created object.
func (client *Client) CommitObjectWithResponse(ctx context.Context, params CommitObjectParams) (_ CommitObjectResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.CommitObjectResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.CommitObject(ctx, params.toRequest(client.header()))
		return err
	})

	if err != nil {
		return CommitObjectResponse{}, Error.Wrap(err)
	}

	return CommitObjectResponse{
		Object: newObjectInfo(response.Object),
	}, nil
}

// GetObjectParams parameters for GetObject method.
type GetObjectParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	Version            []byte

	RedundancySchemePerSegment bool
}

func (params *GetObjectParams) toRequest(header *pb.RequestHeader) *pb.ObjectGetRequest {
	return &pb.ObjectGetRequest{
		Header:                     header,
		Bucket:                     params.Bucket,
		EncryptedObjectKey:         params.EncryptedObjectKey,
		ObjectVersion:              params.Version,
		RedundancySchemePerSegment: params.RedundancySchemePerSegment,
	}
}

// BatchItem returns single item for batch request.
func (params *GetObjectParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectGet{
			ObjectGet: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetObjectParams) IsRetriable() bool { return true }

// GetObjectResponse response for GetObject request.
type GetObjectResponse struct {
	Info RawObjectItem
}

func newGetObjectResponse(response *pb.ObjectGetResponse) GetObjectResponse {
	return GetObjectResponse{
		Info: newObjectInfo(response.Object),
	}
}

func newObjectInfo(object *pb.Object) RawObjectItem {
	if object == nil {
		return RawObjectItem{}
	}

	info := RawObjectItem{
		Bucket:             string(object.Bucket),
		EncryptedObjectKey: object.EncryptedObjectKey,
		Version:            object.ObjectVersion,
		Status:             int32(object.Status),

		StreamID: object.StreamId,

		Created:   object.CreatedAt,
		PlainSize: object.PlainSize,
		Expires:   object.ExpiresAt,
		EncryptedUserData: EncryptedUserData{
			EncryptedMetadata:             object.EncryptedMetadata,
			EncryptedMetadataNonce:        object.EncryptedMetadataNonce,
			EncryptedMetadataEncryptedKey: object.EncryptedMetadataEncryptedKey,
			EncryptedETag:                 object.EncryptedEtag,
		},
	}

	if object.Retention != nil {
		info.Retention = &Retention{
			Mode:        storj.RetentionMode(object.Retention.Mode),
			RetainUntil: object.Retention.RetainUntil,
		}
	}

	if object.LegalHold != nil {
		info.LegalHold = &object.LegalHold.Value
	}

	if object.EncryptionParameters != nil {
		info.EncryptionParameters = storj.EncryptionParameters{
			CipherSuite: storj.CipherSuite(object.EncryptionParameters.CipherSuite),
			BlockSize:   int32(object.EncryptionParameters.BlockSize),
		}
	}

	pbRS := object.RedundancyScheme
	if pbRS != nil {
		info.RedundancyScheme = storj.RedundancyScheme{
			Algorithm:      storj.RedundancyAlgorithm(pbRS.Type),
			ShareSize:      pbRS.ErasureShareSize,
			RequiredShares: int16(pbRS.MinReq),
			RepairShares:   int16(pbRS.RepairThreshold),
			OptimalShares:  int16(pbRS.SuccessThreshold),
			TotalShares:    int16(pbRS.Total),
		}
	}
	return info
}

// GetObject gets single object.
func (client *Client) GetObject(ctx context.Context, params GetObjectParams) (_ RawObjectItem, err error) {
	defer mon.Task()(&ctx)(&err)

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return RawObjectItem{}, ErrObjectNotFound.Wrap(err)
		}
		return RawObjectItem{}, Error.Wrap(err)
	}
	if len(items) != 1 {
		return RawObjectItem{}, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_ObjectGet)
	if !ok {
		return RawObjectItem{}, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	getResponse := newGetObjectResponse(response.ObjectGet)
	return getResponse.Info, nil
}

// GetObjectIPsParams are params for the GetObjectIPs request.
type GetObjectIPsParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
}

// GetObjectIPsResponse is the response from GetObjectIPs.
type GetObjectIPsResponse struct {
	IPPorts             [][]byte
	SegmentCount        int64
	PieceCount          int64
	ReliablePieceCount  int64
	PlacementConstraint uint32
}

func (params *GetObjectIPsParams) toRequest(header *pb.RequestHeader) *pb.ObjectGetIPsRequest {
	return &pb.ObjectGetIPsRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
	}
}

// GetObjectIPs returns the IP addresses of the nodes which hold the object.
func (client *Client) GetObjectIPs(ctx context.Context, params GetObjectIPsParams) (r *GetObjectIPsResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.ObjectGetIPsResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetObjectIPs(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return nil, ErrObjectNotFound.Wrap(err)
		}
		return nil, Error.Wrap(err)
	}

	return &GetObjectIPsResponse{
		IPPorts:             response.Ips,
		SegmentCount:        response.SegmentCount,
		PieceCount:          response.PieceCount,
		ReliablePieceCount:  response.ReliablePieceCount,
		PlacementConstraint: response.PlacementConstraint,
	}, nil
}

// UpdateObjectMetadataParams are params for the UpdateObjectMetadata request.
type UpdateObjectMetadataParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	StreamID           storj.StreamID

	EncryptedUserData
	SetEncryptedETag bool
}

func (params *UpdateObjectMetadataParams) toRequest(header *pb.RequestHeader) *pb.ObjectUpdateMetadataRequest {
	return &pb.ObjectUpdateMetadataRequest{
		Header:                        header,
		Bucket:                        params.Bucket,
		EncryptedObjectKey:            params.EncryptedObjectKey,
		StreamId:                      params.StreamID,
		EncryptedMetadataNonce:        params.EncryptedMetadataNonce,
		EncryptedMetadata:             params.EncryptedMetadata,
		EncryptedMetadataEncryptedKey: params.EncryptedMetadataEncryptedKey,
		EncryptedEtag:                 params.EncryptedETag,
		SetEncryptedEtag:              params.SetEncryptedETag,
	}
}

// UpdateObjectMetadata replaces objects metadata.
func (client *Client) UpdateObjectMetadata(ctx context.Context, params UpdateObjectMetadataParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.UpdateObjectMetadata(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return ErrObjectNotFound.Wrap(err)
		}
	}

	return Error.Wrap(err)
}

// SetObjectLegalHoldParams are params for the SetObjectLegalHold request.
type SetObjectLegalHoldParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	ObjectVersion      []byte

	Enabled bool
}

func (params *SetObjectLegalHoldParams) toRequest(header *pb.RequestHeader) *pb.SetObjectLegalHoldRequest {
	req := &pb.SetObjectLegalHoldRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		ObjectVersion:      params.ObjectVersion,
		Enabled:            params.Enabled,
	}

	return req
}

// SetObjectLegalHold sets legal hold status on the object.
func (client *Client) SetObjectLegalHold(ctx context.Context, params SetObjectLegalHoldParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetObjectLegalHold(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return Error.Wrap(convertErrors(err))
	}

	return Error.Wrap(err)
}

// GetObjectLegalHoldParams are params for the GetObjectLegalHold request.
type GetObjectLegalHoldParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	ObjectVersion      []byte
}

func (params *GetObjectLegalHoldParams) toRequest(header *pb.RequestHeader) *pb.GetObjectLegalHoldRequest {
	return &pb.GetObjectLegalHoldRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		ObjectVersion:      params.ObjectVersion,
	}
}

// GetObjectLegalHold retrieves object's legal hold configuration.
func (client *Client) GetObjectLegalHold(ctx context.Context, params GetObjectLegalHoldParams) (_ bool, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetObjectLegalHoldResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetObjectLegalHold(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return false, Error.Wrap(convertErrors(err))
	}

	return response.Enabled, nil
}

// SetObjectRetentionParams are params for the SetObjectRetention request.
type SetObjectRetentionParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	ObjectVersion      []byte

	Retention                 Retention
	BypassGovernanceRetention bool
}

func (params *SetObjectRetentionParams) toRequest(header *pb.RequestHeader) *pb.SetObjectRetentionRequest {
	req := &pb.SetObjectRetentionRequest{
		Header:                    header,
		Bucket:                    params.Bucket,
		EncryptedObjectKey:        params.EncryptedObjectKey,
		ObjectVersion:             params.ObjectVersion,
		BypassGovernanceRetention: params.BypassGovernanceRetention,
	}

	if params.Retention != (Retention{}) {
		req.Retention = &pb.Retention{
			Mode:        pb.Retention_Mode(params.Retention.Mode),
			RetainUntil: params.Retention.RetainUntil,
		}
	}

	return req
}

// SetObjectRetention sets retention on the object.
func (client *Client) SetObjectRetention(ctx context.Context, params SetObjectRetentionParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetObjectRetention(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return Error.Wrap(convertErrors(err))
	}

	return nil
}

// GetObjectRetentionParams are params for the GetObjectRetention request.
type GetObjectRetentionParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	ObjectVersion      []byte
}

func (params *GetObjectRetentionParams) toRequest(header *pb.RequestHeader) *pb.GetObjectRetentionRequest {
	return &pb.GetObjectRetentionRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		ObjectVersion:      params.ObjectVersion,
	}
}

// GetObjectRetention retrieves object's retention.
func (client *Client) GetObjectRetention(ctx context.Context, params GetObjectRetentionParams) (r *Retention, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetObjectRetentionResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetObjectRetention(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return nil, Error.Wrap(convertErrors(err))
	}

	return &Retention{
		Mode:        storj.RetentionMode(response.Retention.Mode),
		RetainUntil: response.Retention.RetainUntil,
	}, nil
}

func convertErrors(err error) error {
	if err == nil {
		return nil
	}

	message := internal.RootError(err).Error()
	switch {
	case strings.HasPrefix(message, "bucket not found"):
		return ErrBucketNotFound.Wrap(err)
	case strings.HasPrefix(message, "object not found"):
		return ErrObjectNotFound.Wrap(err)
	case strings.HasPrefix(message, "method not allowed"):
		return ErrMethodNotAllowed.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockEndpointsDisabled):
		return ErrLockNotEnabled.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockDisabledForProject):
		return ErrProjectNoLock.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockBucketRetentionConfigurationMissing):
		return ErrBucketNoLock.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockObjectRetentionConfigurationMissing):
		return ErrRetentionNotFound.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockObjectProtected):
		return ErrObjectProtected.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockInvalidObjectState):
		return ErrObjectLockInvalidObjectState.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockInvalidBucketRetentionConfiguration):
		return ErrBucketInvalidObjectLockConfig.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.ObjectLockInvalidBucketState):
		return ErrBucketInvalidStateObjectLock.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.PlacementInvalidValue):
		return ErrInvalidPlacement.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.PlacementConflictingValues):
		return ErrConflictingPlacement.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.TagsNotFound):
		return ErrBucketTagsNotFound.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.TooManyTags):
		return ErrTooManyBucketTags.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.TagKeyInvalid):
		return ErrBucketTagKeyInvalid.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.TagKeyDuplicate):
		return ErrBucketTagKeyDuplicate.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.TagValueInvalid):
		return ErrBucketTagValueInvalid.Wrap(err)
	case errs2.IsRPC(err, rpcstatus.Unimplemented):
		return ErrUnimplemented.Wrap(err)
	default:
		return err
	}
}

// BeginDeleteObjectParams parameters for BeginDeleteObject method.
type BeginDeleteObjectParams struct {
	Bucket                    []byte
	EncryptedObjectKey        []byte
	Version                   []byte
	StreamID                  storj.StreamID
	Status                    int32
	BypassGovernanceRetention bool
}

func (params *BeginDeleteObjectParams) toRequest(header *pb.RequestHeader) *pb.ObjectBeginDeleteRequest {
	return &pb.ObjectBeginDeleteRequest{
		Header:                    header,
		Bucket:                    params.Bucket,
		EncryptedObjectKey:        params.EncryptedObjectKey,
		ObjectVersion:             params.Version,
		StreamId:                  &params.StreamID,
		Status:                    params.Status,
		BypassGovernanceRetention: params.BypassGovernanceRetention,
	}
}

// BatchItem returns single item for batch request.
func (params *BeginDeleteObjectParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectBeginDelete{
			ObjectBeginDelete: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *BeginDeleteObjectParams) IsRetriable() bool { return false }

// BeginDeleteObjectResponse response for BeginDeleteObject request.
type BeginDeleteObjectResponse struct {
}

func newBeginDeleteObjectResponse(response *pb.ObjectBeginDeleteResponse) BeginDeleteObjectResponse {
	return BeginDeleteObjectResponse{}
}

// BeginDeleteObject begins object deletion process.
func (client *Client) BeginDeleteObject(ctx context.Context, params BeginDeleteObjectParams) (_ RawObjectItem, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.ObjectBeginDeleteResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		// response.StreamID is not processed because satellite will always return nil
		response, err = client.client.BeginDeleteObject(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return RawObjectItem{}, ErrObjectNotFound.Wrap(err)
		}
		return RawObjectItem{}, Error.Wrap(err)
	}

	return newObjectInfo(response.Object), nil
}

// DeleteObjectsParams represents parameters for the DeleteObjects method.
type DeleteObjectsParams struct {
	Bucket                    []byte
	Items                     []RawDeleteObjectsItem
	BypassGovernanceRetention bool
	Quiet                     bool
}

// RawDeleteObjectsItem describes the location of an object in a bucket to be deleted.
type RawDeleteObjectsItem struct {
	EncryptedObjectKey []byte
	Version            []byte
}

func (params *DeleteObjectsParams) toRequest(header *pb.RequestHeader) *pb.DeleteObjectsRequest {
	req := &pb.DeleteObjectsRequest{
		Header:                    header,
		Bucket:                    params.Bucket,
		Items:                     make([]*pb.DeleteObjectsRequestItem, 0, len(params.Items)),
		BypassGovernanceRetention: params.BypassGovernanceRetention,
		Quiet:                     params.Quiet,
	}
	for _, item := range params.Items {
		req.Items = append(req.Items, &pb.DeleteObjectsRequestItem{
			EncryptedObjectKey: item.EncryptedObjectKey,
			ObjectVersion:      item.Version,
		})
	}
	return req
}

// BatchItem returns a single item for a batch request.
func (params *DeleteObjectsParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectsDelete{
			ObjectsDelete: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *DeleteObjectsParams) IsRetriable() bool { return false }

// RawDeleteObjectsResultItem represents the result of an individual DeleteObjects deletion.
type RawDeleteObjectsResultItem struct {
	EncryptedObjectKey []byte
	RequestedVersion   []byte

	Removed *DeleteObjectsResultItemRemoved
	Marker  *DeleteObjectsResultItemMarker

	Status storj.DeleteObjectsStatus
}

// DeleteObjectsResultItemRemoved contains information about an object that was removed
// as a result of processing a DeleteObjects request item.
type DeleteObjectsResultItemRemoved struct {
	Version     []byte
	IsCommitted bool
	IsVersioned bool
}

// DeleteObjectsResultItemMarker contains information about a delete marker that was inserted
// as a result of processing a DeleteObjects request item.
type DeleteObjectsResultItemMarker struct {
	Version     []byte
	IsVersioned bool
}

func newDeleteObjectsResponse(pbResponse *pb.DeleteObjectsResponse) []RawDeleteObjectsResultItem {
	resultItems := make([]RawDeleteObjectsResultItem, 0, len(pbResponse.Items))
	for _, pbItem := range pbResponse.Items {
		item := RawDeleteObjectsResultItem{
			EncryptedObjectKey: pbItem.EncryptedObjectKey,
			RequestedVersion:   pbItem.RequestedObjectVersion,
			Status:             storj.DeleteObjectsStatus(pbItem.Status),
		}
		if pbItem.Removed != nil {
			item.Removed = &DeleteObjectsResultItemRemoved{
				Version:     pbItem.Removed.ObjectVersion,
				IsCommitted: isStatusCommitted(pbItem.Removed.Status),
				IsVersioned: isStatusVersioned(pbItem.Removed.Status),
			}
		}
		if pbItem.Marker != nil {
			item.Marker = &DeleteObjectsResultItemMarker{
				Version:     pbItem.Marker.ObjectVersion,
				IsVersioned: isStatusVersioned(pbItem.Marker.Status),
			}
		}
		resultItems = append(resultItems, item)
	}
	return resultItems
}

// DeleteObjects deletes multiple objects from a bucket.
func (client *Client) DeleteObjects(ctx context.Context, params DeleteObjectsParams) (_ []RawDeleteObjectsResultItem, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.DeleteObjectsResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.DeleteObjects(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return newDeleteObjectsResponse(response), nil
}

// ListObjectsParams parameters for ListObjects method.
type ListObjectsParams struct {
	Bucket          []byte
	Delimiter       []byte
	EncryptedPrefix []byte
	EncryptedCursor []byte
	VersionCursor   []byte
	Limit           int32

	IncludeCustomMetadata       bool
	IncludeSystemMetadata       bool
	IncludeETag                 bool
	IncludeETagOrCustomMetadata bool

	Recursive          bool
	Status             int32
	IncludeAllVersions bool
	ArbitraryPrefix    bool
}

func (params *ListObjectsParams) toRequest(header *pb.RequestHeader) *pb.ObjectListRequest {
	return &pb.ObjectListRequest{
		Header:          header,
		Bucket:          params.Bucket,
		Delimiter:       params.Delimiter,
		EncryptedPrefix: params.EncryptedPrefix,
		EncryptedCursor: params.EncryptedCursor,
		Limit:           params.Limit,
		ObjectIncludes: &pb.ObjectListItemIncludes{
			Metadata:                    params.IncludeCustomMetadata,
			ExcludeSystemMetadata:       !params.IncludeSystemMetadata,
			IncludeEtag:                 params.IncludeETag,
			IncludeEtagOrCustomMetadata: params.IncludeETagOrCustomMetadata,
		},
		UseObjectIncludes:  true,
		Recursive:          params.Recursive,
		Status:             pb.Object_Status(params.Status),
		IncludeAllVersions: params.IncludeAllVersions,
		VersionCursor:      params.VersionCursor,
		ArbitraryPrefix:    params.ArbitraryPrefix,
	}
}

// BatchItem returns single item for batch request.
func (params *ListObjectsParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectList{
			ObjectList: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *ListObjectsParams) IsRetriable() bool { return true }

// ListObjectsResponse response for ListObjects request.
type ListObjectsResponse struct {
	Items []RawObjectListItem
	More  bool
}

func newListObjectsResponse(response *pb.ObjectListResponse, encryptedPrefix []byte, delimiter []byte, recursive bool) ListObjectsResponse {
	objects := make([]RawObjectListItem, len(response.Items))
	for i, object := range response.Items {
		objects[i] = RawObjectListItem{
			EncryptedObjectKey: object.EncryptedObjectKey,
			Version:            object.ObjectVersion,
			Status:             int32(object.Status),
			IsLatest:           object.IsLatest,
			CreatedAt:          object.CreatedAt,
			ExpiresAt:          object.ExpiresAt,
			PlainSize:          object.PlainSize,
			EncryptedUserData: EncryptedUserData{
				EncryptedMetadataNonce:        object.EncryptedMetadataNonce,
				EncryptedMetadataEncryptedKey: object.EncryptedMetadataEncryptedKey,
				EncryptedMetadata:             object.EncryptedMetadata,
				EncryptedETag:                 object.EncryptedEtag,
			},
			IsPrefix: object.Status == pb.Object_PREFIX,
		}

		if object.StreamId != nil {
			objects[i].StreamID = *object.StreamId
		}
	}

	return ListObjectsResponse{
		Items: objects,
		More:  response.More,
	}
}

// ListObjects lists objects according to specific parameters.
func (client *Client) ListObjects(ctx context.Context, params ListObjectsParams) (_ []RawObjectListItem, more bool, err error) {
	defer mon.Task()(&ctx)(&err)

	if params.Recursive {
		params.Delimiter = nil
	}

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		return []RawObjectListItem{}, false, Error.Wrap(err)
	}
	if len(items) != 1 {
		return []RawObjectListItem{}, false, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_ObjectList)
	if !ok {
		return []RawObjectListItem{}, false, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	listResponse := newListObjectsResponse(response.ObjectList, params.EncryptedPrefix, params.Delimiter, params.Recursive)
	return listResponse.Items, listResponse.More, Error.Wrap(err)
}

// ListPendingObjectStreamsParams parameters for ListPendingObjectStreams method.
type ListPendingObjectStreamsParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	EncryptedCursor    []byte
	Limit              int32
}

func (params *ListPendingObjectStreamsParams) toRequest(header *pb.RequestHeader) *pb.ObjectListPendingStreamsRequest {
	return &pb.ObjectListPendingStreamsRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		StreamIdCursor:     params.EncryptedCursor,
		Limit:              params.Limit,
	}
}

// BatchItem returns single item for batch request.
func (params *ListPendingObjectStreamsParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectListPendingStreams{
			ObjectListPendingStreams: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *ListPendingObjectStreamsParams) IsRetriable() bool { return true }

// ListPendingObjectStreamsResponse response for ListPendingObjectStreams request.
type ListPendingObjectStreamsResponse struct {
	Items []RawObjectListItem
	More  bool
}

func newListPendingObjectStreamsResponse(response *pb.ObjectListPendingStreamsResponse) ListPendingObjectStreamsResponse {
	objects := make([]RawObjectListItem, len(response.Items))
	for i, object := range response.Items {

		objects[i] = RawObjectListItem{
			EncryptedObjectKey: object.EncryptedObjectKey,
			Version:            object.ObjectVersion,
			Status:             int32(object.Status),
			IsLatest:           false,
			CreatedAt:          object.CreatedAt,
			ExpiresAt:          object.ExpiresAt,
			PlainSize:          object.PlainSize,
			EncryptedUserData: EncryptedUserData{
				EncryptedMetadataEncryptedKey: object.EncryptedMetadataEncryptedKey,
				EncryptedMetadataNonce:        object.EncryptedMetadataNonce,
				EncryptedMetadata:             object.EncryptedMetadata,
				EncryptedETag:                 object.EncryptedEtag,
			},

			IsPrefix: false,
		}

		if object.StreamId != nil {
			objects[i].StreamID = *object.StreamId
		}
	}

	return ListPendingObjectStreamsResponse{
		Items: objects,
		More:  response.More,
	}
}

// ListPendingObjectStreams lists pending objects with the specified object key in the specified bucket.
func (client *Client) ListPendingObjectStreams(ctx context.Context, params ListPendingObjectStreamsParams) (_ ListPendingObjectStreamsResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.ObjectListPendingStreamsResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.ListPendingObjectStreams(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return ListPendingObjectStreamsResponse{}, Error.Wrap(err)
	}

	return newListPendingObjectStreamsResponse(response), nil
}

// SegmentListItem represents listed segment.
type SegmentListItem struct {
	Position          SegmentPosition
	PlainSize         int64
	PlainOffset       int64
	CreatedAt         time.Time
	EncryptedETag     []byte
	EncryptedKeyNonce storj.Nonce
	EncryptedKey      []byte
}

// ListSegmentsParams parameters for ListSegments method.
type ListSegmentsParams struct {
	StreamID []byte
	Cursor   SegmentPosition
	Limit    int32
	Range    StreamRange
}

func (params *ListSegmentsParams) toRequest(header *pb.RequestHeader) *pb.SegmentListRequest {
	return &pb.SegmentListRequest{
		Header:   header,
		StreamId: params.StreamID,
		CursorPosition: &pb.SegmentPosition{
			PartNumber: params.Cursor.PartNumber,
			Index:      params.Cursor.Index,
		},
		Limit: params.Limit,
		Range: params.Range.toProto(),
	}
}

// BatchItem returns single item for batch request.
func (params *ListSegmentsParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentList{
			SegmentList: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *ListSegmentsParams) IsRetriable() bool { return true }

// ListSegmentsResponse response for ListSegments request.
type ListSegmentsResponse struct {
	Items                []SegmentListItem
	More                 bool
	EncryptionParameters storj.EncryptionParameters
}

func newListSegmentsResponse(response *pb.SegmentListResponse) ListSegmentsResponse {
	segments := make([]SegmentListItem, len(response.Items))
	for i, segment := range response.Items {
		segments[i] = SegmentListItem{
			Position: SegmentPosition{
				PartNumber: segment.Position.PartNumber,
				Index:      segment.Position.Index,
			},
			PlainSize:         segment.PlainSize,
			PlainOffset:       segment.PlainOffset,
			CreatedAt:         segment.CreatedAt,
			EncryptedETag:     segment.EncryptedETag,
			EncryptedKeyNonce: segment.EncryptedKeyNonce,
			EncryptedKey:      segment.EncryptedKey,
		}
	}

	ep := storj.EncryptionParameters{}
	if response.EncryptionParameters != nil {
		ep = storj.EncryptionParameters{
			CipherSuite: storj.CipherSuite(response.EncryptionParameters.CipherSuite),
			BlockSize:   int32(response.EncryptionParameters.BlockSize),
		}
	}

	return ListSegmentsResponse{
		Items:                segments,
		More:                 response.More,
		EncryptionParameters: ep,
	}
}

// ListSegments lists segments according to specific parameters.
func (client *Client) ListSegments(ctx context.Context, params ListSegmentsParams) (_ ListSegmentsResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.SegmentListResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.ListSegments(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return ListSegmentsResponse{}, Error.Wrap(err)
	}

	return newListSegmentsResponse(response), nil
}

// BeginSegmentParams parameters for BeginSegment method.
type BeginSegmentParams struct {
	StreamID      storj.StreamID
	Position      SegmentPosition
	MaxOrderLimit int64
}

func (params *BeginSegmentParams) toRequest(header *pb.RequestHeader) *pb.SegmentBeginRequest {
	return &pb.SegmentBeginRequest{
		Header:   header,
		StreamId: params.StreamID,
		Position: &pb.SegmentPosition{
			PartNumber: params.Position.PartNumber,
			Index:      params.Position.Index,
		},
		MaxOrderLimit: params.MaxOrderLimit,
	}
}

// BatchItem returns single item for batch request.
func (params *BeginSegmentParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentBegin{
			SegmentBegin: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *BeginSegmentParams) IsRetriable() bool { return true }

// BeginSegmentResponse response for BeginSegment request.
type BeginSegmentResponse struct {
	SegmentID          storj.SegmentID
	Limits             []*pb.AddressedOrderLimit
	PiecePrivateKey    storj.PiecePrivateKey
	RedundancyStrategy eestream.RedundancyStrategy
	CohortRequirements *pb.CohortRequirements
}

func newBeginSegmentResponse(response *pb.SegmentBeginResponse) (BeginSegmentResponse, error) {
	var rs eestream.RedundancyStrategy
	var err error
	if response.RedundancyScheme != nil {
		rs, err = eestream.NewRedundancyStrategyFromProto(response.RedundancyScheme)
		if err != nil {
			return BeginSegmentResponse{}, err
		}
	}
	return BeginSegmentResponse{
		SegmentID:          response.SegmentId,
		Limits:             response.AddressedLimits,
		PiecePrivateKey:    response.PrivateKey,
		RedundancyStrategy: rs,
		CohortRequirements: response.CohortRequirements,
	}, nil
}

// BeginSegment begins a segment upload.
func (client *Client) BeginSegment(ctx context.Context, params BeginSegmentParams) (_ BeginSegmentResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.SegmentBeginResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.BeginSegment(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return BeginSegmentResponse{}, Error.Wrap(err)
	}

	return newBeginSegmentResponse(response)
}

// RetryBeginSegmentPiecesParams parameters for RetryBeginSegmentPieces method.
type RetryBeginSegmentPiecesParams struct {
	SegmentID         storj.SegmentID
	RetryPieceNumbers []int
}

// BatchItem returns single item for batch request.
func (params *RetryBeginSegmentPiecesParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentBeginRetryPieces{
			SegmentBeginRetryPieces: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *RetryBeginSegmentPiecesParams) IsRetriable() bool { return false }

func (params *RetryBeginSegmentPiecesParams) toRequest(header *pb.RequestHeader) *pb.RetryBeginSegmentPiecesRequest {
	retryPieceNumbers := make([]int32, len(params.RetryPieceNumbers))
	for i, pieceNumber := range params.RetryPieceNumbers {
		retryPieceNumbers[i] = int32(pieceNumber)
	}
	return &pb.RetryBeginSegmentPiecesRequest{
		Header:            header,
		SegmentId:         params.SegmentID,
		RetryPieceNumbers: retryPieceNumbers,
	}
}

// RetryBeginSegmentPiecesResponse response for RetryBeginSegmentPieces request.
type RetryBeginSegmentPiecesResponse struct {
	SegmentID storj.SegmentID
	Limits    []*pb.AddressedOrderLimit
}

func newRetryBeginSegmentPiecesResponse(response *pb.RetryBeginSegmentPiecesResponse) (RetryBeginSegmentPiecesResponse, error) {
	return RetryBeginSegmentPiecesResponse{
		SegmentID: response.SegmentId,
		Limits:    response.AddressedLimits,
	}, nil
}

// RetryBeginSegmentPieces exchanges piece orders.
func (client *Client) RetryBeginSegmentPieces(ctx context.Context, params RetryBeginSegmentPiecesParams) (_ RetryBeginSegmentPiecesResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	if os.Getenv("STORJ_COMPRESSED_BATCH") != "false" {
		return client.batchRetryBeginSegmentPieces(ctx, params)
	}

	var response *pb.RetryBeginSegmentPiecesResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.RetryBeginSegmentPieces(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return RetryBeginSegmentPiecesResponse{}, Error.Wrap(err)
	}

	return newRetryBeginSegmentPiecesResponse(response)
}

// batchRetryBeginSegmentPieces is RetryBeginSegmentPieces but goes through the batch rpc.
func (client *Client) batchRetryBeginSegmentPieces(ctx context.Context, params RetryBeginSegmentPiecesParams) (_ RetryBeginSegmentPiecesResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		return RetryBeginSegmentPiecesResponse{}, Error.Wrap(err)
	}
	if len(items) != 1 {
		return RetryBeginSegmentPiecesResponse{}, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_SegmentBeginRetryPieces)
	if !ok {
		return RetryBeginSegmentPiecesResponse{}, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	return newRetryBeginSegmentPiecesResponse(response.SegmentBeginRetryPieces)
}

// CommitSegmentParams parameters for CommitSegment method.
type CommitSegmentParams struct {
	SegmentID         storj.SegmentID
	Encryption        SegmentEncryption
	SizeEncryptedData int64
	PlainSize         int64
	EncryptedETag     []byte

	UploadResult []*pb.SegmentPieceUploadResult
}

func (params *CommitSegmentParams) toRequest(header *pb.RequestHeader) *pb.SegmentCommitRequest {
	return &pb.SegmentCommitRequest{
		Header:    header,
		SegmentId: params.SegmentID,

		EncryptedKeyNonce: params.Encryption.EncryptedKeyNonce,
		EncryptedKey:      params.Encryption.EncryptedKey,
		SizeEncryptedData: params.SizeEncryptedData,
		PlainSize:         params.PlainSize,
		EncryptedETag:     params.EncryptedETag,
		UploadResult:      params.UploadResult,
	}
}

// BatchItem returns single item for batch request.
func (params *CommitSegmentParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentCommit{
			SegmentCommit: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *CommitSegmentParams) IsRetriable() bool { return false }

// CommitSegment commits an uploaded segment.
func (client *Client) CommitSegment(ctx context.Context, params CommitSegmentParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.CommitSegment(ctx, params.toRequest(client.header()))
		return err
	})

	return Error.Wrap(err)
}

// MakeInlineSegmentParams parameters for MakeInlineSegment method.
type MakeInlineSegmentParams struct {
	StreamID            storj.StreamID
	Position            SegmentPosition
	Encryption          SegmentEncryption
	EncryptedInlineData []byte
	PlainSize           int64
	EncryptedETag       []byte
}

func (params *MakeInlineSegmentParams) toRequest(header *pb.RequestHeader) *pb.SegmentMakeInlineRequest {
	return &pb.SegmentMakeInlineRequest{
		Header:   header,
		StreamId: params.StreamID,
		Position: &pb.SegmentPosition{
			PartNumber: params.Position.PartNumber,
			Index:      params.Position.Index,
		},
		EncryptedKeyNonce:   params.Encryption.EncryptedKeyNonce,
		EncryptedKey:        params.Encryption.EncryptedKey,
		EncryptedInlineData: params.EncryptedInlineData,
		PlainSize:           params.PlainSize,
		EncryptedETag:       params.EncryptedETag,
	}
}

// BatchItem returns single item for batch request.
func (params *MakeInlineSegmentParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentMakeInline{
			SegmentMakeInline: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *MakeInlineSegmentParams) IsRetriable() bool { return false }

// MakeInlineSegment creates an inline segment.
func (client *Client) MakeInlineSegment(ctx context.Context, params MakeInlineSegmentParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.MakeInlineSegment(ctx, params.toRequest(client.header()))
		return err
	})

	return Error.Wrap(err)
}

// DownloadObjectParams parameters for DownloadObject method.
type DownloadObjectParams struct {
	Bucket             []byte
	EncryptedObjectKey []byte
	Version            []byte

	Range StreamRange

	ServerSideCopy bool
}

func (params *DownloadObjectParams) toRequest(header *pb.RequestHeader) *pb.ObjectDownloadRequest {
	return &pb.ObjectDownloadRequest{
		Header:             header,
		Bucket:             params.Bucket,
		EncryptedObjectKey: params.EncryptedObjectKey,
		ObjectVersion:      params.Version,
		Range:              params.Range.toProto(),
		ServerSideCopy:     params.ServerSideCopy,
	}
}

// BatchItem returns single item for batch request.
func (params *DownloadObjectParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_ObjectDownload{
			ObjectDownload: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *DownloadObjectParams) IsRetriable() bool { return true }

// DownloadObjectResponse response for DownloadObject request.
type DownloadObjectResponse struct {
	Object             RawObjectItem
	DownloadedSegments []DownloadSegmentWithRSResponse
	ListSegments       ListSegmentsResponse
}

func newDownloadObjectResponse(response *pb.ObjectDownloadResponse) DownloadObjectResponse {
	downloadedSegments := make([]DownloadSegmentWithRSResponse, 0, len(response.SegmentDownload))
	for _, segmentDownload := range response.SegmentDownload {
		downloadedSegments = append(downloadedSegments, newDownloadSegmentResponseWithRS(segmentDownload))
	}
	return DownloadObjectResponse{
		Object:             newObjectInfo(response.Object),
		DownloadedSegments: downloadedSegments,
		ListSegments:       newListSegmentsResponse(response.SegmentList),
	}
}

// DownloadObject gets object information, lists segments and downloads the first segment.
func (client *Client) DownloadObject(ctx context.Context, params DownloadObjectParams) (_ DownloadObjectResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	if os.Getenv("STORJ_COMPRESSED_BATCH") != "false" {
		return client.batchDownloadObject(ctx, params)
	}

	var response *pb.ObjectDownloadResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.DownloadObject(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return DownloadObjectResponse{}, ErrObjectNotFound.Wrap(err)
		}
		return DownloadObjectResponse{}, Error.Wrap(err)
	}
	return newDownloadObjectResponse(response), nil
}

// batchDownloadObject is DownloadObject but goes through the batch rpc.
func (client *Client) batchDownloadObject(ctx context.Context, params DownloadObjectParams) (_ DownloadObjectResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return DownloadObjectResponse{}, ErrObjectNotFound.Wrap(err)
		}
		return DownloadObjectResponse{}, Error.Wrap(err)
	}
	if len(items) != 1 {
		return DownloadObjectResponse{}, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_ObjectDownload)
	if !ok {
		return DownloadObjectResponse{}, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	return newDownloadObjectResponse(response.ObjectDownload), nil
}

// DownloadSegmentParams parameters for DownloadSegment method.
type DownloadSegmentParams struct {
	StreamID     storj.StreamID
	Position     SegmentPosition
	DesiredNodes int32

	ServerSideCopy bool
}

func (params *DownloadSegmentParams) toRequest(header *pb.RequestHeader) *pb.SegmentDownloadRequest {
	return &pb.SegmentDownloadRequest{
		Header:   header,
		StreamId: params.StreamID,
		CursorPosition: &pb.SegmentPosition{
			PartNumber: params.Position.PartNumber,
			Index:      params.Position.Index,
		},
		DesiredNodes:   params.DesiredNodes,
		ServerSideCopy: params.ServerSideCopy,
	}
}

// BatchItem returns single item for batch request.
func (params *DownloadSegmentParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_SegmentDownload{
			SegmentDownload: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *DownloadSegmentParams) IsRetriable() bool { return true }

// DownloadSegmentResponse response for DownloadSegment request.
type DownloadSegmentResponse struct {
	Info SegmentDownloadResponseInfo

	Limits []*pb.AddressedOrderLimit
}

func newDownloadSegmentResponse(response *pb.SegmentDownloadResponse) DownloadSegmentResponse {
	info := SegmentDownloadResponseInfo{
		SegmentID:           response.SegmentId,
		EncryptedSize:       response.SegmentSize,
		EncryptedInlineData: response.EncryptedInlineData,
		PiecePrivateKey:     response.PrivateKey,
		SegmentEncryption: SegmentEncryption{
			EncryptedKeyNonce: response.EncryptedKeyNonce,
			EncryptedKey:      response.EncryptedKey,
		},
	}
	if response.Next != nil {
		info.Next = SegmentPosition{
			PartNumber: response.Next.PartNumber,
			Index:      response.Next.Index,
		}
	}

	for i := range response.AddressedLimits {
		if response.AddressedLimits[i].Limit == nil {
			response.AddressedLimits[i] = nil
		}
	}
	return DownloadSegmentResponse{
		Info:   info,
		Limits: response.AddressedLimits,
	}
}

// DownloadSegmentWithRSResponse contains information for downloading remote segment or data from an inline segment.
type DownloadSegmentWithRSResponse struct {
	Info   SegmentDownloadInfo
	Limits []*pb.AddressedOrderLimit
}

// SegmentDownloadInfo represents information necessary for downloading segment (inline and remote).
type SegmentDownloadInfo struct {
	SegmentID           storj.SegmentID
	PlainOffset         int64
	PlainSize           int64
	EncryptedSize       int64
	EncryptedInlineData []byte
	PiecePrivateKey     storj.PiecePrivateKey
	SegmentEncryption   SegmentEncryption
	RedundancyScheme    storj.RedundancyScheme
	Position            *SegmentPosition
}

func newDownloadSegmentResponseWithRS(response *pb.SegmentDownloadResponse) DownloadSegmentWithRSResponse {
	info := SegmentDownloadInfo{
		SegmentID:           response.SegmentId,
		PlainOffset:         response.PlainOffset,
		PlainSize:           response.PlainSize,
		EncryptedSize:       response.SegmentSize,
		EncryptedInlineData: response.EncryptedInlineData,
		PiecePrivateKey:     response.PrivateKey,
		SegmentEncryption: SegmentEncryption{
			EncryptedKeyNonce: response.EncryptedKeyNonce,
			EncryptedKey:      response.EncryptedKey,
		},
	}

	if response.Position != nil {
		info.Position = &SegmentPosition{
			PartNumber: response.Position.PartNumber,
			Index:      response.Position.Index,
		}
	}

	if response.RedundancyScheme != nil {
		info.RedundancyScheme = storj.RedundancyScheme{
			Algorithm:      storj.RedundancyAlgorithm(response.RedundancyScheme.Type),
			ShareSize:      response.RedundancyScheme.ErasureShareSize,
			RequiredShares: int16(response.RedundancyScheme.MinReq),
			RepairShares:   int16(response.RedundancyScheme.RepairThreshold),
			OptimalShares:  int16(response.RedundancyScheme.SuccessThreshold),
			TotalShares:    int16(response.RedundancyScheme.Total),
		}
	}

	for i := range response.AddressedLimits {
		if response.AddressedLimits[i].Limit == nil {
			response.AddressedLimits[i] = nil
		}
	}
	return DownloadSegmentWithRSResponse{
		Info:   info,
		Limits: response.AddressedLimits,
	}
}

// TODO replace DownloadSegment with DownloadSegmentWithRS in batch

// DownloadSegmentWithRS gets information for downloading remote segment or data from an inline segment.
func (client *Client) DownloadSegmentWithRS(ctx context.Context, params DownloadSegmentParams) (_ DownloadSegmentWithRSResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return DownloadSegmentWithRSResponse{}, ErrObjectNotFound.Wrap(err)
		}
		return DownloadSegmentWithRSResponse{}, Error.Wrap(err)
	}
	if len(items) != 1 {
		return DownloadSegmentWithRSResponse{}, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_SegmentDownload)
	if !ok {
		return DownloadSegmentWithRSResponse{}, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	return newDownloadSegmentResponseWithRS(response.SegmentDownload), nil
}

// RevokeAPIKey revokes the APIKey provided in the params.
func (client *Client) RevokeAPIKey(ctx context.Context, params RevokeAPIKeyParams) (err error) {
	defer mon.Task()(&ctx)(&err)
	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.RevokeAPIKey(ctx, params.toRequest(client.header()))
		return err
	})
	return Error.Wrap(err)
}

// RevokeAPIKeyParams contain params for a RevokeAPIKey request.
type RevokeAPIKeyParams struct {
	APIKey []byte
}

func (r RevokeAPIKeyParams) toRequest(header *pb.RequestHeader) *pb.RevokeAPIKeyRequest {
	return &pb.RevokeAPIKeyRequest{
		Header: header,
		ApiKey: r.APIKey,
	}
}

// Batch sends multiple requests in one batch.
func (client *Client) Batch(ctx context.Context, requests ...BatchItem) (resp []BatchResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	if os.Getenv("STORJ_COMPRESSED_BATCH") != "false" {
		return client.compressedBatch(ctx, requests...)
	}

	// We can retry if all the requests in the batch are retriable.
	isRetriable := true
	batchItems := make([]*pb.BatchRequestItem, len(requests))
	for i, request := range requests {
		if !request.IsRetriable() {
			isRetriable = false
		}
		batchItems[i] = request.BatchItem()
	}

	if err := client.preprocess(ctx, batchItems); err != nil {
		return nil, Error.Wrap(err)
	}

	withRetry := NoopWithRetry
	if isRetriable {
		withRetry = WithRetry
	}

	var response *pb.BatchResponse
	err = withRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.Batch(ctx, &pb.BatchRequest{
			Header:   client.header(),
			Requests: batchItems,
		})
		return err
	})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	if err := client.postprocess(ctx, response.Responses); err != nil {
		return nil, Error.Wrap(err)
	}

	resp = make([]BatchResponse, len(response.Responses))
	for i, response := range response.Responses {
		resp[i] = MakeBatchResponse(batchItems[i], response)
	}

	return resp, nil
}

// compressedBatch sends multiple requests in one batch supporting compressed responses.
func (client *Client) compressedBatch(ctx context.Context, requests ...BatchItem) (_ []BatchResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	batchItems := make([]*pb.BatchRequestItem, len(requests))
	for i, request := range requests {
		batchItems[i] = request.BatchItem()
	}

	if err := client.preprocess(ctx, batchItems); err != nil {
		return nil, Error.Wrap(err)
	}

	data, err := pb.Marshal(&pb.BatchRequest{
		Header:   client.header(),
		Requests: batchItems,
	})
	compResponse, err := client.client.CompressedBatch(ctx, &pb.CompressedBatchRequest{
		Supported: []pb.CompressedBatchRequest_CompressionType{pb.CompressedBatchRequest_ZSTD},
		Data:      data,
	})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	var respData []byte
	switch compResponse.Selected {
	case pb.CompressedBatchRequest_NONE:
		respData = compResponse.Data
	case pb.CompressedBatchRequest_ZSTD:
		respData, err = zstdDecoder.DecodeAll(compResponse.Data, nil)
	default:
		err = Error.New("unsupported compression type: %v", compResponse.Selected)
	}
	if err != nil {
		return nil, Error.Wrap(err)
	}

	var response pb.BatchResponse
	if err := pb.Unmarshal(respData, &response); err != nil {
		return nil, Error.Wrap(err)
	}

	if err := client.postprocess(ctx, response.Responses); err != nil {
		return nil, Error.Wrap(err)
	}

	resp := make([]BatchResponse, len(response.Responses))
	for i, response := range response.Responses {
		resp[i] = MakeBatchResponse(batchItems[i], response)
	}

	return resp, nil
}

func (client *Client) preprocess(_ context.Context, reqs []*pb.BatchRequestItem) error {
	if client.opts.SatelliteSigner == nil || os.Getenv("STORJ_LITE_DISABLED") != "" {
		return nil
	}

	for _, req := range reqs {
		switch r := req.Request.(type) {
		default: // present to silence gocritic. i like the type switch better.

		case *pb.BatchRequestItem_SegmentBegin:
			r.SegmentBegin.LiteRequest = true
		case *pb.BatchRequestItem_ObjectDownload:
			r.ObjectDownload.LiteRequest = true
		}
	}

	return nil
}

func (client *Client) postprocess(ctx context.Context, resps []*pb.BatchResponseItem) error {
	if client.opts.SatelliteSigner == nil || os.Getenv("STORJ_LITE_DISABLED") != "" {
		return nil
	}

	for _, resp := range resps {
		switch r := resp.Response.(type) {
		default: // present to silence gocritic. i like the type switch better.

		case *pb.BatchResponseItem_SegmentBegin:
			limits := r.SegmentBegin.AddressedLimits

			if len(limits) > 0 && limits[0].Limit != nil && len(limits[0].Limit.SatelliteSignature) == 0 {
				// we know we got a lite response so we have to add in the information to make it
				// valid: the piece id and satellite signature. we then need to repack the segment
				// id to contain the new original order limits and signature.

				var segmentID cursed.SegmentID
				if err := pb.Unmarshal(r.SegmentBegin.SegmentId, &segmentID); err != nil {
					return err
				}

				var err error
				for pieceNum, limit := range limits {
					limit.Limit.PieceId = segmentID.RootPieceId.Derive(limit.Limit.StorageNodeId, int32(pieceNum))
					limits[pieceNum].Limit, err = signing.SignOrderLimit(ctx, client.opts.SatelliteSigner, limit.Limit)
					if err != nil {
						return err
					}
				}

				segmentID.OriginalOrderLimits = limits
				r.SegmentBegin.SegmentId, err = cursed.PackSegmentID(ctx, client.opts.SatelliteSigner, &segmentID)
				if err != nil {
					return err
				}
			}
		case *pb.BatchResponseItem_ObjectDownload:
			// ObjectDownload only sends the first segment of the requested range segment unless the
			// requested range is outside of objects bounds.
			if len(r.ObjectDownload.SegmentDownload) > 0 {
				segmentDownload := r.ObjectDownload.SegmentDownload[0]
				var segmentID cursed.SegmentID
				err := pb.Unmarshal(segmentDownload.SegmentId, &segmentID)
				if err != nil {
					return err
				}

				// segmentID is only set for the purpose of having the root piece ID. The normal object
				// download requests (no lite ones) don't set the segmentID
				segmentDownload.SegmentId = nil

				// These addressed limits don't contain the piece ID, and they aren't signed. We have to set the
				// piece ID and sign the orders.
				for pieceNum, limit := range segmentDownload.AddressedLimits {
					// Download always return a slice of orders limits with length equal to the number of pieces
					// but to download we don't need all the pieces, so the order limits of the unneeded pieces
					// are nil.
					if limit.Limit == nil {
						continue
					}

					limit.Limit.PieceId = segmentID.RootPieceId.Derive(limit.Limit.StorageNodeId, int32(pieceNum))
					limit.Limit, err = signing.SignOrderLimit(ctx, client.opts.SatelliteSigner, limit.Limit)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// SetRawAPIKey sets the client's raw API key. Mainly used for testing.
func (client *Client) SetRawAPIKey(key []byte) {
	client.apiKeyRaw = key
}
