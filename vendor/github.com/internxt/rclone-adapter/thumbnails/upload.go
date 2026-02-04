package thumbnails

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

// GenerateAndPrepare generates a thumbnail and prepares it for upload.
// Returns the thumbnail data as a reader, the size, and any error.
func GenerateAndPrepare(fileType string, originalData []byte) (io.Reader, int64, *Config, error) {
	if !IsSupportedFormat(fileType) {
		return nil, 0, nil, fmt.Errorf("unsupported format: %s", fileType)
	}

	thumbCfg := DefaultConfig()
	thumbData, thumbSize, err := Generate(originalData, thumbCfg)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return bytes.NewReader(thumbData), thumbSize, thumbCfg, nil
}

// CreateThumbnailMetadata creates the metadata struct for registering a thumbnail.
func CreateThumbnailMetadata(fileUUID, bucketID, bucketFile, encryptVersion string, size int64, cfg *Config) CreateThumbnailRequest {
	return CreateThumbnailRequest{
		FileUUID:       fileUUID,
		MaxWidth:       cfg.MaxWidth,
		MaxHeight:      cfg.MaxHeight,
		Type:           cfg.Format,
		Size:           size,
		BucketID:       bucketID,
		BucketFile:     bucketFile,
		EncryptVersion: encryptVersion,
	}
}

// ThumbnailUploadTask represents a task to upload a thumbnail asynchronously.
type ThumbnailUploadTask struct {
	Ctx          context.Context
	FileUUID     string
	FileType     string
	OriginalData []byte
}

// UploadFunc is a function type that handles the actual upload of a thumbnail.
// This allows dependency injection to avoid circular imports.
type UploadFunc func(ctx context.Context, task *ThumbnailUploadTask) error

// ProcessAsync processes a thumbnail upload task asynchronously.
// The uploadFunc parameter should contain the logic to upload the thumbnail
// and register it with the API.
func ProcessAsync(task *ThumbnailUploadTask, uploadFunc UploadFunc) {
	go func() {
		bgCtx := context.Background()

		if err := uploadFunc(bgCtx, task); err != nil {
			fmt.Printf("[WARN] Thumbnail generation failed for %s: %v\n", task.FileUUID, err)
		}
	}()
}
