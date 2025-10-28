package timer

import (
	"sync"
	"time"
)

type Run struct {
	Start  time.Time
	Finish time.Time
}

type Runs []Run

type Timer struct {
	Runs
	*sync.RWMutex
}

func New() *Timer {
	return &Timer{RWMutex: &sync.RWMutex{}}
}

func (t *Timer) Start() time.Time {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()

	n := time.Now()
	t.Runs = append(t.Runs, Run{Start: n})
	return n
}

func (t *Timer) Stop() time.Time {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()

	n := time.Now()
	t.Runs[len(t.Runs)-1].Finish = n
	return n
}

func (t *Timer) Elapsed() time.Duration {
	var elapsed time.Duration
	t.RWMutex.RLock()
	for _, r := range t.Runs {
		var finish time.Time
		if r.Finish.IsZero() {
			finish = time.Now()
		} else {
			finish = r.Finish
		}
		elapsed += finish.Sub(r.Start)
	}

	t.RWMutex.RUnlock()

	return elapsed
}

func (t *Timer) LastStart() time.Time {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	if len(t.Runs) == 0 {
		return time.Time{}
	}
	return t.Runs[len(t.Runs)-1].Start
}

func (t *Timer) Running() bool {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	if len(t.Runs) == 0 {
		return false
	}
	return t.Runs[len(t.Runs)-1].Finish.IsZero()
}

func (t *Timer) Finished() bool {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	if len(t.Runs) == 0 {
		return false
	}
	return !t.Runs[len(t.Runs)-1].Finish.IsZero()
}

func (t *Timer) Started() bool {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	if len(t.Runs) == 0 {
		return false
	}
	return !t.Runs[0].Start.IsZero()
}

func (t *Timer) StartTime() time.Time {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	return t.Runs[0].Start
}

func (t *Timer) FinishTime() time.Time {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	return t.Runs[0].Finish
}
