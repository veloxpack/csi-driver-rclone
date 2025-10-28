package file

import (
	"context"
	"math"
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib"
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
	PartSizes []int64
}

func (b ByteOffset) WithDefaultChunkSize(size *int64, off int64, index int, defaultChunkSize int64) Iterator {
	return func() (OffSet, Iterator, int) {
		// if size is nil or off is still less than size
		if size == nil || off < *size {
			endRange := off + b.PartSizes[index]

			if size != nil && *size > endRange {
				endRange = defaultChunkSize
			}

			// if size is not nil, limit endRange by size
			if size != nil {
				endRange = int64(math.Min(float64(endRange), float64(*size)))
			}

			offset := OffSet{off: off, len: endRange - off}

			off = endRange

			// if there are no more partSizes or off is already more than size, return nil iterator
			if index >= len(lib.PartSizes)-1 || (size != nil && off >= *size) {
				return offset, nil, index + 1
			}

			return offset, b.Resume(size, off, index+1), index
		}

		return OffSet{}, nil, index
	}
}

func (b ByteOffset) BySize(size *int64) Iterator {
	return b.Resume(size, 0, 0)
}

func (b ByteOffset) Resume(size *int64, off int64, index int) Iterator {
	return func() (OffSet, Iterator, int) {
		// if size is nil or off is still less than size
		if size == nil || off < *size {
			endRange := off + b.PartSizes[index]

			// if size is not nil, limit endRange by size
			if size != nil {
				endRange = int64(math.Min(float64(endRange), float64(*size)))
			}

			offset := OffSet{off: off, len: endRange - off}

			off = endRange

			// if there are no more partSizes or off is already more than size, return nil iterator
			if index >= len(lib.PartSizes)-1 || (size != nil && off >= *size) {
				return offset, nil, index
			}

			return offset, b.Resume(size, off, index+1), index
		}

		return OffSet{}, nil, index
	}
}

type Iterator func() (OffSet, Iterator, int)
