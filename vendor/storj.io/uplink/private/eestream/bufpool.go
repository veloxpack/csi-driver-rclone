// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"sync"
	"sync/atomic"

	"storj.io/common/sync2/race2"
)

const globalBufSize = 32 * 1024

var globalPool = sync.Pool{New: func() any { return new([globalBufSize]byte) }}

// A BatchPool is a sync.Pool that deals with batches of erasure shares,
// serialized as []byte slices of a fixed size. The fixed size is the largest
// multiple of the erasure share size that fits in standardBufSize.
type BatchPool struct {
	bufSize int
}

// NewBatchPool creates a BatchPool with the given erasure share size.
func NewBatchPool(shareSize int) *BatchPool {
	return &BatchPool{
		bufSize: (globalBufSize / shareSize) * shareSize,
	}
}

// GetAndClaim returns a batch of the pool. To free the batch, a Dec() call is needed.
func (b *BatchPool) GetAndClaim() *Batch {
	batch := &Batch{
		slice:   globalPool.Get().(*[globalBufSize]byte),
		bufSize: b.bufSize,
	}
	batch.refCount.Store(1)
	return batch
}

// Size returns the buffer size used in this pool.
func (b *BatchPool) Size() int { return b.bufSize }

// A Batch is a reference counted slice of erasure shares. Batches are returned
// by BatchPool.Get with a starting reference count of 1.
type Batch struct {
	slice    *[globalBufSize]byte
	bufSize  int
	refCount atomic.Int32
}

// Slice returns the batch's underlying memory allocation.
func (b *Batch) Slice() []byte { return b.slice[:b.bufSize] }

// Claim adds 1 to the batch reference count and returns true if the batch
// was claimable. See Release.
func (b *Batch) Claim() bool {
	for {
		val := b.refCount.Load()
		if val <= 0 {
			return false
		}
		if b.refCount.CompareAndSwap(val, val+1) {
			return true
		}
	}
}

// Release subtracts 1 from the batch reference count, returning the batch to
// the pool when it hits zero. Future Claim calls will return false once
// the counter drops to zero.
func (b *Batch) Release() {
	res := b.refCount.Add(-1)
	if res <= 0 {
		if res < 0 {
			panic("extra release")
		}
		race2.WriteSlice(b.slice[:])
		globalPool.Put(b.slice)
	}
}
