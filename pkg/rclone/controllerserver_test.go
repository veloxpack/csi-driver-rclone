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
	"fmt"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testRemote     = "s3"
	testRemotePath = "bucket/path"
	testVolumeName = "test-volume"
	testVolumeID   = "s3#test-volume"
)

func initTestController(_ *testing.T) *ControllerServer {
	driver := NewEmptyDriver("")
	driver.volumeLocks = NewVolumeLocks()
	driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	cs := NewControllerServer(driver)
	return cs
}

func TestCreateVolume(t *testing.T) {
	cases := []struct {
		name      string
		req       *csi.CreateVolumeRequest
		resp      *csi.CreateVolumeResponse
		expectErr bool
	}{
		{
			name: "valid request with required parameters",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
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
				Parameters: map[string]string{
					paramRemote:     testRemote,
					paramRemotePath: testRemotePath,
				},
			},
			resp: &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      testVolumeID,
					CapacityBytes: 0,
					VolumeContext: map[string]string{
						paramRemote:     testRemote,
						paramRemotePath: testRemotePath,
					},
				},
			},
			expectErr: false,
		},
		{
			name: "valid request with configData",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
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
				Parameters: map[string]string{
					paramRemote:     testRemote,
					paramRemotePath: testRemotePath,
					paramConfigData: "[s3]\ntype=s3\naccess_key_id=test",
				},
			},
			expectErr: false,
		},
		{
			name: "name empty",
			req: &csi.CreateVolumeRequest{
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
				Parameters: map[string]string{
					paramRemote: testRemote,
				},
			},
			expectErr: true,
		},
		{
			name: "volume capabilities missing",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
				Parameters: map[string]string{
					paramRemote: testRemote,
				},
			},
			expectErr: true,
		},
		{
			name: "block volume capability not supported",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					paramRemote: testRemote,
				},
			},
			expectErr: true,
		},
		{
			name: "remote parameter missing (allowed - can come from secrets)",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
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
				Parameters: map[string]string{},
			},
			resp: &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      testVolumeName, // Uses name as volumeID when remote not provided
					CapacityBytes: 0,
					VolumeContext: map[string]string{},
				},
			},
			expectErr: false,
		},
		{
			name: "remotePath parameter missing (allowed - can come from secrets)",
			req: &csi.CreateVolumeRequest{
				Name: testVolumeName,
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
				Parameters: map[string]string{
					paramRemote: testRemote,
				},
			},
			resp: &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      testVolumeID, // Uses remote#name format when remote is provided
					CapacityBytes: 0,
					VolumeContext: map[string]string{
						paramRemote: testRemote,
					},
				},
			},
			expectErr: false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cs := initTestController(t)
			resp, err := cs.CreateVolume(context.TODO(), test.req)

			if !test.expectErr && err != nil {
				t.Errorf("test %q failed: %v", test.name, err)
			}
			if test.expectErr && err == nil {
				t.Errorf("test %q failed; expected error but got success", test.name)
			}
			if test.resp != nil && !test.expectErr {
				if resp == nil {
					t.Errorf("test %q failed: expected response but got nil", test.name)
				} else {
					// Check volume ID contains remote and name
					if resp.Volume.VolumeId == "" {
						t.Errorf("test %q failed: volume ID is empty", test.name)
					}
					// Check volume context
					if resp.Volume.VolumeContext[paramRemote] != test.req.Parameters[paramRemote] {
						t.Errorf("test %q failed: remote parameter mismatch", test.name)
					}
				}
			}
		})
	}
}

