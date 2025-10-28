package lib

import "io"

type WriterAndAt interface {
	io.WriterAt
	io.Writer
	io.Closer
}

type ProgressWriter struct {
	WriterAndAt
	ProgressWatcher func(int64)
}

func (w ProgressWriter) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = w.WriterAndAt.WriteAt(p, off)

	w.ProgressWatcher(int64(n))
	return n, err
}

func (w ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = w.WriterAndAt.Write(p)

	w.ProgressWatcher(int64(n))
	return n, err
}

func (w ProgressWriter) Close() error {
	return w.WriterAndAt.Close()
}
