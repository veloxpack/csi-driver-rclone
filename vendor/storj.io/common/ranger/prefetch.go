// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package ranger

import (
	"context"
	"io"
)

// ConcatOpts specifies a couple of concatenation options.
type ConcatOpts struct {
	// Prefetch, when true, will support prefetching the next range. Prefetching
	// won't be very useful without a non-zero PrefetchWhenBytesRemaining value.
	Prefetch bool

	// ForceReads only matters if Prefetch is true. If true, not only will the
	// next range be prefetched, the first few bytes will be also.
	ForceReads bool

	// PrefetchWhenBytesRemaining specifies how many bytes should be remaining
	// at most before prefetching the next bit. Prefetch must be true.
	PrefetchWhenBytesRemaining int64
}

// ConcatWithOpts concatenates Rangers with support for prefetching the next
// range if specified.
func ConcatWithOpts(opts ConcatOpts, r ...Ranger) Ranger {
	if opts.Prefetch {
		return Concat(r...)
	}
	return newPrefetchConcatReader(&opts, r...)
}

type prefetchConcatReader struct {
	opts *ConcatOpts
	size int64

	leaf  Ranger
	left  *prefetchConcatReader
	right *prefetchConcatReader
}

func newPrefetchConcatReader(opts *ConcatOpts, r ...Ranger) *prefetchConcatReader {
	switch len(r) {
	case 0:
		return &prefetchConcatReader{
			opts: opts,
			size: 0,
			leaf: ByteRanger(nil),
		}
	case 1:
		return &prefetchConcatReader{
			opts: opts,
			size: r[0].Size(),
			leaf: r[0],
		}
	default:
		mid := len(r) / 2
		rv := &prefetchConcatReader{
			opts:  opts,
			left:  newPrefetchConcatReader(opts, r[:mid]...),
			right: newPrefetchConcatReader(opts, r[mid:]...),
		}
		rv.size = rv.left.Size() + rv.right.Size()
		return rv
	}
}

func (c *prefetchConcatReader) Size() (s int64) {
	return c.size
}

type sizedReadCloser struct {
	io.ReadCloser
	Size int64
}

func (c *prefetchConcatReader) rangeThunks(ctx context.Context, offset, length int64, out []*thunk[sizedReadCloser]) (_ []*thunk[sizedReadCloser]) {
	if c.leaf != nil {
		return append(out, newThunk(func() (sizedReadCloser, error) {
			r, err := c.leaf.Range(ctx, offset, length)
			if err == nil && c.opts.ForceReads {
				r = forceReads(r)
			}
			return sizedReadCloser{ReadCloser: r, Size: length}, err
		}))
	}

	leftSize := c.left.Size()
	if offset+length <= leftSize {
		return c.left.rangeThunks(ctx, offset, length, out)
	}
	if offset >= leftSize {
		return c.right.rangeThunks(ctx, offset-leftSize, length, out)
	}

	out = c.left.rangeThunks(ctx, offset, leftSize-offset, out)
	return c.right.rangeThunks(ctx, 0, length-(leftSize-offset), out)
}

func (c *prefetchConcatReader) Range(ctx context.Context, offset, length int64) (_ io.ReadCloser, err error) {
	defer mon.Task()(&ctx)(&err)
	return newThunkReadCloser(c.opts, c.rangeThunks(ctx, offset, length, nil))
}
