// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package buffer

import (
	"io"
	"sync"
	"sync/atomic"
)

// Backend is a backing store of bytes for a Backend.
type Backend interface {
	io.Writer
	io.ReaderAt
	io.Closer
}

const (
	standardMaxEncryptedSegmentSize = 67254016

	chunkSize = 1024 * 1024
)

var (
	standardPool = sync.Pool{
		New: func() interface{} {
			// TODO: this pool approach is a bit of a bandaid - it would be good to
			// rework this logic to not require this large allocation at all.
			return new([standardMaxEncryptedSegmentSize]byte)
		},
	}

	chunkPool = sync.Pool{
		New: func() interface{} {
			return new([chunkSize]byte)
		},
	}
)

// NewMemoryBackend returns a MemoryBackend with the provided initial
// capacity. It implements the Backend interface.
func NewMemoryBackend(cap int64) (rv *MemoryBackend) {
	rv = &MemoryBackend{}
	if cap == standardMaxEncryptedSegmentSize {
		rv.buf = standardPool.Get().(*[standardMaxEncryptedSegmentSize]byte)[:]
	} else {
		rv.buf = make([]byte, cap)
	}
	return rv
}

// MemoryBackend implements the Backend interface backed by a slice.
type MemoryBackend struct {
	len    int64
	buf    []byte
	closed bool
}

// Write appends the data to the buffer.
func (u *MemoryBackend) Write(p []byte) (n int, err error) {
	if u.closed {
		return 0, io.ErrClosedPipe
	}
	l := atomic.LoadInt64(&u.len)
	n = copy(u.buf[l:], p)
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	atomic.AddInt64(&u.len, int64(n))
	return n, nil
}

// ReadAt reads into the provided buffer p starting at off.
func (u *MemoryBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if u.closed {
		return 0, io.ErrClosedPipe
	}
	l := atomic.LoadInt64(&u.len)
	if off < 0 || off >= l {
		return 0, io.EOF
	}
	return copy(p, u.buf[off:l]), nil
}

// Close releases memory and causes future calls to ReadAt and Write to fail.
func (u *MemoryBackend) Close() error {
	buf := u.buf
	u.buf = nil
	u.closed = true
	if len(buf) == standardMaxEncryptedSegmentSize {
		standardPool.Put((*[standardMaxEncryptedSegmentSize]byte)(buf))
	}
	return nil
}

// NewChunkBackend returns a ChunkBackend with the provided initial capacity.
// Internally it stitchers writes together into small chunks to reduce the size
// of allocations needed for small objects. It implements the Backend interface.
// TODO: evaluate the usefulness of `cap` for the chunk backend.
func NewChunkBackend(cap int64) (rv *ChunkBackend) {
	// TODO: evaluate whether the chunks slice is worth trying to pool. Benchmarks
	// currently show the ChunkBackend has one extra (tiny) allocation but is otherwise
	// barely distinguishable to the MemoryBackend in terms of read/write performance.
	chunks := make([]atomic.Pointer[[chunkSize]byte], chunksNeeded(cap))
	return &ChunkBackend{chunks: chunks, cap: cap}
}

// ChunkBackend implements the Backend interface backed by a chained series of memory-pooled slices.
type ChunkBackend struct {
	end    atomic.Int64
	cap    int64
	chunks []atomic.Pointer[[chunkSize]byte]
	closed bool
}

// Write appends the data to the buffer.
func (u *ChunkBackend) Write(p []byte) (n int, err error) {
	if u.closed {
		return 0, io.ErrClosedPipe
	}

	end := u.end.Load()
	// If writing p exceeds the cap then constrain p so the write
	// no longer exceeds the cap and return ErrShortWrite.
	if end+int64(len(p)) > u.cap {
		p = p[:u.cap-end]
		err = io.ErrShortWrite
	}

	// Calculate the starting chunk position relative to the end
	chunkIdx, chunkOff := chunkPosition(end)

	for len(p) > 0 {
		chunk := u.chunks[chunkIdx].Load()
		if chunk == nil {
			chunk = chunkPool.Get().(*[chunkSize]byte)
			u.chunks[chunkIdx].Store(chunk)
		}
		nchunk := copy(chunk[chunkOff:], p)
		p = p[nchunk:]
		n += nchunk

		chunkIdx++
		chunkOff = 0
	}

	if n > 0 {
		u.end.Add(int64(n))
	}
	return n, err
}

// ReadAt reads into the provided buffer p starting at off.
func (u *ChunkBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if u.closed {
		return 0, io.ErrClosedPipe
	}

	end := u.end.Load()
	if off < 0 || off >= end {
		return 0, io.EOF
	}

	// If the read goes past the end, cap p to prevent read overflow.
	if off+int64(len(p)) > end {
		p = p[:end-off]
	}

	// Calculate the starting chunk position relative to the read offset
	chunkIdx, chunkOff := chunkPosition(off)

	for len(p) > 0 {
		chunk := u.chunks[chunkIdx].Load()
		nchunk := copy(p, chunk[chunkOff:])
		p = p[nchunk:]
		n += nchunk

		chunkIdx++
		chunkOff = 0
	}
	return n, nil
}

// Close releases memory and causes future calls to ReadAt and Write to fail.
func (u *ChunkBackend) Close() error {
	chunks := u.chunks
	u.chunks = nil
	u.closed = true
	for i := 0; i < len(chunks); i++ {
		chunk := chunks[i].Load()
		if chunk == nil {
			break
		}
		chunkPool.Put(chunk)
	}
	return nil
}

func chunksNeeded(n int64) int64 {
	if n == 0 {
		return 0
	}
	return 1 + ((n - 1) / chunkSize)
}

func chunkPosition(pos int64) (index, offset int64) {
	index = pos / chunkSize
	offset = pos - (index * chunkSize)
	return index, offset
}
