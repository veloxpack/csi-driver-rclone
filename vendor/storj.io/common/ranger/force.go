// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package ranger

import (
	"io"
)

type forcedReader struct {
	buf []byte
	err error
	r   io.ReadCloser
}

func (r *forcedReader) Close() error {
	return r.r.Close()
}

func (r *forcedReader) Read(p []byte) (n int, err error) {
	if len(r.buf) > 0 {
		n = copy(p, r.buf)
		r.buf = r.buf[n:]
		if len(r.buf) == 0 {
			r.buf = nil
		}
		return n, nil
	}
	if r.err != nil {
		return 0, r.err
	}
	return r.r.Read(p)
}

func forceReads(r io.ReadCloser) io.ReadCloser {
	var buf [4096]byte
	n, err := r.Read(buf[:])
	return &forcedReader{
		buf: buf[:n],
		err: err,
		r:   r,
	}
}
