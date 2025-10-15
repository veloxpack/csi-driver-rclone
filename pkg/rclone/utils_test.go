/*
Copyright 2025 Veloxpack.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rclone

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

func TestVolumeLocks(t *testing.T) {
	tests := []struct {
		desc      string
		volumeID  string
		expectAcq bool
	}{
		{
			desc:      "Acquire lock on new volume",
			volumeID:  "vol-1",
			expectAcq: true,
		},
		{
			desc:      "Acquire lock on different volume",
			volumeID:  "vol-2",
			expectAcq: true,
		},
	}

	vl := NewVolumeLocks()

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			acquired := vl.TryAcquire(test.volumeID)
			if acquired != test.expectAcq {
				t.Errorf("Expected acquire=%v, got=%v", test.expectAcq, acquired)
			}

			// Clean up
			if acquired {
				vl.Release(test.volumeID)
			}
		})
	}
}

func TestVolumeLocksConcurrency(t *testing.T) {
	vl := NewVolumeLocks()
	volumeID := "test-volume"

	// First acquire should succeed
	acquired1 := vl.TryAcquire(volumeID)
	if !acquired1 {
		t.Error("First acquire should succeed")
	}

	// Second acquire on same volume should fail
	acquired2 := vl.TryAcquire(volumeID)
	if acquired2 {
		t.Error("Second acquire on same volume should fail")
	}

	// Release the lock
	vl.Release(volumeID)

	// Third acquire should succeed after release
	acquired3 := vl.TryAcquire(volumeID)
	if !acquired3 {
		t.Error("Acquire after release should succeed")
	}

	vl.Release(volumeID)
}

func TestVolumeLocksMultipleVolumes(t *testing.T) {
	vl := NewVolumeLocks()

	// Acquire locks on multiple different volumes
	vol1Acquired := vl.TryAcquire("volume-1")
	vol2Acquired := vl.TryAcquire("volume-2")
	vol3Acquired := vl.TryAcquire("volume-3")

	if !vol1Acquired || !vol2Acquired || !vol3Acquired {
		t.Error("All three different volumes should be acquired successfully")
	}

	// Try to acquire already locked volumes
	vol1Reacquire := vl.TryAcquire("volume-1")
	vol2Reacquire := vl.TryAcquire("volume-2")

	if vol1Reacquire || vol2Reacquire {
		t.Error("Reacquiring locked volumes should fail")
	}

	// Release and verify can be reacquired
	vl.Release("volume-1")
	vol1AfterRelease := vl.TryAcquire("volume-1")
	if !vol1AfterRelease {
		t.Error("Volume-1 should be acquirable after release")
	}

	// Clean up
	vl.Release("volume-1")
	vl.Release("volume-2")
	vl.Release("volume-3")
}

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		desc         string
		endpoint     string
		expectScheme string
		expectAddr   string
		expectErr    bool
	}{
		{
			desc:         "Unix socket endpoint",
			endpoint:     "unix:///tmp/csi.sock",
			expectScheme: "unix",
			expectAddr:   "/tmp/csi.sock",
			expectErr:    false,
		},
		{
			desc:         "TCP endpoint",
			endpoint:     "tcp://127.0.0.1:10000",
			expectScheme: "tcp",
			expectAddr:   "127.0.0.1:10000",
			expectErr:    false,
		},
		{
			desc:         "Unix endpoint with caps",
			endpoint:     "UNIX:///var/lib/csi.sock",
			expectScheme: "UNIX",
			expectAddr:   "/var/lib/csi.sock",
			expectErr:    false,
		},
		{
			desc:         "TCP endpoint with caps",
			endpoint:     "TCP://0.0.0.0:9000",
			expectScheme: "TCP",
			expectAddr:   "0.0.0.0:9000",
			expectErr:    false,
		},
		{
			desc:         "Unix endpoint without scheme (backward compatibility)",
			endpoint:     "/tmp/csi.sock",
			expectScheme: "unix",
			expectAddr:   "/tmp/csi.sock",
			expectErr:    false,
		},
		{
			desc:      "Invalid endpoint - empty address",
			endpoint:  "unix://",
			expectErr: true,
		},
		{
			desc:      "Invalid endpoint - wrong scheme",
			endpoint:  "http://127.0.0.1:8080",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			scheme, addr, err := ParseEndpoint(test.endpoint)

			if test.expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if scheme != test.expectScheme {
					t.Errorf("Expected scheme %s, got %s", test.expectScheme, scheme)
				}
				if addr != test.expectAddr {
					t.Errorf("Expected address %s, got %s", test.expectAddr, addr)
				}
			}
		})
	}
}

func TestNewDefaultIdentityServer(t *testing.T) {
	driver := &Driver{
		name:    "test-driver",
		version: "1.0.0",
		nodeID:  "test-node",
	}

	ids := NewDefaultIdentityServer(driver)
	assert.NotNil(t, ids)
	assert.Equal(t, driver, ids.Driver)
}

func TestNewControllerServer(t *testing.T) {
	driver := &Driver{
		name:    "test-driver",
		version: "1.0.0",
	}

	cs := NewControllerServer(driver)
	assert.NotNil(t, cs)
	assert.Equal(t, driver, cs.Driver)
}

func TestNewControllerServiceCapability(t *testing.T) {
	tests := []struct {
		cap csi.ControllerServiceCapability_RPC_Type
	}{
		{cap: csi.ControllerServiceCapability_RPC_UNKNOWN},
		{cap: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME},
		{cap: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME},
		{cap: csi.ControllerServiceCapability_RPC_LIST_VOLUMES},
		{cap: csi.ControllerServiceCapability_RPC_GET_CAPACITY},
		{cap: csi.ControllerServiceCapability_RPC_CLONE_VOLUME},
	}

	for _, test := range tests {
		resp := NewControllerServiceCapability(test.cap)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.GetRpc())
		assert.Equal(t, test.cap, resp.GetRpc().Type)
	}
}

func TestNewNodeServiceCapability(t *testing.T) {
	tests := []struct {
		cap csi.NodeServiceCapability_RPC_Type
	}{
		{cap: csi.NodeServiceCapability_RPC_UNKNOWN},
		{cap: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME},
		{cap: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS},
		{cap: csi.NodeServiceCapability_RPC_EXPAND_VOLUME},
	}

	for _, test := range tests {
		resp := NewNodeServiceCapability(test.cap)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.GetRpc())
		assert.Equal(t, test.cap, resp.GetRpc().Type)
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		desc          string
		method        string
		expectedLevel int32
	}{
		{
			desc:          "Probe method - verbose logging",
			method:        "/csi.v1.Identity/Probe",
			expectedLevel: 8,
		},
		{
			desc:          "NodeGetCapabilities - verbose logging",
			method:        "/csi.v1.Node/NodeGetCapabilities",
			expectedLevel: 8,
		},
		{
			desc:          "NodeGetVolumeStats - verbose logging",
			method:        "/csi.v1.Node/NodeGetVolumeStats",
			expectedLevel: 8,
		},
		{
			desc:          "CreateVolume - normal logging",
			method:        "/csi.v1.Controller/CreateVolume",
			expectedLevel: 2,
		},
		{
			desc:          "NodePublishVolume - normal logging",
			method:        "/csi.v1.Node/NodePublishVolume",
			expectedLevel: 2,
		},
		{
			desc:          "Unknown method - normal logging",
			method:        "/csi.v1.Unknown/SomeMethod",
			expectedLevel: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			level := getLogLevel(test.method)
			if level != test.expectedLevel {
				t.Errorf("Expected log level %d for %s, got %d", test.expectedLevel, test.method, level)
			}
		})
	}
}

func TestSanitizeFlag(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		key      string
		expected string
	}{
		{"lowercase prefix", "s3", "s3-endpoint", "endpoint"},
		{"uppercase prefix", "s3", "S3-endpoint", "endpoint"},
		{"mixed case prefix", "s3", "S3-cache-mode", "cache_mode"},
		{"all uppercase", "s3", "S3-CACHE-MODE", "cache_mode"},
		{"no prefix", "s3", "endpoint", "endpoint"},
		{"with dashes", "s3", "s3-cache-max-age", "cache_max_age"},
		{"empty remote", "", "s3-endpoint", "s3_endpoint"},
		{"empty key", "s3", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFlag(tt.remote, tt.key)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
