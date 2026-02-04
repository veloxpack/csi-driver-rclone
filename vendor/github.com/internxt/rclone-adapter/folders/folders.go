package folders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/errors"
)

// CreateFolder calls the folder creation endpoint with authorization.
// It auto‑fills CreationTime/ModificationTime if empty, checks status, and returns the newly created Folder.
func CreateFolder(ctx context.Context, cfg *config.Config, reqBody CreateFolderRequest) (*Folder, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if reqBody.CreationTime == "" {
		reqBody.CreationTime = now
	}
	if reqBody.ModificationTime == "" {
		reqBody.ModificationTime = now
	}

	endpoint := cfg.Endpoints.Drive().Folders().Create()
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create folder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create folder request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create folder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		return nil, errors.NewHTTPError(resp, "create folder")
	}

	var folder Folder
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, fmt.Errorf("failed to decode create folder response: %w", err)
	}

	return &folder, nil
}

// DeleteFolders deletes a folder by UUID
func DeleteFolder(ctx context.Context, cfg *config.Config, uuid string) error {
	u, err := url.Parse(cfg.Endpoints.Drive().Folders().Delete(uuid))
	if err != nil {
		return fmt.Errorf("failed to parse delete folder URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete folder request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete folder request: %w", err)
	}
	defer resp.Body.Close()

	//Server returns 204 on success
	if resp.StatusCode != 204 {
		return errors.NewHTTPError(resp, "delete folder")
	}

	return nil
}

// ListFolders lists child folders under the given parent UUID.
// Returns a slice of folders or error otherwise
func ListFolders(ctx context.Context, cfg *config.Config, parentUUID string, opts ListOptions) ([]Folder, error) {
	base := cfg.Endpoints.Drive().Folders().ContentFolders(parentUUID)
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list folders URL: %w", err)
	}
	q := u.Query()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	sortField := opts.Sort
	if sortField == "" {
		sortField = "plainName"
	}
	order := opts.Order
	if order == "" {
		order = "ASC"
	}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("sort", sortField)
	q.Set("order", order)

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list folders request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list folders request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "list folders")
	}

	var wrapper struct {
		Folders []Folder `json:"folders"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode list folders response: %w", err)
	}
	return wrapper.Folders, nil
}

// ListFiles lists files under the given parent folder UUID.
// Returns a slice of files or error otherwise
func ListFiles(ctx context.Context, cfg *config.Config, parentUUID string, opts ListOptions) ([]File, error) {
	base := cfg.Endpoints.Drive().Folders().ContentFiles(parentUUID)
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list files URL: %w", err)
	}
	q := u.Query()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	sortField := opts.Sort
	if sortField == "" {
		sortField = "plainName"
	}
	order := opts.Order
	if order == "" {
		order = "ASC"
	}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("sort", sortField)
	q.Set("order", order)

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list files request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list files request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "list files")
	}

	var wrapper struct {
		Files []File `json:"files"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode list files response: %w", err)
	}
	return wrapper.Files, nil
}

// This function will get all of the files in a folder, getting 50 at a time until completed
func ListAllFiles(ctx context.Context, cfg *config.Config, parentUUID string) ([]File, error) {
	var outFiles []File
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := ListFiles(ctx, cfg, parentUUID, ListOptions{Offset: offset})
		if err != nil {
			return nil, fmt.Errorf("failed to list all files at offset %d: %w", offset, err)
		}
		outFiles = append(outFiles, files...)
		offset += 50
		loops += 1
		if len(files) != 50 || loops >= maxLoops {
			break
		}
	}
	return outFiles, nil
}

// This function will get all of the folders in a folder, getting 50 at a time until completed
func ListAllFolders(ctx context.Context, cfg *config.Config, parentUUID string) ([]Folder, error) {
	var outFolders []Folder
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := ListFolders(ctx, cfg, parentUUID, ListOptions{Offset: offset})
		if err != nil {
			return nil, fmt.Errorf("failed to list all folders at offset %d: %w", offset, err)
		}
		outFolders = append(outFolders, files...)
		offset += 50
		loops += 1
		if len(files) != 50 || loops >= maxLoops {
			break
		}
	}
	return outFolders, nil
}
