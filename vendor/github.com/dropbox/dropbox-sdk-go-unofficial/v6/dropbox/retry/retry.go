// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package retry implements the retry and exponential-backoff policy used by the
// Dropbox SDK. A Policy drives which HTTP responses are retried and how long to
// wait between attempts.
package retry

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxRetryAfterSeconds is the largest whole-second Retry-After value that can be
// converted to a time.Duration without overflowing int64. Anything larger is
// treated as "effectively unbounded" so it is rejected by the MaxRetryAfter cap
// rather than wrapping to a negative duration.
const maxRetryAfterSeconds = int64(math.MaxInt64 / int64(time.Second))

// Default backoff and Retry-After bounds applied when a Policy leaves them unset.
const (
	defaultInitialBackoff = 500 * time.Millisecond
	defaultMaxBackoff     = 30 * time.Second
	defaultMaxRetryAfter  = 30 * time.Second
)

// Policy controls automatic retries for a request. MaxRetries is the number of
// attempts after the initial request. A zero MaxRetries disables retries.
type Policy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	MaxRetryAfter  time.Duration
	// Retryable409Tags are top-level Stone error tags that should be retried
	// when an API route returns HTTP 409 Conflict.
	Retryable409Tags []string
}

// Normalized returns a copy of p with invalid or unset fields replaced by their
// defaults.
func (p Policy) Normalized() Policy {
	if p.MaxRetries < 0 {
		p.MaxRetries = 0
	}
	if p.InitialBackoff <= 0 {
		p.InitialBackoff = defaultInitialBackoff
	}
	if p.MaxBackoff <= 0 {
		p.MaxBackoff = defaultMaxBackoff
	}
	if p.MaxRetryAfter <= 0 {
		p.MaxRetryAfter = defaultMaxRetryAfter
	}
	return p
}

// CanRetry reports whether another attempt is permitted after the given
// zero-based attempt, i.e. whether attempt is below MaxRetries.
func (p Policy) CanRetry(attempt int) bool {
	return attempt < p.MaxRetries
}

// Backoff returns the exponential backoff duration for the given zero-based
// attempt, capped at MaxBackoff.
func (p Policy) Backoff(attempt int) time.Duration {
	wait := p.InitialBackoff
	for i := 0; i < attempt; i++ {
		if wait >= p.MaxBackoff || wait > p.MaxBackoff-wait {
			return p.MaxBackoff
		}
		wait *= 2
		if wait >= p.MaxBackoff {
			return p.MaxBackoff
		}
	}
	if wait > p.MaxBackoff {
		return p.MaxBackoff
	}
	return wait
}

// Delay reports whether the HTTP response should be retried and, if so, how long
// to wait before the next attempt. body is the fully read response body.
func (p Policy) Delay(resp *http.Response, body []byte, attempt int) (time.Duration, bool) {
	if attempt >= p.MaxRetries || !p.isRetryableStatus(resp.StatusCode, body) {
		return 0, false
	}

	if wait, ok := retryAfter(resp.Header.Get("Retry-After")); ok {
		return p.cappedRetryAfter(wait)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		if wait, ok := retryAfterBody(body); ok {
			return p.cappedRetryAfter(wait)
		}
	}

	return p.Backoff(attempt), true
}

func (p Policy) isRetryableStatus(statusCode int, body []byte) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable:
		return true
	case http.StatusConflict:
		return isRetryable409(body, p.Retryable409Tags)
	default:
		return false
	}
}

func (p Policy) cappedRetryAfter(wait time.Duration) (time.Duration, bool) {
	if wait > p.MaxRetryAfter {
		return 0, false
	}
	return wait, true
}

func isRetryable409(body []byte, retryableTags []string) bool {
	if len(retryableTags) == 0 {
		return false
	}
	tag, ok := stoneErrorTag(body)
	if !ok {
		return false
	}
	for _, retryableTag := range retryableTags {
		if tag == retryableTag {
			return true
		}
	}
	return false
}

func stoneErrorTag(body []byte) (string, bool) {
	var payload struct {
		ErrorSummary string `json:"error_summary"`
		Error        struct {
			Tag string `json:".tag"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}
	if payload.Error.Tag != "" {
		return payload.Error.Tag, true
	}
	if payload.ErrorSummary == "" {
		return "", false
	}
	tag, _, _ := strings.Cut(payload.ErrorSummary, "/")
	return tag, tag != ""
}

func retryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	if seconds, ok := parseRetryAfterSeconds(value); ok {
		return secondsToDuration(seconds), true
	}

	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	wait := time.Until(when)
	if wait < 0 {
		wait = 0
	}
	return wait, true
}

func retryAfterBody(body []byte) (time.Duration, bool) {
	var payload struct {
		Error struct {
			RetryAfter *uint64 `json:"retry_after"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.Error.RetryAfter == nil {
		return 0, false
	}
	seconds := *payload.Error.RetryAfter
	if seconds > uint64(maxRetryAfterSeconds) {
		return time.Duration(math.MaxInt64), true
	}
	return time.Duration(seconds) * time.Second, true
}

func parseRetryAfterSeconds(value string) (int64, bool) {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return seconds, true
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return r < '0' || r > '9'
	}) == -1 {
		return math.MaxInt64, true
	}
	return 0, false
}

// secondsToDuration converts a whole-second Retry-After value to a Duration,
// clamping negatives to zero and saturating at math.MaxInt64 instead of
// overflowing (which would wrap to a negative duration and bypass MaxRetryAfter).
func secondsToDuration(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	if seconds > maxRetryAfterSeconds {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(seconds) * time.Second
}

// Sleep waits for d or until ctx is done, whichever comes first. It returns the
// context error if the context is cancelled during the wait.
func Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
