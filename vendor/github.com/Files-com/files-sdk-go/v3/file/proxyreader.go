package file

import (
	"io"
	"sync/atomic"
	"time"
)

type ProxyReader interface {
	io.ReadCloser
	Len() int
	BytesRead() int64
	ReadDuration() time.Duration
	Rewind() bool
}

type ProxyReaderAt struct {
	io.ReaderAt
	off               int64
	len               int64
	onRead            func(i int64)
	read              int64
	readDurationNanos int64
	trackReadDuration bool
	closed            atomic.Bool
}

type ProxySectionReader struct {
	section           io.SectionReader
	len               int64
	onRead            func(i int64)
	read              int64
	readDurationNanos int64
	trackReadDuration bool
	closed            bool
}

type ProxyRead struct {
	io.Reader
	len               int64
	onRead            func(i int64)
	read              int64
	readDurationNanos int64
	trackReadDuration bool
	closed            atomic.Bool
}

func newProxySectionReader(readerAt io.ReaderAt, off int64, length int64, onRead func(int64), trackReadDuration bool) *ProxySectionReader {
	section := io.NewSectionReader(readerAt, off, length)
	return &ProxySectionReader{
		section:           *section,
		len:               length,
		onRead:            onRead,
		trackReadDuration: trackReadDuration,
	}
}

func (x *ProxyReaderAt) Rewind() bool {
	if read := x.BytesRead(); read != 0 && x.onRead != nil {
		x.onRead(-read)
	}
	atomic.StoreInt64(&x.read, 0)
	atomic.StoreInt64(&x.readDurationNanos, 0)
	return true
}

func (x *ProxySectionReader) Rewind() bool {
	if read := x.BytesRead(); read != 0 && x.onRead != nil {
		x.onRead(-read)
	}
	if _, err := x.section.Seek(0, io.SeekStart); err != nil {
		return false
	}
	x.read = 0
	x.readDurationNanos = 0
	x.closed = false
	return true
}

func (x *ProxyRead) Rewind() bool {
	ok := atomic.LoadInt64(&x.read) == 0
	if ok {
		atomic.StoreInt64(&x.readDurationNanos, 0)
	}
	return ok
}

func (x *ProxyReaderAt) Len() int {
	return int(x.len)
}

func (x *ProxyRead) Len() int {
	return int(x.len)
}

func (x *ProxySectionReader) Len() int {
	return int(x.len)
}

func (x *ProxyReaderAt) BytesRead() int64 {
	return atomic.LoadInt64(&x.read)
}

func (x *ProxyRead) BytesRead() int64 {
	return atomic.LoadInt64(&x.read)
}

func (x *ProxySectionReader) BytesRead() int64 {
	return x.read
}

func (x *ProxyReaderAt) ReadDuration() time.Duration {
	return time.Duration(atomic.LoadInt64(&x.readDurationNanos))
}

func (x *ProxyRead) ReadDuration() time.Duration {
	return time.Duration(atomic.LoadInt64(&x.readDurationNanos))
}

func (x *ProxySectionReader) ReadDuration() time.Duration {
	return time.Duration(x.readDurationNanos)
}

func (x *ProxyReaderAt) Seek(offset int64, whence int) (int64, error) {
	if x.onRead != nil {
		x.onRead(-(x.BytesRead() - offset)) // rewind progress
	}
	atomic.StoreInt64(&x.read, offset)
	return offset, nil
}

func (x *ProxyReaderAt) Read(p []byte) (int, error) {
	if x.closed.Load() {
		if read := x.BytesRead(); read != 0 && x.onRead != nil {
			x.onRead(-read) // rewind progress
		}
		atomic.StoreInt64(&x.read, 0)
		atomic.StoreInt64(&x.readDurationNanos, 0)
		x.closed.Store(false)
	}

	read := atomic.LoadInt64(&x.read)
	if read >= x.len {
		return 0, io.EOF
	}
	var n int
	var err error
	remaining := x.len - read
	var start time.Time
	if x.trackReadDuration {
		start = time.Now()
	}
	if int64(len(p)) > remaining {
		n, err = x.ReadAt(p[:min(remaining, int64(len(p)))], x.off+read)
	} else {
		n, err = x.ReadAt(p, x.off+read)
	}
	if x.trackReadDuration {
		atomic.AddInt64(&x.readDurationNanos, time.Since(start).Nanoseconds())
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

func (x *ProxySectionReader) Read(p []byte) (int, error) {
	if x.closed {
		if read := x.BytesRead(); read != 0 && x.onRead != nil {
			x.onRead(-read) // rewind progress
		}
		if _, err := x.section.Seek(0, io.SeekStart); err != nil {
			return 0, err
		}
		x.read = 0
		x.readDurationNanos = 0
		x.closed = false
	}

	var start time.Time
	if x.trackReadDuration {
		start = time.Now()
	}
	n, err := x.section.Read(p)
	if x.trackReadDuration {
		x.readDurationNanos += time.Since(start).Nanoseconds()
	}
	if n > 0 {
		x.read += int64(n)
		if x.onRead != nil {
			x.onRead(int64(n))
		}
	}
	return n, err
}

func (x *ProxyRead) Read(p []byte) (int, error) {
	read := atomic.LoadInt64(&x.read)
	if read >= x.len || x.closed.Load() {
		return 0, io.EOF
	}

	var start time.Time
	if x.trackReadDuration {
		start = time.Now()
	}
	n, err := x.Reader.Read(p[:min(x.len-read, int64(len(p)))])
	if x.trackReadDuration {
		atomic.AddInt64(&x.readDurationNanos, time.Since(start).Nanoseconds())
	}

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

func (x *ProxySectionReader) Close() error {
	x.closed = true
	return nil
}

func (x *ProxyRead) Close() error {
	x.closed.Store(true)
	return nil
}
