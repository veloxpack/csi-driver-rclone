package buckets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/errors"
)

type CreateMetaRequest struct {
	Name             string    `json:"name"`
	Bucket           string    `json:"bucket"`
	FileID           *string   `json:"fileId,omitempty"`
	EncryptVersion   string    `json:"encryptVersion"`
	FolderUuid       string    `json:"folderUuid"`
	Size             int64     `json:"size"`
	PlainName        string    `json:"plainName"`
	Type             string    `json:"type"`
	CreationTime     time.Time `json:"creationTime"`
	Date             time.Time `json:"date"`
	ModificationTime time.Time `json:"modificationTime"`
}

type CreateMetaResponse struct {
	UUID           string      `json:"uuid"`
	Name           string      `json:"name"`
	Bucket         string      `json:"bucket"`
	FileID         string      `json:"fileId"`
	EncryptVersion string      `json:"encryptVersion"`
	FolderUuid     string      `json:"folderUuid"`
	Size           json.Number `json:"size"`
	PlainName      string      `json:"plainName"`
	Type           string      `json:"type"`
	Created        string      `json:"created"`
}

func CreateMetaFile(ctx context.Context, cfg *config.Config, name, bucketID string, fileID *string, encryptVersion, folderUuid, plainName, fileType string, size int64, modTime time.Time) (*CreateMetaResponse, error) {
	url := cfg.Endpoints.Drive().Files().Create()
	reqBody := CreateMetaRequest{
		Name:             name,
		Bucket:           bucketID,
		FileID:           fileID,
		EncryptVersion:   encryptVersion,
		FolderUuid:       folderUuid,
		Size:             size,
		PlainName:        plainName,
		Type:             fileType,
		CreationTime:     modTime,
		Date:             modTime,
		ModificationTime: modTime,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create meta request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create meta request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("internxt-version", "v1.0.436")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create meta request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.NewHTTPError(resp, "create meta")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result CreateMetaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create meta response: %w", err)
	}
	return &result, nil
}
