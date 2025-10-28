// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"golang.org/x/exp/slices"

	"storj.io/common/rpc/rpctracing"
	"storj.io/common/sync2"
	"storj.io/infectious"
)

const (
	debugEnabled          = false
	maxStripesAhead       = 256         // might be interesting to test different values later
	inactiveCheckInterval = time.Second // how frequently to check for inactivity
	inactiveCheckMaxCount = 5           // number of inactive checks before triggering
)

// pieceReader represents the stream of shares within one piece.
type pieceReader struct {
	shareNum     int
	source       io.Reader
	sourceCloser io.Closer
	buffer       *StreamingPiece

	backpressureMu  sync.Mutex
	backpressure    sync.Cond
	completedShares int
}

// StripeReader reads from a collection of piece io.ReadClosers in parallel,
// recombining them into a single stream using an ErasureScheme.
type StripeReader struct {
	bundy           *PiecesProgress
	pieces          []pieceReader
	scheme          ErasureScheme
	wg              sync.WaitGroup
	stripeReady     sync2.Event
	returnedStripes int32
	totalStripes    int32
	errorDetection  bool
	runningPieces   atomic.Int32
	inactive        atomic.Bool
}

// NewStripeReader makes a new StripeReader using the provided map of share
// number to io.ReadClosers, an ErasureScheme, the total number of stripes in
// the stream, and whether or not to use the Erasure Scheme's error detection.
func NewStripeReader(readers map[int]io.ReadCloser, scheme ErasureScheme, totalStripes int,
	errorDetection bool) *StripeReader {

	pool := NewBatchPool(scheme.ErasureShareSize())

	totalPieceSize := int64(totalStripes) * int64(scheme.ErasureShareSize())

	pieces := make([]pieceReader, 0, len(readers))
	for shareNum, source := range readers {
		pieces = append(pieces, pieceReader{
			shareNum:     shareNum,
			source:       io.LimitReader(source, totalPieceSize),
			sourceCloser: source,
			buffer:       NewStreamingPiece(scheme.ErasureShareSize(), totalPieceSize, pool),
		})
		piece := &pieces[len(pieces)-1]
		piece.backpressure.L = &piece.backpressureMu
	}

	minimum := int32(scheme.RequiredCount())
	if errorDetection && minimum < int32(len(pieces)) {
		minimum++
	}

	s := &StripeReader{
		bundy:          NewPiecesProgress(minimum, int32(len(pieces))),
		pieces:         pieces,
		scheme:         scheme,
		totalStripes:   int32(totalStripes),
		errorDetection: errorDetection,
	}
	s.start()
	return s
}

// start creates the goroutines to start reading each of the share streams.
func (s *StripeReader) start() {
	if debugEnabled {
		fmt.Println("starting", len(s.pieces), "readers")
	}

	var pwg sync.WaitGroup
	s.runningPieces.Store(int32(len(s.pieces)))

	for idx := range s.pieces {
		s.wg.Add(1)
		pwg.Add(1)
		go func(idx int) {
			defer s.wg.Done()
			defer pwg.Done()

			// whenever a share reader is done, we should wake up the core in case
			// this share reader just exited unsuccessfully and this represents a
			// failure to get enough pieces.
			defer s.stripeReady.Signal()

			// we should mark that there is one less running share reader.
			defer s.runningPieces.Add(-1)

			// do the work.
			s.readShares(idx)
		}(idx)
	}

	done := make(chan struct{})
	go func() {
		pwg.Wait()
		close(done)
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		s1 := s.bundy.ProgressSnapshot(nil)
		var s2 []int32

		t := time.NewTicker(inactiveCheckInterval)
		defer t.Stop()

		match := 0
		for {
			select {
			case <-t.C:
				s2 = s.bundy.ProgressSnapshot(s2[:0])

				if !slices.Equal(s1, s2) {
					match = 0
					s2, s1 = s1, s2
					continue
				}

				match++
				if match == inactiveCheckMaxCount {
					s.inactive.Store(true)
					s.stripeReady.Signal()
					return
				}

			case <-done:
				return
			}
		}
	}()
}

