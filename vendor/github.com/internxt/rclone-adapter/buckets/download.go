package buckets

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/errors"
)

// ShardInfo mirrors the per‑shard info returned by /files/{fileID}/info
type ShardInfo struct {
	Index int    `json:"index"`
	Hash  string `json:"hash"`
	URL   string `json:"url"`
}

// BucketFileInfo is the metadata returned by GET /buckets/{bucketID}/files/{fileID}/info
type BucketFileInfo struct {
	Bucket   string      `json:"bucket"`
	Index    string      `json:"index"`
	Size     int64       `json:"size"`
	Version  int         `json:"version"`
	Created  string      `json:"created"`
	Renewal  string      `json:"renewal"`
	Mimetype string      `json:"mimetype"`
	Filename string      `json:"filename"`
	ID       string      `json:"id"`
	Shards   []ShardInfo `json:"shards"`
}

// GetBucketFileInfo calls the correct /info endpoint and parses its JSON.
func GetBucketFileInfo(ctx context.Context, cfg *config.Config, bucketID, fileID string) (*BucketFileInfo, error) {
	url := cfg.Endpoints.Network().FileInfo(bucketID, fileID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get bucket file info request: %w", err)
	}
	req.Header.Set("Authorization", cfg.BasicAuthHeader)
	req.Header.Set("internxt-version", "1.0")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get bucket file info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.NewHTTPError(resp, "get bucket file info")
	}

	var info BucketFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode bucket file info response: %w", err)
	}
	return &info, nil
}

// DownloadFile downloads and decrypts the first shard of the given file.
func DownloadFile(ctx context.Context, cfg *config.Config, fileID, destPath string) error {
	// 1) fetch file info from the bucket API
	info, err := GetBucketFileInfo(ctx, cfg, cfg.Bucket, fileID)
	if err != nil {
		return fmt.Errorf("failed to get bucket file info: %w", err)
	}

	if info.Size == 0 {
		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create empty file %s: %w", destPath, err)
		}
		return out.Close()
	}

	if len(info.Shards) == 0 {
		return fmt.Errorf("no shards found for file %s", fileID)
	}
	shard := info.Shards[0]

	// 2) derive fileKey+iv using the stored index (hex of random index)
	key, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, info.Index)
	if err != nil {
		return fmt.Errorf("failed to generate file key: %w", err)
	}

	// 3) GET the encrypted shard directly from its presigned URL
	req, err := http.NewRequestWithContext(ctx, "GET", shard.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.NewHTTPError(resp, "shard download")
	}

	// 4) Set up hash computation for encrypted data stream
	// Hash algorithm: RIPEMD-160(SHA-256(encrypted_data))
	var readStream io.Reader = resp.Body
	var sha256Hasher io.Writer
	if !cfg.SkipHashValidation {
		sha256Hasher = sha256.New()
		readStream = io.TeeReader(resp.Body, sha256Hasher)
	}

	// 5) wrap in AES‑CTR decryptor
	decReader, err := DecryptReader(readStream, key, iv)
	if err != nil {
		return fmt.Errorf("failed to create decrypt reader: %w", err)
	}

	// 6) write plaintext to file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, decReader); err != nil {
		return fmt.Errorf("failed to write decrypted data to file: %w", err)
	}

	// 7) Validate hash after download completes
	if !cfg.SkipHashValidation {
		// Compute RIPEMD-160(SHA-256(encrypted_data)) to match web client
		sha256Result := sha256Hasher.(interface{ Sum([]byte) []byte }).Sum(nil)
		computedHash := ComputeFileHash(sha256Result)

		if computedHash != shard.Hash {
			// Clean up corrupted file
			out.Close()
			os.Remove(destPath)
			return fmt.Errorf("hash mismatch for file %s: expected %s, got %s (file removed)",
				fileID, shard.Hash, computedHash)
		}
	}

	return nil
}

