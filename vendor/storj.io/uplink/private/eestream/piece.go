// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"storj.io/common/sync2/race2"
)

var freedBatch = &Batch{}

// A StreamingPiece is an in memory storage location for a stream of bytes being
// operated on by a single producer and a single consumer in atomic units of
// a given erasure share size. The StreamingPiece type must know its full expected size
// up front, and allocates slots for each *BufPool batch of erasure shares
// up to that total size. It will hydrate these slots on demand and free them
// back to the BufPool as they are consumed.
type StreamingPiece struct {
	// bufs is a list of atomic.Pointer[Batch].
	// The value of each slot is either nil, emptyBatch, or a valid Batch
	// slice returned from BatchPool. nil means the batch has yet to be
	// instantiated. freedBatch means that the batch is freed or no longer
	// needed.
	batches             []atomic.Pointer[Batch]
	pool                *BatchPool
	receivedBytesSoFar  int64
	receivedSharesSoFar atomic.Int64
	shareSize           int
	err                 atomic.Value
	completedBatches    atomic.Int32
}

// NewStreamingPiece creates a buffer that uses units of size unitSize, with a total
// amount of bytes of totalSize. It uses pool to hydrate and return buffers in
// its slots.
func NewStreamingPiece(shareSize int, totalSize int64, pool *BatchPool) *StreamingPiece {
	poolSize := int64(pool.Size())
	batches := (totalSize + poolSize - 1) / poolSize

	return &StreamingPiece{
		batches:   make([]atomic.Pointer[Batch], batches+1),
		pool:      pool,
		shareSize: shareSize,
	}
}

// ensureBatch will return a batch for writing that is either a full length
// from the BatchPool, or specifically the freedBatch value, which means that
// the consumer has already indicated all units that would belong to this
// buffer are not needed.
func (b *StreamingPiece) ensureBatch(idx int) *Batch {
	for {
		if batch := b.batches[idx].Load(); batch != nil {
			return batch
		}
		batch := b.pool.GetAndClaim()
		if b.batches[idx].CompareAndSwap(nil, batch) {
			return batch
		}
		batch.Release()
	}
}

// byteToBatch determines which batch and which batch offset a specific byte
// at byteOffset in the overall stream lives.
func (b *StreamingPiece) byteToBatch(byteOffset int64) (batchIdx, batchOffset int) {
	poolSize := int64(b.pool.Size())
	batchIdx = int(byteOffset / poolSize)
	batchOffset = int(byteOffset % poolSize)
	return batchIdx, batchOffset
}

// ReadSharesFrom is going to call r.Read() once, and will return the number of
// full shares that are now newly completely read as a result of this call. If
// r.Read() returns an error or io.EOF, or no data is expected otherwise,
// done will be true. The read error if any is available from Err() or
// ReadShare().
func (b *StreamingPiece) ReadSharesFrom(r io.Reader) (shareCount int, done bool) {
	// find our current buffer
	currentBatchIdx, currentBatchOffset := b.byteToBatch(b.receivedBytesSoFar)

	currentBatch := b.ensureBatch(currentBatchIdx)

	// okay, there are two main cases for the batch we just grabbed:
	// 1) it has already been freed. in this case, currentBatch == freedBatch
	// 2) we can attempt to claim it. this has two subcases:
	//   a) if the claim fails, then it was in the process of being freed.
	//   b) if the claim succeeds, then we know that no one is going to free
	//      it while we have it, at least until we release it.
	// in case 1 and 2a, then we don't have a batch to use. only case 2b is
	// useful for using the returned batch.
	// all of this is because in case 2b, we want to make sure someone doesn't
	// try and free this batch while we're using it.
	if currentBatch == freedBatch || !currentBatch.Claim() {
		// this batch isn't ours, but we still need to read off the stream
		// as long as someone is still interested in this stream, so we need
		// to use a throwaway buffer.
		// we do this instead of using io.Discard or something because this way
		// all of the logic and bookkeeping is exactly the same.
		currentBatch = b.pool.GetAndClaim()
	}
	defer currentBatch.Release()

	currentSlice := currentBatch.Slice()[currentBatchOffset:]
	race2.WriteSlice(currentSlice)

	// okay, read into the current buffer
	n, err := r.Read(currentSlice)
	// keep track of how many bytes we've read from the stream
	b.receivedBytesSoFar += int64(n)

	// we may have only read a partial share last time, so we need to
	// recalculate how many bytes are covered by prior completed ReadShareFrom
	// shareCount returns.
	receivedSharesSoFar := b.receivedSharesSoFar.Load()

	notifiedBytesSoFar := receivedSharesSoFar * int64(b.shareSize)

	// okay, let's see how many completed shares we can tell the caller about.
	unnotifiedBytes := b.receivedBytesSoFar - notifiedBytesSoFar
	unnotifiedShares := unnotifiedBytes / int64(b.shareSize)

	// make a note about how many we've told the caller about
	b.receivedSharesSoFar.Add(unnotifiedShares)

	// keep track of the error if there was a read error.
	if err != nil && !errors.Is(err, io.EOF) {
		b.err.Store(err)
	}

	return int(unnotifiedShares), err != nil
}

