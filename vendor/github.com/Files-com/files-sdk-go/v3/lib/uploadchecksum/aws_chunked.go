package uploadchecksum

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

type AWSChunkedEncoderOptions struct {
	Algorithm Algorithm
	ChunkSize int
}

type AWSChunkedEncoder struct {
	source     io.Reader
	state      State
	headerName string
	chunkSize  int
	buffer     []byte
	out        []byte
	chunk      []byte
	header     []byte
	pendingErr error
	finalSent  bool
}

func NewAWSChunkedEncoder(source io.Reader, options AWSChunkedEncoderOptions) (*AWSChunkedEncoder, error) {
	if source == nil {
		return nil, errors.New("uploadchecksum: source reader is nil")
	}
	algorithm := options.Algorithm.OrBestForPlatform()
	state, err := algorithm.NewState()
	if err != nil {
		return nil, err
	}
	headerName, err := algorithm.TrailerHeader()
	if err != nil {
		return nil, err
	}
	chunkSize := normalizeChunkSize(options.ChunkSize)
	return &AWSChunkedEncoder{
		source:     source,
		state:      state,
		headerName: headerName,
		chunkSize:  chunkSize,
		buffer:     make([]byte, chunkSize),
	}, nil
}

func (e *AWSChunkedEncoder) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	var total int
	for len(p) > 0 {
		if len(e.out) == 0 && len(e.chunk) == 0 {
			if total > 0 {
				return total, nil
			}
			if err := e.fill(); err != nil {
				return 0, err
			}
		}

		if len(e.out) > 0 {
			n := copy(p, e.out)
			e.out = e.out[n:]
			p = p[n:]
			total += n
			continue
		}

		n := copy(p, e.chunk)
		e.chunk = e.chunk[n:]
		p = p[n:]
		total += n
		if len(e.chunk) == 0 {
			e.out = awsChunkedCRLF
		}
	}
	return total, nil
}

func (e *AWSChunkedEncoder) fill() error {
	if e.finalSent {
		return io.EOF
	}
	if e.pendingErr != nil {
		err := e.pendingErr
		e.pendingErr = nil
		if err != io.EOF {
			return err
		}
		return e.writeFinalChunk()
	}

	n, err := e.source.Read(e.buffer)
	if n > 0 {
		_, _ = e.state.Write(e.buffer[:n])
		e.out = e.chunkHeader(n)
		e.chunk = e.buffer[:n]
		if err != nil {
			e.pendingErr = err
		}
		return nil
	}
	if err != nil {
		if err == io.EOF {
			return e.writeFinalChunk()
		}
		return err
	}
	return io.ErrNoProgress
}

func (e *AWSChunkedEncoder) chunkHeader(size int) []byte {
	e.header = strconv.AppendInt(e.header[:0], int64(size), 16)
	e.header = append(e.header, '\r', '\n')
	return e.header
}

func (e *AWSChunkedEncoder) writeFinalChunk() error {
	value, err := e.state.Encoded()
	if err != nil {
		return err
	}
	e.header = append(e.header[:0], "0\r\n"...)
	e.header = append(e.header, e.headerName...)
	e.header = append(e.header, ':')
	e.header = append(e.header, value...)
	e.header = append(e.header, "\r\n\r\n"...)
	e.out = e.header
	e.finalSent = true
	return nil
}

var awsChunkedCRLF = []byte("\r\n")

type AWSChunkedDecoderOptions struct {
	Algorithm Algorithm
}

type AWSChunkedDecoder struct {
	reader     *bufio.Reader
	state      State
	headerName string
	remaining  int64
	needCRLF   bool
	verified   bool
	trailer    string
}

func NewAWSChunkedDecoder(source io.Reader, options AWSChunkedDecoderOptions) (*AWSChunkedDecoder, error) {
	if source == nil {
		return nil, errors.New("uploadchecksum: source reader is nil")
	}
	algorithm := options.Algorithm.Normalize()
	state, err := algorithm.NewState()
	if err != nil {
		return nil, err
	}
	headerName, err := algorithm.TrailerHeader()
	if err != nil {
		return nil, err
	}
	return &AWSChunkedDecoder{
		reader:     bufio.NewReader(source),
		state:      state,
		headerName: headerName,
	}, nil
}

func NewAWSChunkedDecoderForHeaders(source io.Reader, headers http.Header) (*AWSChunkedDecoder, error) {
	algorithm, err := AlgorithmFromHeaders(headers)
	if err != nil {
		return nil, err
	}
	return NewAWSChunkedDecoder(source, AWSChunkedDecoderOptions{Algorithm: algorithm})
}

