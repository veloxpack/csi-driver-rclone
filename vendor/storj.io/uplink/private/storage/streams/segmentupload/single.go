// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package segmentupload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"

	"storj.io/common/encryption"
	"storj.io/common/pb"
	"storj.io/eventkit"
	"storj.io/uplink/private/eestream"
	"storj.io/uplink/private/eestream/scheduler"
	"storj.io/uplink/private/metaclient"
	"storj.io/uplink/private/stalldetection"
	"storj.io/uplink/private/storage/streams/pieceupload"
	"storj.io/uplink/private/storage/streams/segmentupload/cohorts"
	"storj.io/uplink/private/storage/streams/splitter"
	"storj.io/uplink/private/testuplink"
)

var (
	mon             = monkit.Package()
	uploadTask      = mon.TaskNamed("segment-upload")
	evs             = eventkit.Package()
	uploadIDCounter atomic.Int64
)

// generateUploadID creates a unique identifier for tracking upload performance.
func generateUploadID() string {
	id := uploadIDCounter.Add(1)
	return fmt.Sprintf("upload-%d", id)
}

// Scheduler is used to coordinate and constrain resources between
// concurrent segment uploads.
type Scheduler interface {
	Join(ctx context.Context) (scheduler.Handle, bool)
}

// Begin starts a segment upload identified by the segment ID provided in the
// beginSegment response. The returned upload will complete when enough piece
// uploads to fulfill the optimal threshold for the segment redundancy strategy
// plus a small long tail margin. It cancels remaining piece uploads once that
// threshold has been hit.
func Begin(ctx context.Context,
	beginSegment *metaclient.BeginSegmentResponse,
	segment splitter.Segment,
	limitsExchanger pieceupload.LimitsExchanger,
	piecePutter pieceupload.PiecePutter,
	scheduler Scheduler,
	stallDetectionConfig *stalldetection.Config,
) (_ *Upload, err error) {
	defer mon.Task()(&ctx)(&err)

	ctx = testuplink.WithLogWriterContext(ctx, "seg_pos", fmt.Sprint(segment.Position()))
	testuplink.Log(ctx, "Begin upload segment...")
	defer testuplink.Log(ctx, "Done begin upload segment.")

	taskDone := uploadTask(&ctx)
	defer func() {
		if err != nil {
			taskDone(&err)
		}
	}()

	// Join the scheduler so the concurrency can be limited appropriately.
	handle, ok := scheduler.Join(ctx)
	if !ok {
		return nil, errs.New("failed to obtain piece upload handle")
	}
	defer func() {
		if err != nil {
			handle.Done()
		}
	}()

	if beginSegment.RedundancyStrategy.ErasureScheme == nil {
		return nil, errs.New("begin segment response is missing redundancy strategy")
	}
	if beginSegment.PiecePrivateKey.IsZero() {
		return nil, errs.New("begin segment response is missing piece private key")
	}

	optimalThreshold := beginSegment.RedundancyStrategy.OptimalThreshold()
	if optimalThreshold > len(beginSegment.Limits) {
		return nil, errs.New("begin segment response needs at least %d limits to meet optimal threshold but has %d", optimalThreshold, len(beginSegment.Limits))
	}

	uploaderCount := len(beginSegment.Limits)

	mgr := pieceupload.NewManager(
		limitsExchanger,
		&pieceReader{segment, beginSegment.RedundancyStrategy},
		beginSegment.SegmentID,
		beginSegment.Limits,
	)

	wg := new(sync.WaitGroup)
	defer func() {
		if err != nil {
			wg.Wait()
		}
	}()

	// Create a context that we can use to cancel piece uploads when we have enough.
	longTailCtx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	// Set up StallManager.
	var stallManager *pieceupload.StallManager
	var updatedDetectionConfig *stalldetection.Config
	if stallDetectionConfig != nil {
		updatedDetectionConfig = stallDetectionConfig.ValidateAndUpdate(beginSegment.RedundancyStrategy.TotalCount())
		stallManager = pieceupload.NewStallManager()
	}

	// Set up a cohort matcher to track when we've satisfied the cohort requirements.
	var matcher *cohorts.Matcher
	if beginSegment.CohortRequirements == nil {
		matcher = cohorts.NewMatcher(&pb.CohortRequirements{
			Requirement: &pb.CohortRequirements_Literal_{
				Literal: &pb.CohortRequirements_Literal{
					Value: int32(optimalThreshold),
				},
			},
		}, len(beginSegment.Limits))
	} else {
		matcher = cohorts.NewMatcher(beginSegment.CohortRequirements, len(beginSegment.Limits))
	}

	results := make(chan segmentResult, uploaderCount)
	var successful atomic.Int32
	uploadStart := time.Now()
	uploadID := generateUploadID()
	for range uploaderCount {
		res, ok := handle.Get(ctx)
		if !ok {
			return nil, errs.New("failed to obtain piece upload resource")
		}

		wg.Add(1)
		go func() {
			pieceStart := time.Now()
			defer wg.Done()
			// Whether the upload is ultimately successful or not, when this
			// function returns, the scheduler resource MUST be released to
			// allow other piece uploads to take place.
			defer res.Done()
			uploaded, tags, node, stallCount, err := pieceupload.UploadOne(longTailCtx, ctx, mgr, piecePutter, beginSegment.PiecePrivateKey, stallManager)

			// Emit individual stall events
			if stallCount > 0 {
				evs.Event("piece_stalls_during_upload",
					eventkit.String("upload_id", uploadID),
					eventkit.Int64("stall_count", int64(stallCount)),
					eventkit.Duration("duration", time.Since(pieceStart)))
			}

			if err != nil {
				var optimalError pieceupload.OptimalThresholdError
				if errors.As(err, &optimalError) {
					// Convert optimal threshold errors to nil for backward compatibility
					err = nil
				}
				// Keep other errors as-is
			}

			results <- segmentResult{uploaded: uploaded, err: err, stallCount: stallCount}
			if uploaded {
				// Piece upload was successful.
				successfulSoFar := int(successful.Add(1))
				// After BaseUploads successful uploads, calculate the stall threshold.
				if updatedDetectionConfig != nil && successfulSoFar == updatedDetectionConfig.BaseUploads {
					baseUploadsDuration := time.Since(uploadStart)
					detectionFactor := updatedDetectionConfig.Factor
					minStallDuration := updatedDetectionConfig.MinStallDuration
					maxDuration := max(baseUploadsDuration*time.Duration(detectionFactor), minStallDuration)
					stallManager.SetMaxDuration(maxDuration)

					evs.Event("stall_threshold_calculation",
						eventkit.String("uploadID", uploadID),
						eventkit.Duration("baseUploadsDuration", baseUploadsDuration),
						eventkit.Int64("detectionFactor", int64(detectionFactor)),
						eventkit.Duration("minStallDuration", minStallDuration),
						eventkit.Duration("maxDuration", maxDuration))
				}

				// if the matcher says we've satisfied the cohort requirements, then cancel the
				// remaining uploads.
				if matcher.Increment(tags, node) {
					testuplink.Log(ctx, "Satisfied cohort requirements after", successfulSoFar, "pieces")
					cancel()
				}
			}
		}()
	}

	return &Upload{
		ctx:                  ctx,
		taskDone:             taskDone,
		optimalThreshold:     beginSegment.RedundancyStrategy.OptimalThreshold(),
		handle:               handle,
		results:              results,
		cancel:               cancel,
		wg:                   wg,
		mgr:                  mgr,
		segment:              segment,
		uploadStart:          uploadStart,
		uploadID:             uploadID,
		stallDetectionConfig: stallDetectionConfig,
	}, nil
}

