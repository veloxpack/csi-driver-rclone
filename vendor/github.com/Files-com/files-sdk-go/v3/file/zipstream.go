package file

import (
	"bufio"
	"compress/flate"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	zipLocalFileHeaderSignature   = 0x04034b50
	zipCentralDirectorySignature  = 0x02014b50
	zipEndOfCentralDirSignature   = 0x06054b50
	zip64EndOfCentralDirSignature = 0x06064b50
	zipDataDescriptorSignature    = 0x08074b50
	zipMethodStore                = 0
	zipMethodDeflate              = 8
	zipFlagDataDescriptor         = 1 << 3
	zipMaxUint32                  = 1<<32 - 1
)

type zipStream struct {
	reader *bufio.Reader
	done   bool
}

type zipStreamEntryHeader struct {
	name               string
	flags              uint16
	method             uint16
	crc32              uint32
	compressedSize     uint32
	uncompressedSize   uint32
	hasDataDescriptor  bool
	hasTrustedMetadata bool
}

type zipStreamEntryResult struct {
	crc32            uint32
	compressedSize   uint32
	uncompressedSize uint32
}

func newZipStream(r io.Reader) *zipStream {
	return &zipStream{reader: bufio.NewReader(r)}
}

func (z *zipStream) nextHeader() (zipStreamEntryHeader, bool, error) {
	if z.done {
		return zipStreamEntryHeader{}, false, nil
	}

	signature, err := z.readUint32()
	if err != nil {
		return zipStreamEntryHeader{}, false, fmt.Errorf("zip batch: stream missing central directory: %w", err)
	}

	switch signature {
	case zipLocalFileHeaderSignature:
		return z.readLocalHeader()
	case zipCentralDirectorySignature, zipEndOfCentralDirSignature:
		if err := z.readCentralDirectory(signature); err != nil {
			return zipStreamEntryHeader{}, false, err
		}
		z.done = true
		return zipStreamEntryHeader{}, false, nil
	case zip64EndOfCentralDirSignature:
		return zipStreamEntryHeader{}, false, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip64 central directory is not supported")}
	default:
		return zipStreamEntryHeader{}, false, fmt.Errorf("zip batch: unexpected zip signature 0x%x", signature)
	}
}

func (z *zipStream) readLocalHeader() (zipStreamEntryHeader, bool, error) {
	fixed := make([]byte, 26)
	if _, err := io.ReadFull(z.reader, fixed); err != nil {
		return zipStreamEntryHeader{}, false, err
	}

	flags := binary.LittleEndian.Uint16(fixed[2:4])
	method := binary.LittleEndian.Uint16(fixed[4:6])
	crc := binary.LittleEndian.Uint32(fixed[10:14])
	compressedSize := binary.LittleEndian.Uint32(fixed[14:18])
	uncompressedSize := binary.LittleEndian.Uint32(fixed[18:22])
	nameLen := binary.LittleEndian.Uint16(fixed[22:24])
	extraLen := binary.LittleEndian.Uint16(fixed[24:26])

	if compressedSize == zipMaxUint32 || uncompressedSize == zipMaxUint32 {
		return zipStreamEntryHeader{}, false, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip64 entry sizes are not supported")}
	}

	name := make([]byte, nameLen)
	if _, err := io.ReadFull(z.reader, name); err != nil {
		return zipStreamEntryHeader{}, false, err
	}
	if err := z.skip(int64(extraLen)); err != nil {
		return zipStreamEntryHeader{}, false, err
	}

	hasDataDescriptor := flags&zipFlagDataDescriptor != 0
	return zipStreamEntryHeader{
		name:               string(name),
		flags:              flags,
		method:             method,
		crc32:              crc,
		compressedSize:     compressedSize,
		uncompressedSize:   uncompressedSize,
		hasDataDescriptor:  hasDataDescriptor,
		hasTrustedMetadata: !hasDataDescriptor,
	}, true, nil
}

func (z *zipStream) extractEntry(ctx context.Context, header zipStreamEntryHeader, out io.Writer, onBytes func(int64)) (zipStreamEntryResult, error) {
	crc := crc32.NewIEEE()
	counter := &zipCountingByteReader{reader: z.reader}
	var data io.Reader

	switch header.method {
	case zipMethodDeflate:
		inflater := flate.NewReader(counter)
		defer inflater.Close()
		data = inflater
	case zipMethodStore:
		if header.hasDataDescriptor {
			return zipStreamEntryResult{}, zipBatchTripwireError{error: fmt.Errorf("zip batch: stored zip entries with data descriptors are not supported")}
		}
		data = io.LimitReader(counter, int64(header.uncompressedSize))
	default:
		return zipStreamEntryResult{}, zipBatchTripwireError{error: fmt.Errorf("zip batch: unsupported zip compression method %d", header.method)}
	}

	uncompressedSize, err := copyZipEntry(ctx, data, out, crc, onBytes)
	if err != nil {
		return zipStreamEntryResult{}, err
	}

	result := zipStreamEntryResult{
		crc32:            crc.Sum32(),
		compressedSize:   uint32(counter.n),
		uncompressedSize: uint32(uncompressedSize),
	}
	if int64(result.compressedSize) != counter.n || int64(result.uncompressedSize) != uncompressedSize {
		return zipStreamEntryResult{}, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry exceeds supported 32-bit size")}
	}

	expected := zipStreamEntryResult{
		crc32:            header.crc32,
		compressedSize:   header.compressedSize,
		uncompressedSize: header.uncompressedSize,
	}
	if header.hasDataDescriptor {
		expected, err = z.readDataDescriptor(result)
		if err != nil {
			return zipStreamEntryResult{}, err
		}
	}

	if result.crc32 != expected.crc32 {
		return zipStreamEntryResult{}, fmt.Errorf("zip batch: zip entry %q crc mismatch", header.name)
	}
	if result.compressedSize != expected.compressedSize {
		return zipStreamEntryResult{}, fmt.Errorf("zip batch: zip entry %q compressed size mismatch", header.name)
	}
	if result.uncompressedSize != expected.uncompressedSize {
		return zipStreamEntryResult{}, fmt.Errorf("zip batch: zip entry %q size mismatch", header.name)
	}

	return result, nil
}

