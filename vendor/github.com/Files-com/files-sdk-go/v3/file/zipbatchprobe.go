package file

// ZIP batch probe system: phases, sustained windows, verdicts, re-probe scheduling, and the circuit breaker.
//
// Cross-cutting ZIP batch invariants:
//   - zipBatchDownloader fields are walk-goroutine-only unless the field is explicitly atomic or mutex-protected.
//   - Each DownloadStatus is signaled exactly once; a duplicate signal can panic and a missing signal can hang the job.
//   - Every file admission Wait/WaitWithContext has exactly one Done, including retry, fallback, and cancel paths.
//   - Batch-routed files prepare at dispatch; per-file downloads prepare at download.

import (
	"sort"
	"sync"
	"time"

	"github.com/Files-com/files-sdk-go/v3/file/status"
)

type zipBatchTuningConfig struct {
	ProbePhaseAMinCompletions int
	ProbePhaseAMinDuration    time.Duration
	ProbePhaseAHardCap        time.Duration
	ProbePhaseAWindow         time.Duration
	ProbePhaseAWindow2After   int
	ProbePhaseAWindow3After   int
	ProbePhaseBMinCompletions int
	ProbePhaseBWarmup         time.Duration
	ProbePhaseBWindow         time.Duration
	ProbePhaseBEndMinWindow   time.Duration
	ProbePhaseBHardCap        time.Duration
	DefaultReprobeInterval    time.Duration
	MaxReprobeInterval        time.Duration
}

// zipBatchTuning corrals package-level mutable test seams. Production values
// come from issue #3511 sustained-window benchmarks.
var zipBatchTuning = zipBatchTuningConfig{
	ProbePhaseAMinCompletions: 64,
	ProbePhaseAMinDuration:    5 * time.Second,
	ProbePhaseAHardCap:        12 * time.Second,
	ProbePhaseAWindow:         4 * time.Second,
	ProbePhaseAWindow2After:   2000,
	ProbePhaseAWindow3After:   5000,
	ProbePhaseBMinCompletions: 64,
	ProbePhaseBWarmup:         3 * time.Second,
	ProbePhaseBWindow:         6 * time.Second,
	ProbePhaseBEndMinWindow:   3 * time.Second,
	ProbePhaseBHardCap:        15 * time.Second,
	// Production re-probes start at 3m and back off to 30m; a permanent ZIP
	// loser spends at most one capped Phase B window per interval.
	DefaultReprobeInterval: 3 * time.Minute,
	MaxReprobeInterval:     30 * time.Minute,
}

const zipBatchProbeBaselineMaxWindows = 3
const zipBatchCircuitBreakerBatches = 3

type zipBatchProbeDecision string

const (
	zipBatchProbeDecisionNone      zipBatchProbeDecision = "none"
	zipBatchProbeDecisionCommitted zipBatchProbeDecision = "committed"
	zipBatchProbeDecisionDissolved zipBatchProbeDecision = "dissolved"
)

type zipBatchProbePhase int

const (
	zipBatchProbePhaseNone zipBatchProbePhase = iota
	zipBatchProbePhaseA
	zipBatchProbePhaseB
)

type zipBatchProbeRateWindow struct {
	rate        float64
	completions int
	duration    time.Duration
}

type zipBatchProbe struct {
	mu                     sync.Mutex
	phase                  zipBatchProbePhase
	perFileIDs             map[string]struct{}
	zipIDs                 map[string]struct{}
	perFileCompletions     []time.Time
	zipCompletions         []time.Time
	perFileRateWindows     []zipBatchProbeRateWindow
	perFileFirstCompletion time.Time
	zipFirstCompletion     time.Time
	decision               zipBatchProbeDecision
	attempt                int
	zipRate                float64
	perFileRate            float64
	zipRateCompletions     int
	perFileRateCompletions int
	zipRateDuration        time.Duration
	perFileRateDuration    time.Duration
}

func newZipBatchProbe() *zipBatchProbe {
	return &zipBatchProbe{
		perFileIDs: make(map[string]struct{}),
		zipIDs:     make(map[string]struct{}),
		decision:   zipBatchProbeDecisionNone,
	}
}

