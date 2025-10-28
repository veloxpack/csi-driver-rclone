// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package signing

import (
	"context"

	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/common/tracing"
)

// Signee is able to verify that the data signature belongs to the signee.
type Signee interface {
	ID() storj.NodeID
	HashAndVerifySignature(ctx context.Context, data, signature []byte) error
}

// VerifyOrderLimitSignature verifies that the signature inside order limit is valid and  belongs to the satellite.
func VerifyOrderLimitSignature(ctx context.Context, satellite Signee, signed *pb.OrderLimit) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyOrderLimitSignature(ctx, satellite, signed)
}

// VerifyOrderSignature verifies that the signature inside order is valid and belongs to the uplink.
func VerifyOrderSignature(ctx context.Context, uplink Signee, signed *pb.Order) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyOrderSignature(ctx, uplink, signed)
}

// VerifyUplinkOrderSignature verifies that the signature inside order is valid and belongs to the uplink.
func VerifyUplinkOrderSignature(ctx context.Context, publicKey storj.PiecePublicKey, signed *pb.Order) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyUplinkOrderSignature(ctx, publicKey, signed)
}

// VerifyPieceHashSignature verifies that the signature inside piece hash is valid and belongs to the signer, which is either uplink or storage node.
func VerifyPieceHashSignature(ctx context.Context, signee Signee, signed *pb.PieceHash) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyPieceHashSignature(ctx, signee, signed)
}

// VerifyUplinkPieceHashSignature verifies that the signature inside piece hash is valid and belongs to the signer, which is either uplink or storage node.
func VerifyUplinkPieceHashSignature(ctx context.Context, publicKey storj.PiecePublicKey, signed *pb.PieceHash) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyUplinkPieceHashSignature(ctx, publicKey, signed)
}

// VerifyExitCompleted verifies that the signature inside ExitCompleted belongs to the satellite.
func VerifyExitCompleted(ctx context.Context, satellite Signee, signed *pb.ExitCompleted) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyExitCompleted(ctx, satellite, signed)
}

// VerifyExitFailed verifies that the signature inside ExitFailed belongs to the satellite.
func VerifyExitFailed(ctx context.Context, satellite Signee, signed *pb.ExitFailed) (err error) {
	if areSignaturesDisabled(ctx) {
		return nil
	}
	return verifyExitFailed(ctx, satellite, signed)
}

var monVerifyOrderLimitSignature = mon.Task()

func verifyOrderLimitSignature(ctx context.Context, satellite Signee, signed *pb.OrderLimit) (err error) {
	ctx = tracing.WithoutDistributedTracing(ctx)
	defer monVerifyOrderLimitSignature(&ctx)(&err)

	bytes, err := EncodeOrderLimit(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return satellite.HashAndVerifySignature(ctx, bytes, signed.SatelliteSignature)
}

var monVerifyOrderSignature = mon.Task()

func verifyOrderSignature(ctx context.Context, uplink Signee, signed *pb.Order) (err error) {
	defer monVerifyOrderSignature(&ctx)(&err)

	if len(signed.XXX_unrecognized) > 0 {
		return Error.New("unrecognized fields are not allowed")
	}

	bytes, err := EncodeOrder(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return uplink.HashAndVerifySignature(ctx, bytes, signed.UplinkSignature)
}

var monVerifyUplinkOrderSignature = mon.Task()

func verifyUplinkOrderSignature(ctx context.Context, publicKey storj.PiecePublicKey, signed *pb.Order) (err error) {
	ctx = tracing.WithoutDistributedTracing(ctx)
	defer monVerifyUplinkOrderSignature(&ctx)(&err)

	if len(signed.XXX_unrecognized) > 0 {
		return Error.New("unrecognized fields are not allowed")
	}

	bytes, err := EncodeOrder(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return Error.Wrap(publicKey.Verify(bytes, signed.UplinkSignature))
}

var monVerifyPieceHashSignature = mon.Task()

func verifyPieceHashSignature(ctx context.Context, signee Signee, signed *pb.PieceHash) (err error) {
	ctx = tracing.WithoutDistributedTracing(ctx)
	defer monVerifyPieceHashSignature(&ctx)(&err)

	if len(signed.XXX_unrecognized) > 0 {
		return Error.New("unrecognized fields are not allowed")
	}

	bytes, err := EncodePieceHash(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return signee.HashAndVerifySignature(ctx, bytes, signed.Signature)
}

var monVerifyUplinkPieceHashSignature = mon.Task()

func verifyUplinkPieceHashSignature(ctx context.Context, publicKey storj.PiecePublicKey, signed *pb.PieceHash) (err error) {
	ctx = tracing.WithoutDistributedTracing(ctx)
	defer monVerifyUplinkPieceHashSignature(&ctx)(&err)

	if len(signed.XXX_unrecognized) > 0 {
		return Error.New("unrecognized fields are not allowed")
	}

	bytes, err := EncodePieceHash(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return Error.Wrap(publicKey.Verify(bytes, signed.Signature))
}

func verifyExitCompleted(ctx context.Context, satellite Signee, signed *pb.ExitCompleted) (err error) {
	defer mon.Task()(&ctx)(&err)

	bytes, err := EncodeExitCompleted(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return Error.Wrap(satellite.HashAndVerifySignature(ctx, bytes, signed.ExitCompleteSignature))
}

func verifyExitFailed(ctx context.Context, satellite Signee, signed *pb.ExitFailed) (err error) {
	defer mon.Task()(&ctx)(&err)

	bytes, err := EncodeExitFailed(ctx, signed)
	if err != nil {
		return Error.Wrap(err)
	}

	return Error.Wrap(satellite.HashAndVerifySignature(ctx, bytes, signed.ExitFailureSignature))
}
