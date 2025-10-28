package lib

import (
	"context"
	"io"
)

type readerAtCtx struct {
	ctx context.Context
	r   io.ReaderAt
}

type readerCtx struct {
	ctx context.Context
	io.ReadCloser
}

type WithContextReaderAt interface {
	WithContext(context.Context) interface{}
	io.ReaderAt
}

type WithContextReader interface {
	WithContext(context.Context) interface{}
	io.Reader
}

type ReaderAtCloser interface {
	io.ReaderAt
	io.ReadCloser
}

func (r *readerAtCtx) ReadAt(p []byte, off int64) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	withContext, ok := r.r.(WithContextReader)
	if ok {
		r.r = withContext.WithContext(r.ctx).(io.ReaderAt)
	}
	return r.r.ReadAt(p, off)
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	withContext, ok := r.ReadCloser.(WithContextReader)
	if ok {
		r.ReadCloser = withContext.WithContext(r.ctx).(io.ReadCloser)
	}
	return r.ReadCloser.Read(p)
}

// NewReader gets a context-aware io.Reader.
func NewReaderAt(ctx context.Context, r io.ReaderAt) io.ReaderAt {
	return &readerAtCtx{ctx: ctx, r: r}
}

// NewReader gets a context-aware io.Reader.
func NewReader(ctx context.Context, r io.ReadCloser) io.ReadCloser {
	return &readerCtx{ctx: ctx, ReadCloser: r}
}
