package file

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib/uploadchecksum"
)

const (
	uploadV2ChecksumTrailerAlgorithm uploadchecksum.Algorithm = ""
	uploadV2ChecksumTrailerChunkSize                          = uploadchecksum.DefaultChunkSize
)

var errUploadV2ChecksumTrailerUnsignedUploadURL = errors.New("upload v2 checksum trailer requires upload URL signed for AWS trailer headers")

var uploadV2ChecksumTrailerRequiredSignedHeaders = []string{
	"content-encoding",
	"x-amz-content-sha256",
	"x-amz-decoded-content-length",
	"x-amz-sdk-checksum-algorithm",
	"x-amz-trailer",
}

type uploadV2ChecksumTrailer struct {
	enabled       bool
	headers       http.Header
	encodedLength int64
	algorithm     uploadchecksum.Algorithm
}

func (e *uploadV2Engine) prepareChecksumTrailer(part *uploadV2Part) (uploadV2ChecksumTrailer, error) {
	if e == nil || e.u == nil || !e.u.Config.FeatureFlags[files_sdk.FeatureFlagUploadV2ChecksumTrailer] {
		return uploadV2ChecksumTrailer{}, nil
	}
	if !e.checksumTrailerEnabled {
		e.logChecksumTrailerSkipped(part, e.checksumTrailerSkipReason)
		return uploadV2ChecksumTrailer{}, nil
	}
	if part == nil || part.reader == nil {
		e.logChecksumTrailerSkipped(part, "missing_part_reader")
		return uploadV2ChecksumTrailer{}, nil
	}
	if !uploadV2S3ChecksumTrailerSigned(part.upload.UploadUri) {
		e.logChecksumTrailerSkipped(part, "part_url_required_headers_not_signed")
		return uploadV2ChecksumTrailer{}, errUploadV2ChecksumTrailerUnsignedUploadURL
	}

	headers, encodedLength, err := uploadchecksum.S3TrailerHeaders(uploadchecksum.S3TrailerHeadersOptions{
		Algorithm:     uploadV2ChecksumTrailerAlgorithm,
		DecodedLength: int64(part.reader.Len()),
		ChunkSize:     uploadV2ChecksumTrailerChunkSize,
	})
	if err != nil {
		return uploadV2ChecksumTrailer{}, err
	}
	algorithm, err := uploadchecksum.AlgorithmFromHeaders(headers)
	if err != nil {
		return uploadV2ChecksumTrailer{}, err
	}
	trailer := uploadV2ChecksumTrailer{
		enabled:       true,
		headers:       headers,
		encodedLength: encodedLength,
		algorithm:     algorithm,
	}
	e.logChecksumTrailerEnabled(part, trailer)
	return trailer, nil
}

func uploadV2ChecksumTrailerDecision(u *uploadIO, plan uploadV2PartPlan) (bool, string) {
	if u == nil || u.Client == nil || !u.Config.FeatureFlags[files_sdk.FeatureFlagUploadV2ChecksumTrailer] {
		return false, ""
	}
	if !uploadV2ChecksumTrailerSupportedDestination(plan.target) {
		return false, "unsupported_destination"
	}
	if !uploadV2S3ChecksumTrailerSigned(u.FileUploadPart.UploadUri) {
		return false, "required_headers_not_signed"
	}
	return true, ""
}

func uploadV2ChecksumTrailerSupportedDestination(target TransferV2TargetClass) bool {
	return target == uploadV2TargetS3
}

func (t uploadV2ChecksumTrailer) newReader(source io.Reader) (io.Reader, error) {
	if !t.enabled {
		return source, nil
	}
	encoder, err := uploadchecksum.NewAWSChunkedEncoder(source, uploadchecksum.AWSChunkedEncoderOptions{
		Algorithm: t.algorithm,
		ChunkSize: uploadV2ChecksumTrailerChunkSize,
	})
	if err != nil {
		return nil, err
	}
	return encoder, nil
}

func uploadV2S3ChecksumTrailerSigned(uploadURI string) bool {
	parsed, err := url.Parse(uploadURI)
	if err != nil {
		return false
	}
	signedHeaders := parsed.Query().Get("X-Amz-SignedHeaders")
	if signedHeaders == "" {
		signedHeaders = parsed.Query().Get("x-amz-signedheaders")
	}
	if signedHeaders == "" {
		return false
	}

	seen := map[string]struct{}{}
	for _, header := range strings.Split(signedHeaders, ";") {
		header = strings.ToLower(strings.TrimSpace(header))
		if header != "" {
			seen[header] = struct{}{}
		}
	}
	for _, header := range uploadV2ChecksumTrailerRequiredSignedHeaders {
		if _, ok := seen[header]; !ok {
			return false
		}
	}
	return true
}

func (e *uploadV2Engine) logChecksumTrailerEnabled(part *uploadV2Part, trailer uploadV2ChecksumTrailer) {
	if part == nil || part.number != 1 {
		return
	}
	e.u.logUploadV2(map[string]any{
		"timestamp":              time.Now(),
		"event":                  "upload v2 checksum trailer enabled",
		"checksum_algorithm":     string(trailer.algorithm),
		"checksum_chunk_size":    uploadV2ChecksumTrailerChunkSize,
		"checksum_encoded_bytes": trailer.encodedLength,
	})
}

func (e *uploadV2Engine) logChecksumTrailerSkipped(part *uploadV2Part, reason string) {
	if part == nil || part.number != 1 {
		return
	}
	e.u.logUploadV2(map[string]any{
		"timestamp": time.Now(),
		"event":     "upload v2 checksum trailer skipped",
		"reason":    reason,
		"target":    string(e.plan.target),
	})
}

type uploadV2ReadCloser struct {
	io.Reader
	closer io.Closer
}

func (r uploadV2ReadCloser) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}
