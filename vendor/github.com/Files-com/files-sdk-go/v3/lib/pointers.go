package lib

import "time"

func Bool(bool bool) *bool {
	return &bool
}

func UnWrapBool(b *bool) bool {
	if b == nil {
		return false
	}

	return *b
}

func Time(t time.Time) *time.Time {
	return &t
}

func Int64(i int64) *int64 {
	return &i
}

func Ptr[T any](t T) *T {
	return &t
}
