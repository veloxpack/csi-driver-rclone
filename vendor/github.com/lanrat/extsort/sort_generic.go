// Package extsort implements an unstable external sort for all the records in a chan or iterator
package extsort

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"slices"
	"sync"

	"github.com/lanrat/extsort/queue"
	"github.com/lanrat/extsort/tempfile"

	"golang.org/x/sync/errgroup"
)

const (
	// ctxCheckInterval is how many records a loop handles between checks of its context
	// while a non-blocking receive keeps succeeding, since that path never selects on ctx.Done().
	ctxCheckInterval = 1024
	// firstChunkCap is the capacity a chunk starts with before the input has filled a chunk.
	firstChunkCap = 1024
	// mergeBatchSize is how many records a merge worker hands to the final merge at a time.
	// Sending records one per channel operation made the handoffs cost more than the merge.
	mergeBatchSize = 1024
	// mergeBatchBuffer is how many full batches a merge worker can queue for the final merge.
	mergeBatchBuffer = 2
)

// genericChunk represents a collection of any data that can be sorted.
// It holds data in memory before being sorted using slices.SortFunc.
type genericChunk[E any] struct {
	data []E
}

// getChunk retrieves a chunk from the pool and initializes it
func (s *GenericSorter[E]) getChunk() *genericChunk[E] {
	c := s.pools.chunkPool.Get().(*genericChunk[E])

	// Get a slice pointer from the pool
	slicePtr := s.pools.slicePool.Get().(*[]E)
	*slicePtr = (*slicePtr)[:0] // Reset length but keep capacity

	c.data = *slicePtr
	return c
}

// putChunk returns a chunk to the pool for reuse
func (s *GenericSorter[E]) putChunk(c *genericChunk[E]) {
	if c != nil && c.data != nil {
		// Return the slice to the pool
		data := c.data
		c.data = nil // Clear reference before putting
		s.pools.slicePool.Put(&data)

		// Return the chunk to the pool
		s.pools.chunkPool.Put(c)
	}
}

// memoryPools holds sync.Pool instances for memory reuse
type memoryPools struct {
	chunkPool sync.Pool // *chunk objects
	slicePool sync.Pool // []E slices
}

// GenericSorter implements external sorting for any type E using a divide-and-conquer approach.
// It reads input from a channel, splits data into chunks that fit in memory, sorts each chunk,
// saves them to temporary files, then merges all chunks back into a sorted output stream.
// The sorter uses configurable parallelism for both sorting and merging phases,
// and employs memory pools to reduce garbage collection pressure during operation.
type GenericSorter[E any] struct {
	config         Config
	sortCtx        context.Context // shared by the build, sort and save stages
	mergeErrChan   chan error
	newTempWriter  func() (tempfile.TempWriter, error) // called once a second chunk exists
	tempWriter     tempfile.TempWriter
	tempReader     tempfile.TempReader
	input          <-chan E
	chunkChan      chan *genericChunk[E]
	saveChunkChan  chan *genericChunk[E]
	mergeChunkChan chan E
	compareFunc    CompareGeneric[E]
	fromBytes      FromBytesGeneric[E]
	toBytes        ToBytesGeneric[E]
	pools          *memoryPools
	singleChunk    *genericChunk[E] // Holds the single chunk for optimization
	// Set by useBuiltinCodec for the package's own codecs only
	appendBytes     appendBytesFunc[E] // encodes into a reused buffer instead of toBytes
	reuseReadBuffer bool               // fromBytes never keeps its input, so the merge reuses one buffer
}

// appendBytesFunc appends the encoding of v to dst and returns the extended slice.
type appendBytesFunc[E any] func(dst []byte, v E) []byte

// useBuiltinCodec makes the sorter reuse buffers with one of the package's own codecs:
// records are encoded with appendBytes into one reused buffer, and the merge decodes all the
// records of a chunk from one reused buffer. The second is only safe because those fromBytes
// functions never keep the slice they are given; a user's FromBytesGeneric may.
func (s *GenericSorter[E]) useBuiltinCodec(appendBytes appendBytesFunc[E]) {
	s.appendBytes = appendBytes
	s.reuseReadBuffer = true
}

