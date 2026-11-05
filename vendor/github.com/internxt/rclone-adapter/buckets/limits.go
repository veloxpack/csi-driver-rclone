package buckets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/internxt/rclone-adapter/config"
	sdkerrors "github.com/internxt/rclone-adapter/errors"
)

// LimitsResponse mirrors the GET /drive/files/limits payload
type LimitsResponse struct {
	MaxUploadFileSize int64           `json:"maxUploadFileSize"`
	Versioning        json.RawMessage `json:"versioning,omitempty"`
}

// GetFileLimits calls GET /drive/files/limits and returns the account's
// upload limits
func GetFileLimits(ctx context.Context, cfg *config.Config) (*LimitsResponse, error) {
	url := cfg.Endpoints.Drive().Files().Limits()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get file limits request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get file limits request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, sdkerrors.NewHTTPError(resp, "get file limits")
	}

	var limits LimitsResponse
	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return nil, fmt.Errorf("failed to decode get file limits response: %w", err)
	}

	json, err := json.MarshalIndent(&limits, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get file limits response: %w", err)
	}
	fmt.Println("FileLimitsResponse", string(json))

	return &limits, nil
}

// checkUploadSize is a no-op for unknown sizes (plainSize < 0) and fails-open if the limit cannot be fetched.
func checkUploadSize(ctx context.Context, cfg *config.Config, plainSize int64) error {
	if plainSize < 0 {
		return nil
	}
	if cfg == nil || cfg.FileLimits == nil {
		return nil
	}
	maxSize, ok := cfg.FileLimits.GetOrFetch(ctx, func(ctx context.Context) (int64, error) {
		resp, err := GetFileLimits(ctx, cfg)
		if err != nil {
			return 0, err
		}
		return resp.MaxUploadFileSize, nil
	})
	if !ok || maxSize <= 0 {
		return nil
	}
	if plainSize > maxSize {
		return &sdkerrors.FileTooLargeError{Size: plainSize, MaxSize: maxSize}
	}
	return nil
}
