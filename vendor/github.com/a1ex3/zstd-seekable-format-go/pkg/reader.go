package seekable

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"

	"github.com/cespare/xxhash/v2"
	"github.com/google/btree"

	"github.com/a1ex3/zstd-seekable-format-go/pkg/env"
)

type cachedFrame struct {
	m sync.Mutex

	offset uint64
	data   []byte
}

func (f *cachedFrame) replace(offset uint64, data []byte) {
	f.m.Lock()
	defer f.m.Unlock()

	f.offset = offset
	f.data = data
}

func (f *cachedFrame) get() (uint64, []byte) {
	f.m.Lock()
	defer f.m.Unlock()

	return f.offset, f.data
}

// readSeekerEnvImpl is the environment implementation for the io.ReadSeeker.
type readSeekerEnvImpl struct {
	rs io.ReadSeeker
	mu sync.Mutex
}

func (rs *readSeekerEnvImpl) GetFrameByIndex(index env.FrameOffsetEntry) (p []byte, err error) {
	p = make([]byte, index.CompSize)
	off := int64(index.CompOffset)

	switch v := rs.rs.(type) {
	case io.ReaderAt:
		_, err = v.ReadAt(p, off)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	default:
		rs.mu.Lock()
		defer rs.mu.Unlock()

		_, err = v.Seek(off, io.SeekStart)
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(rs.rs, p)
	}

	return
}

func (rs *readSeekerEnvImpl) ReadFooter() ([]byte, error) {
	n, err := rs.rs.Seek(-seekTableFooterOffset, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to: %d: %w", -seekTableFooterOffset, err)
	}

	buf := make([]byte, seekTableFooterOffset)
	_, err = io.ReadFull(rs.rs, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer at: %d: %w", n, err)
	}

	return buf, nil
}

func (rs *readSeekerEnvImpl) ReadSkipFrame(skippableFrameOffset int64) ([]byte, error) {
	n, err := rs.rs.Seek(-skippableFrameOffset, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to: %d: %w", -skippableFrameOffset, err)
	}

	buf := make([]byte, skippableFrameOffset)
	_, err = io.ReadFull(rs.rs, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read skippable frame header at: %d: %w", n, err)
	}
	return buf, nil
}

type readerImpl struct {
	dec   ZSTDDecoder
	index *btree.BTreeG[*env.FrameOffsetEntry]

	checksums bool

	offset int64

	numFrames int64
	endOffset int64

	env    env.REnvironment

	closed atomic.Bool

	// TODO: Add simple LRU cache.
	cachedFrame cachedFrame
}

var (
	_ io.Seeker   = (*readerImpl)(nil)
	_ io.Reader   = (*readerImpl)(nil)
	_ io.ReaderAt = (*readerImpl)(nil)
	_ io.Closer   = (*readerImpl)(nil)
)

type Reader interface {
	// Seek implements io.Seeker interface to randomly access data.
	// This method is NOT goroutine-safe and CAN NOT be called
	// concurrently since it modifies the underlying offset.
	Seek(offset int64, whence int) (int64, error)

	// Read implements io.Reader interface to sequentially access data.
	// This method is NOT goroutine-safe and CAN NOT be called
	// concurrently since it modifies the underlying offset.
	Read(p []byte) (n int, err error)

	// ReadAt implements io.ReaderAt interface to randomly access data.
	// This method is goroutine-safe and can be called concurrently ONLY if
	// the underlying reader supports io.ReaderAt interface.
	ReadAt(p []byte, off int64) (n int, err error)

	// Close implements io.Closer interface free up any resources.
	Close() error
}

// ZSTDDecoder is the decompressor.  Tested with github.com/klauspost/compress/zstd.
type ZSTDDecoder interface {
	DecodeAll(input, dst []byte) ([]byte, error)
}