// toBytesFunc returns the ToBytesGeneric form of appendBytes.
func toBytesFunc[E any](appendBytes appendBytesFunc[E]) ToBytesGeneric[E] {
	return func(v E) ([]byte, error) { return appendBytes(nil, v), nil }
}

// newSorter creates a new GenericSorter instance with the given configuration.
// This internal constructor initializes all channels, applies configuration defaults,
// and sets up memory pools for efficient resource reuse during sorting operations.
func newSorter[E any](input <-chan E, fromBytes FromBytesGeneric[E], toBytes ToBytesGeneric[E], compareFunc CompareGeneric[E], config *Config) *GenericSorter[E] {
	config = mergeConfig(config)
	s := &GenericSorter[E]{
		input:          input,
		compareFunc:    compareFunc,
		fromBytes:      fromBytes,
		toBytes:        toBytes,
		config:         *config,
		chunkChan:      make(chan *genericChunk[E], config.ChanBuffSize),
		saveChunkChan:  make(chan *genericChunk[E], config.NumWorkers*2), // Buffer for workers to avoid deadlock
		mergeChunkChan: make(chan E, config.SortedChanBuffSize),
		mergeErrChan:   make(chan error, 1),
	}
	s.pools = s.initMemoryPools()
	return s
}

// initMemoryPools initializes sync.Pool instances for efficient memory reuse during sorting.
// Creates pools for chunks and their slices to reduce GC pressure
// and improve performance during high-frequency allocation/deallocation cycles.
func (s *GenericSorter[E]) initMemoryPools() *memoryPools {
	pools := &memoryPools{}

	// Pool for chunk objects
	pools.chunkPool = sync.Pool{
		New: func() any {
			return &genericChunk[E]{}
		},
	}

	// Pool for slices - store pointers to slices. New slices start empty and
	// buildChunks grows them, so a small input does not allocate a full ChunkSize slice.
	pools.slicePool = sync.Pool{
		New: func() any {
			return new([]E)
		},
	}

	return pools
}

// Generic creates a new external sorter for any type E and returns the sorter instance,
// output channel with sorted results, and error channel.
//
// Parameters:
//   - input: Channel providing the data to be sorted. This channel MUST be closed when all data has been sent.
//   - fromBytes: Function to deserialize E from bytes when reading from disk
//   - toBytes: Function to serialize E to bytes when writing to disk
//   - compareFunc: Comparison function that returns negative/zero/positive for less/equal/greater
//   - config: Configuration options (nil uses defaults)
//
// The sorting process:
//  1. Reads data from input channel into memory chunks
//  2. Sorts each chunk in parallel using the provided compareFunc
//  3. Saves sorted chunks to temporary files using toBytes serialization
//  4. Merges all chunks back into sorted order using fromBytes deserialization
//
// fromBytes, toBytes and compareFunc are called from several goroutines at once,
// so they must be safe for concurrent use.
//
// Call Sort() on the returned sorter to begin the sorting process.
// Results are delivered via the output channel, errors via the error channel.
// The temporary file is only created once the input spans more than one chunk;
// failure to create it is reported on the error channel. The file is closed and
// removed when the sort completes, fails or is cancelled.
//
// IMPORTANT: The input channel must be closed to signal completion. The Sort() method
// will block until the input channel is closed. Failure to close it will cause a deadlock.
func Generic[E any](input <-chan E, fromBytes FromBytesGeneric[E], toBytes ToBytesGeneric[E], compareFunc CompareGeneric[E], config *Config) (*GenericSorter[E], <-chan E, <-chan error) {
	s := newSorter(input, fromBytes, toBytes, compareFunc, config)
	dir := s.config.TempFilesDir
	s.newTempWriter = func() (tempfile.TempWriter, error) {
		w, err := tempfile.New(dir, true)
		if err != nil {
			return nil, err
		}
		return w, nil
	}
	return s, s.mergeChunkChan, s.mergeErrChan
}

