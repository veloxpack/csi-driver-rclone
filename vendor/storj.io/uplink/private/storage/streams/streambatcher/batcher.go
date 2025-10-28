// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package streambatcher

import (
	"context"
	"sync"
	"time"

	"github.com/zeebo/errs"

	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/uplink/private/metaclient"
)

// Info returns stream information gathered by the Batcher.
type Info struct {
	// CreationDate is the creation date of the stream extracted by the
	// stream ID that is provided to the Batcher or gathered from the
	// BeginObject response.
	CreationDate time.Time

	// PlainSize is the plain-text size of the stream aggregated from all
	// MakeInlineSegment or CommitSegment batch items.
	PlainSize int64

	// Version is object version retrieved from CommitObject batch item.
	Version []byte
}

// Batcher issues batch items related to a single stream. It aggregates
// information about the stream required by callers to commit the stream. It
// also learns the stream ID (unless already provided for part uploads) and
// automatically injects it into batch items that need it.
type Batcher struct {
	miBatcher metaclient.Batcher

	mu       sync.Mutex
	streamID storj.StreamID
	info     Info
}

// New returns a new Batcher that issues batch items for a stream. The streamID
// can be nil (in the case of an object upload) or not (in the case of a part
// upload). The batcher will discover the streamID in the former case when it
// processes a BeginObject.
func New(miBatcher metaclient.Batcher, streamID storj.StreamID) *Batcher {
	return &Batcher{
		miBatcher: miBatcher,
		streamID:  streamID,
	}
}

// Batch issues batch items for a stream. Once the streamID is known, it will
// be injected into batch items that need it. If a BeginObject is issued, the
// stream ID will be gleaned from it. If a BeginObject needs to be issued, it
// must be the first batch item issued by the batcher.
func (s *Batcher) Batch(ctx context.Context, batchItems ...metaclient.BatchItem) ([]metaclient.BatchResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range batchItems {
		switch item := item.(type) {
		case *metaclient.BeginSegmentParams:
			item.StreamID = s.streamID
		case *metaclient.MakeInlineSegmentParams:
			item.StreamID = s.streamID
			s.info.PlainSize += item.PlainSize
		case *metaclient.CommitSegmentParams:
			s.info.PlainSize += item.PlainSize
		case *metaclient.CommitObjectParams:
			item.StreamID = s.streamID
		}
	}

	if len(batchItems) == 0 {
		return nil, errs.New("programmer error: empty batch request")
	}

	resp, err := s.miBatcher.Batch(ctx, batchItems...)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errs.New("programmer error: empty batch response")
	}

	if s.streamID == nil {
		beginObject, err := resp[0].BeginObject()
		if err != nil {
			return nil, errs.New("programmer error: first batch must start with BeginObject: %w", err)
		}
		if beginObject.StreamID.IsZero() {
			return nil, errs.New("stream ID missing from BeginObject response")
		}
		s.streamID = beginObject.StreamID
	}

	for _, response := range resp {
		if response.IsCommitObject() {
			commitObject, err := response.CommitObject()
			if err != nil {
				return nil, errs.New("programmer error: batch must be CommitObject: %w", err)
			}

			s.info = Info{
				CreationDate: commitObject.Object.Created,
				PlainSize:    commitObject.Object.PlainSize,
			}
			if commitObject.Object.Status == int32(pb.Object_COMMITTED_VERSIONED) {
				s.info.Version = commitObject.Object.Version
			}
		}
	}

	return resp, nil
}

// StreamID returns the stream ID either provided to the Batcher or gleaned
// from issuing a BeginObject request.
func (s *Batcher) StreamID() storj.StreamID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streamID
}

// Info returns the stream information gathered by the batch items.
func (s *Batcher) Info() (Info, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.streamID == nil {
		return Info{}, errs.New("stream ID is unexpectedly nil")
	}

	return s.info, nil
}
