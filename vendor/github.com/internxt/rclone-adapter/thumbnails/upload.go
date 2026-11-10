package thumbnails

import (
	"context"
	"log"
)

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
			log.Printf("[WARN] Thumbnail generation failed for %s: %v\n", task.FileUUID, err)
		}
	}()
}
