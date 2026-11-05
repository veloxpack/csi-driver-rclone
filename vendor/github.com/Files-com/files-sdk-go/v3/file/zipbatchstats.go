package file

// ZIP batch stats: counters, immutable snapshots, and shared log helpers.
//
// Cross-cutting ZIP batch invariants:
//   - zipBatchDownloader fields are walk-goroutine-only unless the field is explicitly atomic or mutex-protected.
//   - Each DownloadStatus is signaled exactly once; a duplicate signal can panic and a missing signal can hang the job.
//   - Every file admission Wait/WaitWithContext has exactly one Done, including retry, fallback, and cancel paths.
//   - Batch-routed files prepare at dispatch; per-file downloads prepare at download.

import (
	"sync/atomic"
	"time"
)

type ZipBatchStatsSnapshot struct {
	// BatchesDispatched is the number of ZIP batches admitted for download.
	BatchesDispatched int64
	// BatchFiles is the number of files originally assigned to dispatched ZIP batches.
	BatchFiles int64
	// BatchesDissolved is the number of below-gate or below-MinFiles remainders sent to per-file downloads.
	BatchesDissolved int64
	// DissolvedFiles is the number of files in dissolved batches.
	DissolvedFiles int64
	// CreateRequests is the number of ZIP create POST requests, including 422 shrink retries.
	CreateRequests int64
	// StreamAttempts is the number of ZIP stream download attempts.
	StreamAttempts int64
	// StreamFailures is the number of corrupt or failed ZIP stream attempts.
	StreamFailures int64
	// CleanFinalized is the number of files finalized from clean ZIP extraction.
	CleanFinalized int64
	// SalvageFinalized is the number of files finalized from corrupt-spool salvage.
	SalvageFinalized int64
	// FallbackCreateError is the number of files falling back after ZIP create failures.
	FallbackCreateError int64
	// FallbackTripwire is the number of files falling back after correctness tripwires.
	FallbackTripwire int64
	// FallbackRetriesExhausted is the number of files falling back after ZIP retries are exhausted.
	FallbackRetriesExhausted int64
	// FallbackMissingEntry is the number of files falling back because the ZIP omitted expected entries.
	FallbackMissingEntry int64
	// ProbeZipFiles is the number of probe files routed through ZIP batches.
	ProbeZipFiles int64
	// ProbePerFileFiles is the number of probe files routed through per-file downloads.
	ProbePerFileFiles int64
	// ProbeZipRateMilli is measured ZIP probe throughput in files/sec multiplied by 1000.
	ProbeZipRateMilli int64
	// ProbePerFileRateMilli is measured per-file probe throughput in files/sec multiplied by 1000.
	ProbePerFileRateMilli int64
	// ProbeBaselineWindows is the number of per-file windows used for the latest probe baseline.
	ProbeBaselineWindows int64
	// ProbeDecision is none, committed, or dissolved.
	ProbeDecision string
	// CircuitBreakerTrips is the number of times ZIP batching was disabled after repeated retry-exhausted batches.
	CircuitBreakerTrips int64
	// Reprobes is the number of scheduled ZIP re-probe attempts after a dissolve verdict.
	Reprobes int64
}

// Active reports whether the job used or attempted ZIP batching.
func (s ZipBatchStatsSnapshot) Active() bool {
	return s.BatchesDispatched != 0 ||
		s.BatchesDissolved != 0 ||
		s.CreateRequests != 0 ||
		s.StreamAttempts != 0 ||
		s.CleanFinalized != 0 ||
		s.SalvageFinalized != 0 ||
		s.FallbackFiles() != 0 ||
		s.ProbeZipFiles != 0 ||
		s.ProbePerFileFiles != 0 ||
		s.ProbeBaselineWindows != 0 ||
		s.CircuitBreakerTrips != 0 ||
		s.Reprobes != 0
}

// FallbackFiles returns the total number of files routed from ZIP batches to per-file downloads.
func (s ZipBatchStatsSnapshot) FallbackFiles() int64 {
	return s.FallbackCreateError + s.FallbackTripwire + s.FallbackRetriesExhausted + s.FallbackMissingEntry
}

// StreamRetries returns ZIP stream attempts after the first attempt for each dispatched batch.
func (s ZipBatchStatsSnapshot) StreamRetries() int64 {
	retries := s.StreamAttempts - s.BatchesDispatched
	if retries < 0 {
		return 0
	}
	return retries
}

