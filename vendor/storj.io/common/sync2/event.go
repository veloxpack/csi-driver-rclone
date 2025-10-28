// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information

package sync2

import (
	"context"
	"sync"
)

// Event allows to signal another goroutine of something happening. This
// primitive is useful for signaling a single goroutine to update it's
// internal state.
//
// An Event doesn't need initialization.
// An Event must not be copied after first use.
type Event struct {
	noCopy noCopy //nolint:structcheck

	setup    sync.Once
	signaled chan struct{}
}

// init sets up the initial lock into wait.
func (event *Event) init() {
	event.setup.Do(func() {
		event.signaled = make(chan struct{}, 1)
	})
}

// Signal signals once. Signal guarantees that at least one goroutine is
// released from Wait or the next call to Wait. Multiple signals may be
// coalesced into a single wait release. In other words N signals results in
// 1 to N releases from [Wait] or [Signaled].
func (event *Event) Signal() {
	event.init()
	select {
	case event.signaled <- struct{}{}:
	default:
	}
}

// Wait waits for a signal. Only one goroutine should call [Wait] or
// [Signaled]. The implementation allows concurrent calls, however the exact
// behaviour is hard to reason about.
//
// Returns true when it was not related to context cancellation.
func (event *Event) Wait(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}

	event.init()

	select {
	case <-ctx.Done():
		return false
	case <-event.signaled:
		return true
	}
}

// Signaled returns channel that is notified when a signal happens. Only one
// goroutine should call `Wait` or `Signaled`. The implementation allows
// concurrent calls, however the exact behaviour is hard to reason about.
func (event *Event) Signaled() chan struct{} {
	event.init()
	return event.signaled
}
