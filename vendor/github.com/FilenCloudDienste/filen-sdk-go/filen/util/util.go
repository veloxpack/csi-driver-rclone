package util

import (
	"math"
	"time"
)

// TimestampToTime converts a timestamp to a time.Time.
// It is not known whether the timestamp is in milliseconds or seconds.
func TimestampToTime(timestamp int64) time.Time {
	now := time.Now().UnixMilli()
	if math.Abs(float64(now-timestamp)) < math.Abs(float64(now-timestamp*1000)) {
		// ms timestamps
		return time.UnixMilli(timestamp)
	} else {
		// (legacy) seconds timestamps
		return time.Unix(timestamp, 0)
	}
}