func (b *zipBatchDownloader) registerProbeReporter() {
	if b == nil || b.job == nil || b.params.MinAdvantage < 0 {
		return
	}
	b.job.RegisterFileEvent(func(file JobFile) {
		if file.Status.Is(status.Complete) {
			b.probePerFileComplete(file.Id, time.Now())
		}
	}, status.Ended...)
}

func (b *zipBatchDownloader) markDissolvedPerFile(downloadStatus *DownloadStatus) {
	if b.params.MinAdvantage < 0 || b.params.ReprobeInterval < 0 {
		return
	}
	b.markProbePerFile(downloadStatus, false)
}

func (b *zipBatchDownloader) markProbePerFile(downloadStatus *DownloadStatus, countStats bool) {
	if b.probe == nil {
		return
	}
	b.probe.mu.Lock()
	b.probe.perFileIDs[downloadStatus.Id()] = struct{}{}
	b.probe.mu.Unlock()
	if countStats {
		if stats := b.stats(); stats != nil {
			stats.probePerFileFiles.Add(1)
		}
	}
}

func (b *zipBatchDownloader) markProbeZip(downloadStatus *DownloadStatus) {
	if b.probe == nil {
		return
	}
	b.probe.mu.Lock()
	b.probe.zipIDs[downloadStatus.Id()] = struct{}{}
	b.probe.mu.Unlock()
	if stats := b.stats(); stats != nil {
		stats.probeZipFiles.Add(1)
	}
}

func (b *zipBatchDownloader) shouldReprobe() bool {
	return b.probe != nil && b.params.MinAdvantage >= 0 && b.params.ReprobeInterval >= 0
}

func (b *zipBatchDownloader) startReprobeIfDue(now time.Time) bool {
	if !b.shouldReprobe() || b.walkEnded || b.reprobeAfter.IsZero() || now.Before(b.reprobeAfter) {
		return false
	}
	b.probe.mu.Lock()
	perFileWindows := zipBatchProbeRecentWindows(b.probe.perFileCompletions, now, zipBatchTuning.ProbePhaseAWindow, zipBatchProbeBaselineMaxWindows)
	if len(perFileWindows) == 0 {
		b.probe.mu.Unlock()
		return false
	}
	perFileWindow := zipBatchProbeMedianWindow(perFileWindows)
	b.probe.attempt++
	attempt := b.probe.attempt
	b.probe.phase = zipBatchProbePhaseB
	b.probe.decision = zipBatchProbeDecisionNone
	b.probe.zipIDs = make(map[string]struct{})
	b.probe.zipCompletions = nil
	b.probe.zipFirstCompletion = time.Time{}
	b.probe.perFileRateWindows = perFileWindows
	b.probe.perFileRate = perFileWindow.rate
	b.probe.perFileRateCompletions = perFileWindow.completions
	b.probe.perFileRateDuration = perFileWindow.duration
	b.probe.mu.Unlock()

	b.transition(zipBatchModeProbePhaseB, "reprobe")
	b.reprobeAfter = time.Time{}
	if stats := b.stats(); stats != nil {
		stats.reprobes.Add(1)
		stats.probeBaselineWindows.Store(int64(len(perFileWindows)))
	}
	b.log("probe-start", map[string]interface{}{
		"attempt":                   attempt,
		"phase":                     "zip",
		"reprobe":                   true,
		"baseline_windows":          len(perFileWindows),
		"per_file_rate":             perFileWindow.rate,
		"per_file_rate_completions": perFileWindow.completions,
		"per_file_window_ms":        perFileWindow.duration.Milliseconds(),
		"zip_min_completions":       zipBatchTuning.ProbePhaseBMinCompletions,
		"zip_warmup_seconds":        zipBatchTuning.ProbePhaseBWarmup.Seconds(),
		"zip_window_seconds":        zipBatchTuning.ProbePhaseBWindow.Seconds(),
		"zip_hard_seconds":          zipBatchTuning.ProbePhaseBHardCap.Seconds(),
		"min_advantage":             b.params.MinAdvantage,
	})
	return true
}

func (b *zipBatchDownloader) scheduleReprobe(now time.Time, backoff bool) {
	if !b.shouldReprobe() || b.walkEnded {
		return
	}
	if b.reprobeWait <= 0 {
		b.reprobeWait = b.params.ReprobeInterval
	}
	if backoff {
		b.reprobeWait *= 2
		if b.reprobeWait > zipBatchTuning.MaxReprobeInterval {
			b.reprobeWait = zipBatchTuning.MaxReprobeInterval
		}
	}
	b.reprobeAfter = now.Add(b.reprobeWait)
	b.log("reprobe-scheduled", map[string]interface{}{
		"interval_ms": b.reprobeWait.Milliseconds(),
		"after":       b.reprobeAfter.Format(time.RFC3339Nano),
	})
}

