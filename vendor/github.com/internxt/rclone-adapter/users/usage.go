package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone-adapter/config"
)

type UsageResponse struct {
	Drive int64 `json:"drive"`
}

// GetUsage calls GET {DRIVE_API_URL}/users/usage and returns the account's current usage in bytes.
func GetUsage(ctx context.Context, cfg *config.Config) (*UsageResponse, error) {
	url := cfg.Endpoints.Drive().Users().Usage()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get usage request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get usage request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(body))
	}

	var usage UsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, fmt.Errorf("failed to decode get usage response: %w", err)
	}

	return &usage, nil
}
