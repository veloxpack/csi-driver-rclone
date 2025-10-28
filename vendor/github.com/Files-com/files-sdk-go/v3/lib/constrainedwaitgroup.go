package lib

import (
	"context"
	"sync"
	"sync/atomic"
)

type ConstrainedWorkGroup struct {
	wg   sync.WaitGroup
	sem  chan struct{}
	cond *sync.Cond
}

func NewConstrainedWorkGroup(maxConcurrency int) *ConstrainedWorkGroup {
	return &ConstrainedWorkGroup{
		sem:  make(chan struct{}, maxConcurrency),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (cw *ConstrainedWorkGroup) Wait() {
	cw.wg.Add(1)
	cw.sem <- struct{}{}
}

func (cw *ConstrainedWorkGroup) Done() {
	cw.cond.L.Lock()
	defer cw.cond.L.Unlock()
	cw.wg.Done()
	<-cw.sem
	cw.cond.Signal()
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
		cw:   cw,
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (cw *ConstrainedWorkGroup) WaitForADone() bool {
	cw.cond.L.Lock()
	defer cw.cond.L.Unlock()
	if cw.RunningCount() == 0 {
		return false
	}

	cw.cond.Wait()
	return true
}

type SubWorker struct {
	cw           *ConstrainedWorkGroup
	wg           sync.WaitGroup
	runningCount int32
	cond         *sync.Cond
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
	sw.cond.L.Lock()
	defer sw.cond.L.Unlock()
	atomic.AddInt32(&sw.runningCount, -1)
	sw.wg.Done()
	sw.cw.Done()
	sw.cond.Signal()
}

func (sw *SubWorker) WaitAllDone() {
	sw.wg.Wait()
}

// WaitForADone Blocks until at least one goroutine has completed.
func (sw *SubWorker) WaitForADone() bool {
	sw.cond.L.Lock()
	defer sw.cond.L.Unlock()
	if sw.RunningCount() == 0 {
		return false
	}
	sw.cond.Wait()
	return true
}

func (sw *SubWorker) RunningCount() int {
	return int(atomic.LoadInt32(&sw.runningCount))
}
