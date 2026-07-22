// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package cursed

import (
	"context"

	"github.com/zeebo/errs"

	"storj.io/common/pb"
	"storj.io/common/rpc/rpcstatus"
	"storj.io/common/signing"
	"storj.io/common/storj"
)

// PackSegmentID takes a satellite signer and a segment ID and returns a storj.SegmentID.
func PackSegmentID(ctx context.Context, satellite signing.Signer, satSegmentID *SegmentID) (segmentID storj.SegmentID, err error) {
	defer mon.Task()(&ctx)(&err)

	if satSegmentID == nil {
		return nil, rpcstatus.Error(rpcstatus.Internal, "unable to create segment id")
	}

	// remove satellite signature from limits to reduce response size
	// signature is not needed here, because segment id is signed by satellite
	originalOrderLimits := make([]*pb.AddressedOrderLimit, len(satSegmentID.OriginalOrderLimits))
	for i, alimit := range satSegmentID.OriginalOrderLimits {
		originalOrderLimits[i] = &pb.AddressedOrderLimit{
			StorageNodeAddress: alimit.StorageNodeAddress,
			Limit: &pb.OrderLimit{
				SerialNumber:           alimit.Limit.SerialNumber,
				SatelliteId:            alimit.Limit.SatelliteId,
				UplinkPublicKey:        alimit.Limit.UplinkPublicKey,
				StorageNodeId:          alimit.Limit.StorageNodeId,
				PieceId:                alimit.Limit.PieceId,
				Limit:                  alimit.Limit.Limit,
				PieceExpiration:        alimit.Limit.PieceExpiration,
				Action:                 alimit.Limit.Action,
				OrderExpiration:        alimit.Limit.OrderExpiration,
				OrderCreation:          alimit.Limit.OrderCreation,
				EncryptedMetadataKeyId: alimit.Limit.EncryptedMetadataKeyId,
				EncryptedMetadata:      alimit.Limit.EncryptedMetadata,
				// don't copy satellite signature
			},
		}
	}

	satSegmentID.OriginalOrderLimits = originalOrderLimits

	signedSegmentID, err := SignSegmentID(ctx, satellite, satSegmentID)
	if err != nil {
		return nil, err
	}

	encodedSegmentID, err := pb.Marshal(signedSegmentID)
	if err != nil {
		return nil, err
	}

	segmentID, err = storj.SegmentIDFromBytes(encodedSegmentID)
	if err != nil {
		return nil, err
	}
	return segmentID, nil
}

// SignSegmentID signs the segment ID using the specified signer.
// Signer is a satellite.
func SignSegmentID(ctx context.Context, signer signing.Signer, unsigned *SegmentID) (_ *SegmentID, err error) {
	defer mon.Task()(&ctx)(&err)
	bytes, err := EncodeSegmentID(ctx, unsigned)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	signed := *unsigned
	signed.SatelliteSignature, err = signer.HashAndSign(ctx, bytes)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return &signed, nil
}

// EncodeSegmentID encodes segment ID into bytes for signing.
func EncodeSegmentID(ctx context.Context, segmentID *SegmentID) (_ []byte, err error) {
	defer mon.Task()(&ctx)(&err)
	signature := segmentID.SatelliteSignature
	segmentID.SatelliteSignature = nil
	out, err := pb.Marshal(segmentID)
	segmentID.SatelliteSignature = signature
	return out, err
}