// NewReader returns ZSTD stream reader that can be randomly accessed using uncompressed data offset.
// Ideally, passed io.ReadSeeker should implement io.ReaderAt interface.
func NewReader(rs io.ReadSeeker, decoder ZSTDDecoder, opts ...rOption) (Reader, error) {
	sr := readerImpl{
		dec: decoder,
	}

	for _, o := range opts {
		err := o(&sr)
		if err != nil {
			return nil, err
		}
	}

	if sr.env == nil {
		if rs == nil {
			return nil, fmt.Errorf("nil ReadSeeker and no custom environment supplied")
		}
		sr.env = &readSeekerEnvImpl{
			rs: rs,
		}
	}

	tree, last, err := sr.indexFooter()
	if err != nil {
		return nil, err
	}

	sr.index = tree
	if last != nil {
		sr.endOffset = int64(last.DecompOffset) + int64(last.DecompSize)
		sr.numFrames = last.ID + 1
	} else {
		sr.endOffset = 0
		sr.numFrames = 0
	}

	return &sr, nil
}

func (r *readerImpl) ReadAt(p []byte, off int64) (n int, err error) {
	for m := 0; n < len(p) && err == nil; n += m {
		_, m, err = r.read(p[n:], off+int64(n))
	}
	return
}

func (r *readerImpl) Read(p []byte) (n int, err error) {
	offset, n, err := r.read(p, r.offset)
	if err != nil {
		if errors.Is(err, io.EOF) {
			r.offset = r.endOffset
		}
		return
	}
	r.offset = offset
	return
}

func (r *readerImpl) Close() error {
	if r.closed.CompareAndSwap(false, true) {
		r.cachedFrame.replace(math.MaxUint64, nil)
		r.index = nil
	}
	return nil
}

func (r *readerImpl) read(dst []byte, off int64) (int64, int, error) {
	if r.closed.Load() {
		return 0, 0, fmt.Errorf("reader is closed")
	}

	if off >= r.endOffset {
		return 0, 0, io.EOF
	}
	if off < 0 {
		return 0, 0, fmt.Errorf("offset before the start of the file: %d", off)
	}

	index := r.GetIndexByDecompOffset(uint64(off))
	if index == nil {
		return 0, 0, fmt.Errorf("failed to get index by offset: %d", off)
	}
	if off < int64(index.DecompOffset) || off > int64(index.DecompOffset)+int64(index.DecompSize) {
		return 0, 0, fmt.Errorf("offset outside of index bounds: %d: min: %d, max: %d",
			off, int64(index.DecompOffset), int64(index.DecompOffset)+int64(index.DecompSize))
	}

	var decompressed []byte

	cachedOffset, cachedData := r.cachedFrame.get()
	if cachedOffset == index.DecompOffset && cachedData != nil {
		// fastpath
		decompressed = cachedData
	} else {
		// slowpath
		if index.CompSize > maxDecoderFrameSize {
			return 0, 0, fmt.Errorf("index.CompSize is too big: %d > %d",
				index.CompSize, maxDecoderFrameSize)
		}

		src, err := r.env.GetFrameByIndex(*index)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read compressed data at: %d, %w", index.CompOffset, err)
		}

		if len(src) != int(index.CompSize) {
			return 0, 0, fmt.Errorf("compressed size does not match index at: %d: expected: %d, index: %+v",
				off, len(src), index)
		}

		decompressed, err = r.dec.DecodeAll(src, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to decompress data data at: %d, %w", index.CompOffset, err)
		}

		if r.checksums {
			checksum := uint32((xxhash.Sum64(decompressed) << 32) >> 32)
			if index.Checksum != checksum {
				return 0, 0, fmt.Errorf("checksum verification failed at: %d: expected: %d, actual: %d",
					index.CompOffset, index.Checksum, checksum)
			}
		}
		r.cachedFrame.replace(index.DecompOffset, decompressed)
	}

	if len(decompressed) != int(index.DecompSize) {
		return 0, 0, fmt.Errorf("index corruption: len: %d, expected: %d", len(decompressed), int(index.DecompSize))
	}

	offsetWithinFrame := uint64(off) - index.DecompOffset

	size := uint64(len(decompressed)) - offsetWithinFrame
	if size > uint64(len(dst)) {
		size = uint64(len(dst))
	}

	copy(dst, decompressed[offsetWithinFrame:offsetWithinFrame+size])

	return off + int64(size), int(size), nil
}

func (r *readerImpl) Seek(offset int64, whence int) (int64, error) {
	newOffset := r.offset
	switch whence {
	case io.SeekCurrent:
		newOffset += offset
	case io.SeekStart:
		newOffset = offset
	case io.SeekEnd:
		newOffset = r.endOffset + offset
	default:
		return 0, fmt.Errorf("unknown whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("offset before the start of the file: %d (%d + %d)",
			newOffset, r.offset, offset)
	}

	r.offset = newOffset
	return r.offset, nil
}

