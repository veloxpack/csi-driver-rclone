package file

// ZIP batch orchestration: public params, the eligibility gate, dynamic sizing, and batch dispatch.
//
// Cross-cutting ZIP batch invariants:
//   - zipBatchDownloader fields are walk-goroutine-only unless the field is explicitly atomic or mutex-protected.
//   - Each DownloadStatus is signaled exactly once; a duplicate signal can panic and a missing signal can hang the job.
//   - Every file admission Wait/WaitWithContext has exactly one Done, including retry, fallback, and cancel paths.
//   - Batch-routed files prepare at dispatch; per-file downloads prepare at download.

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Files-com/files-sdk-go/v3/file/status"
)

const (
	// GCP benchmark 2026-07: 366 verified runs, see issue #3511.
	// EligibleSize: 128 KiB. Wide eligibility is safe because routing is probe-guarded:
	// the gate, the sampled median baseline, the sequential zip cohort trial, route-to-winner,
	// re-probe, and the circuit breaker together guarantee per-job that zip is only used where
	// it measures faster (validated across ~900 benchmark runs on datacenter and residential
	// networks). The interim 16 KiB LCD existed only while that machinery was unproven.
	// MinFiles: 500 avoids engaging ZIP batching for small jobs that lose to per-file adaptive downloads.
	// DynamicBatchFloor: 16 keeps early batches small enough to fill all streams fast.
	// MaxFiles: 32 is the dynamic growth ceiling; at low RTT, 20k-file jobs measured fastest
	// with a 32 ceiling (732 f/s vs 634 at 256 and 669 pinned at 16) — small batches win on
	// tail granularity once per-batch setup is amortized. High-RTT may favor larger; revisit
	// with the far-region benchmark.
	// MaxBytes: 25 MiB was never binding in the benchmark and remains a conservative cap.
	// ConcurrentBatches: 64 streams reached the scaling knee before 128 streams.
	defaultZipBatchEligibleSize      = 128 * 1024
	defaultZipBatchMinFiles          = 500
	defaultZipBatchDynamicBatchFloor = 16
	defaultZipBatchMaxFiles          = 32
	defaultZipBatchMaxBytes          = 25 * 1024 * 1024
	defaultZipBatchConcurrentBatches = 64
	// MinAdvantage 1.05 is a route-to-winner bar with a noise guard: per job, use zip
	// whenever it measures faster. The >=1.25x product requirement is a feature-level
	// criterion, proven by forced benchmarks (1.3-2x in winning environments); using it
	// as a per-job routing bar measurably forfeited 1.5-1.8x wins to probe conservatism.
	defaultZipBatchMinAdvantage = 1.05
)

// ZipBatchExtractionMode selects how ZIP batch archives are extracted.
type ZipBatchExtractionMode string

const (
	// ZipBatchExtractionSpool downloads the ZIP archive to a temporary spool file before extraction.
	ZipBatchExtractionSpool ZipBatchExtractionMode = "spool"
	// ZipBatchExtractionStream extracts the ZIP archive directly from the HTTP response stream.
	ZipBatchExtractionStream ZipBatchExtractionMode = "stream"
)

// ZipBatchParams controls batched small-file downloads through the ZIP download endpoint.
type ZipBatchParams struct {
	// Disabled turns ZIP batching off entirely.
	Disabled bool
	// Extraction selects spool or stream extraction. Empty uses the SDK default.
	Extraction ZipBatchExtractionMode
	// EligibleSize is the exclusive per-file size ceiling for ZIP batching. Zero uses the SDK default.
	EligibleSize int64
	// MinFiles is the number of eligible small files required before ZIP batching engages. Zero uses the SDK default.
	MinFiles int
	// MaxFiles is the maximum dynamic or pinned file count in one ZIP batch. Zero uses the SDK default.
	MaxFiles int
	// BatchSize pins the number of files per ZIP batch. Zero uses dynamic growth.
	BatchSize int
	// MaxBytes is the maximum listed file bytes in one ZIP batch. Zero uses the SDK default.
	MaxBytes int64
	// ConcurrentBatches is the maximum number of ZIP streams in flight. Zero uses the SDK default.
	ConcurrentBatches int
	// MinAdvantage is the required ZIP/per-file speedup before committing to ZIP batching. Zero uses the SDK default; negative disables probing and batches unconditionally after MinFiles.
	MinAdvantage float64
	// ReprobeInterval is the delay before re-testing ZIP after a dissolve verdict. Zero uses the SDK default; negative disables re-probing.
	ReprobeInterval time.Duration
}