// MockGeneric creates an external sorter that uses in-memory storage instead of disk files.
// This is primarily useful for testing and benchmarking without filesystem I/O overhead.
// The parameter n specifies the initial capacity of the in-memory buffer.
// All other behavior is identical to Generic(), including that fromBytes, toBytes and
// compareFunc must be safe for concurrent use.
func MockGeneric[E any](input <-chan E, fromBytes FromBytesGeneric[E], toBytes ToBytesGeneric[E], compareFunc CompareGeneric[E], config *Config, n int) (*GenericSorter[E], <-chan E, <-chan error) {
	s := newSorter(input, fromBytes, toBytes, compareFunc, config)
	s.newTempWriter = func() (tempfile.TempWriter, error) {
		return tempfile.Mock(n), nil
	}
	return s, s.mergeChunkChan, s.mergeErrChan
}

// Sort sorts the Sorter's input chan and returns a new sorted chan, and error Chan
// Sort is a chunking operation that runs multiple workers asynchronously
// this blocks while sorting chunks and unblocks when merging
//
// IMPORTANT: The input channel MUST be closed to signal the end of data.
// Sort will continue reading from the input channel until it is closed.
// Failure to close the input channel will cause Sort to hang indefinitely.
//
// NOTE: the context passed to Sort must outlive Sort() returning.
// Merge uses the same context and runs in a goroutine after Sort returns().
// for example, if calling sort in an errGroup, you must pass the group's parent context into sort.
func (s *GenericSorter[E]) Sort(ctx context.Context) {
	// One group for all stages: an error in any stage cancels the others,
	// so a failed save cannot leave the sort workers blocked on saveChunkChan.
	group, groupCtx := errgroup.WithContext(ctx)
	s.sortCtx = groupCtx

	//start creating chunks
	group.Go(s.buildChunks)

	// sort chunks
	var sorters sync.WaitGroup
	sorters.Add(s.config.NumWorkers)
	for i := 0; i < s.config.NumWorkers; i++ {
		group.Go(func() error {
			defer sorters.Done()
			return s.sortChunks()
		})
	}

	// Close saveChunkChan to signal end of chunks once every sort worker has
	// exited, successfully or not, so the save worker always returns.
	group.Go(func() error {
		sorters.Wait()
		close(s.saveChunkChan)
		return nil
	})

	// Start the save worker that will handle single-chunk optimization
	group.Go(s.saveChunksOptimized)

	err := group.Wait()
	if err != nil {
		s.closeTempFiles()
		s.mergeErrChan <- err
		close(s.mergeErrChan)
		close(s.mergeChunkChan)
		return
	}

	// Check if single chunk optimization was used
	if s.singleChunk != nil {
		// Single chunk case - output directly
		go s.outputSingleChunk(ctx)
		return
	}

	// Multiple chunks: read chunks and merge
	// if this errors, it is returned in the errorChan
	go s.mergeNChunks(ctx)
}

// closeTempFiles releases the temp file on paths that never reach the merge.
// The merge closes the reader itself.
func (s *GenericSorter[E]) closeTempFiles() {
	if s.tempReader != nil {
		_ = s.tempReader.Close()
		s.tempReader = nil
	}
	if s.tempWriter != nil {
		_ = s.tempWriter.Close()
		s.tempWriter = nil
	}
}

