// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package sync2

import (
	"time"
)

// WithTimeout calls `do` concurrently and waits for it to complete. If the timeout
// is reached before `do` returns, `onTimeout` will be called; otherwise, `onTimeout`
// will not be called.
//
// Avoid attempting to detect whether a timeout has occurred from within `do`.
// Because `do` runs concurrently with the timeout timer, it may complete at the
// same time as the timeout timer expires, making detection within `do` unreliable.
// Logic specific to the timeout should instead be placed in `onTimeout`.
func WithTimeout(timeout time.Duration, do, onTimeout func()) {
	done := make(chan any)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				done <- err
			}
			close(done)
		}()
		do()
	}()

	select {
	case err, ok := <-done:
		if ok {
			panic(err)
		}
	case <-time.After(timeout):
		defer func() {
			if err, ok := <-done; ok {
				panic(err)
			}
		}()
		onTimeout()
	}
}
