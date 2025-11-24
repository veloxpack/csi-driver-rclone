package seekable

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/a1ex3/zstd-seekable-format-go/pkg/env"
)

var errWriterClosed = errors.New("writer is closed")

// writerEnvImpl is the environment implementation of for the underlying WriteCloser.
type writerEnvImpl struct {
	w io.Writer
}

func (w *writerEnvImpl) WriteFrame(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writerEnvImpl) WriteSeekTable(p []byte) (n int, err error) {
	return w.w.Write(p)
}

type writerImpl struct {
	enc          ZSTDEncoder
	frameEntries []seekTableEntry

	env env.WEnvironment

	mu sync.Mutex
}

var (
	_ io.Writer = (*writerImpl)(nil)
	_ io.Closer = (*writerImpl)(nil)
)

type Writer interface {
	// Write writes a chunk of data as a separate frame into the datastream.
	//
	// Note that Write does not do any coalescing nor splitting of data,
	// so each write will map to a separate ZSTD Frame.
	Write(src []byte) (int, error)

	// Close implement io.Closer interface.  It writes the seek table footer
	// and releases occupied memory.
	//
	// Caller is still responsible to Close the underlying writer.
	Close() (err error)
}

// FrameSource returns one frame of data at a time.
// When there are no more frames, returns nil.
type FrameSource func() ([]byte, error)

// ConcurrentWriter allows writing many frames concurrently
type ConcurrentWriter interface {
	Writer

	// WriteMany writes many frames concurrently
	WriteMany(ctx context.Context, frameSource FrameSource, options ...WriteManyOption) error
}

// ZSTDEncoder is the compressor.  Tested with github.com/klauspost/compress/zstd.
type ZSTDEncoder interface {
	EncodeAll(src, dst []byte) []byte
}

// NewWriter wraps the passed io.Writer and Encoder into and indexed ZSTD stream.
// Resulting stream then can be randomly accessed through the Reader and Decoder interfaces.
func NewWriter(w io.Writer, encoder ZSTDEncoder, opts ...wOption) (ConcurrentWriter, error) {
	sw := writerImpl{
		enc: encoder,
	}

	for _, o := range opts {
		err := o(&sw)
		if err != nil {
			return nil, err
		}
	}

	if sw.env == nil {
		sw.env = &writerEnvImpl{
			w: w,
		}
	}

	return &sw, nil
}

func (s *writerImpl) Write(src []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.env == nil {
		return 0, errWriterClosed
	}

	dst, err := s.Encode(src)
	if err != nil {
		return 0, err
	}

	n, err := s.env.WriteFrame(dst)
	if err != nil {
		return 0, err
	}
	if n != len(dst) {
		return 0, fmt.Errorf("partial write: %d out of %d", n, len(dst))
	}

	return len(src), nil
}

func (s *writerImpl) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.env == nil {
		return errWriterClosed
	}

	err := s.writeSeekTable()
	s.env = nil
	return err
}

type encodeResult struct {
	buf   []byte
	entry seekTableEntry
}

func (s *writerImpl) writeManyEncoder(ctx context.Context, ch chan<- encodeResult, frame []byte) func() error {
	return func() error {
		dst, entry, err := s.encodeOne(frame)
		if err != nil {
			return fmt.Errorf("failed to encode frame: %w", err)
		}

		select {
		case <-ctx.Done():
		// Fulfill our promise
		case ch <- encodeResult{dst, entry}:
			close(ch)
		}

		return nil
	}
}

func (s *writerImpl) writeManyProducer(ctx context.Context, frameSource FrameSource, g *errgroup.Group, queue chan<- chan encodeResult) func() error {
	return func() error {
		for {
			frame, err := frameSource()
			if err != nil {
				return fmt.Errorf("frame source failed: %w", err)
			}
			if frame == nil {
				close(queue)
				return nil
			}
			if len(frame) == 0 {
				// Skip empty frames entirely.
				continue
			}

			// Put a channel on the queue as a sort of promise.
			// This is a nice trick to keep our results ordered, even when compression
			// completes out-of-order.
			ch := make(chan encodeResult, 1)
			select {
			case <-ctx.Done():
				return nil
			case queue <- ch:
			}

			g.Go(s.writeManyEncoder(ctx, ch, frame))
		}
	}
}

func (s *writerImpl) writeManyConsumer(ctx context.Context, callback func(uint32), queue <-chan chan encodeResult) func() error {
	return func() error {
		for {
			var ch <-chan encodeResult
			select {
			case <-ctx.Done():
				return nil
			case ch = <-queue:
			}
			if ch == nil {
				return nil
			}

			// Wait for the block to be complete
			var result encodeResult
			select {
			case <-ctx.Done():
				return nil
			case result = <-ch:
			}

			n, err := s.env.WriteFrame(result.buf)
			if err != nil {
				return fmt.Errorf("failed to write compressed data: %w", err)
			}
			if n != len(result.buf) {
				return fmt.Errorf("partial write: %d out of %d", n, len(result.buf))
			}
			s.frameEntries = append(s.frameEntries, result.entry)

			if callback != nil {
				callback(result.entry.DecompressedSize)
			}
		}
	}
}

func (s *writerImpl) WriteMany(ctx context.Context, frameSource FrameSource, options ...WriteManyOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.env == nil {
		return errWriterClosed
	}

	opts := writeManyOptions{concurrency: runtime.GOMAXPROCS(0)}
	for _, o := range options {
		if err := o(&opts); err != nil {
			return err // no wrap, these should be user-comprehensible
		}
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(opts.concurrency + 2) // reader and writer
	// Add extra room in the queue, so we can keep throughput high even if blocks finish out of order
	queue := make(chan chan encodeResult, opts.concurrency*2)
	g.Go(s.writeManyProducer(gCtx, frameSource, g, queue))
	g.Go(s.writeManyConsumer(gCtx, opts.writeCallback, queue))
	return g.Wait()
}

func (s *writerImpl) writeSeekTable() error {
	seekTableBytes, err := s.EndStream()
	if err != nil {
		return err
	}

	n, err := s.env.WriteSeekTable(seekTableBytes)
	if err != nil {
		return err
	}
	if n != len(seekTableBytes) {
		return fmt.Errorf("partial write: %d out of %d", n, len(seekTableBytes))
	}
	return nil
}
