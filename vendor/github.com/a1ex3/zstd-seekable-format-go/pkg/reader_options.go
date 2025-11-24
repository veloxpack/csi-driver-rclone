package seekable

import (
	"github.com/a1ex3/zstd-seekable-format-go/pkg/env"
)

type rOption func(*readerImpl) error

func WithREnvironment(e env.REnvironment) rOption {
	return func(r *readerImpl) error { r.env = e; return nil }
}