type zipBatchOfferState int

const (
	zipBatchNotEligible zipBatchOfferState = iota
	zipBatchHandled
	zipBatchBatchable
)

type zipBatchMode int

const (
	zipBatchModeLocked zipBatchMode = iota
	zipBatchModeProbePhaseA
	zipBatchModeProbePhaseB
	zipBatchModeCommitted
	zipBatchModeDissolved
)

// zipBatchDownloader owns the walk-goroutine batch gate and dispatch queue.
// It moves from locked to probing or committed after MinFiles opens the gate.
type zipBatchDownloader struct {
	job          *Job
	ctx          context.Context
	params       ZipBatchParams
	signal       chan *DownloadStatus
	pending      []*DownloadStatus
	bytes        int64
	eligibleSeen int
	unlocked     bool
	mode         zipBatchMode
	probe        *zipBatchProbe
	circuitTrips atomic.Int64
	batchSlots   chan struct{}
	walkEnded    bool
	reprobeAfter time.Time
	reprobeWait  time.Duration
}

func (p ZipBatchParams) withDefaults() ZipBatchParams {
	if p.Extraction == "" {
		// Stream extraction benchmarked equal-or-faster than spooling at
		// identical batch settings, with immediate abort on corrupt streams
		// and no spool tmp files; spool remains selectable for diagnostics.
		p.Extraction = ZipBatchExtractionStream
	}
	if p.EligibleSize <= 0 {
		p.EligibleSize = defaultZipBatchEligibleSize
	}
	if p.MinFiles <= 0 {
		p.MinFiles = defaultZipBatchMinFiles
	}
	if p.MaxFiles <= 0 {
		p.MaxFiles = defaultZipBatchMaxFiles
	}
	if p.BatchSize < 0 {
		p.BatchSize = 0
	}
	if p.MaxBytes <= 0 {
		p.MaxBytes = defaultZipBatchMaxBytes
	}
	if p.ConcurrentBatches <= 0 {
		p.ConcurrentBatches = defaultZipBatchConcurrentBatches
	}
	if p.MinAdvantage == 0 {
		p.MinAdvantage = defaultZipBatchMinAdvantage
	}
	if p.ReprobeInterval == 0 {
		p.ReprobeInterval = zipBatchTuning.DefaultReprobeInterval
	}
	return p
}

func zipBatchDisabled(params DownloaderParams, job *Job) bool {
	if params.ZipBatch.Disabled || params.DryRun {
		return true
	}
	if job == nil || job.Manager == nil {
		return false
	}
	return job.Manager.FilePartsManager.DownloadFilesAsSingleStream
}

func newZipBatchDownloader(job *Job, ctx context.Context, params DownloaderParams, signal chan *DownloadStatus) *zipBatchDownloader {
	zipParams := params.ZipBatch.withDefaults()
	zipParams.ConcurrentBatches = zipBatchEffectiveConcurrentBatches(zipParams.ConcurrentBatches, zipBatchFileSlotCapacity(job))
	batcher := &zipBatchDownloader{
		job:         job,
		ctx:         ctx,
		params:      zipParams,
		signal:      signal,
		mode:        zipBatchModeLocked,
		probe:       newZipBatchProbe(),
		batchSlots:  make(chan struct{}, zipParams.ConcurrentBatches),
		reprobeWait: zipParams.ReprobeInterval,
	}
	batcher.registerProbeReporter()
	return batcher
}

