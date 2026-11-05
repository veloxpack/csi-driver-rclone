package uploadchecksum

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultChunkSize = 64 * 1024

	ChecksumAlgorithmCRC32  Algorithm = "CRC32"
	ChecksumAlgorithmCRC32C Algorithm = "CRC32C"

	ContentEncodingAWSChunked             = "aws-chunked"
	ContentSHA256StreamingUnsignedTrailer = "STREAMING-UNSIGNED-PAYLOAD-TRAILER"
)

var (
	ErrUnsupportedAlgorithm = errors.New("uploadchecksum: unsupported checksum algorithm")
	ErrChecksumMismatch     = errors.New("uploadchecksum: checksum mismatch")
	ErrMissingChecksum      = errors.New("uploadchecksum: missing checksum trailer")
	ErrMalformedAWSChunked  = errors.New("uploadchecksum: malformed aws-chunked body")

	crc32Table  = crc32.IEEETable
	crc32cTable = crc32.MakeTable(crc32.Castagnoli)
)

type Algorithm string

func (a Algorithm) Normalize() Algorithm {
	switch Algorithm(strings.ToUpper(strings.TrimSpace(string(a)))) {
	case ChecksumAlgorithmCRC32:
		return ChecksumAlgorithmCRC32
	case "", ChecksumAlgorithmCRC32C:
		return ChecksumAlgorithmCRC32C
	default:
		return Algorithm(strings.ToUpper(strings.TrimSpace(string(a))))
	}
}

func (a Algorithm) OrBestForPlatform() Algorithm {
	if strings.TrimSpace(string(a)) == "" {
		return BestAlgorithmForPlatform()
	}
	return a.Normalize()
}

func (a Algorithm) TrailerHeader() (string, error) {
	switch a.Normalize() {
	case ChecksumAlgorithmCRC32:
		return "x-amz-checksum-crc32", nil
	case ChecksumAlgorithmCRC32C:
		return "x-amz-checksum-crc32c", nil
	default:
		return "", ErrUnsupportedAlgorithm
	}
}

func (a Algorithm) NewState() (State, error) {
	switch a.Normalize() {
	case ChecksumAlgorithmCRC32:
		return State{algorithm: ChecksumAlgorithmCRC32, table: crc32Table}, nil
	case ChecksumAlgorithmCRC32C:
		return State{algorithm: ChecksumAlgorithmCRC32C, table: crc32cTable}, nil
	default:
		return State{}, ErrUnsupportedAlgorithm
	}
}

type State struct {
	algorithm Algorithm
	table     *crc32.Table
	crc       uint32
}

func (s *State) Write(p []byte) (int, error) {
	switch s.algorithm {
	case ChecksumAlgorithmCRC32, ChecksumAlgorithmCRC32C:
		s.crc = crc32.Update(s.crc, s.table, p)
		return len(p), nil
	default:
		return 0, ErrUnsupportedAlgorithm
	}
}

func (s State) Encoded() (string, error) {
	switch s.algorithm {
	case ChecksumAlgorithmCRC32, ChecksumAlgorithmCRC32C:
		var digest [4]byte
		binary.BigEndian.PutUint32(digest[:], s.crc)
		return base64.StdEncoding.EncodeToString(digest[:]), nil
	default:
		return "", ErrUnsupportedAlgorithm
	}
}

func Checksum(p []byte, algorithm Algorithm) (string, error) {
	state, err := algorithm.NewState()
	if err != nil {
		return "", err
	}
	_, _ = state.Write(p)
	return state.Encoded()
}

type S3TrailerHeadersOptions struct {
	Algorithm     Algorithm
	DecodedLength int64
	ChunkSize     int
}

func S3TrailerHeaders(options S3TrailerHeadersOptions) (http.Header, int64, error) {
	algorithm := options.Algorithm.OrBestForPlatform()
	trailerHeader, err := algorithm.TrailerHeader()
	if err != nil {
		return nil, 0, err
	}
	encodedLength, err := AWSChunkedEncodedLength(options.DecodedLength, options.ChunkSize, algorithm)
	if err != nil {
		return nil, 0, err
	}

	headers := http.Header{}
	headers.Set("Content-Encoding", ContentEncodingAWSChunked)
	headers.Set("Content-Length", strconv.FormatInt(encodedLength, 10))
	headers.Set("x-amz-content-sha256", ContentSHA256StreamingUnsignedTrailer)
	headers.Set("x-amz-decoded-content-length", strconv.FormatInt(options.DecodedLength, 10))
	headers.Set("x-amz-sdk-checksum-algorithm", string(algorithm))
	headers.Set("x-amz-trailer", trailerHeader)
	return headers, encodedLength, nil
}
