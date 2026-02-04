package filen

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"github.com/zeebo/blake3"
	"hash"
	"io"
)

// fetchAndDecryptChunk downloads and decrypts a single chunk of a file.
// It retrieves the encrypted chunk from the Filen servers and decrypts it
// using the file's encryption key.
func (api *Filen) fetchAndDecryptChunk(ctx context.Context, file *types.File, chunkIndex int64) ([]byte, error) {
	// could potentially be optimized by accepting a []byte buffer to reuse
	encryptedBytes, err := api.Client.DownloadFileChunk(ctx, file.UUID, file.Region, file.Bucket, chunkIndex)
	if err != nil {
		return nil, fmt.Errorf("downloading chunk %d: %w", chunkIndex, err)
	}
	var decryptedBytes []byte
	if file.Version == 1 {
		decryptedBytes, err = crypto.V1Decrypt(encryptedBytes, file.EncryptionKey.Bytes[:])
	} else {
		decryptedBytes, err = file.EncryptionKey.DecryptData(encryptedBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("decrypting chunk %d: %w", chunkIndex, err)
	}
	return decryptedBytes, nil
}

// ChunkedReader implements io.Reader for sequential chunked file downloads.
// It provides efficient streaming access to files stored in Filen cloud storage
// by downloading chunks in parallel and validating file integrity.
type ChunkedReader struct {
	ctx          context.Context
	file         *types.File // The file being downloaded
	api          *Filen      // API client to use for downloading
	hasher       hash.Hash   // For calculating file hash during download
	currentChunk []byte
	firstIndex   int64 // Starting index to read from
	totalRead    int64 // Total bytes read, -1 if started with an offset
	totalToRead  int64
}

// newChunkedReaderWithOffset creates a new ChunkedReader for sequential reading,
// starting at the specified byte offset and reading up to the specified limit.
// If limit is -1, reads to the end of the file.
func newChunkedReaderWithOffset(ctx context.Context, api *Filen, file *types.File, offset int64, limit int64) *ChunkedReader {
	if limit == -1 {
		limit = file.Size
	} else {
		limit = min(limit, file.Size)
	}

	return &ChunkedReader{
		ctx:          ctx,
		file:         file,
		api:          api,
		hasher:       blake3.New(),
		currentChunk: nil,
		firstIndex:   offset,
		totalRead:    0,
		totalToRead:  limit - offset,
	}
}

// newChunkedReader creates a new ChunkedReader that reads the entire file
// from the beginning.
func newChunkedReader(ctx context.Context, api *Filen, file *types.File) *ChunkedReader {
	return newChunkedReaderWithOffset(ctx, api, file, 0, -1)
}

// Read implements the io.Reader interface, optimized for sequential reading.
// It reads data from the file in chunks, handling prefetching of future chunks
// and validating the file hash as data is read.
func (r *ChunkedReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Check for fetch errors

	read := 0
	for read < len(p) {
		// Check if we've reached EOF
		if r.totalRead >= r.totalToRead {
			err = io.EOF
			break
		}

		toRead := int(min(int64(len(p)-read), r.totalToRead-r.totalRead))

		if r.currentChunk == nil {
			r.currentChunk, err = r.api.fetchAndDecryptChunk(r.ctx, r.file, (r.firstIndex+r.totalRead)/ChunkSize)

			if err != nil {
				err = fmt.Errorf("failed to fetch chunk: %w", err)
				break
			}
			if r.totalRead == 0 {
				r.currentChunk = r.currentChunk[r.firstIndex%ChunkSize:]
			}
		}

		maxToCopy := min(toRead, len(r.currentChunk))

		copiedLen := copy(p[read:read+maxToCopy], r.currentChunk[:maxToCopy])

		if copiedLen == len(r.currentChunk) {
			r.currentChunk = nil
		} else {
			r.currentChunk = r.currentChunk[copiedLen:]
		}

		read += copiedLen
		r.totalRead += int64(copiedLen)
	}

	if r.firstIndex == 0 {
		_, err := r.hasher.Write(p[:read])
		if err != nil {
			return read, err
		}
	}
	return read, err
}

// Close cleans up resources used by the reader and verifies the file hash
// if the entire file was read. It should be called when done with the reader
// to ensure proper cleanup and validation.
func (r *ChunkedReader) Close() error {
	if r.totalRead != r.file.Size {
		// incomplete read
		return nil
	}
	if r.file.Hash != "" {
		h := hex.EncodeToString(r.hasher.Sum(nil))
		if r.file.Hash != h {
			return fmt.Errorf("hash mismatch: expected %s, got %s", r.file.Hash, h)
		}
	}
	// should we be replacing the hash if it's empty?
	return nil
}
