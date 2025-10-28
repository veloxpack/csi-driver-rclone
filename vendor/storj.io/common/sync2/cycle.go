// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information

package sync2

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"golang.org/x/sync/errgroup"
)

// Cycle implements a controllable recurring event.
//
// Cycle control methods PANICS after Close has been called and don't have any
// effect after Stop has been called.
//
// Start or Run (only one of them, not both) must be only called once.
type Cycle struct {
	noCopy noCopy //nolint:structcheck

	stopsent int32
	runexec  int32

	interval time.Duration

	ticker  *time.Ticker
	control chan any

	stopping chan struct{}
	stopped  chan struct{}

	init       sync.Once
	delayStart bool
}

type (
	// cycle control messages.
	cyclePause          struct{}
	cycleContinue       struct{}
	cycleChangeInterval struct{ Interval time.Duration }
	cycleTrigger        struct{ done chan struct{} }
)

// NewCycle creates a new cycle with the specified interval.
func NewCycle(interval time.Duration) *Cycle {
	cycle := &Cycle{}
	cycle.SetInterval(interval)
	return cycle
}

// SetInterval allows to change the interval before starting.
func (cycle *Cycle) SetInterval(interval time.Duration) {
	cycle.interval = interval
}

// SetDelayStart wait interval before first trigger on start/run.
func (cycle *Cycle) SetDelayStart() {
	cycle.delayStart = true
}

func (cycle *Cycle) initialize() {
	cycle.init.Do(func() {
		cycle.stopped = make(chan struct{})
		cycle.stopping = make(chan struct{})
		cycle.control = make(chan any)
	})
}

// Start runs the specified function with an errgroup.
func (cycle *Cycle) Start(ctx context.Context, group *errgroup.Group, fn func(ctx context.Context) error) {
	atomic.CompareAndSwapInt32(&cycle.runexec, 0, 1)
	group.Go(func() error {
		return cycle.Run(ctx, fn)
	})
}

// Run runs the specified in an interval.
//
// Every interval `fn` is started.
// When `fn` is not fast enough, it may skip some of those executions.
//
// Run PANICS if it's called after Stop has been called.
func (cycle *Cycle) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	atomic.CompareAndSwapInt32(&cycle.runexec, 0, 1)
	cycle.initialize()
	defer close(cycle.stopped)

	if cycle.Disabled() {
		return nil
	}
	currentInterval := cycle.interval
	cycle.ticker = time.NewTicker(currentInterval)
	defer cycle.ticker.Stop()

	choreCtx := monkit.ResetContextSpan(ctx)

	if !cycle.delayStart {
		if err := fn(choreCtx); err != nil {
			return err
		}
	}
	for {
		// prioritize stopping messages
		select {
		case <-cycle.stopping:
			return nil

		case <-ctx.Done():
			// handle control messages
			return ctx.Err()

		default:
		}

		// handle other messages as well
		select {

		case message := <-cycle.control:
			// handle control messages

			switch message := message.(type) {

			case cycleChangeInterval:
				currentInterval = message.Interval
				cycle.ticker.Stop()
				cycle.ticker = time.NewTicker(currentInterval)

			case cyclePause:
				cycle.ticker.Stop()
				// ensure we don't have ticks left
				select {
				case <-cycle.ticker.C:
				default:
				}

			case cycleContinue:
				cycle.ticker.Stop()
				cycle.ticker = time.NewTicker(currentInterval)

			case cycleTrigger:
				// trigger the function
				if err := fn(withManualTrigger(choreCtx)); err != nil {
					return err
				}
				if message.done != nil {
					close(message.done)
				}
			}

		case <-cycle.stopping:
			return nil

		case <-ctx.Done():
			// handle control messages
			return ctx.Err()

		case <-cycle.ticker.C:
			// trigger the function
			if err := fn(choreCtx); err != nil {
				return err
			}
		}
	}
}

// Close closes all resources associated with it.
//
// It MUST NOT be called concurrently.
func (cycle *Cycle) Close() {
	cycle.Stop()

	if atomic.LoadInt32(&cycle.runexec) == 1 {
		<-cycle.stopped
	}

	close(cycle.control)
}

// sendControl sends a control message.
func (cycle *Cycle) sendControl(message any) {
	cycle.initialize()
	select {
	case cycle.control <- message:
	case <-cycle.stopped:
	}
}

// Stop stops the cycle permanently.
func (cycle *Cycle) Stop() {
	cycle.initialize()
	if atomic.CompareAndSwapInt32(&cycle.stopsent, 0, 1) {
		close(cycle.stopping)
	}

	if atomic.LoadInt32(&cycle.runexec) == 1 {
		<-cycle.stopped
	}
}

// ChangeInterval allows to change the ticker interval after it has started.
// interval=-1 (disabled loop) is not allowed to be changed (it will panic).
func (cycle *Cycle) ChangeInterval(interval time.Duration) {
	if cycle.Disabled() {
		if interval != -1 {
			panic("Change interval on a disabled cycle is not supported")
		}
		return
	}
	if interval == -1 {
		panic("Cannot disable an already initialized cycle.")
	}
	cycle.sendControl(cycleChangeInterval{interval})
}

// Pause pauses the cycle.
func (cycle *Cycle) Pause() {
	if cycle.Disabled() {
		return
	}
	cycle.sendControl(cyclePause{})
}

// Restart restarts the ticker from 0.
func (cycle *Cycle) Restart() {
	if cycle.Disabled() {
		return
	}
	cycle.sendControl(cycleContinue{})
}

// Trigger ensures that the loop is done at least once. Note that it will not run if
// the cycle is disabled.
// TODO: Trigger should probably run the cycle even if the cycle interval is disabled.
// If it's currently running it waits for the previous to complete and then runs.
func (cycle *Cycle) Trigger() {
	if cycle.Disabled() {
		return
	}
	cycle.sendControl(cycleTrigger{})
}

// TriggerWait ensures that the loop is done at least once and waits for completion.
// Note that it will not run if the cycle is disabled.
// TODO: Trigger should probably run the cycle even if the cycle interval is disabled.
// If it's currently running it waits for the previous to complete and then runs.
func (cycle *Cycle) TriggerWait() {
	if cycle.Disabled() {
		return
	}
	done := make(chan struct{})

	cycle.sendControl(cycleTrigger{done})
	select {
	case <-done:
	case <-cycle.stopped:
	}
}

// Disabled means the cycle is not intended to run (interval is -1).
func (cycle *Cycle) Disabled() bool {
	return cycle.interval == -1
}

type cycleManualTriggerTag struct{}

func withManualTrigger(ctx context.Context) context.Context {
	return context.WithValue(ctx, cycleManualTriggerTag{}, true)
}

// IsManuallyTriggeredCycle returns whether ctx comes from a context
// that was started due to a `Trigger` or `TriggerWait` call in Cycle.
func IsManuallyTriggeredCycle(ctx context.Context) bool {
	val := ctx.Value(cycleManualTriggerTag{})
	if val == nil {
		return false
	}
	triggered, ok := val.(bool)
	return triggered && ok
}