func TestDeleteVolume(t *testing.T) {
	cases := []struct {
		desc        string
		req         *csi.DeleteVolumeRequest
		resp        *csi.DeleteVolumeResponse
		expectedErr error
	}{
		{
			desc: "Valid request",
			req: &csi.DeleteVolumeRequest{
				VolumeId: testVolumeID,
			},
			resp:        &csi.DeleteVolumeResponse{},
			expectedErr: nil,
		},
		{
			desc:        "Volume ID missing",
			req:         &csi.DeleteVolumeRequest{},
			resp:        nil,
			expectedErr: status.Error(codes.InvalidArgument, "volume id is empty"),
		},
		{
			desc: "Valid request with complex volume ID",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "s3#my-bucket/path#volume-name",
			},
			resp:        &csi.DeleteVolumeResponse{},
			expectedErr: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			cs := initTestController(t)
			resp, err := cs.DeleteVolume(context.TODO(), test.req)

			if test.expectedErr == nil && err != nil {
				t.Errorf("test %q failed: %v", test.desc, err)
			}
			if test.expectedErr != nil && err == nil {
				t.Errorf("test %q failed; expected error %v, got success", test.desc, test.expectedErr)
			}
			if !reflect.DeepEqual(resp, test.resp) {
				t.Errorf("test %q failed: got resp %+v, expected %+v", test.desc, resp, test.resp)
			}
		})
	}
}

func TestValidateVolumeCapabilities(t *testing.T) {
	cases := []struct {
		desc        string
		req         *csi.ValidateVolumeCapabilitiesRequest
		expectedErr error
	}{
		{
			desc: "Valid request",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: testVolumeID,
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
			},
			expectedErr: nil,
		},
		{
			desc: "Volume ID missing",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID missing in request"),
		},
		{
			desc: "Volume capabilities missing",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: testVolumeID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "volume capabilities missing in request"),
		},
		{
			desc: "Block volume capability not supported",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: testVolumeID,
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "block volume capability not supported"),
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			cs := initTestController(t)
			resp, err := cs.ValidateVolumeCapabilities(context.TODO(), test.req)

			if test.expectedErr == nil && err != nil {
				t.Errorf("test %q failed: %v", test.desc, err)
			}
			if test.expectedErr != nil && err == nil {
				t.Errorf("test %q failed; expected error %v, got success", test.desc, test.expectedErr)
			}
			if test.expectedErr == nil && resp.Confirmed == nil {
				t.Errorf("test %q failed: expected confirmed capabilities", test.desc)
			}
		})
	}
}

func TestControllerGetCapabilities(t *testing.T) {
	req := &csi.ControllerGetCapabilitiesRequest{}
	cs := initTestController(t)

	resp, err := cs.ControllerGetCapabilities(context.TODO(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Capabilities)
	assert.Equal(t, 1, len(resp.Capabilities))
	assert.Equal(t, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		resp.Capabilities[0].GetRpc().Type)
}

func TestIsValidVolumeCapabilities(t *testing.T) {
	mountVolumeCapabilities := []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
	}
	blockVolumeCapabilities := []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		},
	}

	cases := []struct {
		desc      string
		volCaps   []*csi.VolumeCapability
		expectErr error
	}{
		{
			desc:      "mount volume capabilities",
			volCaps:   mountVolumeCapabilities,
			expectErr: nil,
		},
		{
			desc:      "block volume capabilities not supported",
			volCaps:   blockVolumeCapabilities,
			expectErr: fmt.Errorf("block volume capability not supported"),
		},
		{
			desc:      "empty volume capabilities",
			volCaps:   []*csi.VolumeCapability{},
			expectErr: fmt.Errorf("volume capabilities missing in request"),
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			err := isValidVolumeCapabilities(test.volCaps)
			if !reflect.DeepEqual(err, test.expectErr) {
				t.Errorf("[test: %s] Unexpected error: %v, expected error: %v", test.desc, err, test.expectErr)
			}
		})
	}
}

func TestUnimplementedControllerMethods(t *testing.T) {
	cs := initTestController(t)
	ctx := context.Background()

	// Test ControllerPublishVolume
	_, err := cs.ControllerPublishVolume(ctx, &csi.ControllerPublishVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test ControllerUnpublishVolume
	_, err = cs.ControllerUnpublishVolume(ctx, &csi.ControllerUnpublishVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test ControllerGetVolume
	_, err = cs.ControllerGetVolume(ctx, &csi.ControllerGetVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test ListVolumes
	_, err = cs.ListVolumes(ctx, &csi.ListVolumesRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test GetCapacity
	_, err = cs.GetCapacity(ctx, &csi.GetCapacityRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test ControllerExpandVolume
	_, err = cs.ControllerExpandVolume(ctx, &csi.ControllerExpandVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test ControllerModifyVolume
	_, err = cs.ControllerModifyVolume(ctx, &csi.ControllerModifyVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
}
