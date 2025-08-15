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

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestS3Integration tests S3 backend integration
func TestS3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would be a real integration test with actual S3 credentials
	// For now, we'll just test the structure
	t.Run("S3VolumeCreation", func(t *testing.T) {
		// Test S3 volume creation with real credentials
		// This would require actual AWS credentials in test environment
		t.Skip("Requires actual S3 credentials")
	})

	t.Run("S3VolumeMount", func(t *testing.T) {
		// Test S3 volume mounting
		t.Skip("Requires actual S3 credentials")
	})
}

// TestGCSIntegration tests Google Cloud Storage backend integration
func TestGCSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("GCSVolumeCreation", func(t *testing.T) {
		// Test GCS volume creation with real credentials
		t.Skip("Requires actual GCS credentials")
	})

	t.Run("GCSVolumeMount", func(t *testing.T) {
		// Test GCS volume mounting
		t.Skip("Requires actual GCS credentials")
	})
}

// TestAzureBlobIntegration tests Azure Blob Storage backend integration
func TestAzureBlobIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("AzureBlobVolumeCreation", func(t *testing.T) {
		// Test Azure Blob volume creation with real credentials
		t.Skip("Requires actual Azure credentials")
	})

	t.Run("AzureBlobVolumeMount", func(t *testing.T) {
		// Test Azure Blob volume mounting
		t.Skip("Requires actual Azure credentials")
	})
}

// TestMinIOIntegration tests MinIO S3-compatible backend integration
func TestMinIOIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("MinIOVolumeCreation", func(t *testing.T) {
		// Test MinIO volume creation
		t.Skip("Requires MinIO test environment")
	})

	t.Run("MinIOVolumeMount", func(t *testing.T) {
		// Test MinIO volume mounting
		t.Skip("Requires MinIO test environment")
	})
}

// TestCSIIdentityService tests the CSI Identity service
func TestCSIIdentityService(t *testing.T) {
	conn, err := grpc.Dial("unix:///tmp/csi.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := csi.NewIdentityClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test GetPluginInfo
	info, err := client.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{})
	require.NoError(t, err)
	assert.Equal(t, "rclone.csi.veloxpack.io", info.GetName())
	assert.NotEmpty(t, info.GetVendorVersion())

	// Test GetPluginCapabilities
	caps, err := client.GetPluginCapabilities(ctx, &csi.GetPluginCapabilitiesRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, caps.GetCapabilities())

	// Test Probe
	probe, err := client.Probe(ctx, &csi.ProbeRequest{})
	require.NoError(t, err)
	assert.True(t, probe.GetReady().GetValue())
}

// TestCSIControllerService tests the CSI Controller service
func TestCSIControllerService(t *testing.T) {
	conn, err := grpc.Dial("unix:///tmp/csi.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := csi.NewControllerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test CreateVolume
	req := &csi.CreateVolumeRequest{
		Name: "test-volume",
		Parameters: map[string]string{
			"remote":     "s3",
			"remotePath": "test-bucket",
		},
		VolumeCapabilities: []*csi.VolumeCapability{
			{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
				},
			},
		},
	}

	resp, err := client.CreateVolume(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetVolume().GetVolumeId())
	assert.Equal(t, "test-volume", resp.GetVolume().GetVolumeContext()["remotePath"])
}

// TestCSINodeService tests the CSI Node service
func TestCSINodeService(t *testing.T) {
	conn, err := grpc.Dial("unix:///tmp/csi.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := csi.NewNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test NodeGetInfo
	info, err := client.NodeGetInfo(ctx, &csi.NodeGetInfoRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, info.GetNodeId())

	// Test NodeGetCapabilities
	caps, err := client.NodeGetCapabilities(ctx, &csi.NodeGetCapabilitiesRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, caps.GetCapabilities())
}