type pieceReader struct {
	segment    splitter.Segment
	redundancy eestream.RedundancyStrategy
}

func (r *pieceReader) PieceReader(num int) io.Reader {
	segment := r.segment.Reader()
	stripeSize := r.redundancy.StripeSize()
	paddedData := encryption.PadReader(io.NopCloser(segment), stripeSize)
	return NewEncodedReader(paddedData, r.redundancy, num)
}

type segmentResult struct {
	uploaded   bool
	err        error
	stallCount int
}

// Upload is a segment upload that has been started and returned by the Begin
// method.
type Upload struct {
	ctx              context.Context
	taskDone         func(*error)
	optimalThreshold int
	handle           scheduler.Handle
	results          chan segmentResult
	cancel           context.CancelFunc
	wg               *sync.WaitGroup
	mgr              *pieceupload.Manager
	segment          splitter.Segment

	// Performance tracking fields
	uploadStart          time.Time
	uploadID             string
	stallDetectionConfig *stalldetection.Config

	// Test-only fields for verification
	stallsDetected int
}

// Wait blocks until the segment upload completes. It will be successful as
// long as enough pieces have uploaded successfully.
func (upload *Upload) Wait() (_ *metaclient.CommitSegmentParams, err error) {
	defer upload.taskDone(&err)
	defer upload.handle.Done()
	defer upload.cancel()

	var eg errs.Group
	var successful int
	var totalStalls int
	for i := 0; i < cap(upload.results); i++ {
		result := <-upload.results
		if result.uploaded {
			successful++
		}
		totalStalls += result.stallCount
		eg.Add(result.err)
	}

	// Store stall count for test verification
	upload.stallsDetected = totalStalls

	// The goroutines should all be on their way to exiting since the loop
	// above guarantees they have written their results to the channel. Wait
	// for them to all finish and release the scheduler resource. This is
	// really only necessary for deterministic testing.
	upload.wg.Wait()

	if successful < upload.optimalThreshold {
		err = errs.Combine(errs.New("failed to upload enough pieces (needed at least %d but got %d)", upload.optimalThreshold, successful), eg.Err())
	}
	upload.segment.DoneReading(err)

	// Emit comprehensive upload summary metrics
	totalDuration := time.Since(upload.uploadStart)
	uploaderCount := cap(upload.results)
	stallsDetectedCount := int64(totalStalls)

	stallRate := float64(0)
	if uploaderCount > 0 {
		stallRate = float64(stallsDetectedCount) / float64(uploaderCount)
	}
	emitUploadSummary(upload.uploadID, totalDuration, successful, upload.optimalThreshold, uploaderCount, stallsDetectedCount, stallRate, err == nil, upload.stallDetectionConfig)

	testuplink.Log(upload.ctx, "Done waiting for segment.",
		"successful:", successful,
		"optimal:", upload.optimalThreshold,
		"errs:", eg.Err(),
	)

	if err != nil {
		return nil, err
	}

	info := upload.segment.Finalize()
	segmentID, uploadResults := upload.mgr.Results()

	return &metaclient.CommitSegmentParams{
		SegmentID:         segmentID,
		Encryption:        info.Encryption,
		SizeEncryptedData: info.EncryptedSize,
		PlainSize:         info.PlainSize,
		EncryptedETag:     nil, // encrypted eTag is injected by a different layer
		UploadResult:      uploadResults,
	}, nil
}

