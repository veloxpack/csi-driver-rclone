// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package sync2

import (
	"time"
)

// WithTimeout calls `do` and when the timeout is reached and `do`
// has not finished, it'll call `onTimeout` concurrently.
//
// When WithTimeout returns it's guaranteed to not call onTimeout.
func WithTimeout(timeout time.Duration, do, onTimeout func()) {
	c := make(chan struct{})
	t := time.AfterFunc(timeout, func() {
		defer close(c)
		onTimeout()
	})
	defer func() {
		if !t.Stop() {
			<-c
		}
	}()
	do()
}