func (b *zipBatchDownloader) resetReprobeBackoff() {
	if !b.shouldReprobe() {
		return
	}
	b.reprobeWait = b.params.ReprobeInterval
	b.reprobeAfter = time.Time{}
}

func (b *zipBatchDownloader) currentProbeAttempt() int {
	if b.probe == nil {
		return 0
	}
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	return b.probe.attempt
}

func (b *zipBatchDownloader) probePerFileComplete(id string, completedAt time.Time) {
	if b.probe == nil {
		return
	}
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	if _, ok := b.probe.perFileIDs[id]; !ok {
		return
	}
	delete(b.probe.perFileIDs, id)
	if b.probe.perFileFirstCompletion.IsZero() {
		b.probe.perFileFirstCompletion = completedAt
	}
	b.probe.perFileCompletions = append(b.probe.perFileCompletions, completedAt)
	if b.probe.phase == zipBatchProbePhaseA {
		return
	}
	b.probe.perFileCompletions = zipBatchPruneCompletions(
		b.probe.perFileCompletions,
		completedAt.Add(-zipBatchTuning.ProbePhaseAWindow*time.Duration(zipBatchProbeBaselineMaxWindows)),
	)
}

func (b *zipBatchDownloader) probeZipFinalized(downloadStatus *DownloadStatus) {
	if b.probe == nil || downloadStatus == nil || !downloadStatus.Status().Is(status.Complete) {
		return
	}
	now := time.Now()
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	if b.probe.decision != zipBatchProbeDecisionNone || b.probe.phase != zipBatchProbePhaseB {
		return
	}
	if _, ok := b.probe.zipIDs[downloadStatus.Id()]; !ok {
		return
	}
	delete(b.probe.zipIDs, downloadStatus.Id())
	if b.probe.zipFirstCompletion.IsZero() {
		b.probe.zipFirstCompletion = now
	}
	b.probe.zipCompletions = append(b.probe.zipCompletions, now)
	b.maybeAdvanceProbeLocked(false, now)
}

func (b *zipBatchDownloader) updateProbe(endOfWalk bool) {
	if (b.mode != zipBatchModeProbePhaseA && b.mode != zipBatchModeProbePhaseB) || b.probe == nil {
		return
	}
	b.probe.mu.Lock()
	b.maybeAdvanceProbeLocked(endOfWalk, time.Now())
	b.probe.mu.Unlock()
}

func (b *zipBatchDownloader) maybeAdvanceProbeLocked(endOfWalk bool, now time.Time) {
	if b.probe.decision != zipBatchProbeDecisionNone {
		return
	}
	switch b.probe.phase {
	case zipBatchProbePhaseA:
		b.maybeFinishProbePhaseALocked(endOfWalk, now)
	case zipBatchProbePhaseB:
		b.maybeFinishProbePhaseBLocked(endOfWalk, now)
	}
}

func (b *zipBatchDownloader) maybeFinishProbePhaseALocked(endOfWalk bool, now time.Time) {
	count := len(b.probe.perFileCompletions)
	if endOfWalk {
		perFileRate, perFileCount, perFileDuration := zipBatchProbeTrailingRate(b.probe.perFileCompletions, now, zipBatchTuning.ProbePhaseAWindow)
		b.setProbeDecisionLocked(zipBatchProbeDecisionDissolved, 0, perFileRate, 0, perFileCount, 0, perFileDuration, "phase-a-end")
		return
	}
	if b.probe.perFileFirstCompletion.IsZero() {
		return
	}
	elapsed := now.Sub(b.probe.perFileFirstCompletion)
	if (count >= zipBatchTuning.ProbePhaseAMinCompletions && elapsed >= zipBatchTuning.ProbePhaseAMinDuration) ||
		elapsed >= zipBatchTuning.ProbePhaseAHardCap {
		b.finishProbePhaseAWindowLocked(now, count)
	}
}

