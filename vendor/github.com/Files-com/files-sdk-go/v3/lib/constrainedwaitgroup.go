package lib

import (
	"context"
	"sync"
	"sync/atomic"
)

type ConstrainedWorkGroup struct {
	wg     sync.WaitGroup
	sem    chan struct{}
	mu     sync.Mutex
	notify chan struct{}
}

func NewConstrainedWorkGroup(maxConcurrency int) *ConstrainedWorkGroup {
	return &ConstrainedWorkGroup{
		sem:    make(chan struct{}, maxConcurrency),
		notify: make(chan struct{}),
	}
}

func (cw *ConstrainedWorkGroup) Wait() {
	cw.wg.Add(1)
	cw.sem <- struct{}{}
}

func (cw *ConstrainedWorkGroup) Done() {
	cw.wg.Done()
	<-cw.sem
	cw.mu.Lock()
	cw.signalDoneLocked()
	cw.mu.Unlock()
}

func (cw *ConstrainedWorkGroup) WaitAllDone() {
	cw.wg.Wait()
}

func (cw *ConstrainedWorkGroup) RunningCount() int {
	return len(cw.sem)
}

func (cw *ConstrainedWorkGroup) Max() int {
	return cap(cw.sem)
}

func (cw *ConstrainedWorkGroup) WaitWithContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case cw.sem <- struct{}{}:
		cw.wg.Add(1)
		return true
	}
}

func (cw *ConstrainedWorkGroup) RemainingCapacity() int {
	return cap(cw.sem) - len(cw.sem)
}

func (cw *ConstrainedWorkGroup) NewSubWorker() ConcurrencyManager {
	return &SubWorker{
		cw:     cw,
		notify: make(chan struct{}),
	}
}

func (cw *ConstrainedWorkGroup) WaitForADone() bool {
	return cw.WaitForADoneWithContext(context.Background())
}

func (cw *ConstrainedWorkGroup) WaitForADoneWithContext(ctx context.Context) bool {
	cw.mu.Lock()
	if cw.RunningCount() == 0 {
		cw.mu.Unlock()
		return false
	}
	notify := cw.notify
	cw.mu.Unlock()

	select {
	case <-ctx.Done():
		return false
	case <-notify:
		return true
	}
}

func (cw *ConstrainedWorkGroup) signalDoneLocked() {
	close(cw.notify)
	cw.notify = make(chan struct{})
}

type SubWorker struct {
	cw           *ConstrainedWorkGroup
	wg           sync.WaitGroup
	runningCount int32
	mu           sync.Mutex
	notify       chan struct{}
}

func (sw *SubWorker) Wait() {
	sw.cw.Wait()
	sw.wg.Add(1)
	atomic.AddInt32(&sw.runningCount, 1)
}

func (sw *SubWorker) WaitWithContext(ctx context.Context) bool {
	if sw.cw.WaitWithContext(ctx) {
		sw.wg.Add(1)
		atomic.AddInt32(&sw.runningCount, 1)
		return true
	}
	return false
}

func (sw *SubWorker) Done() {
	sw.mu.Lock()
	atomic.AddInt32(&sw.runningCount, -1)
	sw.wg.Done()
	sw.signalDoneLocked()
	sw.mu.Unlock()
	sw.cw.Done()
}

func (sw *SubWorker) WaitAllDone() {
	sw.wg.Wait()
}

// WaitForADone Blocks until at least one goroutine has completed.
func (sw *SubWorker) WaitForADone() bool {
	return sw.WaitForADoneWithContext(context.Background())
}

func (sw *SubWorker) WaitForADoneWithContext(ctx context.Context) bool {
	sw.mu.Lock()
	if sw.RunningCount() == 0 {
		sw.mu.Unlock()
		return false
	}
	notify := sw.notify
	sw.mu.Unlock()

	select {
	case <-ctx.Done():
		return false
	case <-notify:
		return true
	}
}

func (sw *SubWorker) RunningCount() int {
	return int(atomic.LoadInt32(&sw.runningCount))
}

func (sw *SubWorker) signalDoneLocked() {
	close(sw.notify)
	sw.notify = make(chan struct{})
}
