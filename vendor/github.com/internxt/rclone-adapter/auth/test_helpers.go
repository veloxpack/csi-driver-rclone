package auth

import (
	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/endpoints"
)

// newTestConfig creates a test config with the given mock server URL and token.
// The HTTPClient is properly configured with the centralized header transport.
func newTestConfig(mockServerURL, token string) *config.Config {
	cfg := &config.Config{
		Token:     token,
		Endpoints: endpoints.NewConfig(mockServerURL),
	}
	cfg.ApplyDefaults()
	return cfg
}
