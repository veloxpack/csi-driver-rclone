package buckets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"context"

	"github.com/internxt/rclone-adapter/config"
)

// ChunkUploadSession holds the state for a chunked upload session
// where the caller (rclone) controls concurrency and buffer management
type ChunkUploadSession struct {
	cfg        *config.Config
	encIndex   string
	sha256Hash hash.Hash
	startResp  *StartUploadResp
	uploadID   string
	uuid       string
	totalSize  int64
	chunkSize  int64
	numParts   int64
	fileKey []byte
	iv      []byte
}

// NewChunkUploadSession initializes encryption and starts the multipart
// upload session on the Internxt network. The caller specifies totalSize
// and chunkSize
func NewChunkUploadSession(ctx context.Context, cfg *config.Config, totalSize, chunkSize int64) (*ChunkUploadSession, error) {
	var ph [32]byte
	if _, err := rand.Read(ph[:]); err != nil {
		return nil, fmt.Errorf("cannot generate random index: %w", err)
	}
	plainIndex := hex.EncodeToString(ph[:])

	fileKey, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, plainIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file key: %w", err)
	}

	numParts := (totalSize + chunkSize - 1) / chunkSize
	if totalSize == 0 {
		numParts = 0
	}

	s := &ChunkUploadSession{
		cfg:        cfg,
		encIndex:   plainIndex,
		sha256Hash: sha256.New(),
		totalSize:  totalSize,
		chunkSize:  chunkSize,
		numParts:   numParts,
		fileKey: fileKey,
		iv:      iv,
	}

	specs := []UploadPartSpec{{Index: 0, Size: totalSize}}
	s.startResp, err = StartUploadMultipart(ctx, cfg, cfg.Bucket, specs, int(numParts))
	if err != nil {
		return nil, fmt.Errorf("failed to start multipart upload: %w", err)
	}

	if len(s.startResp.Uploads) != 1 {
		return nil, fmt.Errorf("expected 1 upload entry, got %d", len(s.startResp.Uploads))
	}

	uploadInfo := s.startResp.Uploads[0]
	if len(uploadInfo.URLs) != int(numParts) {
		return nil, fmt.Errorf("expected %d URLs, got %d", numParts, len(uploadInfo.URLs))
	}

	s.uploadID = uploadInfo.UploadId
	s.uuid = uploadInfo.UUID

	return s, nil
}

// UploadChunk uploads encrypted data to the presigned URL for the given
// partIndex. Returns the ETag from the server
func (s *ChunkUploadSession) UploadChunk(ctx context.Context, partIndex int, data io.ReadSeeker, size int64) (string, error) {
	if partIndex < 0 || partIndex >= len(s.startResp.Uploads[0].URLs) {
		return "", fmt.Errorf("part index %d out of range [0, %d)", partIndex, len(s.startResp.Uploads[0].URLs))
	}

	uploadURL := s.startResp.Uploads[0].URLs[partIndex]
	result, err := Transfer(ctx, s.cfg, uploadURL, data, size)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk %d: %w", partIndex, err)
	}
	return result.ETag, nil
}

// Finish computes the final file hash (RIPEMD-160(SHA-256(encrypted_data)))
// and completes the multipart upload on the Internxt network
func (s *ChunkUploadSession) Finish(ctx context.Context, parts []CompletedPart) (*FinishUploadResp, error) {
	sha256Result := s.sha256Hash.Sum(nil)
	overallHash := ComputeFileHash(sha256Result)

	shard := MultipartShard{
		UUID:     s.uuid,
		Hash:     overallHash,
		UploadId: s.uploadID,
		Parts:    parts,
	}

	return FinishMultipartUpload(ctx, s.cfg, s.cfg.Bucket, s.encIndex, shard)
}

// NewCipherAtOffset returns an AES-256-CTR cipher.Stream positioned at byteOffset.
// Handles both block-aligned and non-aligned offsets.
func (s *ChunkUploadSession) NewCipherAtOffset(byteOffset int64) (cipher.Stream, error) {
	blockNum := byteOffset / int64(aes.BlockSize)
	adjustedIV := AddToIV(s.iv, blockNum)
	stream, err := NewAES256CTRCipher(s.fileKey, adjustedIV)
	if err != nil {
		return nil, err
	}

	if partial := int(byteOffset % int64(aes.BlockSize)); partial > 0 {
		throwaway := make([]byte, partial)
		stream.XORKeyStream(throwaway, throwaway)
	}
	return stream, nil
}

// HashEncryptedData feeds already-encrypted bytes into the session's SHA-256 hasher.
// Caller must ensure data is fed in sequential byte order.
func (s *ChunkUploadSession) HashEncryptedData(data []byte) {
	s.sha256Hash.Write(data)
}

// URLs returns the presigned upload URLs for all parts
func (s *ChunkUploadSession) URLs() []string {
	if s.startResp == nil || len(s.startResp.Uploads) == 0 {
		return nil
	}
	return s.startResp.Uploads[0].URLs
}

// UUID returns the upload session UUID
func (s *ChunkUploadSession) UUID() string {
	return s.uuid
}

// EncIndex returns the encryption index for metadata creation
func (s *ChunkUploadSession) EncIndex() string {
	return s.encIndex
}
