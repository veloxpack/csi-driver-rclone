package file

import (
	"context"
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/samber/lo"
)

type OffSet struct {
	off int64
	len int64
}

type Part struct {
	OffSet
	files_sdk.EtagsParam
	bytes    int64
	requests []time.Time
	error
	number     int
	processing bool
	context.Context
	context.CancelFunc
	*sync.RWMutex
	final bool
	files_sdk.FileUploadPart
	ProxyReader
}

func (p *Part) Done() *Part {
	p.processing = false
	return p
}

func (p *Part) Start(ctx ...context.Context) *Part {
	p.Touch()
	p.processing = true
	if len(ctx) == 1 {
		p.WithContext(ctx[0])
	}
	return p
}

func (p *Part) WithContext(ctx context.Context) *Part {
	p.RWMutex = &sync.RWMutex{}
	p.Context, p.CancelFunc = context.WithCancel(ctx)
	return p
}

func (p *Part) Touch() {
	p.requests = append(p.requests, time.Now())
}

func (p *Part) Successful() bool {
	return p.bytes == p.len && p.error == nil
}

func (p *Part) Clear() {
	p.bytes = 0
	p.error = nil
}

func (p *Part) SetError(err error) {
	p.Lock()
	defer p.Unlock()
	p.error = err
}

func (p *Part) Err() error {
	p.RLock()
	defer p.RUnlock()
	if p.error != nil {
		return p.error
	}

	return p.Context.Err()
}

type Parts []*Part

func (p Parts) SuccessfulBytes() (b int64) {
	for _, part := range p {
		if part.Successful() {
			b += part.bytes
		}
	}

	return b
}

type ByteOffset struct {
	PartSizes         []int64
	OverrideChunkSize int64
}

func (b ByteOffset) BySize(size *int64) Iterator {
	return b.Resume(size, 0, 0)
}

// Resume creates an iterator that generates file chunks starting from a given offset.
// When OverrideChunkSize is set (or computed for large files), use a constant-size iterator.
// Otherwise, use the PartSizes-driven iterator.
func (b ByteOffset) Resume(size *int64, off int64, index int) Iterator {
	// Compute dynamic override for large files (>10) when not explicitly set.
	// OK to assign on value receiver: only the closure needs this adjusted value.
	if b.OverrideChunkSize == 0 && size != nil && *size > 10*1024*1024 {
		b.OverrideChunkSize = lo.Clamp((*size+9999)/10000, 10*1024*1024, 525*1024*1024)
	}

	// Iterator for constant-size chunks (override mode).
	// Ignores PartSizes and advances by a fixed size until end (when size known).
	if b.OverrideChunkSize > 0 {
		return func() (OffSet, Iterator, int) {
			// If end reached (only meaningful with known size), stop.
			if size != nil && off >= *size {
				return OffSet{}, nil, index
			}

			partSize := b.OverrideChunkSize
			// Cap to remaining bytes when the file size is known.
			if size != nil {
				remaining := *size - off
				if remaining <= 0 {
					return OffSet{}, nil, index
				}
				if partSize > remaining {
					partSize = remaining
				}
			}

			offset := OffSet{off: off, len: partSize}
			nextOff := off + partSize

			// Stop when we hit the end; otherwise continue with the same index (override mode).
			if size != nil && nextOff >= *size {
				return offset, nil, index
			}
			// Increment index to track chunk number for external use (logging, progress, etc.)
			return offset, b.Resume(size, nextOff, index+1), index
		}
	}

	// Iterator driven by PartSizes (no override).
	// Uses slice entries to define each chunk, advancing index per part.
	return func() (OffSet, Iterator, int) {
		// If size is known and we're done, stop.
		if size != nil && off >= *size {
			return OffSet{}, nil, index
		}

		// If no size is known, PartSizes must define iteration; stop when exhausted.
		if len(b.PartSizes) == 0 || index >= len(b.PartSizes) {
			return OffSet{}, nil, index
		}

		partSize := b.PartSizes[index]
		// Cap to remaining bytes when file size is known.
		if size != nil {
			remaining := *size - off
			if remaining <= 0 {
				return OffSet{}, nil, index
			}
			if partSize > remaining {
				partSize = remaining
			}
		}

		offset := OffSet{off: off, len: partSize}
		nextOff := off + partSize

		// Terminate if: unknown size and last PartSizes entry, or known size end.
		if (size == nil && index >= len(b.PartSizes)-1) || (size != nil && nextOff >= *size) {
			return offset, nil, index
		}

		return offset, b.Resume(size, nextOff, index+1), index
	}
}

type Iterator func() (OffSet, Iterator, int)