// buildChunks reads data from the input chan to builds chunks and pushes them to chunkChan
func (s *GenericSorter[E]) buildChunks() error {
	defer close(s.chunkChan) // if this is not called on error, causes a deadlock

	// Set once a chunk fills: the input spans several chunks, so new chunk slices
	// are allocated at their full size instead of grown.
	spansChunks := false
	for inputOpen := true; inputOpen; {
		c := s.getChunk()
	fill:
		for i := 0; i < s.config.ChunkSize; i++ {
			var rec E
			var ok bool
			// Try a non-blocking receive first: unlike the select below it does not lock
			// the context's channel, so a steady input costs one channel operation per record
			select {
			case rec, ok = <-s.input:
				if i%ctxCheckInterval == ctxCheckInterval-1 && s.sortCtx.Err() != nil {
					s.putChunk(c) // Return unused chunk to pool
					return s.sortCtx.Err()
				}
			default:
				select {
				case rec, ok = <-s.input:
				case <-s.sortCtx.Done():
					s.putChunk(c) // Return unused chunk to pool
					return s.sortCtx.Err()
				}
			}
			if !ok {
				inputOpen = false
				break fill
			}
			if len(c.data) == cap(c.data) {
				c.data = s.growChunk(c.data, spansChunks)
			}
			c.data = append(c.data, rec)
		}
		if len(c.data) == s.config.ChunkSize {
			spansChunks = true
		}
		if len(c.data) == 0 {
			// the chunk is empty, return it to pool
			s.putChunk(c)
			break
		}

		select {
		// chunk is now full, or holds the last records
		case s.chunkChan <- c:
		case <-s.sortCtx.Done():
			s.putChunk(c) // Return unused chunk to pool
			return s.sortCtx.Err()
		}
	}

	return nil
}

// growChunk returns data with room for at least one more record, up to ChunkSize. The first
// chunk doubles from firstChunkCap, so a small input only allocates what it needs; once the
// input has filled a chunk (full), a chunk grows straight to ChunkSize.
func (s *GenericSorter[E]) growChunk(data []E, full bool) []E {
	newCap := s.config.ChunkSize
	if !full {
		newCap = min(newCap, max(2*cap(data), firstChunkCap))
	}
	grown := make([]E, len(data), newCap) // exact capacity: append could grow past ChunkSize
	copy(grown, data)
	return grown
}

// sortChunks is a worker for sorting the data stored in a chunk prior to save
func (s *GenericSorter[E]) sortChunks() error {
	for {
		select {
		case b, more := <-s.chunkChan:
			if more {
				// Create channels to communicate completion and errors
				sortDone := make(chan error, 1)

				// Run sort in a separate goroutine
				go func() {
					defer func() {
						// Recover from panics in comparison function
						if r := recover(); r != nil {
							sortDone <- NewComparisonError(r, "sortChunks")
						} else {
							sortDone <- nil // Success
						}
					}()
					slices.SortFunc(b.data, s.compareFunc)
				}()

				// Wait for either sort completion or context cancellation
				select {
				case sortErr := <-sortDone:
					if sortErr != nil {
						// Sort failed due to panic
						s.putChunk(b) // Return chunk to pool
						return sortErr
					}
					// Sort completed successfully, proceed to save
					select {
					case s.saveChunkChan <- b:
					case <-s.sortCtx.Done():
						return s.sortCtx.Err()
					}
				case <-s.sortCtx.Done():
					// Context cancelled while sorting - abandon this chunk
					return s.sortCtx.Err()
				}
			} else {
				return nil
			}
		case <-s.sortCtx.Done():
			return s.sortCtx.Err()
		}
	}
}

// outputSingleChunk handles the single-chunk optimization by directly outputting
// the sorted chunk without any disk I/O. This provides significant performance
// benefits for small datasets that fit entirely in memory.
func (s *GenericSorter[E]) outputSingleChunk(ctx context.Context) {
	defer close(s.mergeChunkChan)
	defer close(s.mergeErrChan)

	// Use the chunk collected by collectSingleChunk
	chunk := s.singleChunk
	if chunk == nil {
		// No chunk collected - this shouldn't happen but handle gracefully
		return
	}

	// Output each item in the sorted chunk directly
	for _, item := range chunk.data {
		select {
		case s.mergeChunkChan <- item:
		case <-ctx.Done():
			s.mergeErrChan <- ctx.Err()
			return
		}
	}

	// Return chunk to pool
	s.putChunk(chunk)
	s.singleChunk = nil // Clear reference
}

