// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package ranger

import (
	"io"
	"sync"

	"github.com/zeebo/errs"
)

type thunkResponse[T io.Closer] struct {
	result   T
	err      error
	panicVal any
}

// A thunk represents some amount of background work that will create
// a T type.
type thunk[T io.Closer] struct {
	triggerOnce sync.Once
	mtx         sync.Mutex
	work        func() (T, error)
	ch          chan thunkResponse[T]
}

// newThunk makes a thunk that calls work to generate a T type. work is not
// called until either Trigger or Result is called.
func newThunk[T io.Closer](work func() (T, error)) *thunk[T] {
	return &thunk[T]{
		work: work,
	}
}

// Trigger initiates the work, if it hasn't already been initiated.
func (t *thunk[T]) Trigger() {
	t.triggerOnce.Do(func() {
		t.mtx.Lock()
		defer t.mtx.Unlock()
		t.trigger()
	})
}

func (t *thunk[T]) trigger() {
	work := t.work
	t.work = nil
	if work == nil {
		return
	}

	ch := make(chan thunkResponse[T], 1)
	t.ch = ch

	go func() {
		res, panicVal, err := func() (val T, panicVal any, err error) {
			defer func() {
				panicVal = recover()
			}()
			val, err = work()
			return
		}()

		ch <- thunkResponse[T]{
			result:   res,
			err:      err,
			panicVal: panicVal,
		}
	}()
}

// Result waits until work completes, triggering work if necessary. Close will
// not interrupt Result. If you need to cancel the work, the work should be
// cancelable externally.
func (t *thunk[T]) Result() (rv T, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.trigger()

	ch := t.ch
	t.ch = nil
	if ch == nil {
		return rv, errs.New("Result called with no work left to do")
	}

	resp := <-ch
	if resp.panicVal != nil {
		panic(resp.panicVal)
	}
	return resp.result, resp.err
}

// Close shuts down the thunk. If work hasn't happened yet, Close prevents
// work from happening. If work has started but not finished, close waits
// for that work and closes the result type. If work has already finished and
// Result has already been called, Close does nothing. If Result is in the
// process of being called, Close will not interrupt Result and will wait for
// Result to finish.
func (t *thunk[T]) Close() error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if t.work != nil {
		t.work = nil
		return nil
	}

	ch := t.ch
	t.ch = nil
	if ch == nil {
		return nil
	}

	resp := <-ch
	if resp.panicVal != nil {
		panic(resp.panicVal)
	}
	if resp.err != nil {
		return resp.err
	}
	return resp.result.Close()
}
