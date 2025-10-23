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
	"os"
	"path"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/rc" //nolint:misspell // Don't include misspell when running golangci-lint - unknwon is the package author's username
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"

	_ "github.com/rclone/rclone/backend/azureblob"
	_ "github.com/rclone/rclone/backend/b2"
	_ "github.com/rclone/rclone/backend/box"
	_ "github.com/rclone/rclone/backend/drive"
	_ "github.com/rclone/rclone/backend/dropbox"
	_ "github.com/rclone/rclone/backend/ftp"
	_ "github.com/rclone/rclone/backend/googlecloudstorage"
	_ "github.com/rclone/rclone/backend/local"
	_ "github.com/rclone/rclone/backend/onedrive"
	_ "github.com/rclone/rclone/backend/s3"
	_ "github.com/rclone/rclone/backend/sftp"
	_ "github.com/rclone/rclone/backend/swift"
	_ "github.com/rclone/rclone/backend/webdav"
	_ "github.com/rclone/rclone/cmd/mount2"
)

const (
	paramCacheDir    = "cache-dir"
	paramBackendType = "remoteType"
)

// mountContext stores context information for each mount with direct rclone objects
type mountContext struct {
	mountPoint     *mountlib.MountPoint // Direct access to rclone mount point
	remoteName     string               // Created remote name (for backwards compatibility)
	loadedSections []string             // Config sections loaded for nested remotes
	cancel         context.CancelFunc   // Context cancellation for VFS goroutines
}

// NodeServer implements the CSI Node service
type NodeServer struct {
	Driver       *Driver
	mounter      mount.Interface
	mountContext map[string]*mountContext
	mu           sync.RWMutex
	configMu     sync.Mutex // Protects concurrent config operations
	csi.UnimplementedNodeServer
}

// getMountContext retrieves mount context for a given target path
func (ns *NodeServer) getMountContext(targetPath string) *mountContext {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	if mc, ok := ns.mountContext[targetPath]; ok {
		return mc
	}
	return nil
}

// setMountContext stores mount context for a given target path
func (ns *NodeServer) setMountContext(targetPath string, mc *mountContext) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if ns.mountContext == nil {
		ns.mountContext = make(map[string]*mountContext)
	}
	ns.mountContext[targetPath] = mc
}

// deleteMountContext removes mount context for a given target path
func (ns *NodeServer) deleteMountContext(targetPath string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	delete(ns.mountContext, targetPath)
}

