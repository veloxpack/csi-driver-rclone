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
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
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
	Remount    bool
}

// Driver is the main driver structure
type Driver struct {
	name        string
	remount     bool
	nodeID      string
	version     string
	endpoint    string
	ns          *NodeServer
	server      NonBlockingGRPCServer
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
		remount:  options.Remount,
	}

	d.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		// csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	})

	d.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_UNKNOWN,
		csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
		csi.NodeServiceCapability_RPC_VOLUME_MOUNT_GROUP,
	})

	d.volumeLocks = NewVolumeLocks()

	return d
}

// NewNodeServer creates a new node server
func NewNodeServer(d *Driver, mounter mount.Interface, stateManager *MountStateManager) *NodeServer {
	return &NodeServer{
		Driver:            d,
		mounter:           mounter,
		mountStateManager: stateManager,
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

	var stateManager *MountStateManager

	if d.remount {
		// Initialize mount state manager
		stateManager, err = NewMountStateManager()
		if err != nil {
			klog.Fatalf("Failed to initialize mount state manager: %v", err)
		}
	}

	// Initialize node server
	d.ns = NewNodeServer(d, mounter, stateManager)

	// Remount all saved states on boot
	if stateManager != nil {
		if err := d.ns.RemountAllStates(ctx); err != nil {
			klog.Warningf("Failed to remount all states on boot: %v", err)
			// Don't fail driver startup if remount fails - mounts may have been cleaned up
		}
	}

	// Initialize metrics collector with NodeServer reference
	if err := initMetricsCollector(ctx, d.nodeID, d.name, d.endpoint, d.ns); err != nil {
		klog.Fatalf("Failed to initialize CSI metrics collector: %v", err)
	}

	s := NewNonBlockingGRPCServer()
	d.server = s
	s.Start(d.endpoint,
		NewDefaultIdentityServer(d),
		NewControllerServer(d),
		d.ns,
		testMode)
	s.Wait()
}

// Shutdown performs graceful shutdown of the driver
func (d *Driver) Shutdown(ctx context.Context) error {
	klog.Info("Starting driver shutdown...")

	// 1. Stop accepting new gRPC requests
	if d.server != nil {
		klog.Info("Stopping gRPC server...")
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			d.server.Stop()
			close(done)
		}()

		select {
		case <-done:
			klog.Info("gRPC server stopped gracefully")
		case <-shutdownCtx.Done():
			klog.Warning("gRPC server shutdown timeout, forcing stop")
			d.server.ForceStop()
		}
	}

	// 2. Unmount all active mounts
	if d.ns != nil {
		klog.Info("Unmounting all volumes...")
		if err := d.ns.UnmountAll(ctx); err != nil {
			klog.Errorf("Failed to unmount all volumes: %v", err)
			return fmt.Errorf("failed to unmount volumes: %w", err)
		}
		klog.Info("All volumes unmounted successfully")
	}

	klog.Info("Driver shutdown complete")
	return nil
}

// DumpMountInfo logs information about all active mounts
func (d *Driver) DumpMountInfo() {
	if d.ns == nil {
		klog.Warning("NodeServer not initialized")
		return
	}

	d.ns.mu.RLock()
	defer d.ns.mu.RUnlock()

	if len(d.ns.mountContext) == 0 {
		klog.Info("No active mounts")
		return
	}

	klog.Infof("Active mounts: %d", len(d.ns.mountContext))
	for targetPath, mc := range d.ns.mountContext {
		healthy := "unknown"
		vfsStats := "unavailable"

		if mc.mountPoint != nil && mc.mountPoint.VFS != nil {
			stats := mc.mountPoint.VFS.Stats()
			if errors, ok := stats["errors"]; ok {
				if errCount, ok := errors.(int); ok && errCount == 0 {
					healthy = "healthy"
				} else {
					healthy = fmt.Sprintf("unhealthy (errors: %d)", errCount)
				}
			}

			// Get cache stats if available
			if diskCache, ok := stats["diskCache"].(map[string]interface{}); ok {
				inProgress, _ := diskCache["uploadsInProgress"].(int)
				queued, _ := diskCache["uploadsQueued"].(int)
				vfsStats = fmt.Sprintf("uploads: %d in-progress, %d queued", inProgress, queued)
			}
		}

		klog.Infof("  Mount: %s", targetPath)
		klog.Infof("    Remote: %s", mc.remoteName)
		klog.Infof("    Health: %s", healthy)
		klog.Infof("    VFS: %s", vfsStats)
		klog.Infof("    Loaded remotes: %d", len(mc.remotes))
	}
}

// ForceCacheSync forces VFS cache sync on all active mounts
func (d *Driver) ForceCacheSync(ctx context.Context) error {
	if d.ns == nil {
		return fmt.Errorf("NodeServer not initialized")
	}

	d.ns.mu.RLock()
	mountContexts := make(map[string]*mountContext)
	for targetPath, mc := range d.ns.mountContext {
		mountContexts[targetPath] = mc
	}
	d.ns.mu.RUnlock()

	if len(mountContexts) == 0 {
		klog.Info("No active mounts to sync")
		return nil
	}

	klog.Infof("Forcing cache sync on %d mounts...", len(mountContexts))

	for targetPath, mc := range mountContexts {
		klog.V(2).Infof("Syncing cache for %s", targetPath)
		waitForVFSCacheSync(mc)
		klog.V(2).Infof("Cache sync complete for %s", targetPath)
	}

	klog.Info("All cache syncs completed")
	return nil
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
