package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/errors"
)

// FileMeta represents file metadata from GET /files/{uuid}/meta
type FileMeta struct {
	ID               int64       `json:"id"`
	UUID             string      `json:"uuid"`
	FileID           string      `json:"fileId"`
	PlainName        string      `json:"plainName"`
	Type             string      `json:"type"`
	Size             json.Number `json:"size"`
	Bucket           string      `json:"bucket"`
	FolderID         int64       `json:"folderId"`
	FolderUUID       string      `json:"folderUuid"`
	EncryptVersion   string      `json:"encryptVersion"`
	UserID           int64       `json:"userId"`
	CreationTime     time.Time   `json:"creationTime"`
	ModificationTime time.Time   `json:"modificationTime"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	Status           string      `json:"status"`
}

// FileExistenceCheck represents a file to check for existence
type FileExistenceCheck struct {
	PlainName    string `json:"plainName"`
	Type         string `json:"type"`
	OriginalFile any    `json:"originalFile"`
}

// FileExistenceResult represents the response for existence check
type FileExistenceResult struct {
	Exists    bool   `json:"exists"`
	Status    string `json:"status,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	PlainName string `json:"plainName"`
	Type      string `json:"type,omitempty"`
}

// FileExists returns true if the file exists based on either Exists field or Status field
func (f *FileExistenceResult) FileExists() bool {
	return f.Exists || f.Status == "EXISTS"
}

// CheckFilesExistenceRequest is the request payload
type CheckFilesExistenceRequest struct {
	Files []FileExistenceCheck `json:"files"`
}

// CheckFilesExistenceResponse is the response
type CheckFilesExistenceResponse struct {
	Files []FileExistenceResult `json:"existentFiles"`
}

// CheckFilesExistence checks if files exist in a folder (batch operation)
func CheckFilesExistence(ctx context.Context, cfg *config.Config, folderUUID string, files []FileExistenceCheck) (*CheckFilesExistenceResponse, error) {
	endpoint := cfg.Endpoints.Drive().Folders().CheckFilesExistence(folderUUID)

	reqBody := CheckFilesExistenceRequest{Files: files}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal existence check request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create existence check request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute existence check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.NewHTTPError(resp, "check files existence")
	}

	var result CheckFilesExistenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode existence check response: %w", err)
	}

	return &result, nil
}

// DeleteFile deletes a file by UUID
func DeleteFile(ctx context.Context, cfg *config.Config, uuid string) error {
	u, err := url.Parse(cfg.Endpoints.Drive().Files().Delete(uuid))
	if err != nil {
		return fmt.Errorf("failed to parse delete file URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete file request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete file request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewHTTPError(resp, "delete file")
	}

	return nil
}

// RenameFile renames a file by UUID with the given new name and optional type.
func RenameFile(ctx context.Context, cfg *config.Config, fileUUID, newPlainName, newType string) error {
	endpoint := cfg.Endpoints.Drive().Files().Meta(fileUUID)

	payload := map[string]string{
		"plainName": newPlainName,
	}
	if newType != "" {
		payload["type"] = newType
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal rename file request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create rename file request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute rename file request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewHTTPError(resp, "rename file")
	}

	return nil
}

func GetFileMeta(ctx context.Context, cfg *config.Config, fileUUID string) (*FileMeta, error) {
	endpoint := cfg.Endpoints.Drive().Files().Meta(fileUUID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get file meta request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get file meta request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "get file meta")
	}

	var result FileMeta
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode get file meta response: %w", err)
	}

	return &result, nil
}