// saveChunksOptimized handles both single-chunk and multi-chunk cases
// For single chunk: stores it in memory to avoid disk I/O
// For multiple chunks: saves all chunks to disk normally
func (s *GenericSorter[E]) saveChunksOptimized() error {
	// Get the first chunk with context checking
	var firstChunk *genericChunk[E]
	var ok bool
	select {
	case firstChunk, ok = <-s.saveChunkChan:
		if !ok {
			// Channel closed, no chunks at all
			return nil
		}
	case <-s.sortCtx.Done():
		return s.sortCtx.Err()
	}

	// Try to get a second chunk with context checking
	var secondChunk *genericChunk[E]
	select {
	case secondChunk, ok = <-s.saveChunkChan:
		if !ok {
			// Channel closed after first chunk - single chunk optimization
			s.singleChunk = firstChunk
			return nil
		}
	case <-s.sortCtx.Done():
		s.putChunk(firstChunk) // Return to pool before exiting
		return s.sortCtx.Err()
	}

	// We have at least 2 chunks - use multi-chunk path, which needs the temp file
	tempWriter, err := s.newTempWriter()
	if err != nil {
		s.putChunk(firstChunk)
		s.putChunk(secondChunk)
		return err
	}
	s.tempWriter = tempWriter

	// Save the first chunk
	if err := s.saveChunk(firstChunk); err != nil {
		s.putChunk(secondChunk) // Return to pool
		return err
	}

	// Save the second chunk
	if err := s.saveChunk(secondChunk); err != nil {
		return err
	}

	// Continue saving any remaining chunks with context checking
	for {
		select {
		case chunk, ok := <-s.saveChunkChan:
			if !ok {
				// Channel closed, we're done, unless it closed because another stage failed
				if err := s.sortCtx.Err(); err != nil {
					return err
				}
				// Finalize the temp writer and save it for reading
				tempReader, err := s.tempWriter.Save()
				if err != nil {
					return err
				}
				// The reader now owns the file
				s.tempReader = tempReader
				s.tempWriter = nil
				return nil
			}
			if err := s.saveChunk(chunk); err != nil {
				return err
			}
		case <-s.sortCtx.Done():
			return s.sortCtx.Err()
		}
	}
}

// saveChunk writes a sorted chunk as the next section of the temp file, each record as a
// uvarint length followed by its encoding, and returns the chunk to the pool.
func (s *GenericSorter[E]) saveChunk(b *genericChunk[E]) error {
	defer s.putChunk(b)

	var err error
	if s.appendBytes != nil {
		err = s.writeAppended(b.data)
	} else {
		err = s.writeEncoded(b.data)
	}
	if err != nil {
		return err
	}
	if _, err := s.tempWriter.Next(); err != nil {
		return NewDiskError(err, "next chunk", "")
	}
	return nil
}

// writeEncoded writes records encoded by toBytes.
func (s *GenericSorter[E]) writeEncoded(records []E) error {
	var header [binary.MaxVarintLen64]byte
	for _, d := range records {
		raw, err := s.encode(d)
		if err != nil {
			return err
		}
		n := binary.PutUvarint(header[:], uint64(len(raw)))
		if _, err := s.tempWriter.Write(header[:n]); err != nil {
			return NewDiskError(err, "write size header", "")
		}
		if _, err := s.tempWriter.Write(raw); err != nil {
			return NewDiskError(err, "write data", "")
		}
	}
	return nil
}

// writeAppended writes records encoded by appendBytes into one reused buffer. Each record
// is appended after room for the longest length header, and the header is then put just
// before it, so a record and its header go out in a single Write.
func (s *GenericSorter[E]) writeAppended(records []E) error {
	const room = binary.MaxVarintLen64
	buf := make([]byte, room, 256)
	var header [room]byte
	for _, d := range records {
		buf = s.appendBytes(buf[:room], d)
		n := binary.PutUvarint(header[:], uint64(len(buf)-room))
		start := room - n
		copy(buf[start:], header[:n])
		if _, err := s.tempWriter.Write(buf[start:]); err != nil {
			return NewDiskError(err, "write data", "")
		}
	}
	return nil
}

// encode serializes one record with toBytes, converting both a returned error
// and a panic into a SerializationError.
func (s *GenericSorter[E]) encode(d E) (raw []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			raw = nil
			err = NewSerializationError(r, "saveChunk")
		}
	}()
	raw, err = s.toBytes(d)
	if err != nil {
		return nil, NewSerializationError(err, "saveChunk")
	}
	return raw, nil
}

