// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package piecestore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/zeebo/errs"

	"storj.io/common/identity"
	"storj.io/common/memory"
	"storj.io/common/pb"
	"storj.io/common/rpc"
	"storj.io/common/storj"
)

// NoiseEnabled indicates whether Noise is enabled in this build.
var NoiseEnabled = os.Getenv("STORJ_NOISE_DISABLED_EXPERIMENTAL") != "true"

var errMessageTimeout = errors.New("message timeout")

const (
	// retainMessageLimit defines the max size which can be sent via normal request/response protocol.
	// 4MB is the absolute max, defined by DRPC RPC, but we need a few bytes for protobuf overhead / filter header.
	retainMessageLimit = 4100050
)

var (
	// Error is the default error class for piecestore client.
	Error = errs.Class("piecestore")

	// CloseError is the error class used for errors generated during a
	// stream close in a piecestore client.
	//
	// Errors of this type should also be wrapped with Error, for backwards
	// compatibility.
	CloseError = errs.Class("piecestore close")
)

// Config defines piecestore client parameters for upload and download.
type Config struct {
	DownloadBufferSize int64
	UploadBufferSize   int64

	InitialStep      int64
	MaximumStep      int64
	MaximumChunkSize int32

	MessageTimeout time.Duration
}

// DefaultConfig are the default params used for upload and download.
var DefaultConfig = Config{
	DownloadBufferSize: 256 * memory.KiB.Int64(),
	UploadBufferSize:   64 * memory.KiB.Int64(),

	InitialStep:      256 * memory.KiB.Int64(),
	MaximumStep:      550 * memory.KiB.Int64(),
	MaximumChunkSize: 16 * memory.KiB.Int32(),

	MessageTimeout: 10 * time.Minute,
}

// Client implements uploading, downloading and deleting content from a piecestore.
type Client struct {
	client         pb.DRPCPiecestoreClient
	replaySafe     pb.DRPCReplaySafePiecestoreClient
	nodeURL        storj.NodeURL
	conn           *rpc.Conn
	config         Config
	UploadHashAlgo pb.PieceHashAlgorithm
}

// Dial dials the target piecestore endpoint.
func Dial(ctx context.Context, dialer rpc.Dialer, nodeURL storj.NodeURL, config Config) (*Client, error) {
	conn, err := dialer.DialNodeURL(ctx, nodeURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &Client{
		client:  pb.NewDRPCPiecestoreClient(conn),
		nodeURL: nodeURL,
		conn:    conn,
		config:  config,
	}, nil
}

// DialReplaySafe dials the target piecestore endpoint for replay safe request types.
func DialReplaySafe(ctx context.Context, dialer rpc.Dialer, nodeURL storj.NodeURL, config Config) (*Client, error) {
	conn, err := dialer.DialNode(ctx, nodeURL, rpc.DialOptions{ReplaySafe: NoiseEnabled})
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &Client{
		replaySafe: pb.NewDRPCReplaySafePiecestoreClient(conn),
		nodeURL:    nodeURL,
		conn:       conn,
		config:     config,
	}, nil

}

// Retain uses a bloom filter to tell the piece store which pieces to keep.
func (client *Client) Retain(ctx context.Context, req *pb.RetainRequest) (err error) {
	defer mon.Task()(&ctx)(&err)

	// TODO: eventually we will switch to the new protocol by default (because it has hash)
	// but it requires more changes in storj.io/storj/storagenode first
	// until that, we can use the old protocol, by default.
	if len(req.Filter) <= retainMessageLimit {
		_, err = client.client.Retain(ctx, req)
		return err
	}

	// it's a big message, we will use the RetainBig, stream-based protocol

	// hash is calculated on the full message, splitting error can also be detected.
	hasher := pb.NewHashFromAlgorithm(pb.PieceHashAlgorithm_BLAKE3)
	hasher.Write(req.Filter)
	hash := hasher.Sum([]byte{})

	stream, err := client.client.RetainBig(ctx)
	if err != nil {
		return Error.Wrap(err)
	}
	var lastMessage bool
	for i := 0; i <= len(req.Filter); i += retainMessageLimit {

		endOffset := i + retainMessageLimit

		if endOffset > len(req.Filter) {
			lastMessage = true
			endOffset = len(req.Filter)

		}
		req := &pb.RetainRequest{
			CreationDate: req.CreationDate,
			Filter:       req.Filter[i:endOffset],
		}

		if lastMessage {
			req.HashAlgorithm = pb.PieceHashAlgorithm_BLAKE3
			req.Hash = hash
		}

		err = stream.Send(req)
		if err != nil {
			// this is too big, no hope to send with the legacy endpoint
			return Error.Wrap(errs.Combine(err, stream.CloseSend()))
		}
	}
	err = stream.Close()
	return Error.Wrap(err)
}

// RetainRequestFromStream is the inverse logic of Client.Retain method, which splits the retain messages to smaller chunks.
// strictly speaking, it's a server side code, but it's easier to maintain and test here, as it should be the
// opposite of the Retain method.
func RetainRequestFromStream(stream pb.DRPCPiecestore_RetainBigStream) (pb.RetainRequest, error) {
	resp := pb.RetainRequest{}
	for {
		req, err := stream.Recv()
		if err != nil {
			return resp, err
		}
		if resp.CreationDate.IsZero() {
			resp.CreationDate = req.CreationDate
		}
		resp.Filter = append(resp.Filter, req.Filter...)
		if len(req.Hash) > 0 {
			hasher := pb.NewHashFromAlgorithm(req.HashAlgorithm)
			_, err := hasher.Write(resp.Filter)
			if err != nil {
				return resp, errs.Wrap(err)
			}
			if !bytes.Equal(req.Hash, hasher.Sum(nil)) {
				return resp, errs.New("Hash mismatch")
			}
			break
		}
	}
	return resp, stream.Close()
}

// Close closes the underlying connection.
func (client *Client) Close() error {
	return client.conn.Close()
}

// GetPeerIdentity gets the connection's peer identity. This doesn't work
// on Noise-based connections.
func (client *Client) GetPeerIdentity() (*identity.PeerIdentity, error) {
	return client.conn.PeerIdentity()
}

// NodeURL returns the Node we dialed.
func (client *Client) NodeURL() storj.NodeURL { return client.nodeURL }

// next allocation step find the next trusted step.
func (client *Client) nextOrderStep(previous int64) int64 {
	// TODO: ensure that this is frame idependent
	next := previous * 3 / 2
	if next > client.config.MaximumStep {
		next = client.config.MaximumStep
	}
	return next
}

// ignoreEOF is an utility func for ignoring EOF error, when it's not important.
func ignoreEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