type zipBatchStats struct {
	batchesDispatched        atomic.Int64
	batchFiles               atomic.Int64
	batchesDissolved         atomic.Int64
	dissolvedFiles           atomic.Int64
	createRequests           atomic.Int64
	streamAttempts           atomic.Int64
	streamFailures           atomic.Int64
	cleanFinalized           atomic.Int64
	salvageFinalized         atomic.Int64
	fallbackCreateError      atomic.Int64
	fallbackTripwire         atomic.Int64
	fallbackRetriesExhausted atomic.Int64
	fallbackMissingEntry     atomic.Int64
	probeZipFiles            atomic.Int64
	probePerFileFiles        atomic.Int64
	probeZipRateMilli        atomic.Int64
	probePerFileRateMilli    atomic.Int64
	probeBaselineWindows     atomic.Int64
	probeDecision            atomic.Value
	circuitBreakerTrips      atomic.Int64
	reprobes                 atomic.Int64
}

func (s *zipBatchStats) snapshot() ZipBatchStatsSnapshot {
	if s == nil {
		return ZipBatchStatsSnapshot{}
	}
	decision, _ := s.probeDecision.Load().(string)
	if decision == "" {
		decision = string(zipBatchProbeDecisionNone)
	}
	return ZipBatchStatsSnapshot{
		BatchesDispatched:        s.batchesDispatched.Load(),
		BatchFiles:               s.batchFiles.Load(),
		BatchesDissolved:         s.batchesDissolved.Load(),
		DissolvedFiles:           s.dissolvedFiles.Load(),
		CreateRequests:           s.createRequests.Load(),
		StreamAttempts:           s.streamAttempts.Load(),
		StreamFailures:           s.streamFailures.Load(),
		CleanFinalized:           s.cleanFinalized.Load(),
		SalvageFinalized:         s.salvageFinalized.Load(),
		FallbackCreateError:      s.fallbackCreateError.Load(),
		FallbackTripwire:         s.fallbackTripwire.Load(),
		FallbackRetriesExhausted: s.fallbackRetriesExhausted.Load(),
		FallbackMissingEntry:     s.fallbackMissingEntry.Load(),
		ProbeZipFiles:            s.probeZipFiles.Load(),
		ProbePerFileFiles:        s.probePerFileFiles.Load(),
		ProbeZipRateMilli:        s.probeZipRateMilli.Load(),
		ProbePerFileRateMilli:    s.probePerFileRateMilli.Load(),
		ProbeBaselineWindows:     s.probeBaselineWindows.Load(),
		ProbeDecision:            decision,
		CircuitBreakerTrips:      s.circuitBreakerTrips.Load(),
		Reprobes:                 s.reprobes.Load(),
	}
}

// ZipBatchStats returns ZIP batch activity counters for the job.
func (j *Job) ZipBatchStats() ZipBatchStatsSnapshot {
	if j == nil {
		return ZipBatchStatsSnapshot{}
	}
	return j.zipBatchStats.snapshot()
}

// ScanDuration returns the elapsed time between the job's scan start and scan end signals.
func (j *Job) ScanDuration() time.Duration {
	if j == nil || j.scanStartedAt.IsZero() || j.scanEndedAt.IsZero() {
		return 0
	}
	return j.scanEndedAt.Sub(j.scanStartedAt)
}

func (b *zipBatchDownloader) log(event string, fields map[string]interface{}) {
	if b == nil || b.job == nil {
		return
	}
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["event"] = event
	b.job.Config.LogPath("zip-batch", fields)
}

func (b *zipBatchDownloader) stats() *zipBatchStats {
	if b == nil || b.job == nil {
		return nil
	}
	return b.job.zipBatchStats
}

func zipBatchStatusesSize(statuses []*DownloadStatus) int64 {
	var total int64
	for _, downloadStatus := range statuses {
		total += downloadStatus.Size()
	}
	return total
}

func (b *zipBatchDownloader) recordFallback(reason zipBatchFallbackReason, count int) {
	stats := b.stats()
	if stats == nil || count == 0 {
		return
	}
	switch reason {
	case zipBatchFallbackCreateError:
		stats.fallbackCreateError.Add(int64(count))
	case zipBatchFallbackTripwire:
		stats.fallbackTripwire.Add(int64(count))
	case zipBatchFallbackRetriesExhausted:
		stats.fallbackRetriesExhausted.Add(int64(count))
	case zipBatchFallbackMissingEntry:
		stats.fallbackMissingEntry.Add(int64(count))
	}
}