func zipBatchFileSlotCapacity(job *Job) int {
	if job == nil {
		return 1
	}
	if job.Manager == nil {
		job.SetManager(nil)
	}
	return job.Manager.FilesManager.Max()
}

func zipBatchEffectiveConcurrentBatches(requested, fileSlots int) int {
	if fileSlots < 1 {
		fileSlots = 1
	}
	capacity := fileSlots * 3 / 4
	if capacity < 1 {
		capacity = 1
	}
	if requested < 1 || requested > capacity {
		return capacity
	}
	return requested
}

func (b *zipBatchDownloader) offer(downloadStatus *DownloadStatus) bool {
	switch b.classify(downloadStatus) {
	case zipBatchNotEligible:
		return false
	case zipBatchHandled:
		return true
	}
	b.eligibleSeen++
	if b.mode == zipBatchModeLocked {
		b.appendPending(downloadStatus)
		if b.eligibleSeen >= b.params.MinFiles {
			b.unlock()
			b.routeLockedPending(false)
		}
		return true
	}
	b.routeEligible(downloadStatus)
	return true
}

func (b *zipBatchDownloader) unlock() {
	b.unlocked = true
	if b.params.MinAdvantage < 0 {
		b.transition(zipBatchModeCommitted, "force")
		return
	}
	b.transition(zipBatchModeProbePhaseA, "min-files")
	b.probe.mu.Lock()
	b.probe.attempt = 1
	b.probe.phase = zipBatchProbePhaseA
	b.probe.perFileRateWindows = nil
	b.probe.mu.Unlock()
	if stats := b.stats(); stats != nil {
		stats.probeBaselineWindows.Store(0)
	}
	b.log("probe-start", map[string]interface{}{
		"attempt":                  1,
		"phase":                    "per-file",
		"baseline_windows":         0,
		"per_file_min_completions": zipBatchTuning.ProbePhaseAMinCompletions,
		"per_file_min_seconds":     zipBatchTuning.ProbePhaseAMinDuration.Seconds(),
		"per_file_hard_seconds":    zipBatchTuning.ProbePhaseAHardCap.Seconds(),
		"per_file_window_seconds":  zipBatchTuning.ProbePhaseAWindow.Seconds(),
		"per_file_window2_after":   zipBatchTuning.ProbePhaseAWindow2After,
		"per_file_window3_after":   zipBatchTuning.ProbePhaseAWindow3After,
		"zip_min_completions":      zipBatchTuning.ProbePhaseBMinCompletions,
		"zip_warmup_seconds":       zipBatchTuning.ProbePhaseBWarmup.Seconds(),
		"zip_window_seconds":       zipBatchTuning.ProbePhaseBWindow.Seconds(),
		"zip_hard_seconds":         zipBatchTuning.ProbePhaseBHardCap.Seconds(),
		"min_advantage":            b.params.MinAdvantage,
	})
}

func (b *zipBatchDownloader) routeLockedPending(endOfWalk bool) {
	statuses := b.pending
	b.pending = nil
	b.bytes = 0
	for _, downloadStatus := range statuses {
		b.routeEligible(downloadStatus)
	}
	if endOfWalk {
		b.flush(true)
	}
}

func (b *zipBatchDownloader) routeEligible(downloadStatus *DownloadStatus) {
	b.applyCircuitBreaker()
	b.updateProbe(false)
	b.applyProbeState(false)
	switch b.mode {
	case zipBatchModeDissolved:
		if b.startReprobeIfDue(time.Now()) {
			b.routeProbePhaseB(downloadStatus)
			return
		}
		b.markDissolvedPerFile(downloadStatus)
		enqueueIndexedDownloadDirect(b.job, b.ctx, downloadStatus, b.signal)
	case zipBatchModeProbePhaseA:
		b.routeProbePhaseA(downloadStatus)
	case zipBatchModeProbePhaseB:
		b.routeProbePhaseB(downloadStatus)
	default:
		b.appendPending(downloadStatus)
		b.flush(false)
	}
}