// ReadShare returns the byte slice that references the read data in a buffer
// representing the share with index shareIdx. Note that shareIdx is not
// the Reed Solomon Share Number, since all shares in this buffer share the same
// Reed Solomon Share Number. If a share at shareIdx cannot be returned, it will
// return an error, which may be a read error determined by ReadSharesFrom.
// The release callback must be released when the share is done being read from.
func (b *StreamingPiece) ReadShare(shareIdx int) (data []byte, release func(), err error) {
	// first, let's see if we even have data for this unit. have we read this
	// far yet?
	receivedSharesSoFar := b.receivedSharesSoFar.Load()
	if int64(shareIdx) >= receivedSharesSoFar {
		if err, ok := b.err.Load().(error); ok {
			// oh, there's a stored error. let's return that, that probably says
			// what happened.
			return nil, nil, err
		}
		return nil, nil, Error.New("read past end of buffer: %w", io.ErrUnexpectedEOF)
	}

	// find our buffer and buffer offset for this unit
	byteOffset := int64(shareIdx) * int64(b.shareSize)
	if debugEnabled {
		fmt.Println("buffer reading byte offset", byteOffset, "for share", shareIdx)
	}
	batchIdx, batchOffset := b.byteToBatch(byteOffset)

	// okay, let's go find our batch
	batch := b.batches[batchIdx].Load()
	if batch == nil {
		// huh! someone asked for a batch that we haven't received yet, but we
		// checked up top. some major problem with bookkeeping happened.
		return nil, nil, Error.New("unreachable - this batch should be hydrated")
	}
	if batch == freedBatch || !batch.Claim() {
		// this buffer was already marked as completed, so we probably returned it
		// back to the BufPool.
		return nil, nil, Error.New("read completed buffer")
	}

	data = batch.Slice()[batchOffset:][:b.shareSize]
	race2.ReadSlice(data)

	// okay we have the data.
	if debugEnabled {
		fmt.Println("buffer reading unit", shareIdx, "from", batchIdx, batchOffset, fmt.Sprintf("%x", data[:3]))
	}
	return data, batch.Release, nil
}

// Err returns the last error encountered during reading.
func (b *StreamingPiece) Err() error {
	if err, ok := b.err.Load().(error); ok {
		return err
	}
	return nil
}

// MarkCompleted tells the StreamingPiece to return some internal batches back to the
// BatchPool, since we don't need them anymore. It will assume that none of the
// first sharesCompleted units will be asked for again.
func (b *StreamingPiece) MarkCompleted(sharesCompleted int) {
	// okay figure out which buffer is needed for unit index unitsCompleted
	// (this is the next incomplete unit). This will be completedBatches, and
	// the batch with index completedBatches is still in use! Everything before
	// it is free to reclaim though.
	completedBatches, _ := b.byteToBatch(int64(sharesCompleted) * int64(b.shareSize))

	for {
		// what do we think we've already marked completed?
		oldCompletedBatches := int(b.completedBatches.Load())
		if completedBatches <= oldCompletedBatches {
			// already done
			break
		}
		// okay, let's mark all of the ones we don't think we've marked completed
		// before as completed.
		for i := oldCompletedBatches; i < completedBatches; i++ {
			if batch := b.batches[i].Swap(freedBatch); batch != nil && batch != freedBatch {
				// a live batch! let's return to the BufPool.
				batch.Release()
			}
		}
		// okay, let's see if we're racing with any other MarkCompleteds and if
		// we need to rethink what we've done here. (we wouldn't want to stomp on
		// another MarkCompleted that did more than us).
		if b.completedBatches.CompareAndSwap(int32(oldCompletedBatches), int32(completedBatches)) {
			// we're okay, we're done.
			break
		}
	}
}