func (b *zipBatchDownloader) finishProbePhaseAWindowLocked(now time.Time, completions int) {
	window := zipBatchProbeTrailingWindow(b.probe.perFileCompletions, now, zipBatchTuning.ProbePhaseAWindow)
	b.probe.perFileRateWindows = append(b.probe.perFileRateWindows, window)
	if stats := b.stats(); stats != nil {
		stats.probeBaselineWindows.Store(int64(len(b.probe.perFileRateWindows)))
	}
	b.log("probe-phase-a-window", map[string]interface{}{
		"attempt":                   b.probe.attempt,
		"baseline_windows":          len(b.probe.perFileRateWindows),
		"per_file_completions":      completions,
		"per_file_rate_completions": window.completions,
		"per_file_rate":             window.rate,
		"per_file_window_ms":        window.duration.Milliseconds(),
	})
	b.probe.perFileCompletions = nil
	b.probe.perFileFirstCompletion = time.Time{}
	if b.needsAnotherProbePhaseAWindowLocked() {
		return
	}
	median := zipBatchProbeMedianWindow(b.probe.perFileRateWindows)
	b.probe.perFileRate = median.rate
	b.probe.perFileRateCompletions = median.completions
	b.probe.perFileRateDuration = median.duration
	b.probe.phase = zipBatchProbePhaseB
	b.log("probe-phase-b-start", map[string]interface{}{
		"attempt":                   b.probe.attempt,
		"baseline_windows":          len(b.probe.perFileRateWindows),
		"per_file_rate_completions": b.probe.perFileRateCompletions,
		"per_file_rate":             b.probe.perFileRate,
		"per_file_window_ms":        b.probe.perFileRateDuration.Milliseconds(),
	})
}

func (b *zipBatchDownloader) needsAnotherProbePhaseAWindowLocked() bool {
	switch len(b.probe.perFileRateWindows) {
	case 1:
		return b.eligibleSeen > zipBatchTuning.ProbePhaseAWindow2After
	case 2:
		return b.eligibleSeen > zipBatchTuning.ProbePhaseAWindow3After
	default:
		return false
	}
}

func (b *zipBatchDownloader) maybeFinishProbePhaseBLocked(endOfWalk bool, now time.Time) {
	if b.probe.zipFirstCompletion.IsZero() {
		if endOfWalk {
			b.decideProbePhaseBLocked(now, "phase-b-end")
		}
		return
	}
	warmupEnd := b.probe.zipFirstCompletion.Add(zipBatchTuning.ProbePhaseBWarmup)
	postWarmupElapsed := now.Sub(warmupEnd)
	postWarmupCompletions := zipBatchProbeCompletionCount(b.probe.zipCompletions, warmupEnd, now)
	switch {
	case postWarmupCompletions >= zipBatchTuning.ProbePhaseBMinCompletions && postWarmupElapsed >= zipBatchTuning.ProbePhaseBWindow:
		b.decideProbePhaseBLocked(now, "phase-b-full")
	case now.Sub(b.probe.zipFirstCompletion) >= zipBatchTuning.ProbePhaseBHardCap:
		b.decideProbePhaseBLocked(now, "phase-b-time")
	case endOfWalk:
		if postWarmupElapsed >= zipBatchTuning.ProbePhaseBEndMinWindow {
			b.decideProbePhaseBLocked(now, "phase-b-end")
		} else {
			b.setProbeDecisionLocked(zipBatchProbeDecisionDissolved, 0, b.probe.perFileRate, 0, b.probe.perFileRateCompletions, 0, b.probe.perFileRateDuration, "phase-b-end")
		}
	}
}

func (b *zipBatchDownloader) decideProbePhaseBLocked(now time.Time, reason string) {
	warmupEnd := b.probe.zipFirstCompletion.Add(zipBatchTuning.ProbePhaseBWarmup)
	zipRate, zipRateCount, zipRateDuration := zipBatchProbeWarmTrailingRate(b.probe.zipCompletions, warmupEnd, now, zipBatchTuning.ProbePhaseBWindow)
	perFileRate := b.probe.perFileRate
	perFileRateCount := b.probe.perFileRateCompletions
	perFileRateDuration := b.probe.perFileRateDuration
	decision := zipBatchProbeDecisionDissolved
	if zipRateCount > 0 && zipRate >= b.params.MinAdvantage*perFileRate {
		decision = zipBatchProbeDecisionCommitted
	}
	b.setProbeDecisionLocked(decision, zipRate, perFileRate, zipRateCount, perFileRateCount, zipRateDuration, perFileRateDuration, reason)
}