// readShares is the method that does the actual work of reading an individual
// share stream.
func (s *StripeReader) readShares(idx int) {
	r := &s.pieces[idx]
	stripesSoFar := 0
	for {
		// see if we can fill this index's buffer with data from r.source.
		shares, done := r.buffer.ReadSharesFrom(r.source)

		// did we get any shares?
		if shares > 0 {
			// yay!
			stripesSoFar += shares
			if debugEnabled {
				fmt.Println(idx, "read", shares, "shares")
			}
			// tell the bundy clock
			if s.bundy.SharesCompleted(idx, int32(shares)) {
				// oh hey, bundy says we just changed the situation and we should wake
				// up the core.
				if debugEnabled {
					fmt.Println(idx, "bundy counter says", shares, "is ready")
				}
				s.stripeReady.Signal()
			}
		} else if debugEnabled {
			fmt.Println(idx, "read 0 shares?")
		}

		// will we get any more shares?
		if done {
			if debugEnabled {
				fmt.Println(idx, "done")
			}
			break
		}

		r.backpressure.L.Lock()
		// how far ahead are we? are we too far ahead of the core? if so, let's
		// wait. the core will mark us completed if things are closing.
		for stripesSoFar > r.completedShares+maxStripesAhead &&
			r.completedShares < int(s.totalStripes) {
			r.backpressure.Wait()
		}
		r.backpressure.L.Unlock()
	}
}

// markCompleted updates the pieceReader's accounting of how far ahead it
// is from the core, and also tells the *StreamingPiece whether it can free up some
// internal buffers.
func (r *pieceReader) markCompleted(stripes int) {
	r.backpressure.L.Lock()
	defer r.backpressure.L.Unlock()
	r.buffer.MarkCompleted(stripes)
	if stripes > r.completedShares {
		r.completedShares = stripes
	}
	// the pieceReader might be asleep. let's wake it up.
	r.backpressure.Signal()
}

// Close does *not* close the readers it received in the constructor.
// Close does *not* wait for reader goroutines to shut down. See CloseAndWait
// if you want other behavior. Close mimics the older eestream.StripeReader
// behavior.
func (s *StripeReader) Close() error {
	for idx := range s.pieces {
		s.wg.Add(1)
		go func(idx int) {
			defer s.wg.Done()
			r := &s.pieces[idx]
			r.markCompleted(int(s.totalStripes))
		}(idx)
	}
	return nil
}

// CloseAndWait closes all readers and waits for all goroutines.
func (s *StripeReader) CloseAndWait() error {
	for idx := range s.pieces {
		s.wg.Add(1)
		go func(idx int) {
			defer s.wg.Done()
			r := &s.pieces[idx]
			_ = r.sourceCloser.Close()
			r.markCompleted(int(s.totalStripes))
		}(idx)
	}
	s.wg.Wait()
	return nil
}

func (s *StripeReader) combineErrs() error {
	var errstrings []string
	for idx := range s.pieces {
		if err := s.pieces[idx].buffer.Err(); err != nil && !errors.Is(err, io.EOF) {
			errstrings = append(errstrings, fmt.Sprintf("\nerror retrieving piece %02d: %v", s.pieces[idx].shareNum, err))
		}
	}
	if len(errstrings) > 0 {
		sort.Strings(errstrings)
		return Error.New("failed to download segment: %s", strings.Join(errstrings, ""))
	}
	return Error.New("programmer error: no errors to combine")
}

var backcompatMon = monkit.ScopeNamed("storj.io/storj/uplink/eestream")
var monReadStripeTask = mon.Task()

