package filen

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zeebo/blake3"
	"golang.org/x/sync/errgroup"
	"hash"
	"io"
	"strconv"

	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
)

// FileUpload encapsulates the state of an ongoing file upload.
// It tracks the file being uploaded and maintains the upload context,
// encryption key, and hash calculation.
type FileUpload struct {
	types.IncompleteFile           // The file being uploaded
	UploadKey            string    // Random key for this upload session
	Hasher               hash.Hash // For calculating file hash during upload
}

// NewFileUpload creates a new FileUpload structure from an IncompleteFile.
// It initializes a new random upload key and hash calculator for the upload process.
func (api *Filen) NewFileUpload(file *types.IncompleteFile) *FileUpload {
	return &FileUpload{
		IncompleteFile: *file,
		UploadKey:      crypto.GenerateRandomString(32),
		Hasher:         blake3.New(),
	}
}

// UploadChunk encrypts and uploads a single chunk of a file.
// This function handles the encryption of the chunk before sending it to the server.
// It returns storage details (bucket and region) for the uploaded chunk.
func (api *Filen) UploadChunk(ctx context.Context, fu *FileUpload, chunkIndex int, data []byte) (*client.V3UploadResponse, error) {
	data = fu.EncryptionKey.EncryptData(data)
	response, err := api.Client.PostV3Upload(ctx, fu.UUID, chunkIndex, fu.ParentUUID, fu.UploadKey, data)
	if err != nil {
		return nil, fmt.Errorf("upload chunk %d: %w", chunkIndex, err)
	}
	return response, nil
}

// makeEmptyRequestFromUploaderNoMeta creates a basic upload request without metadata.
// This is used as a foundation for both regular and empty file uploads.
func (api *Filen) makeEmptyRequestFromUploaderWithMetaAndSize(fu *FileUpload, meta string, size string) (*client.V3UploadEmptyRequest, error) {
	metaKey, err := fu.EncryptionKey.ToMasterKey()
	if err != nil {
		return nil, err
	}

	return &client.V3UploadEmptyRequest{
		UUID:       fu.UUID,
		Name:       metaKey.EncryptMeta(fu.Name),
		NameHashed: api.HashFileName(fu.Name),
		Size:       metaKey.EncryptMeta(size),
		Parent:     fu.ParentUUID,
		MimeType:   metaKey.EncryptMeta(fu.MimeType),
		Metadata:   api.EncryptMeta(meta),
		Version:    api.FileEncryptionVersion,
	}, nil
}

// makeEmptyRequestFromUploader creates a complete upload request for an empty file.
// It includes encrypted metadata and file hash information.
func (api *Filen) makeEmptyRequestFromUploader(fu *FileUpload, fileHash string) (*client.V3UploadEmptyRequest, error) {
	metadata := fu.GetRawMeta(api.FileEncryptionVersion)
	metadata.Size = 0
	metadata.Hash = fileHash

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal file metadata: %w", err)
	}
	return api.makeEmptyRequestFromUploaderWithMetaAndSize(fu, string(metadataStr), "0")
}

// makeRequestFromUploader creates a complete upload request for a non-empty file.
// It includes encrypted metadata, file size, chunk count, and hash information.
func (api *Filen) makeRequestFromUploader(fu *FileUpload, size int64, fileHash string) (*client.V3UploadDoneRequest, error) {
	metadata := fu.GetRawMeta(api.FileEncryptionVersion)
	metadata.Size = size
	metadata.Hash = fileHash

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal file metadata: %w", err)
	}
	emptyRequest, err := api.makeEmptyRequestFromUploaderWithMetaAndSize(fu, string(metadataStr), strconv.FormatInt(size, 10))

	if err != nil {
		return nil, fmt.Errorf("make empty request from uploader: %w", err)
	}

	return &client.V3UploadDoneRequest{
		V3UploadEmptyRequest: *emptyRequest,
		Chunks:               (size + ChunkSize - 1) / ChunkSize,
		UploadKey:            fu.UploadKey,
		Rm:                   crypto.GenerateRandomString(32),
	}, nil
}