func (r *readerImpl) indexFooter() (*btree.BTreeG[*env.FrameOffsetEntry], *env.FrameOffsetEntry, error) {
	// read seekTableFooter
	buf, err := r.env.ReadFooter()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read footer: %w", err)
	}
	if len(buf) < seekTableFooterOffset {
		return nil, nil, fmt.Errorf("footer is too small: %d", len(buf))
	}

	// parse seekTableFooter
	footer := seekTableFooter{}
	err = footer.UnmarshalBinary(buf[len(buf)-seekTableFooterOffset:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse footer %+v: %w", buf, err)
	}

	r.checksums = footer.SeekTableDescriptor.ChecksumFlag

	// read SeekTableEntries
	seekTableEntrySize := int64(8)
	if footer.SeekTableDescriptor.ChecksumFlag {
		seekTableEntrySize += 4
	}

	skippableFrameOffset := seekTableFooterOffset + seekTableEntrySize*int64(footer.NumberOfFrames)
	skippableFrameOffset += frameSizeFieldSize
	skippableFrameOffset += skippableMagicNumberFieldSize

	if skippableFrameOffset > maxDecoderFrameSize {
		return nil, nil, fmt.Errorf("frame offset is too big: %d > %d",
			skippableFrameOffset, maxDecoderFrameSize)
	}

	buf, err = r.env.ReadSkipFrame(skippableFrameOffset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read footer: %w", err)
	}

	if len(buf) < frameSizeFieldSize+skippableMagicNumberFieldSize+seekTableFooterOffset {
		return nil, nil, fmt.Errorf("skip frame is too small: %d", len(buf))
	}

	// parse SeekTableEntries
	magic := binary.LittleEndian.Uint32(buf[0:4])
	if magic != skippableFrameMagic+seekableTag {
		return nil, nil, fmt.Errorf("skippable frame magic mismatch %d vs %d",
			magic, skippableFrameMagic+seekableTag)
	}

	expectedFrameSize := int64(len(buf)) - frameSizeFieldSize - skippableMagicNumberFieldSize
	frameSize := int64(binary.LittleEndian.Uint32(buf[4:8]))
	if frameSize != expectedFrameSize {
		return nil, nil, fmt.Errorf("skippable frame size mismatch: expected: %d, actual: %d",
			expectedFrameSize, frameSize)
	}

	if frameSize > maxDecoderFrameSize {
		return nil, nil, fmt.Errorf("frame is too big: %d > %d", frameSize, maxDecoderFrameSize)
	}

	return r.indexSeekTableEntries(buf[8:len(buf)-seekTableFooterOffset], uint64(seekTableEntrySize))
}

func (r *readerImpl) indexSeekTableEntries(p []byte, entrySize uint64) (
	*btree.BTreeG[*env.FrameOffsetEntry], *env.FrameOffsetEntry, error,
) {
	if uint64(len(p))%entrySize != 0 {
		return nil, nil, fmt.Errorf("seek table size is not multiple of %d", entrySize)
	}

	// TODO: make fan-out tunable?
	t := btree.NewG(8, env.Less)
	entry := seekTableEntry{}
	var compOffset, decompOffset uint64

	var last *env.FrameOffsetEntry
	var i int64
	for indexOffset := uint64(0); indexOffset < uint64(len(p)); indexOffset += entrySize {
		err := entry.UnmarshalBinary(p[indexOffset : indexOffset+entrySize])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse entry %+v at: %d: %w",
				p[indexOffset:indexOffset+entrySize], indexOffset, err)
		}

		last = &env.FrameOffsetEntry{
			ID:           i,
			CompOffset:   compOffset,
			DecompOffset: decompOffset,
			CompSize:     entry.CompressedSize,
			DecompSize:   entry.DecompressedSize,
			Checksum:     entry.Checksum,
		}
		t.ReplaceOrInsert(last)
		compOffset += uint64(entry.CompressedSize)
		decompOffset += uint64(entry.DecompressedSize)
		i++
	}

	return t, last, nil
}
