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
	"github.com/internxt/rclone-adapter/consistency"
	"github.com/internxt/rclone-adapter/errors"
	"github.com/internxt/rclone-adapter/internal/batch"
)

// CreateFolder calls the folder creation endpoint with authorization.
// It auto‑fills CreationTime/ModificationTime if empty, checks status, and returns the newly created Folder.
// The folder UUID is tracked via the consistency package so that subsequent
// operations on this folder automatically wait for eventual consistency.
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

	consistency.TrackFolder(folder.UUID)

	return &folder, nil
}

// DeleteFolder deletes a folder by UUID.
func DeleteFolder(ctx context.Context, cfg *config.Config, uuid string) error {
	if err := consistency.AwaitFolder(ctx, uuid); err != nil {
		return err
	}

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

// RenameFolder renames a folder by UUID with the given new name.
func RenameFolder(ctx context.Context, cfg *config.Config, folderUUID, newPlainName string) error {
	if err := consistency.AwaitFolder(ctx, folderUUID); err != nil {
		return err
	}

	endpoint := cfg.Endpoints.Drive().Folders().Meta(folderUUID)

	payload := map[string]string{
		"plainName": newPlainName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal rename folder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create rename folder request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute rename folder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewHTTPError(resp, "rename folder")
	}

	return nil
}

// MoveFolder moves a folder to a new destination folder, optionally renaming it.
// If newName is empty, it is omitted and the server keeps the current name.
func MoveFolder(ctx context.Context, cfg *config.Config, folderUUID, destinationFolderUUID, newName string) error {
	if err := consistency.AwaitFolder(ctx, folderUUID); err != nil {
		return err
	}

	endpoint := cfg.Endpoints.Drive().Folders().Move(folderUUID)

	payload := map[string]string{
		"destinationFolder": destinationFolderUUID,
	}
	if newName != "" {
		payload["name"] = newName
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal move folder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create move folder request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute move folder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewHTTPError(resp, "move folder")
	}

	return nil
}

// CheckFoldersExistence returns the child folders of parentUUID whose
// plainName exactly matches one of plainNames. Deleted and removed folders
// are not returned. Any number of names may be passed: they are sent in as
// many requests as the service's request limits need.
func CheckFoldersExistence(ctx context.Context, cfg *config.Config, parentUUID string, plainNames []string) ([]Folder, error) {
	batches, err := batch.Split(plainNames, maxExistenceNames, batch.MaxBodyBytes, func(names []string) any {
		return CheckFoldersExistenceRequest{PlainNames: names}
	})
	if err != nil {
		return nil, fmt.Errorf("check folders existence: %w", err)
	}
	if err := consistency.AwaitFolder(ctx, parentUUID); err != nil {
		return nil, err
	}

	var found []Folder
	for _, names := range batches {
		folders, err := checkFoldersExistence(ctx, cfg, parentUUID, names)
		if err != nil {
			return nil, err
		}
		found = append(found, folders...)
	}
	return found, nil
}

// checkFoldersExistence makes one existence check request for plainNames.
func checkFoldersExistence(ctx context.Context, cfg *config.Config, parentUUID string, plainNames []string) ([]Folder, error) {
	endpoint := cfg.Endpoints.Drive().Folders().CheckFoldersExistence(parentUUID)
	body, err := json.Marshal(CheckFoldersExistenceRequest{PlainNames: plainNames})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal check folders existence request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create check folders existence request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute check folders existence request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.NewHTTPError(resp, "check folders existence")
	}

	var result CheckFoldersExistenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode check folders existence response: %w", err)
	}
	return result.Folders, nil
}

// ListFolders lists one page of child folders under the given parent UUID,
// sorted by plainName. Only existing folders are returned.
// It returns the folders and the cursor for the next page, which is empty
// when there are no more pages.
func ListFolders(ctx context.Context, cfg *config.Config, parentUUID string, opts ListOptions) ([]Folder, string, error) {
	return listPage[Folder](ctx, cfg, parentUUID, cfg.Endpoints.Drive().Folders().ContentFolders(parentUUID), "folders", opts)
}

// ListFiles lists one page of files under the given parent folder UUID,
// sorted by plainName. Only existing files are returned.
// It returns the files and the cursor for the next page, which is empty
// when there are no more pages.
func ListFiles(ctx context.Context, cfg *config.Config, parentUUID string, opts ListOptions) ([]File, string, error) {
	return listPage[File](ctx, cfg, parentUUID, cfg.Endpoints.Drive().Folders().ContentFiles(parentUUID), "files", opts)
}

// listPage fetches one page from a cursor paginated content endpoint whose
// response holds the items under key and the next page cursor.
func listPage[T any](ctx context.Context, cfg *config.Config, parentUUID, endpoint, key string, opts ListOptions) ([]T, string, error) {
	if err := consistency.AwaitFolder(ctx, parentUUID); err != nil {
		return nil, "", err
	}
	op := "list " + key

	u, err := listURL(endpoint, opts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse %s URL: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create %s request: %w", op, err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute %s request: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.NewHTTPError(resp, op)
	}

	var body map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, "", fmt.Errorf("failed to decode %s response: %w", op, err)
	}
	var items []T
	if raw := body[key]; raw != nil {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&items); err != nil {
			return nil, "", fmt.Errorf("failed to decode %s response: %w", op, err)
		}
	}
	if items == nil {
		return nil, "", fmt.Errorf("%s response is missing the %s array", op, key)
	}
	var next string
	if raw := body["nextCursor"]; raw != nil {
		if err := json.Unmarshal(raw, &next); err != nil {
			return nil, "", fmt.Errorf("failed to decode %s response: %w", op, err)
		}
	}
	return items, next, nil
}

// ListAllFiles gets all of the files in a folder, following the cursor until
// the last page.
func ListAllFiles(ctx context.Context, cfg *config.Config, parentUUID string) ([]File, error) {
	files, err := listAllPages(func(cursor string) ([]File, string, error) {
		return ListFiles(ctx, cfg, parentUUID, ListOptions{Cursor: cursor})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all files: %w", err)
	}
	return files, nil
}

// ListAllFolders gets all of the folders in a folder, following the cursor
// until the last page.
func ListAllFolders(ctx context.Context, cfg *config.Config, parentUUID string) ([]Folder, error) {
	folders, err := listAllPages(func(cursor string) ([]Folder, string, error) {
		return ListFolders(ctx, cfg, parentUUID, ListOptions{Cursor: cursor})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all folders: %w", err)
	}
	return folders, nil
}

// listAllPages calls fetchPage with the cursor of the previous page until no
// cursor is returned. A repeated cursor is treated as an error so a
// misbehaving server cannot make it loop forever.
func listAllPages[T any](fetchPage func(cursor string) ([]T, string, error)) ([]T, error) {
	var all []T
	var cursor string
	seen := make(map[string]struct{})
	for page := 0; ; page++ {
		items, next, err := fetchPage(cursor)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		all = append(all, items...)
		if next == "" {
			return all, nil
		}
		if _, ok := seen[next]; ok {
			return nil, fmt.Errorf("page %d: server returned an already visited cursor", page)
		}
		seen[next] = struct{}{}
		cursor = next
	}
}

// listURL builds a cursor paginated list URL from base and opts.
func listURL(base string, opts ListOptions) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = MaxPageSize
	}
	limit = min(max(limit, MinPageSize), MaxPageSize)
	order := opts.Order
	if order == "" {
		order = "ASC"
	}

	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("order", order)
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