func (d *AWSChunkedDecoder) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if d.verified {
		return 0, io.EOF
	}
	for d.remaining == 0 {
		if d.needCRLF {
			if err := d.readCRLF(); err != nil {
				return 0, err
			}
			d.needCRLF = false
		}
		size, err := d.readChunkSize()
		if err != nil {
			return 0, err
		}
		if size == 0 {
			return 0, d.readAndVerifyTrailers()
		}
		d.remaining = size
	}

	limit := len(p)
	if int64(limit) > d.remaining {
		limit = int(d.remaining)
	}
	n, err := d.reader.Read(p[:limit])
	if n > 0 {
		_, _ = d.state.Write(p[:n])
		d.remaining -= int64(n)
		if d.remaining == 0 {
			d.needCRLF = true
		}
		return n, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrMalformedAWSChunked, err)
	}
	return 0, nil
}

func (d *AWSChunkedDecoder) Verify() error {
	_, err := io.Copy(io.Discard, d)
	if err == io.EOF {
		return nil
	}
	return err
}

func (d *AWSChunkedDecoder) TrailerValue() string {
	return d.trailer
}

func AWSChunkedEncodedLength(decodedLength int64, chunkSize int, algorithm Algorithm) (int64, error) {
	if decodedLength < 0 {
		return 0, errors.New("uploadchecksum: decoded length cannot be negative")
	}
	headerName, err := algorithm.Normalize().TrailerHeader()
	if err != nil {
		return 0, err
	}
	chunkSize = normalizeChunkSize(chunkSize)
	var encoded int64
	remaining := decodedLength
	for remaining > 0 {
		n := int64(chunkSize)
		if remaining < n {
			n = remaining
		}
		encoded += int64(len(strconv.FormatInt(n, 16))) + 2 + n + 2
		remaining -= n
	}
	valueLen := encodedChecksumLength(algorithm.Normalize())
	encoded += int64(len("0\r\n") + len(headerName) + len(":") + valueLen + len("\r\n\r\n"))
	return encoded, nil
}

func IsAWSChunked(headers http.Header) bool {
	for _, value := range headers.Values("Content-Encoding") {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), ContentEncodingAWSChunked) {
				return true
			}
		}
	}
	return false
}

func AlgorithmFromHeaders(headers http.Header) (Algorithm, error) {
	if trailer := headers.Get("x-amz-trailer"); trailer != "" {
		return AlgorithmFromTrailerHeader(trailer)
	}
	if algorithm := headers.Get("x-amz-sdk-checksum-algorithm"); algorithm != "" {
		normalized := Algorithm(algorithm).Normalize()
		if _, err := normalized.TrailerHeader(); err != nil {
			return "", err
		}
		return normalized, nil
	}
	return "", ErrUnsupportedAlgorithm
}

func AlgorithmFromTrailerHeader(value string) (Algorithm, error) {
	for _, trailer := range strings.Split(value, ",") {
		switch strings.ToLower(strings.TrimSpace(trailer)) {
		case "x-amz-checksum-crc32":
			return ChecksumAlgorithmCRC32, nil
		case "x-amz-checksum-crc32c":
			return ChecksumAlgorithmCRC32C, nil
		}
	}
	return "", ErrUnsupportedAlgorithm
}

func (d *AWSChunkedDecoder) readCRLF() error {
	cr, err := d.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("%w: missing chunk terminator", ErrMalformedAWSChunked)
	}
	lf, err := d.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("%w: missing chunk terminator", ErrMalformedAWSChunked)
	}
	if cr != '\r' || lf != '\n' {
		return fmt.Errorf("%w: invalid chunk terminator", ErrMalformedAWSChunked)
	}
	return nil
}

func (d *AWSChunkedDecoder) readChunkSize() (int64, error) {
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("%w: missing chunk size", ErrMalformedAWSChunked)
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if index := strings.IndexByte(line, ';'); index >= 0 {
		line = line[:index]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, fmt.Errorf("%w: empty chunk size", ErrMalformedAWSChunked)
	}
	size, err := strconv.ParseInt(line, 16, 64)
	if err != nil || size < 0 {
		return 0, fmt.Errorf("%w: invalid chunk size", ErrMalformedAWSChunked)
	}
	return size, nil
}

func (d *AWSChunkedDecoder) readAndVerifyTrailers() error {
	trailers, err := textproto.NewReader(d.reader).ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("%w: invalid trailers", ErrMalformedAWSChunked)
	}
	got := trailers.Get(d.headerName)
	if got == "" {
		return ErrMissingChecksum
	}
	want, err := d.state.Encoded()
	if err != nil {
		return err
	}
	got = strings.TrimSpace(got)
	d.trailer = got
	if got != want {
		return fmt.Errorf("%w: %s got %s want %s", ErrChecksumMismatch, d.headerName, got, want)
	}
	d.verified = true
	return io.EOF
}

func normalizeChunkSize(chunkSize int) int {
	if chunkSize <= 0 {
		return DefaultChunkSize
	}
	return chunkSize
}

func encodedChecksumLength(algorithm Algorithm) int {
	switch algorithm.Normalize() {
	case ChecksumAlgorithmCRC32:
		return 8
	case ChecksumAlgorithmCRC32C:
		return 8
	default:
		return 0
	}
}
