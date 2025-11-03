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
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
)

const (
	// DefaultDriverName is the default name of the driver
	DefaultDriverName = "rclone.csi.veloxpack.io"

	// CSI parameter keys injected by external-provisioner
	pvcNameKey           = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey      = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey            = "csi.storage.k8s.io/pv/name"
	pvcNameMetadata      = "${pvc.metadata.name}"
	pvcNamespaceMetadata = "${pvc.metadata.namespace}"
	pvNameMetadata       = "${pv.metadata.name}"
)

// DriverOptions defines driver parameters specified in driver deployment
type DriverOptions struct {
	NodeID     string
	DriverName string
	Endpoint   string
}

// Driver is the main driver structure
type Driver struct {
	name        string
	nodeID      string
	version     string
	endpoint    string
	ns          *NodeServer
	cscap       []*csi.ControllerServiceCapability
	nscap       []*csi.NodeServiceCapability
	volumeLocks *VolumeLocks
}

// NewDriver creates a new driver instance
func NewDriver(options *DriverOptions) *Driver {
	klog.V(2).Infof("Driver: %v version: %v", options.DriverName, driverVersion)

	// Initialize rclone logging to redirect to klog
	InitRcloneLogging()

	d := &Driver{
		name:     options.DriverName,
		version:  driverVersion,
		nodeID:   options.NodeID,
		endpoint: options.Endpoint,
	}

	d.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		// csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	})

	d.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_UNKNOWN,
		csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
	})

	d.volumeLocks = NewVolumeLocks()

	return d
}

// NewNodeServer creates a new node server
func NewNodeServer(d *Driver, mounter mount.Interface) *NodeServer {
	return &NodeServer{
		Driver:  d,
		mounter: mounter,
	}
}

// Run starts the CSI driver
func (d *Driver) Run(testMode bool) {
	versionMeta, err := GetVersionYAML(d.name)
	if err != nil {
		klog.Fatalf("%v", err)
	}
	klog.V(2).Infof("\nDRIVER INFORMATION:\n-------------------\n%s\n\nStreaming logs below:", versionMeta)

	// Initialize rclone core components
	ctx := context.Background()

	// Initialize global options
	if err := fs.GlobalOptionsInit(); err != nil {
		klog.Fatalf("Failed to initialize rclone global options: %v", err)
	}

	// Start accounting (bandwidth limiting, stats, TPS limiting)
	accounting.Start(ctx)

	klog.V(2).Info("Rclone core initialization complete")

	mounter := mount.New("")
	d.ns = NewNodeServer(d, mounter)

	// Initialize metrics collector with NodeServer reference
	if err := initMetricsCollector(ctx, d.nodeID, d.name, d.endpoint, d.ns); err != nil {
		klog.Fatalf("Failed to initialize CSI metrics collector: %v", err)
	}

	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint,
		NewDefaultIdentityServer(d),
		NewControllerServer(d),
		d.ns,
		testMode)
	s.Wait()
}

// AddControllerServiceCapabilities adds controller service capabilities
func (d *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	csc := make([]*csi.ControllerServiceCapability, 0, len(cl))
	for _, c := range cl {
		csc = append(csc, NewControllerServiceCapability(c))
	}
	d.cscap = csc
}

// AddNodeServiceCapabilities adds node service capabilities
func (d *Driver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	nsc := make([]*csi.NodeServiceCapability, 0, len(nl))
	for _, n := range nl {
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	d.nscap = nsc
}

// replaceWithMap replaces template variables in str with values from the map
// This enables dynamic path substitution using PVC/PV metadata
func replaceWithMap(str string, m map[string]string) string {
	for k, v := range m {
		if k != "" {
			str = strings.ReplaceAll(str, k, v)
		}
	}
	return str
}
