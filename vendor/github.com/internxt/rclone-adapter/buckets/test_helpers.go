package buckets

import (
	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/endpoints"
)

// newTestConfig creates a test config with the given mock server URL.
// The HTTPClient is properly configured with the centralized header transport.
func newTestConfig(mockServerURL string) *config.Config {
	cfg := &config.Config{
		Mnemonic:        TestMnemonic,
		Bucket:          TestBucket1,
		Token:           TestToken,
		BasicAuthHeader: TestBasicAuth,
		Endpoints:       endpoints.NewConfig(mockServerURL),
	}
	cfg.ApplyDefaults()
	return cfg
}

// newTestConfigWithBucket creates a test config with the given bucket.
// Useful for tests that need a specific bucket but no immediate endpoint.
// Call setEndpoints() if you need to set endpoints later.
func newTestConfigWithBucket(bucket string) *config.Config {
	cfg := &config.Config{
		Mnemonic: TestMnemonic,
		Bucket:   bucket,
	}
	cfg.ApplyDefaults()
	return cfg
}

// newEmptyTestConfig creates an empty test config with defaults applied.
// Useful for tests that don't need specific configuration values.
func newEmptyTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.ApplyDefaults()
	return cfg
}

// setEndpoints sets the endpoints on a config and reapplies defaults.
// Use this after creating a config when you need to set endpoints later.
func setEndpoints(cfg *config.Config, mockServerURL string) {
	cfg.Endpoints = endpoints.NewConfig(mockServerURL)
	cfg.ApplyDefaults()
}

// newTestConfigWithSetup creates a test config with optional setup function.
// The setup function is called after creating the config but before applying defaults.
// Useful for tests that need to modify the config before defaults are applied.
func newTestConfigWithSetup(mockServerURL string, setup func(*config.Config)) *config.Config {
	cfg := &config.Config{
		Mnemonic:        TestMnemonic,
		Bucket:          TestBucket1,
		Token:           TestToken,
		BasicAuthHeader: TestBasicAuth,
		Endpoints:       endpoints.NewConfig(mockServerURL),
	}
	if setup != nil {
		setup(cfg)
	}
	cfg.ApplyDefaults()
	return cfg
}
