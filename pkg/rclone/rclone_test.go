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
	mount "k8s.io/mount-utils"
)

const (
	fakeNodeID = "fakeNodeID"
)

// NewEmptyDriver creates a test driver with optional empty fields for testing error cases
func NewEmptyDriver(emptyField string) *Driver {
	var d *Driver
	switch emptyField {
	case "version":
		d = &Driver{
			name:    DefaultDriverName,
			version: "",
			nodeID:  fakeNodeID,
		}
	case "name":
		d = &Driver{
			name:    "",
			version: driverVersion,
			nodeID:  fakeNodeID,
		}
	default:
		d = &Driver{
			name:    DefaultDriverName,
			version: driverVersion,
			nodeID:  fakeNodeID,
		}
	}
	d.volumeLocks = NewVolumeLocks()
	return d
}

func TestNewEmptyDriver(t *testing.T) {
	d := NewEmptyDriver("version")
	assert.Empty(t, d.version)
	assert.Equal(t, DefaultDriverName, d.name)

	d = NewEmptyDriver("name")
	assert.Empty(t, d.name)
	assert.Equal(t, driverVersion, d.version)

	d = NewEmptyDriver("")
	assert.NotEmpty(t, d.name)
	assert.NotEmpty(t, d.version)
}

func TestNewDriver(t *testing.T) {
	tests := []struct {
		desc    string
		options *DriverOptions
	}{
		{
			desc: "Create driver with default name",
			options: &DriverOptions{
				NodeID:     "test-node",
				DriverName: DefaultDriverName,
				Endpoint:   "unix:///tmp/csi.sock",
			},
		},
		{
			desc: "Create driver with custom name",
			options: &DriverOptions{
				NodeID:     "custom-node",
				DriverName: "custom.rclone.csi.veloxpack.io",
				Endpoint:   "tcp://127.0.0.1:10000",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			driver := NewDriver(test.options)

			assert.NotNil(t, driver)
			assert.Equal(t, test.options.DriverName, driver.name)
			assert.Equal(t, test.options.NodeID, driver.nodeID)
			assert.Equal(t, test.options.Endpoint, driver.endpoint)
			assert.Equal(t, driverVersion, driver.version)
			assert.NotNil(t, driver.volumeLocks)
			assert.NotNil(t, driver.cscap)
			assert.NotNil(t, driver.nscap)

			// Verify controller capabilities
			assert.Equal(t, 2, len(driver.cscap))
			assert.Equal(t, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				driver.cscap[0].GetRpc().Type)
			assert.Equal(t, csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
				driver.cscap[1].GetRpc().Type)

			// Verify node capabilities
			assert.Equal(t, 1, len(driver.nscap))
			assert.Equal(t, csi.NodeServiceCapability_RPC_UNKNOWN,
				driver.nscap[0].GetRpc().Type)
		})
	}
}

func TestNewNodeServer(t *testing.T) {
	driver := &Driver{
		name:    DefaultDriverName,
		version: driverVersion,
		nodeID:  "test-node",
	}

	mounter := mount.NewFakeMounter([]mount.MountPoint{})

	ns := NewNodeServer(driver, mounter)
	assert.NotNil(t, ns)
	assert.Equal(t, driver, ns.Driver)
	assert.Equal(t, mounter, ns.mounter)
}

func TestDriverAddControllerServiceCapabilities(t *testing.T) {
	driver := &Driver{}

	tests := []struct {
		desc string
		caps []csi.ControllerServiceCapability_RPC_Type
	}{
		{
			desc: "Single capability",
			caps: []csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			},
		},
		{
			desc: "Multiple capabilities",
			caps: []csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
			},
		},
		{
			desc: "Empty capabilities",
			caps: []csi.ControllerServiceCapability_RPC_Type{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			driver.AddControllerServiceCapabilities(test.caps)

			assert.Equal(t, len(test.caps), len(driver.cscap))
			for i, cap := range test.caps {
				assert.Equal(t, cap, driver.cscap[i].GetRpc().Type)
			}
		})
	}
}

func TestDriverAddNodeServiceCapabilities(t *testing.T) {
	driver := &Driver{}

	tests := []struct {
		desc string
		caps []csi.NodeServiceCapability_RPC_Type
	}{
		{
			desc: "Single capability",
			caps: []csi.NodeServiceCapability_RPC_Type{
				csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
			},
		},
		{
			desc: "Multiple capabilities",
			caps: []csi.NodeServiceCapability_RPC_Type{
				csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
				csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
			},
		},
		{
			desc: "Empty capabilities",
			caps: []csi.NodeServiceCapability_RPC_Type{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			driver.AddNodeServiceCapabilities(test.caps)

			assert.Equal(t, len(test.caps), len(driver.nscap))
			for i, cap := range test.caps {
				assert.Equal(t, cap, driver.nscap[i].GetRpc().Type)
			}
		})
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Successful run in test mode",
			testFunc: func(_ *testing.T) {
				d := NewEmptyDriver("")
				d.endpoint = "tcp://127.0.0.1:0"
				d.Run(true)
			},
		},
		{
			name: "Successful run with node ID missing",
			testFunc: func(_ *testing.T) {
				d := NewEmptyDriver("")
				d.endpoint = "tcp://127.0.0.1:0"
				d.nodeID = ""
				d.Run(true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestDefaultDriverName(t *testing.T) {
	expectedName := "rclone.csi.veloxpack.io"
	if DefaultDriverName != expectedName {
		t.Errorf("Expected DefaultDriverName to be %s, got %s", expectedName, DefaultDriverName)
	}
}

func TestDriverFields(t *testing.T) {
	options := &DriverOptions{
		NodeID:     "node-1",
		DriverName: "test.driver",
		Endpoint:   "unix:///var/lib/csi.sock",
	}

	driver := NewDriver(options)

	// Verify all fields are properly initialized
	assert.Equal(t, "test.driver", driver.name)
	assert.Equal(t, "node-1", driver.nodeID)
	assert.Equal(t, driverVersion, driver.version)
	assert.Equal(t, "unix:///var/lib/csi.sock", driver.endpoint)
	assert.NotNil(t, driver.volumeLocks)
	assert.Nil(t, driver.ns) // ns is nil until Run() is called

	// Verify capabilities were set
	assert.Greater(t, len(driver.cscap), 0)
	assert.Greater(t, len(driver.nscap), 0)
}