// mergeNChunks runs asynchronously in the background feeding data to getNext
// sends errors to s.mergeErrorChan. Uses parallel merging for better performance.
func (s *GenericSorter[E]) mergeNChunks(ctx context.Context) {
	// Deferred calls run last-in first-out: the error channel closes before the output channel.
	defer close(s.mergeChunkChan)
	defer close(s.mergeErrChan)

	if s.tempReader == nil {
		return
	}

	var err error
	if s.tempReader.Size() <= s.config.NumWorkers {
		// For small number of chunks, use single-threaded merge
		err = s.mergeNChunksSingleThreaded(ctx)
	} else {
		// Use parallel merging for many chunks
		err = s.mergeNChunksParallel(ctx)
	}

	// Release the temp file before signalling completion
	if closeErr := s.tempReader.Close(); closeErr != nil && err == nil {
		err = NewDiskError(closeErr, "close temp file", "")
	}
	s.tempReader = nil

	if err != nil {
		s.mergeErrChan <- err
	}
}

// mergeNChunksSingleThreaded is the original single-threaded implementation
func (s *GenericSorter[E]) mergeNChunksSingleThreaded(ctx context.Context) (err error) {
	// A panicking compareFunc must not crash the process from this goroutine
	defer func() {
		if r := recover(); r != nil {
			err = NewComparisonError(r, "mergeNChunksSingleThreaded")
		}
	}()

	pq := queue.NewPriorityQueue(func(a, b *mergeFile[E]) int {
		return s.compareFunc(a.nextRec, b.nextRec)
	})

	for i := 0; i < s.tempReader.Size(); i++ {
		merge := s.newMergeFile(i)
		_, ok, err := merge.getNext() // start the merge by preloading the values
		if err != nil {
			return err
		}
		if !ok {
			continue // empty chunk
		}
		pq.Push(merge)
	}

	for pq.Len() > 0 {
		merge := pq.Peek()
		rec, more, err := merge.getNext()
		if err != nil {
			return err
		}
		if more {
			pq.PeekUpdate()
		} else {
			pq.Pop()
		}
		// check for err in context just in case
		select {
		case s.mergeChunkChan <- rec:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// mergeNChunksParallel implements parallel k-way merging with robust cancellation
func (s *GenericSorter[E]) mergeNChunksParallel(ctx context.Context) error {
	numChunks := s.tempReader.Size()
	numWorkers := s.config.NumWorkers

	// Create a cancellable context for all merge operations
	mergeCtx, mergeCancel := context.WithCancel(ctx)
	defer mergeCancel() // Ensure all goroutines stop when this function returns

	// Create a stream for each worker to pass its merged records to the final merge in batches
	streams := make([]mergeStream[E], numWorkers)
	for i := range streams {
		streams[i] = newMergeStream[E]()
	}

	// Error collection
	errChan := make(chan error, numWorkers+1) // +1 for final merge errors
	var mergeErr error
	var errOnce sync.Once

	// Start workers
	var wg sync.WaitGroup
	chunksPerWorker := (numChunks + numWorkers - 1) / numWorkers
	workersStarted := 0

	for i := 0; i < numWorkers; i++ {
		startChunk := i * chunksPerWorker
		endChunk := (i + 1) * chunksPerWorker
		if endChunk > numChunks {
			endChunk = numChunks
		}
		if startChunk >= numChunks {
			break
		}
		workersStarted++
		wg.Add(1)

		go func(stream mergeStream[E], start, end int) {
			defer wg.Done()
			defer close(stream.batches) // Each worker closes its own channel

			if err := s.mergeWorkerSimple(mergeCtx, start, end, stream); err != nil {
				errChan <- err
				mergeCancel() // Cancel all operations on error
			}
		}(streams[i], startChunk, endChunk)
	}

	// Start error collector with wait group for synchronization
	var errorCollectorWg sync.WaitGroup
	errorCollectorWg.Add(1)
	go func() {
		defer errorCollectorWg.Done()
		for err := range errChan {
			if err != nil {
				errOnce.Do(func() {
					mergeErr = err
					mergeCancel() // Cancel all operations on first error
				})
			}
		}
	}()

	// Start final merge in a goroutine to avoid blocking
	var finalMergeWg sync.WaitGroup
	finalMergeWg.Add(1)
	go func() {
		defer finalMergeWg.Done()
		if err := s.finalMergeSimple(mergeCtx, streams[:workersStarted]); err != nil {
			errChan <- err
			mergeCancel() // Stop the workers, which may be blocked sending to the final merge
		}
	}()

	// Wait for all workers to complete
	wg.Wait()

	// Wait for final merge to complete
	finalMergeWg.Wait()

	close(errChan) // Signal error collector to stop

	// Wait for error collector to finish processing all errors
	errorCollectorWg.Wait()

	// Return any collected error (now safe to read mergeErr)
	if mergeErr != nil {
		return mergeErr
	}
	return ctx.Err()
}

// mergeStream carries one merge worker's records to the final merge in batches.
type mergeStream[E any] struct {
	batches chan []E // batches of records in merge order, closed when the worker stops
	free    chan []E // used-up batches handed back to the worker for reuse
}

func newMergeStream[E any]() mergeStream[E] {
	return mergeStream[E]{
		batches: make(chan []E, mergeBatchBuffer),
		// room for every batch a worker has: those queued, the one it fills and the one being merged
		free: make(chan []E, mergeBatchBuffer+2),
	}
}

// emptyBatch returns a batch to fill, reusing one the final merge handed back if there is one.
func (m mergeStream[E]) emptyBatch() []E {
	select {
	case batch := <-m.free:
		return batch[:0]
	default:
		return make([]E, 0, mergeBatchSize)
	}
}

// send queues a batch for the final merge. Once ctx is done it returns ctx's error instead,
// so a worker stops at its next batch after a cancellation.
func (m mergeStream[E]) send(ctx context.Context, batch []E) error {
	if err := ctx.Err(); err != nil {
		return err // the select below picks at random when the channel has room too
	}
	select {
	case m.batches <- batch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// mergeWorkerSimple merges a subset of chunks with proper context handling
func (s *GenericSorter[E]) mergeWorkerSimple(ctx context.Context, startChunk, endChunk int, output mergeStream[E]) (err error) {
	// A panicking compareFunc must not crash the process from this goroutine
	defer func() {
		if r := recover(); r != nil {
			err = NewComparisonError(r, "mergeWorkerSimple")
		}
	}()

	pq := queue.NewPriorityQueue(func(a, b *mergeFile[E]) int {
		return s.compareFunc(a.nextRec, b.nextRec)
	})

	// Initialize merge files for this worker's chunk range
	for i := startChunk; i < endChunk; i++ {
		merge := s.newMergeFile(i)
		_, ok, err := merge.getNext()
		if err != nil {
			return err
		}
		if !ok {
			continue // empty chunk
		}
		pq.Push(merge)
	}

	// Merge this worker's chunks, checking ctx as each batch is sent
	batch := output.emptyBatch()
	for pq.Len() > 0 {
		merge := pq.Peek()
		rec, more, err := merge.getNext()
		if err != nil {
			return err
		}
		if more {
			pq.PeekUpdate()
		} else {
			pq.Pop()
		}

		batch = append(batch, rec)
		if len(batch) == mergeBatchSize {
			if err := output.send(ctx, batch); err != nil {
				return err
			}
			batch = output.emptyBatch()
		}
	}
	if len(batch) > 0 {
		return output.send(ctx, batch)
	}
	return nil
}

// finalMergeSimple performs streaming merge with simpler synchronization.
// It returns nil when ctx is cancelled; the caller reports the cancellation.
func (s *GenericSorter[E]) finalMergeSimple(ctx context.Context, streams []mergeStream[E]) (err error) {
	// A panicking compareFunc must not crash the process from this goroutine
	defer func() {
		if r := recover(); r != nil {
			err = NewComparisonError(r, "finalMergeSimple")
		}
	}()

	pq := queue.NewPriorityQueue(func(a, b *channelMergeSource[E]) int {
		return s.compareFunc(a.nextRec, b.nextRec)
	})

	// Initialize sources
	for _, stream := range streams {
		source := &channelMergeSource[E]{stream: stream}
		if source.getNextSimple() {
			pq.Push(source)
		}
	}

	// Perform final streaming merge with proper context handling
	for pq.Len() > 0 {
		// Check if context is cancelled before each iteration
		if ctx.Err() != nil {
			return nil
		}

		source := pq.Peek()

		// Try to send with context cancellation support
		select {
		case s.mergeChunkChan <- source.nextRec:
			// Successfully sent, try to get next from this source
			if source.getNextSimple() {
				pq.PeekUpdate()
			} else {
				pq.Pop()
			}
		case <-ctx.Done():
			// Context cancelled, exit immediately
			return nil
		}
	}
	return nil
}

// channelMergeSource represents a source of sorted data from a merge worker's stream
type channelMergeSource[E any] struct {
	stream  mergeStream[E]
	batch   []E // the batch being merged
	pos     int // index in batch of the record after nextRec
	nextRec E
}

// getNextSimple advances to the next record, receiving the next batch once the current one
// is used up. It reads without context: the worker closes its channel when it stops.
func (c *channelMergeSource[E]) getNextSimple() bool {
	for c.pos == len(c.batch) {
		c.releaseBatch()
		batch, ok := <-c.stream.batches
		if !ok {
			return false
		}
		c.batch, c.pos = batch, 0
	}
	c.nextRec = c.batch[c.pos]
	c.pos++
	return true
}

// releaseBatch hands the used-up batch back to the worker, or drops it if the worker
// already has enough spare batches.
func (c *channelMergeSource[E]) releaseBatch() {
	if c.batch == nil {
		return
	}
	clear(c.batch) // the records were sent on; don't keep them reachable from a spare batch
	select {
	case c.stream.free <- c.batch:
	default:
	}
	c.batch = nil
}

// newMergeFile returns a mergeFile reading section i of the temp file.
func (s *GenericSorter[E]) newMergeFile(i int) *mergeFile[E] {
	return &mergeFile[E]{
		fromBytes: s.fromBytes,
		reader:    s.tempReader.Read(i),
		reuseBuf:  s.reuseReadBuffer,
	}
}

// mergeFile represents each sorted chunk on disk and its next value
type mergeFile[E any] struct {
	nextRec   E
	fromBytes FromBytesGeneric[E]
	reader    *bufio.Reader
	reuseBuf  bool   // fromBytes never keeps its input, so every record can be read into buf
	buf       []byte // holds the last record read when reuseBuf is set
}

// getNext returns the next value from the sorted chunk on disk.
// The first call will return nil while the struct is initialized.
// It handles deserialization errors by wrapping them in DeserializationError instances.
func (m *mergeFile[E]) getNext() (E, bool, error) {
	old := m.nextRec

	n, err := binary.ReadUvarint(m.reader)
	if err == io.EOF {
		return old, false, nil // clean end of the chunk
	}
	if err != nil {
		return old, false, err
	}
	var newRecBytes []byte
	if m.reuseBuf {
		if uint64(cap(m.buf)) < n {
			m.buf = make([]byte, int(n))
		}
		newRecBytes = m.buf[:n]
	} else {
		newRecBytes = make([]byte, int(n))
	}
	if _, err := io.ReadFull(m.reader, newRecBytes); err != nil {
		if err == io.EOF {
			// a length header without its payload is a truncated record, not the end of the chunk
			err = io.ErrUnexpectedEOF
		}
		return old, false, err
	}

	m.nextRec, err = m.decode(newRecBytes)
	if err != nil {
		return old, true, err
	}

	return old, true, nil
}

// decode deserializes one record with fromBytes, converting both a returned error
// and a panic into a DeserializationError.
func (m *mergeFile[E]) decode(d []byte) (rec E, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewDeserializationError(r, len(d), "getNext")
		}
	}()
	rec, err = m.fromBytes(d)
	if err != nil {
		return rec, NewDeserializationError(err, len(d), "getNext")
	}
	return rec, nil
}
