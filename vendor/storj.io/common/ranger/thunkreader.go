// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package ranger

import (
	"errors"
	"io"
	"sync"

	"github.com/zeebo/errs"
)

type thunkReadCloser struct {
	opts *ConcatOpts

	// only modified/accessed by Read
	currentBytesLeft int64
	currentReader    io.Reader
	next             *thunk[sizedReadCloser]

	// Modified/accessed by Read/Close
	mtx           sync.Mutex
	remaining     []*thunk[sizedReadCloser]
	currentCloser io.Closer
}

func newThunkReadCloser(opts *ConcatOpts, thunks []*thunk[sizedReadCloser]) (*thunkReadCloser, error) {
	t := thunkReadCloser{opts: opts}
	if len(thunks) > 0 {
		resp, err := thunks[0].Result()
		if err != nil {
			return nil, err
		}
		t.currentReader = resp.ReadCloser
		t.currentCloser = resp.ReadCloser
		t.currentBytesLeft = resp.Size

		t.remaining = thunks[1:]
		if len(t.remaining) > 0 {
			t.next = t.remaining[0]
		}
	}
	return &t, nil
}

func (t *thunkReadCloser) Read(p []byte) (n int, err error) {
	current := t.currentReader
	if current == nil {
		return 0, io.EOF
	}

	n, err = current.Read(p)
	t.currentBytesLeft -= int64(n)

	next := t.next
	if t.currentBytesLeft < t.opts.PrefetchWhenBytesRemaining && next != nil {
		next.Trigger()
	}

	if errors.Is(err, io.EOF) {
		err = t.advance()
	}
	return n, err
}

func (t *thunkReadCloser) advance() error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if t.currentCloser == nil {
		return errs.New("already closed")
	}

	var eg errs.Group

	eg.Add(t.currentCloser.Close())
	t.currentReader = nil
	t.currentCloser = nil
	t.currentBytesLeft = 0
	t.next = nil

	if len(t.remaining) > 0 {
		next, err := t.remaining[0].Result()
		t.remaining = t.remaining[1:]
		if err != nil {
			eg.Add(err)
		} else {
			t.currentReader = next.ReadCloser
			t.currentCloser = next.ReadCloser
			t.currentBytesLeft = next.Size
		}
		if len(t.remaining) > 0 {
			t.next = t.remaining[0]
		}
	}

	return eg.Err()
}

func (t *thunkReadCloser) Close() error {
	t.mtx.Lock()
	currentCloser, remaining := t.currentCloser, t.remaining
	t.currentCloser, t.remaining = nil, nil
	t.mtx.Unlock()
	var eg errs.Group
	if currentCloser != nil {
		eg.Add(currentCloser.Close())
	}
	for _, next := range remaining {
		eg.Add(next.Close())
	}
	return eg.Err()
}