func zipBatchProbeTrailingRate(completions []time.Time, now time.Time, window time.Duration) (float64, int, time.Duration) {
	rateWindow := zipBatchProbeTrailingWindow(completions, now, window)
	return rateWindow.rate, rateWindow.completions, rateWindow.duration
}

func zipBatchProbeTrailingWindow(completions []time.Time, now time.Time, window time.Duration) zipBatchProbeRateWindow {
	if window <= 0 {
		window = time.Millisecond
	}
	return zipBatchProbeWindowRate(completions, now.Add(-window), now)
}

func zipBatchProbeRecentWindows(completions []time.Time, now time.Time, window time.Duration, maxWindows int) []zipBatchProbeRateWindow {
	if window <= 0 {
		window = time.Millisecond
	}
	if maxWindows <= 0 {
		return nil
	}
	windows := make([]zipBatchProbeRateWindow, 0, maxWindows)
	for i := 0; i < maxWindows; i++ {
		end := now.Add(-time.Duration(i) * window)
		rateWindow := zipBatchProbeWindowRate(completions, end.Add(-window), end)
		if rateWindow.completions > 0 {
			windows = append(windows, rateWindow)
		}
	}
	return windows
}

func zipBatchProbeWindowRate(completions []time.Time, start, end time.Time) zipBatchProbeRateWindow {
	duration := end.Sub(start)
	if duration <= 0 {
		return zipBatchProbeRateWindow{}
	}
	count := zipBatchProbeCompletionCount(completions, start, end)
	return zipBatchProbeRateWindow{
		rate:        float64(count) / duration.Seconds(),
		completions: count,
		duration:    duration,
	}
}

func zipBatchProbeMedianWindow(windows []zipBatchProbeRateWindow) zipBatchProbeRateWindow {
	if len(windows) == 0 {
		return zipBatchProbeRateWindow{}
	}
	rates := make([]float64, 0, len(windows))
	var completions int
	var duration time.Duration
	for _, window := range windows {
		rates = append(rates, window.rate)
		completions += window.completions
		duration += window.duration
	}
	sort.Float64s(rates)
	rate := rates[len(rates)/2]
	if len(rates)%2 == 0 {
		rate = (rates[len(rates)/2-1] + rate) / 2
	}
	return zipBatchProbeRateWindow{rate: rate, completions: completions, duration: duration}
}

func zipBatchProbeWarmTrailingRate(completions []time.Time, warmupEnd, now time.Time, window time.Duration) (float64, int, time.Duration) {
	if now.Before(warmupEnd) {
		return 0, 0, 0
	}
	start := now.Add(-window)
	if start.Before(warmupEnd) {
		start = warmupEnd
	}
	duration := now.Sub(start)
	if duration <= 0 {
		return 0, 0, 0
	}
	count := zipBatchProbeCompletionCount(completions, start, now)
	return float64(count) / duration.Seconds(), count, duration
}

func zipBatchProbeCompletionCount(completions []time.Time, start, end time.Time) int {
	count := 0
	for _, completedAt := range completions {
		if !completedAt.Before(start) && !completedAt.After(end) {
			count++
		}
	}
	return count
}

func zipBatchPruneCompletions(completions []time.Time, after time.Time) []time.Time {
	if len(completions) == 0 {
		return completions
	}
	idx := 0
	for idx < len(completions) && completions[idx].Before(after) {
		idx++
	}
	if idx == 0 {
		return completions
	}
	if idx == len(completions) {
		return completions[:0]
	}
	copy(completions, completions[idx:])
	return completions[:len(completions)-idx]
}