// completeUpload finalizes the upload of a non-empty file.
// It sends the final metadata to the server and constructs the completed File object.
func (api *Filen) completeUpload(ctx context.Context, fu *FileUpload, bucket string, region string, size int64) (*types.File, error) {
	fileHash := hex.EncodeToString(fu.Hasher.Sum(nil))
	uploadRequest, err := api.makeRequestFromUploader(fu, size, fileHash)
	if err != nil {
		return nil, fmt.Errorf("make request from uploader: %w", err)
	}
	_, err = api.Client.PostV3UploadDone(ctx, *uploadRequest)
	if err != nil {
		return nil, fmt.Errorf("complete upload: %w", err)
	}

	return &types.File{
		IncompleteFile: fu.IncompleteFile,
		Size:           size,
		Region:         region,
		Bucket:         bucket,
		Chunks:         uploadRequest.Chunks,
		Hash:           fileHash,
		Version:        api.FileEncryptionVersion,
	}, nil
}

// completeUploadEmpty finalizes the upload of an empty (zero-byte) file.
// It sends the appropriate metadata to the server and constructs the completed File object.
func (api *Filen) completeUploadEmpty(ctx context.Context, fu *FileUpload) (*types.File, error) {
	fileHash := hex.EncodeToString(fu.Hasher.Sum(nil))
	uploadRequest, err := api.makeEmptyRequestFromUploader(fu, fileHash)
	if err != nil {
		return nil, fmt.Errorf("make request from uploader: %w", err)
	}
	_, err = api.Client.PostV3UploadEmpty(ctx, *uploadRequest)
	if err != nil {
		return nil, fmt.Errorf("complete upload: %w", err)
	}

	return &types.File{
		IncompleteFile: fu.IncompleteFile,
		Size:           0,
		Region:         "",
		Bucket:         "",
		Chunks:         0,
		Hash:           fileHash,
	}, nil
}

// UploadFile uploads a file to Filen cloud storage.
// It uses the provided IncompleteFile for metadata and reads file content from the io.Reader.
// The upload process happens in parallel chunks for efficiency.
// Once upload completes, it returns a File object representing the uploaded file.
//
// This method handles:
// - Chunking the file into manageable pieces
// - Parallel upload of chunks
// - Calculating file hash for integrity verification
// - Finalizing the upload with appropriate metadata
// - Updating search indexes and shared parent metadata
func (api *Filen) UploadFile(ctx context.Context, file *types.IncompleteFile, r io.Reader) (*types.File, error) {
	fu := api.NewFileUpload(file)
	bucketAndRegion := make(chan client.V3UploadResponse, 1)

	size, err := api.uploadFileChunks(ctx, fu, bucketAndRegion, file, r)
	if err != nil {
		return nil, err
	}

	completeFile, err := api.CompleteFileUpload(ctx, fu, bucketAndRegion, size)
	if err != nil {
		return nil, err
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return api.updateItemWithMaybeSharedParent(gCtx, completeFile) })
	g.Go(func() error { return api.updateSearchHashes(gCtx, completeFile) })
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return completeFile, nil
}

func (api *Filen) uploadFileChunks(ctx context.Context, fu *FileUpload, bucketAndRegion chan client.V3UploadResponse, file *types.IncompleteFile, r io.Reader) (int64, error) {
	size := int64(0)
	for i := 0; ; i++ {
		data := make([]byte, ChunkSize, ChunkSize+file.EncryptionKey.Cipher.Overhead())
		lastChunk := false
		for j := 0; j < ChunkSize; {
			read, err := r.Read(data[j:])
			if err != nil && err != io.EOF {
				return 0, err
			}
			j += read
			size += int64(read)
			if err == io.EOF {
				data = data[:j]
				lastChunk = true
				break
			}
		}

		if len(data) > 0 {
			fu.Hasher.Write(data)
			resp, err := api.UploadChunk(ctx, fu, i, data)
			if err != nil {
				return 0, err
			}
			select { // only care about getting this once
			case bucketAndRegion <- *resp:
			default:
			}
		}
		if lastChunk {
			break
		}
	}

	return size, nil
}

func (api *Filen) CompleteFileUpload(ctx context.Context, fu *FileUpload, bucketAndRegion chan client.V3UploadResponse, size int64) (*types.File, error) {
	var (
		completeFile *types.File
		err          error
	)
	if size == 0 {
		completeFile, err = api.completeUploadEmpty(ctx, fu)
		if err != nil {
			return nil, err
		}
	} else {
		select {
		case resp, ok := <-bucketAndRegion:
			if !ok {
				return nil, fmt.Errorf("no chunks successfully uploaded")
			}
			completeFile, err = api.completeUpload(ctx, fu, resp.Bucket, resp.Region, size)
			if err != nil {
				return nil, fmt.Errorf("complete upload: %w", err)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("context done %w", context.Cause(ctx))
		}
	}
	return completeFile, nil
}
