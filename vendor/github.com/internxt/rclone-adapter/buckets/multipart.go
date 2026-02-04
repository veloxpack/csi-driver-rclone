package buckets

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/internxt/rclone-adapter/config"
)

// chunkBufferPool reuses memory buffers for chunk encryption to reduce GC pressure
// Uses sync.Pool without pre-allocation - buffers are only created when actually needed
var chunkBufferPool = sync.Pool{
	New: func() any {
		// Return nil - allocate on-demand
		return nil
	},
}

// multipartUploadState holds the state for a single multipart upload session
type multipartUploadState struct {
	cfg            *config.Config
	plainIndex     string
	encIndex       string
	fileKey        []byte
	iv             []byte
	cipher         cipher.Stream
	totalSize      int64
	chunkSize      int64
	numParts       int64
	startResp      *StartUploadResp
	maxConcurrency int
	uploadId       string
	uuid           string
}

// encryptedChunk represents a chunk that has been encrypted and is ready for upload
type encryptedChunk struct {
	index      int
	data       []byte
	err        error
	bufferRefs []*[]byte
}

// uploadResult holds the result of a single chunk upload
type uploadResult struct {
	index int
	etag  string
	err   error
}

// newMultipartUploadState initializes encryption parameters and cipher for multipart upload
func newMultipartUploadState(cfg *config.Config, plainSize int64) (*multipartUploadState, error) {
	var ph [32]byte
	if _, err := rand.Read(ph[:]); err != nil {
		return nil, fmt.Errorf("cannot generate random index: %w", err)
	}

	plainIndex := hex.EncodeToString(ph[:])

	fileKey, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, plainIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file key: %w", err)
	}

	cipherStream, err := NewAES256CTRCipher(fileKey, iv)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	chunkSize := int64(config.DefaultChunkSize)
	numParts := (plainSize + chunkSize - 1) / chunkSize

	return &multipartUploadState{
		cfg:            cfg,
		plainIndex:     plainIndex,
		encIndex:       plainIndex,
		fileKey:        fileKey,
		iv:             iv,
		cipher:         cipherStream,
		totalSize:      plainSize,
		chunkSize:      chunkSize,
		numParts:       numParts,
		maxConcurrency: config.DefaultMaxConcurrency,
	}, nil
}

// executeMultipartUpload orchestrates the entire multipart upload process
func (s *multipartUploadState) executeMultipartUpload(ctx context.Context, reader io.Reader) (*MultipartShard, error) {
	specs := []UploadPartSpec{{Index: 0, Size: s.totalSize}}

	var err error
	s.startResp, err = StartUploadMultipart(ctx, s.cfg, s.cfg.Bucket, specs, int(s.numParts))
	if err != nil {
		return nil, fmt.Errorf("failed to start multipart upload: %w", err)
	}

	if len(s.startResp.Uploads) != 1 {
		return nil, fmt.Errorf("expected 1 upload entry, got %d", len(s.startResp.Uploads))
	}

	uploadInfo := s.startResp.Uploads[0]
	if len(uploadInfo.URLs) != int(s.numParts) {
		return nil, fmt.Errorf("expected %d URLs, got %d", s.numParts, len(uploadInfo.URLs))
	}

	s.uploadId = uploadInfo.UploadId
	s.uuid = uploadInfo.UUID

	completedParts, overallHash, err := s.encryptAndUploadPipelined(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt and upload chunks: %w", err)
	}

	return &MultipartShard{
		UUID:     s.uuid,
		Hash:     overallHash,
		UploadId: s.uploadId,
		Parts:    completedParts,
	}, nil
}