func (b *zipBatchDownloader) routeProbePhaseA(downloadStatus *DownloadStatus) {
	b.markProbePerFile(downloadStatus, true)
	enqueueIndexedDownloadDirect(b.job, b.ctx, downloadStatus, b.signal)
}

func (b *zipBatchDownloader) routeProbePhaseB(downloadStatus *DownloadStatus) {
	b.markProbeZip(downloadStatus)
	b.appendPending(downloadStatus)
	b.flush(false)
}

func (b *zipBatchDownloader) appendPending(downloadStatus *DownloadStatus) {
	b.pending = append(b.pending, downloadStatus)
	b.bytes += downloadStatus.Size()
}

func (m zipBatchMode) String() string {
	switch m {
	case zipBatchModeLocked:
		return "locked"
	case zipBatchModeProbePhaseA:
		return "probe-phase-a"
	case zipBatchModeProbePhaseB:
		return "probe-phase-b"
	case zipBatchModeCommitted:
		return "committed"
	case zipBatchModeDissolved:
		return "dissolved"
	default:
		return "unknown"
	}
}

func (b *zipBatchDownloader) transition(to zipBatchMode, reason string) {
	if b == nil || b.mode == to {
		return
	}
	from := b.mode
	event := "mode-transition"
	if !zipBatchTransitionAllowed(from, to, reason) {
		event = "illegal-mode-transition"
	}
	b.log(event, map[string]interface{}{
		"from":   from.String(),
		"to":     to.String(),
		"reason": reason,
	})
	b.mode = to
}

func zipBatchTransitionAllowed(from, to zipBatchMode, reason string) bool {
	if reason == "circuit-breaker" && to == zipBatchModeDissolved {
		return true
	}
	switch from {
	case zipBatchModeLocked:
		return to == zipBatchModeProbePhaseA || to == zipBatchModeCommitted
	case zipBatchModeProbePhaseA:
		return to == zipBatchModeProbePhaseB || to == zipBatchModeDissolved
	case zipBatchModeProbePhaseB:
		return to == zipBatchModeCommitted || to == zipBatchModeDissolved
	case zipBatchModeDissolved:
		return to == zipBatchModeProbePhaseB
	default:
		return false
	}
}

func (b *zipBatchDownloader) classify(downloadStatus *DownloadStatus) zipBatchOfferState {
	if downloadStatus.error != nil ||
		downloadStatus.fsFile == nil ||
		downloadStatus.FileInfo == nil ||
		downloadStatus.FileInfo.IsDir() ||
		downloadStatus.dryRun ||
		downloadStatus.Size() < 0 ||
		downloadStatus.Size() >= b.params.EligibleSize {
		return zipBatchNotEligible
	}
	if _, ok := b.job.CompletedPaths[downloadStatus.LocalPath()]; ok {
		return zipBatchNotEligible
	}
	if ignoreDownloadJob(b.job, downloadStatus) {
		return zipBatchNotEligible
	}
	if untrusted, ok := downloadStatus.FileInfo.(UntrustedSize); ok && untrusted.SizeTrust() == UntrustedSizeValue {
		return zipBatchNotEligible
	}
	return zipBatchBatchable
}

func (b *zipBatchDownloader) flushEnd() {
	b.walkEnded = true
	if b.mode == zipBatchModeProbePhaseA || b.mode == zipBatchModeProbePhaseB {
		b.updateProbe(true)
	}
	b.applyCircuitBreaker()
	b.applyProbeState(true)
	b.flush(true)
}

