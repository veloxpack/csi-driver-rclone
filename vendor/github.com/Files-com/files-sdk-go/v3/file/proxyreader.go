package file

import (
	"io"
	"sync/atomic"
)

type ProxyReader interface {
	io.ReadCloser
	Len() int
	BytesRead() int64
	Rewind() bool
}

type ProxyReaderAt struct {
	io.ReaderAt
	off    int64
	len    int64
	onRead func(i int64)
	read   int64
	closed atomic.Bool
}

type ProxyRead struct {
	io.Reader
	len    int64
	onRead func(i int64)
	read   int64
	closed atomic.Bool
}

func (x *ProxyReaderAt) Rewind() bool {
	x.onRead(-x.BytesRead())
	atomic.StoreInt64(&x.read, 0)
	return true
}

func (x *ProxyRead) Rewind() bool {
	return atomic.LoadInt64(&x.read) == 0
}

func (x *ProxyReaderAt) Len() int {
	return int(x.len)
}

func (x *ProxyRead) Len() int {
	return int(x.len)
}

func (x *ProxyReaderAt) BytesRead() int64 {
	return atomic.LoadInt64(&x.read)
}

func (x *ProxyRead) BytesRead() int64 {
	return atomic.LoadInt64(&x.read)
}

func (x *ProxyReaderAt) Seek(offset int64, whence int) (int64, error) {
	x.onRead(-(x.BytesRead() - offset)) // rewind progress
	atomic.StoreInt64(&x.read, offset)
	return offset, nil
}

func (x *ProxyReaderAt) Read(p []byte) (int, error) {
	if x.closed.Load() {
		x.onRead(-x.BytesRead()) // rewind progress
		atomic.StoreInt64(&x.read, 0)
		x.closed.Store(false)
	}

	if x.BytesRead() == x.len {
		return 0, io.EOF
	}
	var n int
	var err error
	if int64(len(p)) > x.len-x.BytesRead() {
		n, err = x.ReadAt(p[:min(x.len-x.BytesRead(), int64(len(p)))], x.off+x.BytesRead())
	} else {
		n, err = x.ReadAt(p, x.off+x.BytesRead())
	}

	if err != nil {
		return n, err
	}

	atomic.AddInt64(&x.read, int64(n))
	if x.onRead != nil {
		x.onRead(int64(n))
	}
	return n, nil
}

func (x *ProxyRead) Read(p []byte) (int, error) {
	if x.BytesRead() == x.len || x.closed.Load() {
		return 0, io.EOF
	}

	n, err := x.Reader.Read(p[:min(x.len-x.BytesRead(), int64(len(p)))])

	if err != nil {
		return n, err
	}

	atomic.AddInt64(&x.read, int64(n))
	if x.onRead != nil {
		x.onRead(int64(n))
	}

	return n, err
}

func (x *ProxyReaderAt) Close() error {
	x.closed.Store(true)
	return nil
}

func (x *ProxyRead) Close() error {
	x.closed.Store(true)
	return nil
}