// encryptAndUploadPipelined encrypts chunks and uploads them concurrently
func (s *multipartUploadState) encryptAndUploadPipelined(ctx context.Context, reader io.Reader) ([]CompletedPart, string, error) {
	chunkChan := make(chan encryptedChunk, s.maxConcurrency)

	var uploadWg sync.WaitGroup

	results := make(chan uploadResult, s.numParts)

	semaphore := make(chan struct{}, s.maxConcurrency)

	// Compute hash: RIPEMD-160(SHA-256(encrypted_data)) - matches web client
	overallHasher := sha256.New()
	var hashMutex sync.Mutex
	var encryptErr error

	// Start encryption goroutine
	go func() {
		defer close(chunkChan)

		for i := int64(0); i < s.numParts; i++ {
			select {
			case <-ctx.Done():
				encryptErr = ctx.Err()
				chunkChan <- encryptedChunk{index: int(i), err: encryptErr}
				return
			default:
			}

			chunkSize := s.chunkSize
			if i == s.numParts-1 {
				chunkSize = s.totalSize - (i * s.chunkSize)
			}

			// Get buffers from pool or allocate at exact size needed
			var plainBufPtr, encryptedBufPtr *[]byte

			if poolBuf := chunkBufferPool.Get(); poolBuf != nil {
				plainBufPtr = poolBuf.(*[]byte)
				// Resize if buffer is too small
				if int64(cap(*plainBufPtr)) < chunkSize {
					buf := make([]byte, chunkSize)
					plainBufPtr = &buf
				}
			} else {
				buf := make([]byte, chunkSize)
				plainBufPtr = &buf
			}

			if poolBuf := chunkBufferPool.Get(); poolBuf != nil {
				encryptedBufPtr = poolBuf.(*[]byte)
				if int64(cap(*encryptedBufPtr)) < chunkSize {
					buf := make([]byte, chunkSize)
					encryptedBufPtr = &buf
				}
			} else {
				buf := make([]byte, chunkSize)
				encryptedBufPtr = &buf
			}

			plainChunk := (*plainBufPtr)[:chunkSize]
			n, err := io.ReadFull(reader, plainChunk)
			if err != nil && err != io.ErrUnexpectedEOF {
				chunkBufferPool.Put(plainBufPtr)
				chunkBufferPool.Put(encryptedBufPtr)
				encryptErr = fmt.Errorf("failed to read chunk %d: %w", i, err)
				chunkChan <- encryptedChunk{index: int(i), err: encryptErr}
				return
			}
			plainChunk = plainChunk[:n]

			encryptedData := (*encryptedBufPtr)[:len(plainChunk)]
			s.cipher.XORKeyStream(encryptedData, plainChunk)

			overallHasher.Write(encryptedData)

			chunkChan <- encryptedChunk{
				index:      int(i),
				data:       encryptedData,
				err:        nil,
				bufferRefs: []*[]byte{plainBufPtr, encryptedBufPtr},
			}
		}
	}()

	// Start upload workers
	for chunk := range chunkChan {
		if chunk.err != nil {
			for remaining := range chunkChan {
				for _, bufPtr := range remaining.bufferRefs {
					chunkBufferPool.Put(bufPtr)
				}
			}
			return nil, "", chunk.err
		}

		uploadWg.Add(1)
		go func(ch encryptedChunk) {
			defer uploadWg.Done()

			defer func() {
				for _, bufPtr := range ch.bufferRefs {
					chunkBufferPool.Put(bufPtr)
				}
			}()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			etag, err := s.uploadChunkWithRetry(ctx, ch.index, ch.data)

			results <- uploadResult{
				index: ch.index,
				etag:  etag,
				err:   err,
			}
		}(chunk)
	}

	go func() {
		uploadWg.Wait()
		close(results)
	}()

	parts := make([]CompletedPart, s.numParts)
	var firstError error

	resultsCollected := 0
	for result := range results {
		resultsCollected++
		if result.err != nil && firstError == nil {
			firstError = result.err
		}
		if result.err == nil {
			parts[result.index] = CompletedPart{
				PartNumber: result.index + 1,
				ETag:       result.etag,
			}
		}
	}

	if firstError != nil {
		return nil, "", firstError
	}

	hashMutex.Lock()
	// Compute RIPEMD-160(SHA-256) to match web client
	sha256Result := overallHasher.Sum(nil)
	overallHash := ComputeFileHash(sha256Result)
	hashMutex.Unlock()

	return parts, overallHash, nil
}

// uploadChunkWithRetry uploads a single chunk with exponential backoff retry
func (s *multipartUploadState) uploadChunkWithRetry(ctx context.Context, partIndex int, encryptedData []byte) (string, error) {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	uploadURL := s.startResp.Uploads[0].URLs[partIndex]

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context cancellation before retry
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return "", ctx.Err()
			}
		}

		result, err := Transfer(ctx, s.cfg, uploadURL, bytes.NewReader(encryptedData), int64(len(encryptedData)))
		if err == nil {
			return result.ETag, nil
		}

		lastErr = err
		if !isRetryableError(err) {
			break
		}
	}

	return "", fmt.Errorf("chunk %d upload failed after %d retries: %w", partIndex+1, maxRetries, lastErr)
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	if contains(errStr, "400") || contains(errStr, "401") || contains(errStr, "403") || contains(errStr, "404") {
		return false
	}

	return true
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