func (b *zipBatchDownloader) flush(endOfWalk bool) {
	if len(b.pending) == 0 {
		return
	}

	if b.mode == zipBatchModeLocked {
		if endOfWalk {
			statuses := b.pending
			b.pending = nil
			b.bytes = 0
			b.dissolve(statuses)
		}
		return
	}
	if b.mode == zipBatchModeDissolved {
		statuses := b.pending
		b.pending = nil
		b.bytes = 0
		b.dissolve(statuses)
		return
	}

	for len(b.pending) > 0 {
		if !endOfWalk && len(b.pending) < b.batchSize() && b.bytes < b.params.MaxBytes {
			return
		}
		count, bytes := b.nextBatch()
		statuses := b.pending[:count]
		b.pending = b.pending[count:]
		b.bytes -= bytes
		b.dispatch(statuses)
	}
	b.bytes = 0
}

func (b *zipBatchDownloader) batchSize() int {
	size := b.params.BatchSize
	if size <= 0 {
		size = b.eligibleSeen / (2 * b.params.ConcurrentBatches)
		if size < defaultZipBatchDynamicBatchFloor {
			size = defaultZipBatchDynamicBatchFloor
		}
	}
	if size > b.params.MaxFiles {
		size = b.params.MaxFiles
	}
	if size < 1 {
		size = 1
	}
	return size
}

func (b *zipBatchDownloader) nextBatch() (int, int64) {
	batchSize := b.batchSize()
	var bytes int64
	for i, downloadStatus := range b.pending {
		size := downloadStatus.Size()
		if i > 0 && (i+1 > batchSize || bytes+size > b.params.MaxBytes) {
			return i, bytes
		}
		bytes += size
		if i+1 >= batchSize || bytes >= b.params.MaxBytes {
			return i + 1, bytes
		}
	}
	return len(b.pending), bytes
}

func (b *zipBatchDownloader) dissolve(statuses []*DownloadStatus) {
	if stats := b.stats(); stats != nil {
		stats.batchesDissolved.Add(1)
		stats.dissolvedFiles.Add(int64(len(statuses)))
	}
	b.log("dissolved", map[string]interface{}{
		"files": len(statuses),
		"bytes": zipBatchStatusesSize(statuses),
	})
	for _, downloadStatus := range statuses {
		b.markDissolvedPerFile(downloadStatus)
		enqueueIndexedDownloadDirect(b.job, b.ctx, downloadStatus, b.signal)
	}
}

func (b *zipBatchDownloader) dispatch(statuses []*DownloadStatus) {
	select {
	case <-b.ctx.Done():
		b.cancel(statuses)
		return
	case b.batchSlots <- struct{}{}:
	}
	b.applyProbeState(false)
	b.applyCircuitBreaker()
	if b.mode == zipBatchModeDissolved {
		<-b.batchSlots
		b.dissolve(statuses)
		return
	}

	if !b.job.fileAdmissionManager().WaitWithContext(b.ctx) {
		<-b.batchSlots
		b.cancel(statuses)
		return
	}

	if stats := b.stats(); stats != nil {
		stats.batchesDispatched.Add(1)
		stats.batchFiles.Add(int64(len(statuses)))
	}
	b.log("dispatched", map[string]interface{}{
		"files":              len(statuses),
		"bytes":              zipBatchStatusesSize(statuses),
		"batch_size":         b.batchSize(),
		"concurrent_batches": b.params.ConcurrentBatches,
		"extraction":         string(b.params.Extraction),
	})

	go func() {
		var fallback []zipBatchFallback
		func() {
			defer func() {
				b.job.fileAdmissionManager().Done()
				<-b.batchSlots
			}()
			fallback = b.run(statuses)
			if zipBatchHasFallbackReason(fallback, zipBatchFallbackRetriesExhausted) {
				b.noteRetriesExhaustedBatch()
			}
		}()
		b.enqueueFallbacks(fallback)
	}()
}

func (b *zipBatchDownloader) cancel(statuses []*DownloadStatus) {
	for _, downloadStatus := range statuses {
		b.job.UpdateStatus(status.Canceled, downloadStatus, nil)
		b.signal <- downloadStatus
	}
}