func copyZipEntry(ctx context.Context, data io.Reader, out io.Writer, crc io.Writer, onBytes func(int64)) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, readErr := data.Read(buf)
		if n > 0 {
			if _, err := crc.Write(buf[:n]); err != nil {
				return written, err
			}
			wn, writeErr := out.Write(buf[:n])
			if wn > 0 {
				written += int64(wn)
				if onBytes != nil {
					onBytes(int64(wn))
				}
			}
			if writeErr != nil {
				return written, writeErr
			}
			if wn != n {
				return written, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
}

func (z *zipStream) readDataDescriptor(actual zipStreamEntryResult) (zipStreamEntryResult, error) {
	first, err := z.readUint32()
	if err != nil {
		return zipStreamEntryResult{}, err
	}
	second, err := z.readUint32()
	if err != nil {
		return zipStreamEntryResult{}, err
	}
	third, err := z.readUint32()
	if err != nil {
		return zipStreamEntryResult{}, err
	}

	noSignature := zipStreamEntryResult{
		crc32:            first,
		compressedSize:   second,
		uncompressedSize: third,
	}
	if zipStreamDescriptorMatches(noSignature, actual) {
		return noSignature, nil
	}

	if first == zipDataDescriptorSignature {
		fourth, err := z.readUint32()
		if err != nil {
			return zipStreamEntryResult{}, err
		}
		withSignature := zipStreamEntryResult{
			crc32:            second,
			compressedSize:   third,
			uncompressedSize: fourth,
		}
		if zipStreamDescriptorMatches(withSignature, actual) {
			return withSignature, nil
		}
		return zipStreamEntryResult{}, fmt.Errorf("zip batch: zip data descriptor mismatch")
	}

	return noSignature, nil
}

func zipStreamDescriptorMatches(candidate zipStreamEntryResult, actual zipStreamEntryResult) bool {
	return candidate.crc32 == actual.crc32 &&
		candidate.compressedSize == actual.compressedSize &&
		candidate.uncompressedSize == actual.uncompressedSize &&
		candidate.compressedSize != zipMaxUint32 &&
		candidate.uncompressedSize != zipMaxUint32
}

func (z *zipStream) readCentralDirectory(firstSignature uint32) error {
	signature := firstSignature
	for signature == zipCentralDirectorySignature {
		fixed := make([]byte, 42)
		if _, err := io.ReadFull(z.reader, fixed); err != nil {
			return err
		}
		nameLen := binary.LittleEndian.Uint16(fixed[24:26])
		extraLen := binary.LittleEndian.Uint16(fixed[26:28])
		commentLen := binary.LittleEndian.Uint16(fixed[28:30])
		if err := z.skip(int64(nameLen) + int64(extraLen) + int64(commentLen)); err != nil {
			return err
		}
		next, err := z.readUint32()
		if err != nil {
			return fmt.Errorf("zip batch: stream missing end of central directory: %w", err)
		}
		signature = next
	}

	if signature == zip64EndOfCentralDirSignature {
		return zipBatchTripwireError{error: fmt.Errorf("zip batch: zip64 central directory is not supported")}
	}
	if signature != zipEndOfCentralDirSignature {
		return fmt.Errorf("zip batch: unexpected zip central directory signature 0x%x", signature)
	}

	eocd := make([]byte, 18)
	if _, err := io.ReadFull(z.reader, eocd); err != nil {
		return err
	}
	commentLen := binary.LittleEndian.Uint16(eocd[16:18])
	if err := z.skip(int64(commentLen)); err != nil {
		return err
	}
	if _, err := z.reader.Peek(1); err != io.EOF {
		if err == nil {
			return fmt.Errorf("zip batch: zip stream has trailing data")
		}
		return err
	}
	return nil
}

func (z *zipStream) readUint32() (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(z.reader, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func (z *zipStream) skip(n int64) error {
	if n == 0 {
		return nil
	}
	_, err := io.CopyN(io.Discard, z.reader, n)
	return err
}

type zipCountingByteReader struct {
	reader *bufio.Reader
	n      int64
}

func (r *zipCountingByteReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.n += int64(n)
	return n, err
}

func (r *zipCountingByteReader) ReadByte() (byte, error) {
	// flate uses ReadByte to avoid over-reading past compressed data; removing
	// this breaks stream positioning before the data descriptor.
	b, err := r.reader.ReadByte()
	if err == nil {
		r.n++
	}
	return b, err
}
