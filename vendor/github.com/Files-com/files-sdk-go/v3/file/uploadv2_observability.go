package file

import (
	"sync"
	"time"
)

const uploadV2SlowPartThreshold = 10 * time.Second

type uploadV2SchedulerStats struct {
	mu sync.Mutex

	directScheduledParts     int
	preparedParts            int
	dispatchedPreparedParts  int
	preparedBytes            int64
	preparedMemoryBytes      int64
	runwayPartCapBlocks      int
	runwayByteCapBlocks      int
	maxReadyParts            int
	maxReadyBytes            int64
	readerBuildCount         int
	readerBuildErrors        int
	readerBuildBytes         int64
	readerBuildDuration      time.Duration
	adaptiveWaitCount        int
	adaptiveWaitCanceled     int
	adaptiveWaitDuration     time.Duration
	globalWaitCount          int
	globalWaitCanceled       int
	globalWaitDuration       time.Duration
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
}

func (s *uploadV2SchedulerStats) recordDirectScheduled() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directScheduledParts++
}

func (s *uploadV2SchedulerStats) recordPrepared(bytes int64, memoryBytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preparedParts++
	s.preparedBytes += bytes
	s.preparedMemoryBytes += memoryBytes
}

func (s *uploadV2SchedulerStats) recordDispatchedPrepared() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dispatchedPreparedParts++
}

func (s *uploadV2SchedulerStats) recordRunwayPartCapBlock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runwayPartCapBlocks++
}

func (s *uploadV2SchedulerStats) recordRunwayByteCapBlock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runwayByteCapBlocks++
}

func (s *uploadV2SchedulerStats) recordReadyDepth(parts int, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if parts > s.maxReadyParts {
		s.maxReadyParts = parts
	}
	if bytes > s.maxReadyBytes {
		s.maxReadyBytes = bytes
	}
}

func (s *uploadV2SchedulerStats) recordReaderBuild(bytes int64, duration time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readerBuildCount++
	s.readerBuildBytes += bytes
	s.readerBuildDuration += duration
	if err != nil {
		s.readerBuildErrors++
	}
}

func (s *uploadV2SchedulerStats) recordAdaptiveWait(duration time.Duration, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adaptiveWaitCount++
	s.adaptiveWaitDuration += duration
	if !ok {
		s.adaptiveWaitCanceled++
	}
}

func (s *uploadV2SchedulerStats) recordGlobalWait(duration time.Duration, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalWaitCount++
	s.globalWaitDuration += duration
	if !ok {
		s.globalWaitCanceled++
	}
}

func (s *uploadV2SchedulerStats) recordUploadURLRefresh(duration time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadURLRefreshCount++
	s.uploadURLRefreshDuration += duration
	if err != nil {
		s.uploadURLRefreshErrors++
	}
}

func (s *uploadV2SchedulerStats) recordHTTPCall(duration time.Duration, bytes int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.httpCallCount++
	s.httpCallDuration += duration
	s.httpCallBytes += bytes
	if err != nil {
		s.httpCallErrors++
	}
}

func (s *uploadV2SchedulerStats) recordDirectAttempt() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directAttempts++
}

func (s *uploadV2SchedulerStats) recordDirectSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directSuccesses++
}

func (s *uploadV2SchedulerStats) recordDirectFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directFailures++
}

func (s *uploadV2SchedulerStats) recordDirectDisabled() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.directDisabled++
}

func (s *uploadV2SchedulerStats) recordPartComplete(bytes int64, duration time.Duration, err error) {
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
	if duration >= uploadV2SlowPartThreshold {
		s.partSlowCount++
		s.partSlowDuration += duration
		if duration > s.partSlowMaxDuration {
			s.partSlowMaxDuration = duration
		}
	}
}

func (s *uploadV2SchedulerStats) addLogAttrs(attrs map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs["scheduler_direct_scheduled_parts"] = s.directScheduledParts
	attrs["scheduler_prepared_parts"] = s.preparedParts
	attrs["scheduler_dispatched_prepared_parts"] = s.dispatchedPreparedParts
	attrs["scheduler_prepared_bytes"] = s.preparedBytes
	attrs["scheduler_prepared_memory_bytes"] = s.preparedMemoryBytes
	attrs["scheduler_runway_part_cap_blocks"] = s.runwayPartCapBlocks
	attrs["scheduler_runway_byte_cap_blocks"] = s.runwayByteCapBlocks
	attrs["scheduler_max_ready_parts"] = s.maxReadyParts
	attrs["scheduler_max_ready_bytes"] = s.maxReadyBytes
	attrs["scheduler_reader_build_count"] = s.readerBuildCount
	attrs["scheduler_reader_build_errors"] = s.readerBuildErrors
	attrs["scheduler_reader_build_bytes"] = s.readerBuildBytes
	attrs["scheduler_reader_build_duration_ns"] = s.readerBuildDuration.Nanoseconds()
	attrs["scheduler_adaptive_wait_count"] = s.adaptiveWaitCount
	attrs["scheduler_adaptive_wait_canceled"] = s.adaptiveWaitCanceled
	attrs["scheduler_adaptive_wait_duration_ns"] = s.adaptiveWaitDuration.Nanoseconds()
	attrs["scheduler_global_wait_count"] = s.globalWaitCount
	attrs["scheduler_global_wait_canceled"] = s.globalWaitCanceled
	attrs["scheduler_global_wait_duration_ns"] = s.globalWaitDuration.Nanoseconds()
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
	attrs["scheduler_part_slow_threshold_ns"] = uploadV2SlowPartThreshold.Nanoseconds()
	attrs["scheduler_part_slow_count"] = s.partSlowCount
	attrs["scheduler_part_slow_duration_ns"] = s.partSlowDuration.Nanoseconds()
	attrs["scheduler_part_slow_max_duration_ns"] = s.partSlowMaxDuration.Nanoseconds()
}

func (e *uploadV2Engine) logUploadV2SchedulerSummary(err error) {
	attrs := map[string]any{
		"timestamp":     time.Now(),
		"event":         "upload v2 scheduler summary",
		"bytes_written": e.u.bytesWritten,
		"success":       err == nil,
	}
	if err != nil {
		attrs["error"] = err.Error()
	}
	e.stats.addLogAttrs(attrs)
	e.u.logUploadV2(attrs)
}