// DownloadFileStream returns a ReadCloser that streams the decrypted contents
// of the file with the given UUID. The caller must close the returned ReadCloser.
// It takes an optional range header in the format of either "bytes=100-199" or "bytes=100-".
func DownloadFileStream(ctx context.Context, cfg *config.Config, fileUUID string, optionalRange ...string) (io.ReadCloser, error) {
	rangeValue := ""
	if len(optionalRange) > 0 {
		rangeValue = optionalRange[0]
	}

	// 1) Fetch file info (including shards and index)
	info, err := GetBucketFileInfo(ctx, cfg, cfg.Bucket, fileUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket file info: %w", err)
	}

	if info.Size == 0 {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}

	if len(info.Shards) == 0 {
		return nil, fmt.Errorf("no shards found for file %s", fileUUID)
	}
	shard := info.Shards[0]

	// 2) Derive fileKey and IV from the stored index
	key, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, info.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file key: %w", err)
	}

	// 3) Calculate the IV for the requested range
	if rangeValue != "" {
		startByte, endByte, err := getStartByteAndEndByte(rangeValue)
		if err != nil {
			return nil, fmt.Errorf("invalid range: %w", err)
		}

		// Ensure AES block alignment for correct decryption
		// Find the nearest block and call this function again with the adjusted range, then discard the unwanted bytes before returning
		if offset := startByte % 16; offset != 0 {
			alignedStart := startByte - offset
			var adjustedRange string
			if endByte == -1 {
				adjustedRange = fmt.Sprintf("bytes=%d-", alignedStart)
			} else {
				adjustedRange = fmt.Sprintf("bytes=%d-%d", alignedStart, endByte)
			}

			stream, err := DownloadFileStream(ctx, cfg, fileUUID, adjustedRange)
			if err != nil {
				return nil, fmt.Errorf("failed to download aligned stream: %w", err)
			}

			// Discard unwanted bytes and return the requested range exactly
			if _, err := io.CopyN(io.Discard, stream, int64(offset)); err != nil {
				stream.Close()
				return nil, fmt.Errorf("failed to discard offset bytes: %w", err)
			}
			return stream, nil
		}

		adjustIV(iv, startByte/16)
	}

	// 4) Download the encrypted shard, include the Range header if any
	req, err := http.NewRequestWithContext(ctx, "GET", shard.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if rangeValue != "" {
		req.Header.Set("Range", rangeValue)
	}

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute download stream request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httpErr := errors.NewHTTPError(resp, "shard download stream")
		resp.Body.Close()
		return nil, httpErr
	}

	// 5) Set up hash computation for full downloads only (range requests skip validation)
	// Hash algorithm: RIPEMD-160(SHA-256(encrypted_data)) - matches web client
	var readStream io.Reader = resp.Body

	if rangeValue == "" && !cfg.SkipHashValidation {
		// Full download - validate hash on Close()
		sha256Hasher := sha256.New()
		readStream = io.TeeReader(resp.Body, sha256Hasher)

		decReader, err := DecryptReader(readStream, key, iv)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create decrypt reader: %w", err)
		}

		// Return validating reader that checks hash when closed
		return &hashValidatingReader{
			Reader:       decReader,
			body:         resp.Body,
			sha256Hasher: sha256Hasher,
			expectedHash: shard.Hash,
			fileUUID:     fileUUID,
		}, nil
	}

	// Range request or validation skipped - no hash check
	decReader, err := DecryptReader(readStream, key, iv)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to create decrypt reader: %w", err)
	}

	// 6) Return a ReadCloser that closes the HTTP body when closed
	return struct {
		io.Reader
		io.Closer
	}{Reader: decReader, Closer: resp.Body}, nil
}

// This will return the startByte and endByte of a range header in these formats: "bytes=100-199" or "bytes=100-"
// In the case of the "bytes=100-" the returned endByte will be -1.
// Formats like "bytes=-200" and "bytes=0-99,200-299" are not supported.
func getStartByteAndEndByte(rangeHeader string) (int, int, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	rangePart := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangePart, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	startByte, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start byte in Range header: %w", err)
	}

	// Handle optional endByte
	if parts[1] == "" {
		return startByte, -1, nil
	}

	endByte, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end byte in Range header: %w", err)
	}

	return startByte, endByte, nil
}

// adjustIV increments the IV based on the given block index.
func adjustIV(iv []byte, blockIndex int) {
	for i := 0; i < blockIndex; i++ {
		for j := len(iv) - 1; j >= 0; j-- {
			iv[j]++
			if iv[j] != 0 {
				break
			}
		}
	}
}

// hashValidatingReader wraps a reader and validates the hash on Close().
// It computes RIPEMD-160(SHA-256(encrypted_data)) and compares it
// to the expected hash when the stream is closed
type hashValidatingReader struct {
	io.Reader
	body         io.Closer
	sha256Hasher io.Writer
	expectedHash string
	fileUUID     string
	validated    bool
}

// Close closes the underlying body and validates the computed hash.
// NOTE: The hash is only valid if the ENTIRE stream was read before calling Close().
// If only partial data was read, the hash will be incorrect.
func (h *hashValidatingReader) Close() error {
	if h.body == nil {
		return nil
	}

	// Validate hash BEFORE closing body (in case we need to drain remaining data)
	if !h.validated {
		h.validated = true

		// IMPORTANT: Drain any remaining data in the stream to ensure complete hash
		// This happens if the caller didn't read the entire stream
		remaining, err := io.Copy(io.Discard, h.Reader)
		if err != nil {
			h.body.Close()
			return fmt.Errorf("failed to drain remaining stream data: %w", err)
		}

		// Compute RIPEMD-160(SHA-256(encrypted_data)) to match web client
		sha256Result := h.sha256Hasher.(interface{ Sum([]byte) []byte }).Sum(nil)
		computedHash := ComputeFileHash(sha256Result)

		if computedHash != h.expectedHash {
			h.body.Close()
			return fmt.Errorf("hash mismatch for file %s: expected %s, got %s (remaining bytes: %d)",
				h.fileUUID, h.expectedHash, computedHash, remaining)
		}
	}

	// Close underlying body
	return h.body.Close()
}
