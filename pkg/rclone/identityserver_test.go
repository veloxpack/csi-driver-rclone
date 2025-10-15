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
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetPluginInfo(t *testing.T) {
	req := csi.GetPluginInfoRequest{}
	emptyNameDriver := NewEmptyDriver("name")
	emptyVersionDriver := NewEmptyDriver("version")

	tests := []struct {
		desc        string
		driver      *Driver
		expectedErr error
	}{
		{
			desc:        "Successful Request",
			driver:      NewEmptyDriver(""),
			expectedErr: nil,
		},
		{
			desc:        "Driver name missing",
			driver:      emptyNameDriver,
			expectedErr: status.Error(codes.Unavailable, "Driver name not configured"),
		},
		{
			desc:        "Driver version missing",
			driver:      emptyVersionDriver,
			expectedErr: status.Error(codes.Unavailable, "Driver is missing version"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			fakeIdentityServer := IdentityServer{
				Driver: test.driver,
			}
			resp, err := fakeIdentityServer.GetPluginInfo(context.Background(), &req)

			if !reflect.DeepEqual(err, test.expectedErr) {
				t.Errorf("Unexpected error: %v\nExpected: %v", err, test.expectedErr)
			}

			if test.expectedErr == nil {
				if resp == nil {
					t.Error("Expected non-nil response for successful request")
				} else {
					if resp.Name != test.driver.name {
						t.Errorf("Expected name %s, got %s", test.driver.name, resp.Name)
					}
					if resp.VendorVersion != test.driver.version {
						t.Errorf("Expected version %s, got %s", test.driver.version, resp.VendorVersion)
					}
				}
			}
		})
	}
}

func TestProbe(t *testing.T) {
	d := NewEmptyDriver("")
	req := csi.ProbeRequest{}
	fakeIdentityServer := IdentityServer{
		Driver: d,
	}

	resp, err := fakeIdentityServer.Probe(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Ready)
	assert.Equal(t, true, resp.Ready.Value)
}

func TestProbeWithDifferentDriverStates(t *testing.T) {
	tests := []struct {
		desc        string
		driver      *Driver
		expectReady bool
	}{
		{
			desc:        "Driver with empty name",
			driver:      NewEmptyDriver("name"),
			expectReady: false,
		},
		{
			desc:        "Driver with empty version",
			driver:      NewEmptyDriver("version"),
			expectReady: false,
		},
		{
			desc:        "Driver with all fields",
			driver:      NewEmptyDriver(""),
			expectReady: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req := csi.ProbeRequest{}
			fakeIdentityServer := IdentityServer{
				Driver: test.driver,
			}

			resp, err := fakeIdentityServer.Probe(context.Background(), &req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, resp.Ready)
			assert.Equal(t, test.expectReady, resp.Ready.Value)
		})
	}
}

func TestGetPluginCapabilities(t *testing.T) {
	expectedCap := []*csi.PluginCapability{
		{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
				},
			},
		},
		{
			Type: &csi.PluginCapability_VolumeExpansion_{
				VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
					Type: csi.PluginCapability_VolumeExpansion_ONLINE,
				},
			},
		},
	}

	d := NewEmptyDriver("")
	fakeIdentityServer := IdentityServer{
		Driver: d,
	}
	req := csi.GetPluginCapabilitiesRequest{}

	resp, err := fakeIdentityServer.GetPluginCapabilities(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedCap, resp.Capabilities)
}

func TestGetPluginCapabilitiesWithMultipleDrivers(t *testing.T) {
	tests := []struct {
		desc   string
		driver *Driver
	}{
		{
			desc:   "Default driver",
			driver: NewEmptyDriver(""),
		},
		{
			desc:   "Custom name driver",
			driver: &Driver{name: "custom.driver", version: "1.0.0", volumeLocks: NewVolumeLocks()},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			fakeIdentityServer := IdentityServer{
				Driver: test.driver,
			}
			req := csi.GetPluginCapabilitiesRequest{}

			resp, err := fakeIdentityServer.GetPluginCapabilities(context.Background(), &req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, resp.Capabilities)
			assert.Equal(t, 2, len(resp.Capabilities))

			// Verify it has the CONTROLLER_SERVICE capability
			hasControllerService := false
			hasVolumeExpansion := false
			for _, cap := range resp.Capabilities {
				if cap.GetService() != nil && cap.GetService().Type == csi.PluginCapability_Service_CONTROLLER_SERVICE {
					hasControllerService = true
				}
				if cap.GetVolumeExpansion() != nil && cap.GetVolumeExpansion().Type == csi.PluginCapability_VolumeExpansion_ONLINE {
					hasVolumeExpansion = true
				}
			}
			assert.True(t, hasControllerService, "Should have CONTROLLER_SERVICE capability")
			assert.True(t, hasVolumeExpansion, "Should have ONLINE volume expansion capability")
		})
	}
}