// ReadStripes returns 1 or more stripes. out is overwritten.
func (s *StripeReader) ReadStripes(ctx context.Context, nextStripe int64, out []byte) (_ []byte, count int, err error) {
	defer monReadStripeTask(&ctx)(&err)
	ctx = rpctracing.WithoutDistributedTracing(ctx)

	if nextStripe != int64(s.returnedStripes) {
		return nil, 0, Error.New("unexpected next stripe")
	}

	// first, some memory management. do we have a place to write the results,
	// and how many stripes can we write?
	if cap(out) <= 0 {
		out = make([]byte, 0, globalBufSize)
	}
	maxStripes := int32(cap(out) / s.scheme.StripeSize())
	if debugEnabled {
		fmt.Println("core initial stripe calc", maxStripes, s.returnedStripes, s.totalStripes)
	}
	if s.returnedStripes+maxStripes > s.totalStripes {
		maxStripes = s.totalStripes - s.returnedStripes
	}
	if maxStripes <= 0 {
		return nil, 0, io.EOF
	}

	if debugEnabled {
		fmt.Println("core downloading", maxStripes, "at stripe size", s.scheme.StripeSize(), "with cap", cap(out))
	}

	// okay, let's tell the bundy clock we just want one new stripe. hopefully
	// we get more than just 1.
	requiredWatermark := s.returnedStripes + 1
	s.bundy.SetStripesNeeded(requiredWatermark)

	// if the bundy clock wakes up, we're going to find the lowest watermark
	// with the neededShares number of shares per stripe. since we're essentially doing
	// a min operation, let's start stripesFound at the highest value we want it,
	// and we will lower it as we inspect the pieceSharesReceived on the bundy clock.
	stripesFound := s.returnedStripes + maxStripes

	ready := make([]int, 0, len(s.pieces))

	for {
		// check if we were woken from quiescence. if so, error out.
		if s.inactive.Load() {
			return nil, 0, ErrInactive.New("")
		}

		// okay let's tell the bundy clock we're awake and it should be okay to
		// wake us up again next time we sleep.
		s.bundy.AcknowledgeNewStripes()

		// let's also load the number of running pieces first before we go evaluate
		// their work to avoid a race.
		runningPieces := s.runningPieces.Load()

		// see how many are ready
		ready = ready[:0]
		for idx := range s.pieces {
			watermark := s.bundy.PieceSharesReceived(idx)
			if watermark >= requiredWatermark {
				ready = append(ready, idx)
				if watermark < stripesFound {
					// keep stripesFound at the smallest watermark
					stripesFound = watermark
				}
			}
		}
		if debugEnabled {
			fmt.Println("core found", len(ready), "ready")
		}

		// how many were ready? if we cleared the current neededShares, we can break
		// out of our condition variable for loop
		if int32(len(ready)) >= s.bundy.NeededShares() {
			if debugEnabled {
				fmt.Println("core bundy says that's enough. hooray")
			}
			// hooray!
			break
		}

		// not enough ready.
		// okay, were there enough running share readers at the start still so that
		// we could potentially still have enough ready in the future?
		if runningPieces+int32(len(ready)) < s.bundy.NeededShares() {
			// nope. we need to give up.
			backcompatMon.Meter("download_stripe_failed_not_enough_pieces_uplink").Mark(1) //mon:locked
			return nil, 0, s.combineErrs()
		}

		if debugEnabled {
			fmt.Println("core", len(ready), "ready not enough for", s.bundy.NeededShares(), ", sleeping")
		}

		// let's wait for the bundy clock to tell a share reader to wake us up.
		if !s.stripeReady.Wait(ctx) {
			return nil, 0, ctx.Err()
		}
	}

	// okay, we have a enough share readers ready.

	// some pre-allocated working memory for erasure share calls.
	fecShares := make([]infectious.Share, 0, len(ready))

	// we're going to loop through the stripesFound - s.returnedStripes new
	// stripes we have available.
	for stripe := int(s.returnedStripes); stripe < int(stripesFound); stripe++ {
		stripeOffset := (stripe - int(s.returnedStripes)) * s.scheme.StripeSize()
		if debugEnabled {
			fmt.Println("core piecing together stripe", stripe, "and writing at offset", stripeOffset)
		}

		outslice := out[stripeOffset : stripeOffset+s.scheme.StripeSize()]

		fecShares = fecShares[:0]
		var releases []func()

		for _, idx := range ready {
			data, release, err := s.pieces[idx].buffer.ReadShare(stripe)
			if err != nil {
				return nil, 0, Error.New("unexpected error: %w", err)
			}
			releases = append(releases, release)
			fecShares = append(fecShares, infectious.Share{
				Number: s.pieces[idx].shareNum,
				Data:   data})
		}

		if s.errorDetection {
			_, err = s.scheme.Decode(outslice, fecShares)
		} else {
			err = s.scheme.Rebuild(fecShares, func(r infectious.Share) {
				copy(outslice[r.Number*len(r.Data):(r.Number+1)*len(r.Data)], r.Data)
			})
		}

		for _, release := range releases {
			release()
		}

		if err != nil {
			if needsMoreShares(err) {
				if s.bundy.IncreaseNeededShares() {
					// just start over now
					return s.ReadStripes(ctx, nextStripe, out)
				}
			}
			return nil, 0, Error.New("error decoding data: %w", err)
		}
	}

	// okay, we're about to say we got a bunch of shares, so let's tell all the
	// share readers to raise their watermark of what's done.
	for idx := range s.pieces {
		s.pieces[idx].markCompleted(int(stripesFound))
	}

	stripes := stripesFound - s.returnedStripes
	s.returnedStripes = stripesFound

	if debugEnabled {
		fmt.Println("core returned", int(stripes)*s.scheme.StripeSize(), "bytes and", stripes, "stripes")
	}

	return out[:int(stripes)*s.scheme.StripeSize()], int(stripes), nil
}

func needsMoreShares(err error) bool {
	return errors.Is(err, infectious.NotEnoughShares) ||
		errors.Is(err, infectious.TooManyErrors)
}