// emitUploadSummary emits a comprehensive upload summary event
func emitUploadSummary(uploadID string, totalDuration time.Duration, successful, optimalThreshold, uploaderCount int, stallsDetected int64, stallRate float64, success bool, stallConfig *stalldetection.Config) {
	var stall_enabled, dynamic_base_uploads bool
	var base_uploads, factor int64
	var min_stall_duration time.Duration

	if stallConfig == nil {
		stall_enabled = false
		base_uploads = 0
		factor = 0
		min_stall_duration = 0
		dynamic_base_uploads = false
	} else {
		stall_enabled = true
		base_uploads = int64(stallConfig.BaseUploads)
		factor = int64(stallConfig.Factor)
		min_stall_duration = stallConfig.MinStallDuration
		dynamic_base_uploads = stallConfig.DynamicBaseUploads
	}
	evs.Event("segment_upload_summary",
		eventkit.String("upload_id", uploadID),
		eventkit.Duration("total_duration", totalDuration),
		eventkit.Int64("successful_pieces", int64(successful)),
		eventkit.Int64("optimal_threshold", int64(optimalThreshold)),
		eventkit.Int64("uploader_count", int64(uploaderCount)),
		eventkit.Bool("success", success),
		eventkit.Bool("stall_enabled", stall_enabled),
		eventkit.Int64("stalls_detected", stallsDetected),
		eventkit.Float64("stall_rate", stallRate),
		eventkit.Int64("stall_base_uploads", base_uploads),
		eventkit.Int64("stall_factor", factor),
		eventkit.Duration("stall_min_duration", min_stall_duration),
		eventkit.Bool("stall_dynamic_base_uploads", dynamic_base_uploads),
	)
}