// NodePublishVolume mounts the rclone volume using direct rclone library integration
//
//nolint:lll,gocyclo // Complex function but necessary for CSI spec compliance and error handling
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	// Acquire lock for this volume operation
	lockKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	if acquired := ns.Driver.volumeLocks.TryAcquire(lockKey); !acquired {
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.Driver.volumeLocks.Release(lockKey)

	// Get mount options from VolumeCapability (CSI standard)
	readOnly := req.GetReadonly()
	mountOptions := volCap.GetMount().GetMountFlags()
	if readOnly {
		mountOptions = append(mountOptions, "ro")
	}

	// Merge secrets with volume context (volumeContext overrides secrets)
	params := ns.Driver.rcloneOtherParams
	if cacheDirPrefix, ok := params[paramCacheDir]; ok {
		params[paramCacheDir] = path.Join(cacheDirPrefix, targetPath)
	}

	// First, load values from secrets (defaults)
	secrets := req.GetSecrets()
	if secrets != nil {
		for k, v := range secrets {
			params[k] = v
		}
		klog.V(4).Infof("Loaded %d parameters from secrets", len(secrets))
	}

	// Then, merge with volumeContext (overrides secrets)
	volumeContext := req.GetVolumeContext()
	for k, v := range volumeContext {
		params[k] = v
	}

	// Extract reserved parameters
	remoteName := params[paramRemote]
	remotePath := params[paramRemotePath]
	configData := params[paramConfigData]
	remoteType := params[paramBackendType]

	if remoteName == "" {
		return nil, status.Error(codes.InvalidArgument, "remote is required (provide via volumeAttributes or secrets)")
	}

	if configData == "" && remoteType == "" {
		return nil, status.Error(codes.InvalidArgument, "either configData or remoteType must be provided")
	}

	// Remove reserved parameters from params, leaving only backend-specific config
	delete(params, paramRemote)
	delete(params, paramRemotePath)
	delete(params, paramConfigData)
	delete(params, paramBackendType)

	// Create target directory if it doesn't exist
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		// Check if already mounted
		if !notMnt {
			klog.V(2).Infof("Target path %s is already mounted", targetPath)
			return &csi.NodePublishVolumeResponse{}, nil
		}
	}

	// Ensure target directory has correct permissions
	if err := os.Chmod(targetPath, 0755); err != nil {
		klog.Warningf("Failed to set permissions on target path %s: %v", targetPath, err)
	}

	// If already mounted, verify the mount is valid
	if !notMnt {
		if _, err := os.ReadDir(targetPath); err == nil {
			klog.V(4).Infof("Volume %s already mounted to %s and accessible", volumeID, targetPath)
			return &csi.NodePublishVolumeResponse{}, nil
		}

		// Mount appears to exist but is not accessible - recover
		klog.Warningf("Mount point %s appears mounted but is not accessible (err: %v), attempting recovery", targetPath, err)

		if err := ns.mounter.Unmount(targetPath); err != nil {
			klog.Errorf("Failed to unmount corrupted mount point %s: %v", targetPath, err)
			return nil, status.Errorf(codes.Internal, "corrupted mount could not be cleaned up: %v", err)
		}

		klog.V(2).Infof("Successfully unmounted corrupted mount point %s, will remount", targetPath)
	}

	klog.V(2).Infof("NodePublishVolume: mounting %s:%s at %s", remoteName, remotePath, targetPath)

	var loadedSections []string
	var fsPath string

	updateRemoteOpts := config.UpdateRemoteOpt{
		NonInteractive: true,
		NoObscure:      false,
	}

	// Generate dynamic rclone config
	if configData == "" && remoteType != "" {
		klog.V(2).Infof("Generating dynmaic rcone config for remote type: %s", remoteType)

		// Extract remote params
		remoteParams := extractRemoteTypeParams(params, remoteType)

		if len(remoteParams) > 0 {
			configData = generateRecloneConfigFromParams(remoteParams, remoteType, remoteName)
			klog.V(4).Infof("Generated configData: %d bytes", len(configData))
		}
	}

	if configData == "" {
		return nil, status.Error(codes.InvalidArgument, "failed to parse configData")
	}

	// Handle configData - supports both single remotes and nested remotes (crypt, alias, chunker, union, etc.)
	if configData != "" {
		// Parse ALL sections from configData to support nested remotes
		allSections, err := parseAllConfigSections(configData)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse configData: %v", err)
		}

		klog.V(4).Infof("Parsed %d config sections from configData", len(allSections))

		// Load all sections into rclone's in-memory config storage
		// This allows nested remotes (crypt->alias->s3) to resolve automatically via fs.NewFs
		ns.configMu.Lock()
		for sectionName, sectionData := range allSections {
			for key, value := range sectionData {
				// Set backend config
				config.LoadedData().SetValue(sectionName, key, value)

				// Get params for a given backend type
				if key == "type" && len(params) > 0 {
					remoteParams := extractRemoteTypeParams(params, value)

					if len(remoteParams) > 0 {
						// Set the remaining values (params)
						config.UpdateRemote(ctx, sectionName, remoteParams, updateRemoteOpts)
					}
				}
			}
			loadedSections = append(loadedSections, sectionName)
			klog.V(4).Infof("Loaded config section: %s with %d keys", sectionName, len(sectionData))
		}
		ns.configMu.Unlock()

		// Build fsPath using the remote name directly - fs.NewFs will resolve nested remotes
		if remotePath != "" {
			fsPath = fmt.Sprintf("%s:%s", remoteName, remotePath)
		} else {
			fsPath = fmt.Sprintf("%s:", remoteName)
		}

		klog.V(2).Infof("Using configData with %d sections, resolving remote: %s", len(allSections), fsPath)
	}

	// Ensure cleanup on failure
	var mountSuccess bool
	defer func() {
		if !mountSuccess && len(loadedSections) > 0 {
			ns.configMu.Lock()
			for _, section := range loadedSections {
				// For configData mode, delete from in-memory storage
				config.LoadedData().DeleteSection(section)
			}
			ns.configMu.Unlock()
			klog.V(4).Infof("Cleaned up %d config sections after failure", len(loadedSections))
		}
	}()

	// Initialize filesystem
	rcloneFs, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initialize filesystem: %v", err)
	}

	// Create mount options mapper
	mapper := NewMountOptionsMapper(ns.Driver.rcloneMountOptions, ns.Driver.rcloneVFSOptions)

	// Parse mount options and apply them
	rcloneMountOpts, rcloneVFSOptions, err := mapper.ParseMountOptions(mountOptions)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse mount options: %v", err)
	}

	// Set read-only if specified in volume capability
	if readOnly {
		rcloneVFSOptions.ReadOnly = true
	}

	// Set device name if not already set
	if rcloneMountOpts.DeviceName == "" {
		rcloneMountOpts.DeviceName = fsPath
	}

	// Get mount function
	mountType, mountFn := mountlib.ResolveMountMethod("")
	if mountFn == nil {
		return nil, status.Error(codes.Internal, "no mount method available (FUSE not installed?)")
	}

	klog.V(4).Infof("Using mount method: %s", mountType)

	// Create mount point
	mountPoint := mountlib.NewMountPoint(mountFn, targetPath, rcloneFs, rcloneMountOpts, rcloneVFSOptions)

	// Create context with cancellation for VFS goroutines
	_, cancel := context.WithCancel(context.Background())

	// Mount the filesystem
	_, err = mountPoint.Mount()
	if err != nil {
		cancel()
		return nil, status.Errorf(codes.Internal, "failed to mount: %v", err)
	}

	mountSuccess = true

	// Store mount context
	ns.setMountContext(targetPath, &mountContext{
		mountPoint:     mountPoint,
		remoteName:     remoteName,
		loadedSections: loadedSections,
		cancel:         cancel,
	})

	klog.V(2).Infof("Successfully mounted volume %s to %s (remote: %s)", volumeID, targetPath, remoteName)
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the rclone volume using direct stats access
//
//nolint:lll
func (ns *NodeServer) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	// Acquire lock for this volume operation
	lockKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	if acquired := ns.Driver.volumeLocks.TryAcquire(lockKey); !acquired {
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.Driver.volumeLocks.Release(lockKey)

	klog.V(2).Infof("NodeUnpublishVolume: unmounting volume %s from %s", volumeID, targetPath)

	// Get mount context
	mc := ns.getMountContext(targetPath)
	if mc != nil && mc.mountPoint != nil {
		// Wait for VFS cache sync with improved error handling
		klog.V(2).Infof("Waiting for VFS cache sync (remote: %s)", mc.remoteName)

		timeout := time.Now().Add(2 * time.Minute) // Further reduced timeout for better responsiveness
		retryCount := 0
		maxRetries := 5

		for time.Now().Before(timeout) && retryCount < maxRetries {
			allClear := true

			// Check VFS cache uploads with improved error handling
			stats := mc.mountPoint.VFS.Stats()
			if diskCache, ok := stats["diskCache"].(rc.Params); ok {
				uploadsInProgress, _ := diskCache["uploadsInProgress"].(int)
				uploadsQueued, _ := diskCache["uploadsQueued"].(int)

				if uploadsInProgress > 0 || uploadsQueued > 0 {
					klog.V(4).Infof("Waiting for VFS cache uploads (in progress: %d, queued: %d, retry: %d/%d)", uploadsInProgress, uploadsQueued, retryCount+1, maxRetries)
					allClear = false
				}
			} else {
				klog.Warningf("Failed to get VFS cache stats, retry %d/%d", retryCount+1, maxRetries)
				allClear = false
			}

			if allClear {
				break
			}

			retryCount++
			// Exponential backoff for better performance
			sleepDuration := time.Duration(retryCount) * 2 * time.Second
			if sleepDuration > 10*time.Second {
				sleepDuration = 10 * time.Second
			}
			time.Sleep(sleepDuration)
		}

		if retryCount >= maxRetries {
			klog.Warningf("VFS cache sync timeout after %d retries, proceeding with unmount", maxRetries)
		}

		klog.V(2).Infof("Cache sync complete, proceeding with unmount")

		// Unmount using mountPoint's built-in unmount
		if err := mc.mountPoint.Unmount(); err != nil {
			klog.Errorf("Failed to unmount via mountPoint: %v, will try standard unmount", err)
		} else {
			klog.V(4).Infof("Successfully unmounted via mountPoint.Unmount()")
		}

		// Cancel context to stop VFS goroutines
		if mc.cancel != nil {
			mc.cancel()
		}

		// Clean up loaded config sections (for both configData and legacy modes)
		if len(mc.loadedSections) > 0 {
			ns.configMu.Lock()
			for _, section := range mc.loadedSections {
				// Try both methods to ensure cleanup works for all modes
				config.LoadedData().DeleteSection(section)
				config.DeleteRemote(section)
			}
			ns.configMu.Unlock()
			klog.V(4).Infof("Deleted %d config sections", len(mc.loadedSections))
		}
	}

	// Remove mount context
	ns.deleteMountContext(targetPath)

	// Use k8s mounter as fallback for cleanup
	klog.V(2).Infof("Performing final unmount cleanup for %s", targetPath)
	var err error
	extensiveMountPointCheck := true
	forceUnmounter, ok := ns.mounter.(mount.MounterForceUnmounter)
	if ok {
		klog.V(4).Infof("Using force unmount with 30s timeout")
		err = mount.CleanupMountWithForce(targetPath, forceUnmounter, extensiveMountPointCheck, 30*time.Second)
	} else {
		klog.V(4).Infof("Using standard cleanup")
		err = mount.CleanupMountPoint(targetPath, ns.mounter, extensiveMountPointCheck)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount target %q: %v", targetPath, err)
	}

	klog.V(2).Infof("Successfully unmounted volume %s from %s", volumeID, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeStageVolume is not implemented (rclone doesn't require staging)
//
//nolint:lll
func (ns *NodeServer) NodeStageVolume(_ context.Context, _ *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeUnstageVolume is not implemented (rclone doesn't require staging)
//
//nolint:lll
func (ns *NodeServer) NodeUnstageVolume(_ context.Context, _ *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetInfo returns info about the node
func (ns *NodeServer) NodeGetInfo(_ context.Context, _ *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

// NodeGetCapabilities returns the capabilities of the node
//
//nolint:lll
func (ns *NodeServer) NodeGetCapabilities(_ context.Context, _ *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

// NodeGetVolumeStats returns volume stats (not implemented)
//
//nolint:lll
func (ns *NodeServer) NodeGetVolumeStats(_ context.Context, _ *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeExpandVolume is not implemented
//
//nolint:lll
func (ns *NodeServer) NodeExpandVolume(_ context.Context, _ *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
