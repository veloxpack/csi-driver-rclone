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
		// On abort, signal the buffer before acquiring e.mu: a producer may
		// be blocked inside Write holding e.mu while waiting for the reader
		// to drain the cursor, and would otherwise never release it.
		e.sbuf.DoneWriting(err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// On normal completion this flushes the final encrypted block and
	// signals with the flush result. On abort the signal already happened
	// above and this call is discarded by the buffer.
	e.sbuf.DoneWriting(e.wrc.Close())
}
