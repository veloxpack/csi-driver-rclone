// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package stalldetection

import (
	"os"
	"strconv"
	"strings"
	"time"

	"storj.io/eventkit"
)

var evs = eventkit.Package()

// Config contains configuration for stall detection and handling.
type Config struct {
	// BaseUploads is the number of successful uploads that need to
	// occur before we start considering whether remaining uploads have stalled.
	BaseUploads int

	// Factor determines how much longer than the fastest uploads the remaining
	// uploads are allowed to take. Once BaseUploads successful uploads have occurred,
	// remaining uploads will be considered stalled if they take longer than the
	// BaseUploads'th duration times Factor.
	Factor int

	// MinStallDuration is the minimum amount of time that can be considered a stall.
	MinStallDuration time.Duration

	// DynamicBaseUploads if set to true, will use a dynamically calculated BaseUploads value
	// based on the erasure coding parameters instead of the configured BaseUploads value.
	// The original BaseUploads value will be ignored.
	DynamicBaseUploads bool
}

// defaultConfig provides default values for stall detection.
var defaultConfig = Config{
	BaseUploads:      3,
	Factor:           2,
	MinStallDuration: 10 * time.Second,
}

// ConfigFromEnv creates a config from the environment variables.
func ConfigFromEnv() (config *Config) {
	if !isEnabled() {
		return nil
	}
	return &Config{
		BaseUploads:        getEnvInt("STORJ_STALL_DETECTION_BASE_UPLOADS", defaultConfig.BaseUploads),
		Factor:             getEnvInt("STORJ_STALL_DETECTION_FACTOR", defaultConfig.Factor),
		MinStallDuration:   getEnvDuration("STORJ_STALL_DETECTION_MIN_DURATION", defaultConfig.MinStallDuration),
		DynamicBaseUploads: strings.ToLower(os.Getenv("STORJ_STALL_DETECTION_DYNAMIC_BASE")) == "true",
	}
}

// ValidateAndUpdate sets config to defaults if includes invalid parameters.
// Sets BaseUploads from RS scheme if DyanmicBaseUploads is true.
func (c *Config) ValidateAndUpdate(totalNodes int) (config *Config) {
	configCopy := *c
	config = &configCopy

	if config.DynamicBaseUploads && totalNodes > 1 {
		config.BaseUploads = totalNodes / 2
	}
	if config.Factor < 1 {
		config.Factor = defaultConfig.Factor
	}
	if config.BaseUploads < 1 {
		config.BaseUploads = defaultConfig.BaseUploads
	}
	if config.MinStallDuration <= 0 {
		config.MinStallDuration = defaultConfig.MinStallDuration
	}

	evs.Event("stall-detection-config-validate-update",
		eventkit.Int64("base_uploads", int64(config.BaseUploads)),
		eventkit.Int64("factor", int64(config.Factor)),
		eventkit.Duration("min_stall_duration", config.MinStallDuration),
		eventkit.Bool("dynamic_base_uploads", config.DynamicBaseUploads),
		eventkit.Int64("total_nodes", int64(totalNodes)),
	)
	return config
}

// isEnabled checks if stall detection is enabled via environment variables
func isEnabled() bool {
	return strings.ToLower(os.Getenv("STORJ_STALL_DETECTION_ENABLED")) == "true"
}

// Helper functions for environment variable parsing
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}
