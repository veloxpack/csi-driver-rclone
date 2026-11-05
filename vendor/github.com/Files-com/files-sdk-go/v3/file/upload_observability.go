package file

import (
	"sync"
	"time"
)

const uploadV1SlowPartThreshold = 10 * time.Second

type uploadV1SchedulerStats struct {
	enabled bool
	mu      sync.Mutex

	scheduledParts           int
	waitCount                int
	waitCanceled             int
	waitDuration             time.Duration
	maxRunningParts          int
	uploadURLRefreshCount    int
	uploadURLRefreshErrors   int
	uploadURLRefreshDuration time.Duration
	httpCallCount            int
	httpCallErrors           int
	httpCallBytes            int64
	httpCallDuration         time.Duration
	directAttempts           int
	directSuccesses          int
	directFailures           int
	directDisabled           int
	partCompleteCount        int
	partCompleteErrors       int
	partCompleteBytes        int64
	partCompleteDuration     time.Duration
	partCompleteMaxDuration  time.Duration
	partSlowCount            int
	partSlowDuration         time.Duration
	partSlowMaxDuration      time.Duration
	finalizeCount            int
	finalizeErrors           int
	finalizeDuration         time.Duration
}

func (s *uploadV1SchedulerStats) enable(enabled bool) {
	s.enabled = enabled
}

func (s *uploadV1SchedulerStats) recordScheduled(waitDuration time.Duration, ok bool, running int) {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitCount++
	s.waitDuration += waitDuration
	if !ok {
		s.waitCanceled++
		return
	}
	s.scheduledParts++
	if running > s.maxRunningParts {
		s.maxRunningParts = running
	}
}

func (s *uploadV1SchedulerStats) recordUploadURLRefresh(duration time.Duration, err error) {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadURLRefreshCount++
	s.uploadURLRefreshDuration += duration
	if err != nil {
		s.uploadURLRefreshErrors++
	}
}

func (s *uploadV1SchedulerStats) recordHTTPCall(duration time.Duration, bytes int64, err error) {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.httpCallCount++
	s.httpCallDuration += duration
	s.httpCallBytes += bytes
	if err != nil {
		s.httpCallErrors++
	}
}

func (s *uploadV1SchedulerStats) recordDirectAttempt() {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directAttempts++
}

func (s *uploadV1SchedulerStats) recordDirectSuccess() {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directSuccesses++
}

func (s *uploadV1SchedulerStats) recordDirectFailure() {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directFailures++
}

func (s *uploadV1SchedulerStats) recordDirectDisabled() {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directDisabled++
}

func (s *uploadV1SchedulerStats) recordPartComplete(bytes int64, duration time.Duration, err error) {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.partCompleteCount++
	s.partCompleteBytes += bytes
	s.partCompleteDuration += duration
	if duration > s.partCompleteMaxDuration {
		s.partCompleteMaxDuration = duration
	}
	if err != nil {
		s.partCompleteErrors++
	}
	if duration >= uploadV1SlowPartThreshold {
		s.partSlowCount++
		s.partSlowDuration += duration
		if duration > s.partSlowMaxDuration {
			s.partSlowMaxDuration = duration
		}
	}
}

func (s *uploadV1SchedulerStats) recordFinalize(duration time.Duration, err error) {
	if !s.enabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finalizeCount++
	s.finalizeDuration += duration
	if err != nil {
		s.finalizeErrors++
	}
}

func (s *uploadV1SchedulerStats) addLogAttrs(attrs map[string]any) bool {
	if !s.enabled {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs["scheduler_scheduled_parts"] = s.scheduledParts
	attrs["scheduler_wait_count"] = s.waitCount
	attrs["scheduler_wait_canceled"] = s.waitCanceled
	attrs["scheduler_wait_duration_ns"] = s.waitDuration.Nanoseconds()
	attrs["scheduler_max_running_parts"] = s.maxRunningParts
	attrs["scheduler_upload_url_refresh_count"] = s.uploadURLRefreshCount
	attrs["scheduler_upload_url_refresh_errors"] = s.uploadURLRefreshErrors
	attrs["scheduler_upload_url_refresh_duration_ns"] = s.uploadURLRefreshDuration.Nanoseconds()
	attrs["scheduler_http_call_count"] = s.httpCallCount
	attrs["scheduler_http_call_errors"] = s.httpCallErrors
	attrs["scheduler_http_call_bytes"] = s.httpCallBytes
	attrs["scheduler_http_call_duration_ns"] = s.httpCallDuration.Nanoseconds()
	attrs["scheduler_direct_attempts"] = s.directAttempts
	attrs["scheduler_direct_successes"] = s.directSuccesses
	attrs["scheduler_direct_failures"] = s.directFailures
	attrs["scheduler_direct_disabled"] = s.directDisabled
	attrs["scheduler_part_complete_count"] = s.partCompleteCount
	attrs["scheduler_part_complete_errors"] = s.partCompleteErrors
	attrs["scheduler_part_complete_bytes"] = s.partCompleteBytes
	attrs["scheduler_part_complete_duration_ns"] = s.partCompleteDuration.Nanoseconds()
	attrs["scheduler_part_complete_max_duration_ns"] = s.partCompleteMaxDuration.Nanoseconds()
	attrs["scheduler_part_slow_threshold_ns"] = uploadV1SlowPartThreshold.Nanoseconds()
	attrs["scheduler_part_slow_count"] = s.partSlowCount
	attrs["scheduler_part_slow_duration_ns"] = s.partSlowDuration.Nanoseconds()
	attrs["scheduler_part_slow_max_duration_ns"] = s.partSlowMaxDuration.Nanoseconds()
	attrs["scheduler_finalize_count"] = s.finalizeCount
	attrs["scheduler_finalize_errors"] = s.finalizeErrors
	attrs["scheduler_finalize_duration_ns"] = s.finalizeDuration.Nanoseconds()
	return true
}

func (u *uploadIO) logUploadV1SchedulerSummary(err error) {
	attrs := map[string]any{
		"timestamp":     time.Now(),
		"event":         "upload v1 scheduler summary",
		"bytes_written": u.bytesWritten,
		"success":       err == nil,
	}
	if err != nil {
		attrs["error"] = err.Error()
	}
	if !u.uploadV1Stats.addLogAttrs(attrs) {
		return
	}
	u.LogPath(u.Path, attrs)
}
