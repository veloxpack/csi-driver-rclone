// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package pieceupload

import (
	"context"
	"sync/atomic"
	"time"

	"storj.io/common/sync2"
	"storj.io/common/time2"
)

// StallManager is a type that helps manage stalls with backoff logic.
type StallManager struct {
	maxDurationFence sync2.Fence
	maxDuration      atomic.Int64
}

// NewStallManager makes a new stall manager.
func NewStallManager() *StallManager {
	return &StallManager{}
}

// SetMaxDuration sets the max duration any child context is allowed to go. Can only be set once per upload.
func (manager *StallManager) SetMaxDuration(duration time.Duration) {
	if duration > 0 {
		// Only set if not already set
		if manager.maxDuration.CompareAndSwap(0, int64(duration)) {
			manager.maxDurationFence.Release()
		}
	}
}

// Watch makes sure the current context doesn't go longer than SetMaxDuration.
// If no SetMaxDuration is set, it waits for the fence to be released when SetMaxDuration is called later.
// cleanup should be called when the context is no longer needed.
func (manager *StallManager) Watch(parent context.Context) (ctx context.Context, cleanup func()) {
	if duration := time.Duration(manager.maxDuration.Load()); duration > 0 {
		return context.WithTimeout(parent, duration)
	}

	start := time2.Now(parent)
	ctx, cancel := context.WithCancel(parent)

	go func() {
		defer cancel()

		// Wait for the max duration fence to be set, or for the context to be cancelled.
		if !manager.maxDurationFence.Wait(ctx) {
			// The context is cancelled before the max duration is set, we just return.
			return
		}

		// Determine the remaining time to wait.
		duration := time.Duration(manager.maxDuration.Load())
		elapsed := time2.Since(parent, start)
		if elapsed >= duration {
			return
		}
		remaining := duration - elapsed

		// Wait for the remaining time or until the context is cancelled.
		time2.Sleep(ctx, remaining)
	}()

	return ctx, cancel
}
