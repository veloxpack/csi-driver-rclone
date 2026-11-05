// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package splitter

import (
	"io"
	"sync"

	"storj.io/uplink/private/storage/streams/buffer"
)

type encryptedBuffer struct {
	sbuf *buffer.Buffer
	wrc  io.WriteCloser

	mu    sync.Mutex
	plain int64
}

func newEncryptedBuffer(sbuf *buffer.Buffer, wrc io.WriteCloser) *encryptedBuffer {
	return &encryptedBuffer{
		sbuf: sbuf,
		wrc:  wrc,
	}
}

func (e *encryptedBuffer) Reader() io.Reader     { return e.sbuf.Reader() }
func (e *encryptedBuffer) DoneReading(err error) { e.sbuf.DoneReading(err) }

func (e *encryptedBuffer) Write(p []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	n, err := e.wrc.Write(p)
	e.plain += int64(n)
	return n, err
}

func (e *encryptedBuffer) PlainSize() int64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.plain
}

func (e *encryptedBuffer) DoneWriting(err error) {
	if err != nil {
		// Abort path: a producer may be blocked inside Write holding e.mu
		// while waiting for the reader to drain the cursor. Signal the
		// buffer up front so Cursor.WaitWrite returns (the loop's
		// doneWriting case), letting the producer release e.mu. Without
		// this, the e.mu.Lock below deadlocks whenever the upload is
		// aborted while the producer is pushing into a backpressured
		// segment buffer.
		e.sbuf.DoneWriting(err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	cerr := e.wrc.Close()
	if err == nil {
		// On the abort path e.sbuf.DoneWriting(err) was already called
		// above to unblock the producer; cursor.DoneWriting is idempotent
		// and would discard cerr anyway, so only signal here on the normal
		// completion path where cerr carries the final flush result.
		e.sbuf.DoneWriting(cerr)
	}
}
