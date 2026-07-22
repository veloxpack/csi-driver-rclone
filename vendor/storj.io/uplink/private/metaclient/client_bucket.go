// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package metaclient

import (
	"context"

	"storj.io/common/errs2"
	"storj.io/common/pb"
	"storj.io/common/rpc/rpcstatus"
	"storj.io/common/storj"
)

// CreateBucketParams parameters for CreateBucket method.
type CreateBucketParams struct {
	Name              []byte
	Placement         []byte
	ObjectLockEnabled bool
}

func (params *CreateBucketParams) toRequest(header *pb.RequestHeader) *pb.BucketCreateRequest {
	return &pb.BucketCreateRequest{
		Header:            header,
		Name:              params.Name,
		Placement:         params.Placement,
		ObjectLockEnabled: params.ObjectLockEnabled,
	}
}

// BatchItem returns single item for batch request.
func (params *CreateBucketParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketCreate{
			BucketCreate: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *CreateBucketParams) IsRetriable() bool { return false }

// CreateBucketResponse response for CreateBucket request.
type CreateBucketResponse struct {
	Bucket Bucket
}

func newCreateBucketResponse(response *pb.BucketCreateResponse) (CreateBucketResponse, error) {
	bucket, err := convertProtoToBucket(response.Bucket)
	if err != nil {
		return CreateBucketResponse{}, err
	}
	return CreateBucketResponse{
		Bucket: bucket,
	}, nil
}

// CreateBucket creates a new bucket.
func (client *Client) CreateBucket(ctx context.Context, params CreateBucketParams) (respBucket Bucket, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.BucketCreateResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.CreateBucket(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return Bucket{}, Error.Wrap(convertErrors(err))
	}

	respBucket, err = convertProtoToBucket(response.Bucket)
	if err != nil {
		return Bucket{}, Error.Wrap(convertErrors(err))
	}
	return respBucket, nil
}

// GetBucketParams parameters for GetBucketParams method.
type GetBucketParams struct {
	Name []byte
}

func (params *GetBucketParams) toRequest(header *pb.RequestHeader) *pb.BucketGetRequest {
	return &pb.BucketGetRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for batch request.
func (params *GetBucketParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGet{
			BucketGet: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketParams) IsRetriable() bool { return true }

// GetBucketResponse response for GetBucket request.
type GetBucketResponse struct {
	Bucket Bucket
}

func newGetBucketResponse(response *pb.BucketGetResponse) (GetBucketResponse, error) {
	bucket, err := convertProtoToBucket(response.Bucket)
	if err != nil {
		return GetBucketResponse{}, err
	}
	return GetBucketResponse{
		Bucket: bucket,
	}, nil
}

// GetBucket returns a bucket.
func (client *Client) GetBucket(ctx context.Context, params GetBucketParams) (respBucket Bucket, err error) {
	defer mon.Task()(&ctx)(&err)

	var items []BatchResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		// TODO(moby) make sure bucket not found is properly handled
		items, err = client.Batch(ctx, &params)
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return Bucket{}, ErrBucketNotFound.Wrap(err)
		}
		return Bucket{}, Error.Wrap(err)
	}
	if len(items) != 1 {
		return Bucket{}, Error.New("unexpected number of responses: %d", len(items))
	}
	response, ok := items[0].pbResponse.(*pb.BatchResponseItem_BucketGet)
	if !ok {
		return Bucket{}, Error.New("unexpected response type: %T", items[0].pbResponse)
	}

	respBucket, err = convertProtoToBucket(response.BucketGet.Bucket)
	if err != nil {
		return Bucket{}, Error.Wrap(err)
	}
	return respBucket, nil
}

// GetBucketLocationParams parameters for GetBucketLocation method.
type GetBucketLocationParams struct {
	Name []byte
}

func (params *GetBucketLocationParams) toRequest(header *pb.RequestHeader) *pb.GetBucketLocationRequest {
	return &pb.GetBucketLocationRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for batch request.
func (params *GetBucketLocationParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGetLocation{
			BucketGetLocation: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketLocationParams) IsRetriable() bool { return true }

// GetBucketLocationResponse response for GetBucketLocation request.
type GetBucketLocationResponse struct {
	Location []byte
}

// GetBucketLocation returns a bucket location.
func (client *Client) GetBucketLocation(ctx context.Context, params GetBucketLocationParams) (_ GetBucketLocationResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetBucketLocationResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetBucketLocation(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return GetBucketLocationResponse{}, Error.Wrap(err)
	}

	return GetBucketLocationResponse{
		Location: response.Location,
	}, nil
}

// GetBucketTaggingParams represents parameters for the GetBucketTagging method.
type GetBucketTaggingParams struct {
	Name []byte
}

func (params *GetBucketTaggingParams) toRequest(header *pb.RequestHeader) *pb.GetBucketTaggingRequest {
	return &pb.GetBucketTaggingRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for a batch request.
func (params *GetBucketTaggingParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGetTagging{
			BucketGetTagging: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketTaggingParams) IsRetriable() bool { return true }

// GetBucketTaggingResponse represents the response to a GetBucketTagging request.
type GetBucketTaggingResponse struct {
	Tags []BucketTag
}

// GetBucketTagging returns the set of tags placed on a bucket.
func (client *Client) GetBucketTagging(ctx context.Context, params GetBucketTaggingParams) (_ GetBucketTaggingResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetBucketTaggingResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetBucketTagging(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return GetBucketTaggingResponse{}, Error.Wrap(convertErrors(err))
	}

	tags := make([]BucketTag, 0, len(response.Tags))
	for _, protoTag := range response.Tags {
		tags = append(tags, BucketTag{
			Key:   string(protoTag.Key),
			Value: string(protoTag.Value),
		})
	}

	return GetBucketTaggingResponse{
		Tags: tags,
	}, nil
}

// SetBucketTaggingParams represents parameters for the SetBucketTagging method.
type SetBucketTaggingParams struct {
	Name []byte
	Tags []BucketTag
}

func (params *SetBucketTaggingParams) toRequest(header *pb.RequestHeader) *pb.SetBucketTaggingRequest {
	pbTags := make([]*pb.BucketTag, 0, len(params.Tags))
	for _, tag := range params.Tags {
		pbTags = append(pbTags, &pb.BucketTag{
			Key:   []byte(tag.Key),
			Value: []byte(tag.Value),
		})
	}
	return &pb.SetBucketTaggingRequest{
		Header: header,
		Name:   params.Name,
		Tags:   pbTags,
	}
}

// BatchItem returns single item for a batch request.
func (params *SetBucketTaggingParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketSetTagging{
			BucketSetTagging: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *SetBucketTaggingParams) IsRetriable() bool { return false }

// SetBucketTagging places a set of tags on a bucket.
func (client *Client) SetBucketTagging(ctx context.Context, params SetBucketTaggingParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetBucketTagging(ctx, params.toRequest(client.header()))
		return err
	})
	return Error.Wrap(convertErrors(err))
}

// GetBucketVersioningParams parameters for GetBucketVersioning method.
type GetBucketVersioningParams struct {
	Name []byte
}

func (params *GetBucketVersioningParams) toRequest(header *pb.RequestHeader) *pb.GetBucketVersioningRequest {
	return &pb.GetBucketVersioningRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for batch request.
func (params *GetBucketVersioningParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGetVersioning{
			BucketGetVersioning: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketVersioningParams) IsRetriable() bool { return true }

// GetBucketVersioningResponse response for GetBucketVersioning request.
type GetBucketVersioningResponse struct {
	Versioning int32
}

// GetBucketVersioning returns a bucket versioning state.
func (client *Client) GetBucketVersioning(ctx context.Context, params GetBucketVersioningParams) (_ GetBucketVersioningResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetBucketVersioningResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetBucketVersioning(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return GetBucketVersioningResponse{}, Error.Wrap(convertErrors(err))
	}

	return GetBucketVersioningResponse{
		Versioning: response.Versioning,
	}, nil
}

// SetBucketVersioningParams parameters for SetBucketVersioning method.
type SetBucketVersioningParams struct {
	Name       []byte
	Versioning bool
}

func (params *SetBucketVersioningParams) toRequest(header *pb.RequestHeader) *pb.SetBucketVersioningRequest {
	return &pb.SetBucketVersioningRequest{
		Header:     header,
		Name:       params.Name,
		Versioning: params.Versioning,
	}
}

// BatchItem returns single item for batch request.
func (params *SetBucketVersioningParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketSetVersioning{
			BucketSetVersioning: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *SetBucketVersioningParams) IsRetriable() bool { return false }

// SetBucketVersioning attempts to enable/disable versioning for a bucket.
func (client *Client) SetBucketVersioning(ctx context.Context, params SetBucketVersioningParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetBucketVersioning(ctx, params.toRequest(client.header()))
		return err
	})
	return Error.Wrap(convertErrors(err))
}

// GetBucketObjectLockConfigurationParams parameters for GetBucketObjectLockConfiguration method.
type GetBucketObjectLockConfigurationParams struct {
	Name []byte
}

func (params *GetBucketObjectLockConfigurationParams) toRequest(header *pb.RequestHeader) *pb.GetBucketObjectLockConfigurationRequest {
	return &pb.GetBucketObjectLockConfigurationRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for batch request.
func (params *GetBucketObjectLockConfigurationParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGetObjectLockConfiguration{
			BucketGetObjectLockConfiguration: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketObjectLockConfigurationParams) IsRetriable() bool { return true }

// GetBucketObjectLockConfigurationResponse response for GetBucketObjectLockConfiguration request.
type GetBucketObjectLockConfigurationResponse struct {
	Enabled bool
	*DefaultRetention
}

// GetBucketObjectLockConfiguration returns a bucket object lock configuration.
func (client *Client) GetBucketObjectLockConfiguration(ctx context.Context, params GetBucketObjectLockConfigurationParams) (_ GetBucketObjectLockConfigurationResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetBucketObjectLockConfigurationResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetBucketObjectLockConfiguration(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return GetBucketObjectLockConfigurationResponse{}, Error.Wrap(convertErrors(err))
	}

	if response.Configuration == nil {
		return GetBucketObjectLockConfigurationResponse{}, ErrBucketNoLock.New("missing object lock configuration")
	}
	var defaultRetention *DefaultRetention
	if response.Configuration.DefaultRetention != nil {
		defaultRetention = &DefaultRetention{
			Mode:  storj.RetentionMode(response.Configuration.DefaultRetention.Mode),
			Years: response.Configuration.DefaultRetention.GetYears(),
			Days:  response.Configuration.DefaultRetention.GetDays(),
		}
	}
	return GetBucketObjectLockConfigurationResponse{
		Enabled:          response.Configuration.Enabled,
		DefaultRetention: defaultRetention,
	}, nil
}

// SetBucketObjectLockConfigurationParams parameters for SetBucketObjectLockConfiguration method.
type SetBucketObjectLockConfigurationParams struct {
	Name    []byte
	Enabled bool
	*DefaultRetention
}

func (params *SetBucketObjectLockConfigurationParams) toRequest(header *pb.RequestHeader) *pb.SetBucketObjectLockConfigurationRequest {
	configuration := &pb.ObjectLockConfiguration{
		Enabled: params.Enabled,
	}

	if params.DefaultRetention != nil {
		defaultRetention := &pb.DefaultRetention{
			Mode: pb.Retention_Mode(params.DefaultRetention.Mode),
		}

		if params.DefaultRetention.Days > 0 {
			defaultRetention.Duration = &pb.DefaultRetention_Days{Days: params.DefaultRetention.Days}
		} else if params.DefaultRetention.Years > 0 {
			defaultRetention.Duration = &pb.DefaultRetention_Years{Years: params.DefaultRetention.Years}
		}

		configuration.DefaultRetention = defaultRetention
	}

	return &pb.SetBucketObjectLockConfigurationRequest{
		Header:        header,
		Name:          params.Name,
		Configuration: configuration,
	}
}

// BatchItem returns single item for batch request.
func (params *SetBucketObjectLockConfigurationParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketSetObjectLockConfiguration{
			BucketSetObjectLockConfiguration: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *SetBucketObjectLockConfigurationParams) IsRetriable() bool { return false }

// SetBucketObjectLockConfiguration updates a bucket object lock configuration.
func (client *Client) SetBucketObjectLockConfiguration(ctx context.Context, params SetBucketObjectLockConfigurationParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetBucketObjectLockConfiguration(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return Error.Wrap(convertErrors(err))
	}

	return nil
}

// DeleteBucketParams parameters for DeleteBucket method.
type DeleteBucketParams struct {
	Name                      []byte
	DeleteAll                 bool
	BypassGovernanceRetention bool
}

func (params *DeleteBucketParams) toRequest(header *pb.RequestHeader) *pb.BucketDeleteRequest {
	return &pb.BucketDeleteRequest{
		Header: header,
		Name:   params.Name,

		DeleteAll:                 params.DeleteAll,
		BypassGovernanceRetention: params.BypassGovernanceRetention,
	}
}

// BatchItem returns single item for batch request.
func (params *DeleteBucketParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketDelete{
			BucketDelete: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *DeleteBucketParams) IsRetriable() bool { return false }

// DeleteBucket deletes a bucket.
func (client *Client) DeleteBucket(ctx context.Context, params DeleteBucketParams) (_ Bucket, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.BucketDeleteResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		// TODO(moby) make sure bucket not found is properly handled
		response, err = client.client.DeleteBucket(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		if errs2.IsRPC(err, rpcstatus.NotFound) {
			return Bucket{}, ErrBucketNotFound.Wrap(err)
		}
		return Bucket{}, Error.Wrap(err)
	}

	respBucket, err := convertProtoToBucket(response.Bucket)
	if err != nil {
		return Bucket{}, Error.Wrap(err)
	}
	return respBucket, nil
}

// ListBucketsParams parameters for ListBucketsParams method.
type ListBucketsParams struct {
	ListOpts BucketListOptions
}

func (params *ListBucketsParams) toRequest(header *pb.RequestHeader) *pb.BucketListRequest {
	return &pb.BucketListRequest{
		Header:    header,
		Cursor:    []byte(params.ListOpts.Cursor),
		Limit:     int32(params.ListOpts.Limit),
		Direction: params.ListOpts.Direction,
	}
}

// BatchItem returns single item for batch request.
func (params *ListBucketsParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketList{
			BucketList: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *ListBucketsParams) IsRetriable() bool { return true }

// ListBucketsResponse response for ListBucket request.
type ListBucketsResponse struct {
	BucketList BucketList
}

func newListBucketsResponse(response *pb.BucketListResponse) ListBucketsResponse {
	bucketList := BucketList{
		More: response.More,
	}
	bucketList.Items = make([]Bucket, len(response.Items))
	for i, item := range response.GetItems() {
		bucketList.Items[i] = Bucket{
			Name:    string(item.Name),
			Created: item.CreatedAt,
		}
	}
	return ListBucketsResponse{
		BucketList: bucketList,
	}
}

// ListBuckets lists buckets.
func (client *Client) ListBuckets(ctx context.Context, params ListBucketsParams) (_ BucketList, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.BucketListResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.ListBuckets(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return BucketList{}, Error.Wrap(err)
	}

	resultBucketList := BucketList{
		More: response.GetMore(),
	}
	resultBucketList.Items = make([]Bucket, len(response.GetItems()))
	for i, item := range response.GetItems() {
		resultBucketList.Items[i] = Bucket{
			Name:        string(item.GetName()),
			Created:     item.GetCreatedAt(),
			Attribution: string(item.GetUserAgent()),
		}
	}
	return resultBucketList, nil
}

func convertProtoToBucket(pbBucket *pb.Bucket) (bucket Bucket, err error) {
	if pbBucket == nil {
		return Bucket{}, nil
	}

	return Bucket{
		Name:    string(pbBucket.GetName()),
		Created: pbBucket.GetCreatedAt(),
	}, nil
}

// GetBucketNotificationConfigurationParams parameters for GetBucketNotificationConfiguration method.
type GetBucketNotificationConfigurationParams struct {
	Name []byte
}

func (params *GetBucketNotificationConfigurationParams) toRequest(header *pb.RequestHeader) *pb.GetBucketNotificationConfigurationRequest {
	return &pb.GetBucketNotificationConfigurationRequest{
		Header: header,
		Name:   params.Name,
	}
}

// BatchItem returns single item for batch request.
func (params *GetBucketNotificationConfigurationParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketGetNotificationConfiguration{
			BucketGetNotificationConfiguration: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *GetBucketNotificationConfigurationParams) IsRetriable() bool { return true }

// GetBucketNotificationConfigurationResponse response for GetBucketNotificationConfiguration request.
type GetBucketNotificationConfigurationResponse struct {
	Configuration *BucketNotificationConfiguration
}

// GetBucketNotificationConfiguration returns a bucket notification configuration.
func (client *Client) GetBucketNotificationConfiguration(ctx context.Context, params GetBucketNotificationConfigurationParams) (_ GetBucketNotificationConfigurationResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	var response *pb.GetBucketNotificationConfigurationResponse
	err = WithRetry(ctx, func(ctx context.Context) error {
		response, err = client.client.GetBucketNotificationConfiguration(ctx, params.toRequest(client.header()))
		return err
	})
	if err != nil {
		return GetBucketNotificationConfigurationResponse{}, Error.Wrap(convertErrors(err))
	}

	var config *BucketNotificationConfiguration
	if response.Configuration != nil {
		filter := FilterRule{}
		if response.Configuration.Filter != nil {
			filter.Prefix = response.Configuration.Filter.Prefix
			filter.Suffix = response.Configuration.Filter.Suffix
		}

		config = &BucketNotificationConfiguration{
			ID:         response.Configuration.Id,
			TopicName:  response.Configuration.TopicName,
			Events:     response.Configuration.Events,
			FilterRule: filter,
		}
	}

	return GetBucketNotificationConfigurationResponse{
		Configuration: config,
	}, nil
}

// SetBucketNotificationConfigurationParams parameters for SetBucketNotificationConfiguration method.
type SetBucketNotificationConfigurationParams struct {
	Name          []byte
	Configuration *BucketNotificationConfiguration
}

func (params *SetBucketNotificationConfigurationParams) toRequest(header *pb.RequestHeader) *pb.SetBucketNotificationConfigurationRequest {
	var pbConfig *pb.NotificationConfiguration
	if params.Configuration != nil {
		pbConfig = &pb.NotificationConfiguration{
			Id:        params.Configuration.ID,
			TopicName: params.Configuration.TopicName,
			Events:    params.Configuration.Events,
			Filter: &pb.FilterRule{
				Prefix: params.Configuration.FilterRule.Prefix,
				Suffix: params.Configuration.FilterRule.Suffix,
			},
		}
	}

	return &pb.SetBucketNotificationConfigurationRequest{
		Header:        header,
		Name:          params.Name,
		Configuration: pbConfig,
	}
}

// BatchItem returns single item for batch request.
func (params *SetBucketNotificationConfigurationParams) BatchItem() *pb.BatchRequestItem {
	return &pb.BatchRequestItem{
		Request: &pb.BatchRequestItem_BucketSetNotificationConfiguration{
			BucketSetNotificationConfiguration: params.toRequest(nil),
		},
	}
}

// IsRetriable returns true if the request can be retried when a kind of connection error happens, otherwise false.
func (params *SetBucketNotificationConfigurationParams) IsRetriable() bool { return false }

// SetBucketNotificationConfiguration sets a bucket notification configuration.
func (client *Client) SetBucketNotificationConfiguration(ctx context.Context, params SetBucketNotificationConfigurationParams) (err error) {
	defer mon.Task()(&ctx)(&err)

	err = WithRetry(ctx, func(ctx context.Context) error {
		_, err = client.client.SetBucketNotificationConfiguration(ctx, params.toRequest(client.header()))
		return err
	})
	return Error.Wrap(convertErrors(err))
}
