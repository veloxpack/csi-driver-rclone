package buckets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/errors"
)

// UploadPartSpec defines each part's index and size for the start call
type UploadPartSpec struct {
	Index int   `json:"index"`
	Size  int64 `json:"size"`
}

type startUploadReq struct {
	Uploads []UploadPartSpec `json:"uploads"`
}

type UploadPart struct {
	Index    int      `json:"index"`
	UUID     string   `json:"uuid"`
	URL      string   `json:"url"`
	URLs     []string `json:"urls"`
	UploadId string   `json:"UploadId"`
}

type StartUploadResp struct {
	Uploads []UploadPart `json:"uploads"`
}

// StartUpload reserves all parts at once
func StartUpload(ctx context.Context, cfg *config.Config, bucketID string, parts []UploadPartSpec) (*StartUploadResp, error) {
	url := cfg.Endpoints.Network().StartUpload(bucketID)
	url += fmt.Sprintf("?multiparts=%d", len(parts))
	reqBody := startUploadReq{Uploads: parts}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", cfg.BasicAuthHeader)
	req.Header.Set("internxt-version", "1.0")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if cfg.BasicAuthHeader != "" {
	}

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.NewHTTPError(resp, "start upload")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result StartUploadResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// StartUploadMultipart starts a multipart upload session with explicit part count
func StartUploadMultipart(ctx context.Context, cfg *config.Config, bucketID string, parts []UploadPartSpec, numParts int) (*StartUploadResp, error) {
	url := cfg.Endpoints.Network().StartUpload(bucketID)
	url += fmt.Sprintf("?multiparts=%d", numParts)
	reqBody := startUploadReq{Uploads: parts}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", cfg.BasicAuthHeader)
	req.Header.Set("internxt-version", "1.0")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.NewHTTPError(resp, "start multipart upload")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result StartUploadResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
