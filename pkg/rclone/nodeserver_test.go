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
	"context"
	"errors"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getTestNodeServer() (NodeServer, error) {
	d := NewEmptyDriver("")
	mounter, err := NewFakeMounter()
	if err != nil {
		return NodeServer{}, errors.New("failed to get fake mounter")
	}
	return NodeServer{
		Driver:  d,
		mounter: mounter,
	}, nil
}

func TestNodeGetInfo(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	req := &csi.NodeGetInfoRequest{}
	resp, err := ns.NodeGetInfo(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fakeNodeID, resp.NodeId)
}

func TestNodeGetCapabilities(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	req := &csi.NodeGetCapabilitiesRequest{}
	resp, err := ns.NodeGetCapabilities(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Capabilities)
	assert.Equal(t, 1, len(resp.Capabilities))
	assert.Equal(t, csi.NodeServiceCapability_RPC_UNKNOWN, resp.Capabilities[0].GetRpc().Type)
}

func TestNodePublishVolumeValidation(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	tests := []struct {
		desc        string
		req         *csi.NodePublishVolumeRequest
		expectedErr error
	}{
		{
			desc:        "Volume ID missing",
			req:         &csi.NodePublishVolumeRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID missing in request"),
		},
		{
			desc: "Target path missing",
			req: &csi.NodePublishVolumeRequest{
				VolumeId: testVolumeID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "Target path not provided"),
		},
		{
			desc: "Volume capability missing",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   testVolumeID,
				TargetPath: "/mnt/test",
			},
			expectedErr: status.Error(codes.InvalidArgument, "Volume capability missing in request"),
		},
		{
			desc: "Remote parameter missing in volume context",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   testVolumeID,
				TargetPath: "/mnt/test",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "remote is required (provide via volumeAttributes or secrets)"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := ns.NodePublishVolume(context.Background(), test.req)
			if err == nil {
				t.Errorf("Expected error but got nil")
			}
			if status.Code(err) != status.Code(test.expectedErr) {
				t.Errorf("Expected error code %v, got %v (error: %v)", status.Code(test.expectedErr), status.Code(err), err)
			}
		})
	}
}

func TestNodeUnpublishVolumeValidation(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	tests := []struct {
		desc        string
		req         *csi.NodeUnpublishVolumeRequest
		expectedErr error
	}{
		{
			desc:        "Volume ID missing",
			req:         &csi.NodeUnpublishVolumeRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID missing in request"),
		},
		{
			desc: "Target path missing",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId: testVolumeID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "Target path missing in request"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := ns.NodeUnpublishVolume(context.Background(), test.req)
			if err == nil {
				t.Errorf("Expected error but got nil")
			}
			if status.Code(err) != status.Code(test.expectedErr) {
				t.Errorf("Expected error code %v, got %v", status.Code(test.expectedErr), status.Code(err))
			}
		})
	}
}

func TestNodeServerMountContext(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	targetPath := "/mnt/test-volume"

	// Initially, no mount context should exist
	mc := ns.getMountContext(targetPath)
	assert.Nil(t, mc)

	// Create a test mount context
	testMC := &mountContext{
		remoteName: "test-remote",
	}
	ns.setMountContext(targetPath, testMC)

	// Verify it was set
	mc = ns.getMountContext(targetPath)
	assert.NotNil(t, mc)
	assert.Equal(t, "test-remote", mc.remoteName)

	// Delete the mount context
	ns.deleteMountContext(targetPath)

	// Verify it was deleted
	mc = ns.getMountContext(targetPath)
	assert.Nil(t, mc)
}

func TestSanitizeRemoteName(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "Simple alphanumeric",
			input:    "test-volume-123",
			expected: "test-volume-123",
		},
		{
			desc:     "With special characters",
			input:    "test@volume#name",
			expected: "test_volume_name",
		},
		{
			desc:     "With spaces",
			input:    "test volume name",
			expected: "test_volume_name",
		},
		{
			desc:     "Long name gets truncated",
			input:    "this-is-a-very-long-volume-name-that-exceeds-the-maximum-length",
			expected: "this-is-a-very-long-volume-name-",
		},
		{
			desc:     "Mixed case with symbols",
			input:    "Test-Volume_Name!@#",
			expected: "Test-Volume_Name___",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := sanitizeRemoteName(test.input)
			assert.Equal(t, test.expected, result)
			assert.LessOrEqual(t, len(result), 32, "sanitized name should be <= 32 characters")
		})
	}
}

func TestParseConfigData(t *testing.T) {
	tests := []struct {
		desc       string
		configData string
		remoteName string
		expectErr  bool
	}{
		{
			desc: "Valid INI config",
			configData: `[s3]
type = s3
provider = AWS
access_key_id = test`,
			remoteName: "s3",
			expectErr:  false,
		},
		{
			desc:       "Invalid INI format",
			configData: "not valid ini",
			remoteName: "s3",
			expectErr:  true,
		},
		{
			desc: "Remote not found",
			configData: `[s3]
type = s3`,
			remoteName: "gcs",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result, err := parseConfigData(test.configData, test.remoteName)

			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUnimplementedNodeMethods(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	ctx := context.Background()

	// Test NodeStageVolume
	_, err = ns.NodeStageVolume(ctx, &csi.NodeStageVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test NodeUnstageVolume
	_, err = ns.NodeUnstageVolume(ctx, &csi.NodeUnstageVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test NodeGetVolumeStats
	_, err = ns.NodeGetVolumeStats(ctx, &csi.NodeGetVolumeStatsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test NodeExpandVolume
	_, err = ns.NodeExpandVolume(ctx, &csi.NodeExpandVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
}
