// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package pieceupload

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spacemonkeygo/monkit/v3"

	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/uplink/private/testuplink"
)

var mon = monkit.Package()

// PiecePutter puts pieces.
type PiecePutter interface {
	// PutPiece puts a piece using the given limit and private key. The
	// operation can be cancelled using the longTailCtx or uploadCtx is
	// cancelled.
	PutPiece(longTailCtx, uploadCtx context.Context, limit *pb.AddressedOrderLimit, privateKey storj.PiecePrivateKey, data io.ReadCloser) (hash *pb.PieceHash, deprecated *struct{}, err error)
}

var (
	errTryAgain = errors.New("try upload again")
)

// StallDetectedError indicates a piece was taking too long so the stall manager cancelled it.
type StallDetectedError struct{}

func (e StallDetectedError) Error() string { return "piece stall detected, context cancelled" }

// OptimalThresholdError indicates a redundant piece cancellation because the segment reached its optimal threshold
// (the long tail of piece upload options was cancelled).
type OptimalThresholdError struct{}

func (e OptimalThresholdError) Error() string {
	return "cancelled due to optimal threshold (long-tail cancellation)"
}

// UploadOne uploads one piece from the manager using the given private key. If
// it fails, it will attempt to upload another until either the upload context,
// or the long tail context is cancelled.
// The stallManager parameter is optional. If nil, no stall detection will be used.
// Returns: success (bool), tags (map), node (byte), stallCount (int), err (error)
// stallCount indicates how many pieces were cancelled due to stall detection during this upload attempt.
func UploadOne(
	longTailCtx context.Context,
	uploadCtx context.Context,
	manager *Manager,
	putter PiecePutter,
	privateKey storj.PiecePrivateKey,
	stallManager *StallManager,
) (success bool, tags map[string]string, node byte, stallCount int, err error) {
	defer mon.Task()(&longTailCtx)(&err)
	// If the long tail context is cancelled, then return a nil error.
	defer func() {
		if longTailCtx.Err() != nil {
			err = nil
		}
	}()

	var stalls int
	for {
		piece, limit, tags, node, done, err := manager.NextPiece(longTailCtx)
		if err != nil {
			return false, nil, 0, stalls, err
		}

		var pieceID string
		if limit.Limit != nil {
			pieceID = limit.Limit.PieceId.String()
		}

		var address, noise string
		if limit.StorageNodeAddress != nil {
			address = fmt.Sprintf("%-21s", limit.StorageNodeAddress.Address)
			noise = fmt.Sprintf("%-5t", limit.StorageNodeAddress.NoiseInfo != nil)
		}

		logCtx := testuplink.WithLogWriterContext(uploadCtx,
			"piece_id", pieceID,
			"address", address,
			"noise", noise,
		)

		success, err := func() (bool, error) {
			// Handle stall contexts based on stalls availability
			var ctx context.Context
			var cleanup func()

			if stallManager != nil {
				ctx, cleanup = stallManager.Watch(longTailCtx)
				testuplink.Log(logCtx, "Uploading piece (with stall detection)...")
			} else {
				// No stall manager, use longTailCtx directly
				ctx = longTailCtx
				cleanup = func() {}
				testuplink.Log(logCtx, "Uploading piece...")
			}
			defer cleanup()

			hash, _, err := putter.PutPiece(ctx, uploadCtx, limit, privateKey, io.NopCloser(piece))
			testuplink.Log(logCtx, "Done uploading piece. err:", err)

			// Track success or failure
			done(hash, err == nil)

			if err == nil {
				return true, nil
			}

			if err := uploadCtx.Err(); err != nil {
				return false, err
			}

			if ctx.Err() != nil {
				// If we have a stall manager, we need to distinguish between
				// stall detection timeout vs optimal threshold (long-tail) cancellation.
				if stallManager != nil && longTailCtx.Err() == nil {
					// Stall manager cancellation: this piece was taking longer than the
					// stall detection threshold, so this piece upload has been cancelled
					// so that the segment upload can try a different piece more quickly if necessary.
					testuplink.Log(logCtx, "Stall Detected - retrying with different node")
					return false, StallDetectedError{}
				} else {
					// Long-tail Cancellation: the outer segment upload reached the optimal threshold,
					// and so this piece has been canceled as it is now redundant.
					testuplink.Log(logCtx, "Long tail cancellation due to optimal threshold")
					return false, OptimalThresholdError{}
				}
			}

			return false, errTryAgain
		}()

		// Retry logic: continue the loop for transient errors and stall detection
		if errors.Is(err, errTryAgain) {
			continue
		}
		var stallError StallDetectedError
		if errors.As(err, &stallError) {
			// Stall detected - increment counter and retry with a different node
			stalls++
			continue
		}
		return success, tags, node, stalls, err
	}
}
