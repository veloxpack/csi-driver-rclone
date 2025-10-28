// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"fmt"
	"sync/atomic"
)

// PiecesProgress is an interesting concurrency primitive we don't know what
// to call, but it's kind of like those old clocks that employees would clock
// in and out of. Imagine that Looney Toons episode where Sam the Sheepdog and
// Wile E Coyote are clocking in and out.
//
// In our case, PiecesProgress is two dimensional:
//   - There is the neededShares vs total dimension - this is the number of shares
//     that are necessary for Reed Solomon construction, for instance.
//   - There is the current watermark of what stripe specific pieces are working
//     on.
//
// A bunch of piece readers will be keeping PiecesProgress updated with how
// many shares they have downloaded within their piece. When a piece reader
// tells PiecesProgress that they have downloaded n shares, and that was the
// trigger that means we now have the neededShares necessary shares at a certain
// watermark or stripe, PiecesProgress will tell the piece reader to wake
// up the combining layer.
//
// This data structure is designed to cause these wakeups to happen as little
// as possible.
type PiecesProgress struct {
	neededShares        *atomic.Int32
	stripesNeeded       *atomic.Int32
	newStripeReady      *atomic.Int32
	pieceSharesReceived []atomic.Int32
}

// NewPiecesProgress constructs PiecesProgress with a neededShares number of
// necessary shares per stripe, out of total total shares. PiecesProgress
// doesn't care about how many stripes there are but will keep track of which
// stripe each share reader is on.
func NewPiecesProgress(minimum, total int32) *PiecesProgress {
	// In order to keep all the mutable data on a single cache line, we
	// store all values in a single memory slice. the first few values are
	// named and get dedicated pointers, and the rest are pieceSharesReceived.
	memory := make([]atomic.Int32, total+3)
	memory[0].Store(minimum)
	return &PiecesProgress{
		neededShares:        &memory[0],
		stripesNeeded:       &memory[1],
		newStripeReady:      &memory[2],
		pieceSharesReceived: memory[3:],
	}
}

// ProgressSnapshot returns a snapshot of the current progress. No locks are held
// so it doesn't represent a single point in time in the presence of concurrent
// mutations.
func (y *PiecesProgress) ProgressSnapshot(out []int32) []int32 {
	for i := range y.pieceSharesReceived {
		out = append(out, y.pieceSharesReceived[i].Load())
	}
	return out
}

// SetStripesNeeded tells PiecesProgress what neededShares stripe is needed next.
func (y *PiecesProgress) SetStripesNeeded(required int32) {
	y.stripesNeeded.Store(required)
}

// IncreaseNeededShares tells PiecesProgress that going forward, we need more
// shares per watermark value.
func (y *PiecesProgress) IncreaseNeededShares() bool {
	for {
		min := y.NeededShares()
		if min >= int32(len(y.pieceSharesReceived)) {
			return false
		}
		if y.neededShares.CompareAndSwap(min, min+1) {
			return true
		}
	}
}

// NeededShares returns the current value of the number of shares required at a given
// watermark.
func (y *PiecesProgress) NeededShares() int32 {
	return y.neededShares.Load()
}

// PieceSharesReceived returns the current watermark for reader idx.
func (y *PiecesProgress) PieceSharesReceived(idx int) int32 {
	if idx < 0 || idx >= len(y.pieceSharesReceived) {
		return 0
	}
	return y.pieceSharesReceived[idx].Load()
}

// AcknowledgeNewStripes tells PiecesProgress that the combiner has woken up and
// new alarms are okay to trigger.
func (y *PiecesProgress) AcknowledgeNewStripes() {
	y.newStripeReady.Store(0)
}

// SharesCompleted adds some read events to a given index. If SharesCompleted
// returns true, then the calling reader should wake up the combiner.
func (y *PiecesProgress) SharesCompleted(idx int, delta int32) bool {
	if idx < 0 || idx >= len(y.pieceSharesReceived) {
		if debugEnabled {
			fmt.Println("bundy", idx, "out of range")
		}
		return false
	}

	val := y.pieceSharesReceived[idx].Add(delta)

	stripeNeeded := y.stripesNeeded.Load()
	stripeReady := y.newStripeReady.Load() != 0

	if debugEnabled {
		fmt.Println("bundy", idx, "delta", delta, "counter", val, "stripeNeed", stripeNeeded, "newStripeReady", stripeReady)
	}

	// it's only possible to wake if the value we stored could increase the
	// number of shares larger than the required amount, and there is the potential
	// of a wake available.
	if val < stripeNeeded || stripeReady {
		return false
	}

	// count all of the pieceSharesReceived that are larger than the required share. if we don't
	// have enough, no need to wake yet.
	c := int32(0)
	for i := 0; i < len(y.pieceSharesReceived); i++ {
		if y.pieceSharesReceived[i].Load() >= stripeNeeded {
			c++
		}
	}
	if debugEnabled {
		fmt.Println("bundy", idx, "found", c)
	}
	if c < y.neededShares.Load() {
		return false
	}

	if debugEnabled {
		fmt.Println("bundy", idx, "attempting wake claim")
	}
	// we are sure there's enough shares required. acquire explicit ownership of notification.
	return y.newStripeReady.CompareAndSwap(0, 1)
}