func (b *zipBatchDownloader) setProbeDecisionLocked(decision zipBatchProbeDecision, zipRate float64, perFileRate float64, zipCompletions, perFileCompletions int, zipDuration, perFileDuration time.Duration, reason string) {
	b.probe.decision = decision
	b.probe.zipRate = zipRate
	b.probe.perFileRate = perFileRate
	b.probe.zipRateCompletions = zipCompletions
	b.probe.perFileRateCompletions = perFileCompletions
	b.probe.zipRateDuration = zipDuration
	b.probe.perFileRateDuration = perFileDuration
	baselineWindows := len(b.probe.perFileRateWindows)
	if stats := b.stats(); stats != nil {
		stats.probeDecision.Store(string(decision))
		stats.probeZipRateMilli.Store(int64(zipRate * 1000))
		stats.probePerFileRateMilli.Store(int64(perFileRate * 1000))
		stats.probeBaselineWindows.Store(int64(baselineWindows))
	}
	b.log("probe-decision", map[string]interface{}{
		"attempt":                   b.probe.attempt,
		"decision":                  string(decision),
		"reason":                    reason,
		"baseline_windows":          baselineWindows,
		"zip_rate_completions":      zipCompletions,
		"per_file_rate_completions": perFileCompletions,
		"zip_window_ms":             zipDuration.Milliseconds(),
		"per_file_window_ms":        perFileDuration.Milliseconds(),
		"zip_rate":                  zipRate,
		"per_file_rate":             perFileRate,
		"min_advantage":             b.params.MinAdvantage,
	})
}

func (b *zipBatchDownloader) currentProbeDecision() zipBatchProbeDecision {
	if b.probe == nil {
		return zipBatchProbeDecisionNone
	}
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	return b.probe.decision
}

func (b *zipBatchDownloader) currentProbePhase() zipBatchProbePhase {
	if b.probe == nil {
		return zipBatchProbePhaseNone
	}
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	return b.probe.phase
}

func (b *zipBatchDownloader) applyProbeState(endOfWalk bool) {
	if b.mode != zipBatchModeProbePhaseA && b.mode != zipBatchModeProbePhaseB {
		return
	}
	switch b.currentProbeDecision() {
	case zipBatchProbeDecisionCommitted:
		b.transition(zipBatchModeCommitted, "probe-committed")
		b.resetReprobeBackoff()
		b.flush(endOfWalk)
	case zipBatchProbeDecisionDissolved:
		b.transition(zipBatchModeDissolved, "probe-dissolved")
		b.flush(true)
		b.scheduleReprobe(time.Now(), b.currentProbeAttempt() > 1)
		return
	}
	if b.mode == zipBatchModeProbePhaseA && b.currentProbePhase() == zipBatchProbePhaseB {
		b.transition(zipBatchModeProbePhaseB, "probe-phase-b")
	}
}

func (b *zipBatchDownloader) applyCircuitBreaker() {
	if b.circuitTrips.Load() < zipBatchCircuitBreakerBatches || b.mode == zipBatchModeDissolved {
		return
	}
	now := time.Now()
	b.noteProbeDissolved(now, "circuit-breaker")
	b.transition(zipBatchModeDissolved, "circuit-breaker")
	b.flush(true)
	b.scheduleReprobe(now, false)
}

func (b *zipBatchDownloader) noteProbeDissolved(now time.Time, reason string) {
	if !b.shouldReprobe() {
		return
	}
	b.probe.mu.Lock()
	defer b.probe.mu.Unlock()
	if b.probe.decision == zipBatchProbeDecisionDissolved {
		return
	}
	perFileRate := b.probe.perFileRate
	perFileCount := b.probe.perFileRateCompletions
	perFileDuration := b.probe.perFileRateDuration
	if perFileCount == 0 {
		perFileWindows := zipBatchProbeRecentWindows(b.probe.perFileCompletions, now, zipBatchTuning.ProbePhaseAWindow, zipBatchProbeBaselineMaxWindows)
		perFileWindow := zipBatchProbeMedianWindow(perFileWindows)
		b.probe.perFileRateWindows = perFileWindows
		perFileRate, perFileCount, perFileDuration = perFileWindow.rate, perFileWindow.completions, perFileWindow.duration
	}
	b.setProbeDecisionLocked(zipBatchProbeDecisionDissolved, b.probe.zipRate, perFileRate, b.probe.zipRateCompletions, perFileCount, b.probe.zipRateDuration, perFileDuration, reason)
}

func (b *zipBatchDownloader) noteRetriesExhaustedBatch() {
	if b.circuitTrips.Add(1) != zipBatchCircuitBreakerBatches {
		return
	}
	if stats := b.stats(); stats != nil {
		stats.circuitBreakerTrips.Add(1)
	}
	b.log("circuit-breaker", map[string]interface{}{
		"retries_exhausted_batches": zipBatchCircuitBreakerBatches,
	})
}
