package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/internxt/rclone-adapter/config"
	sdkerrors "github.com/internxt/rclone-adapter/errors"
)

type LimitResponse struct {
	MaxSpaceBytes int64 `json:"maxSpaceBytes"`
}

// GetLimit calls {DRIVE_API_URL}/users/limit and returns the maximum available storage of the account.
func GetLimit(ctx context.Context, cfg *config.Config) (*LimitResponse, error) {
	url := cfg.Endpoints.Drive().Users().Limit()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get limit request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get limit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, sdkerrors.NewHTTPError(resp, "get limit")
	}

	var limit LimitResponse
	if err := json.NewDecoder(resp.Body).Decode(&limit); err != nil {
		return nil, fmt.Errorf("failed to decode get limit response: %w", err)
	}

	return &limit, nil
}
