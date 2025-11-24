package seekable

import (
	"fmt"

	"github.com/a1ex3/zstd-seekable-format-go/pkg/env"
)

type wOption func(*writerImpl) error

func WithWEnvironment(e env.WEnvironment) wOption {
	return func(w *writerImpl) error { w.env = e; return nil }
}

type writeManyOptions struct {
	concurrency   int
	writeCallback func(uint32)
}

type WriteManyOption func(options *writeManyOptions) error

func WithConcurrency(concurrency int) WriteManyOption {
	return func(options *writeManyOptions) error {
		if concurrency < 1 {
			return fmt.Errorf("concurrency must be positive: %d", concurrency)
		}
		options.concurrency = concurrency
		return nil
	}
}

func WithWriteCallback(cb func(size uint32)) WriteManyOption {
	return func(options *writeManyOptions) error {
		options.writeCallback = cb
		return nil
	}
}
