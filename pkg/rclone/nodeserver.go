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
	"strings"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/configstruct"
	"github.com/rclone/rclone/fs/log"
	"github.com/rclone/rclone/fs/rc" //nolint:misspell // Don't include misspell when running golangci-lint - unknwon is the package author's username
	"github.com/rclone/rclone/lib/atexit"
	"github.com/rclone/rclone/vfs/vfscommon"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
)

const (
	paramCacheDir         = "cache_dir"
	paramTmpDir           = "temp_dir"
	paramLogLevel         = "log_level"
	paramMountType        = "mount_type"
	paramBackendType      = "remoteType"
	paramBackendTypeKey   = "type"
	paramDisabledFeatures = "disable"
)

// reservedParams contains parameter names that should not be passed to rclone backend
var reservedParams = map[string]bool{
	paramRemote:      true,
	paramRemotePath:  true,
	paramConfigData:  true,
	paramBackendType: true,
}

// mountContext stores context information for each mount with direct rclone objects
type mountContext struct {
	context.Context
	mountPoint *mountlib.MountPoint // Direct access to rclone mount point
	remoteName string               // Created remote name (for backwards compatibility)
	remotes    []string             // Remotes loaded for nested remotes
	cancel     context.CancelFunc   // Context cancellation for VFS goroutines
	ctx        context.Context      // Context for mount goroutines
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

// validatePublishVolumeRequest validates the NodePublishVolumeRequest
func validatePublishVolumeRequest(req *csi.NodePublishVolumeRequest) error {
	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return status.Error(codes.InvalidArgument, "Target path not provided")
	}
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	return nil
}

// validateUnpublishVolumeRequest validates the NodeUnpublishVolumeRequest
func validateUnpublishVolumeRequest(req *csi.NodeUnpublishVolumeRequest) error {
	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	return nil
}

// publishVolumeParams holds parameters for volume publishing
type publishVolumeParams struct {
	remoteName string
	remotePath string
	configData string
	remoteType string
	params     map[string]string
}

// setRcloneConfigFlags sets global rclone configuration flags
func setRcloneConfigFlags(ctx context.Context, ci *fs.ConfigInfo, params map[string]string) error {
	// Set cache directory if provided
	if cacheDir, ok := params[paramCacheDir]; ok {
		if err := config.SetCacheDir(cacheDir); err != nil {
			return status.Errorf(codes.Internal, "failed to set cache directory: %v", err)
		}
		klog.V(4).Infof("Set rclone cache directory to: %s", cacheDir)
	}

	// Set tmp directory if provided
	if tempDir, ok := params[paramTmpDir]; ok {
		if err := config.SetTempDir(tempDir); err != nil {
			return status.Errorf(codes.Internal, "failed to set temp directory: %v", err)
		}
		klog.V(4).Infof("Set rclone temp directory to: %s", tempDir)
	}

	// Set log level if provided
	if logLevel, ok := params[paramLogLevel]; ok {
		log.Handler.SetLevel(fs.LogLevelToSlog(fs.LogLevelDebug))
		klog.V(4).Infof("Set rclone log level to: %s", logLevel)
	}

	// Get Rclone config
	configMap := configmap.Simple{}

	// Set disabled features if provided
	if disableFeatures, ok := params[paramDisabledFeatures]; ok {
		ci.DisableFeatures = strings.Split(disableFeatures, ",")
	}

	// Set all golbal
	for key, value := range params {
		if opt := fs.ConfigOptionsInfo.Get(key); opt != nil {
			configMap.Set(key, value)
		}
	}

	// Apply the changes to the global config
	if err := configstruct.Set(configMap, ci); err != nil {
		return fmt.Errorf("failed to update global config: %v", err)
	}

	// CRITICAL: Call Reload to make changes take effect
	if err := ci.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload config changes: %v", err)
	}

	return nil
}

// mergeVolumeParameters merges driver params, secrets, and volume context
func (ns *NodeServer) mergeVolumeParameters(req *csi.NodePublishVolumeRequest) map[string]string {
	params := make(map[string]string)

	// TODO: Implement automatic cache directory generation for performance optimization
	//
	// Option 1: Shared cache per PVC (recommended for performance)
	// - Pods using same PVC share cache (warm cache on pod restart/replacement)
	// - Different PVCs remain isolated
	// - Cache lifecycle tied to PVC
	//
	// Implementation:
	//   volumeID := req.GetVolumeId()
	//   params[paramCacheDir] = filepath.Join("/var/lib/rclone-cache", volumeID)
	//   params[paramTmpDir] = filepath.Join("/tmp/rclone-temp", volumeID)
	//   klog.V(4).Infof("Using shared cache for volume %s: %s", volumeID, params[paramCacheDir])
	//
	// Option 2: Shared cache per remote location (maximum sharing)
	//   remoteName := req.GetVolumeContext()[paramRemote]
	//   remotePath := req.GetVolumeContext()[paramRemotePath]
	//   cacheKey := fmt.Sprintf("%s-%s", remoteName, remotePath)
	//   cacheHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cacheKey)))[:16]
	//   params[paramCacheDir] = filepath.Join("/var/lib/rclone-cache", cacheHash)
	//
	// Option 3: Per-pod cache (maximum isolation, no sharing)
	//   uniqueID := fmt.Sprintf("%s-%d", req.GetVolumeId(), time.Now().UnixNano())
	//   params[paramCacheDir] = filepath.Join("/tmp/rclone-cache", uniqueID)
	//
	// Currently: Users must specify cache_dir and temp_dir in volume attributes/secrets
	// If not specified, rclone uses its default locations (may cause conflicts)

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

	return params
}

// extractPublishParams extracts and validates required parameters
func extractPublishParams(params map[string]string) (*publishVolumeParams, error) {
	pvp := &publishVolumeParams{
		remoteName: params[paramRemote],
		remotePath: params[paramRemotePath],
		configData: params[paramConfigData],
		remoteType: params[paramBackendType],
		params:     make(map[string]string),
	}

	if pvp.remoteName == "" {
		return nil, status.Error(codes.InvalidArgument, "remote is required (provide via volumeAttributes or secrets)")
	}

	if pvp.configData == "" && pvp.remoteType == "" {
		return nil, status.Error(codes.InvalidArgument, "either configData or remoteType must be provided")
	}

	// Copy all params except reserved ones
	for k, v := range params {
		if !reservedParams[k] {
			rcloneKey := normalizeRcloneFlag(k)
			if rcloneKey == "" {
				return nil, status.Errorf(codes.InvalidArgument, "invalid parameter name: %s", k)
			}

			pvp.params[rcloneKey] = v
		}
	}

	return pvp, nil
}

// prepareTargetDirectory ensures the target directory exists and is not already mounted
func (ns *NodeServer) prepareTargetDirectory(targetPath string, volumeID string) error {
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		// Check if already mounted
		if !notMnt {
			klog.V(2).Infof("Target path %s is already mounted", targetPath)
			return nil // Signal that mount already exists
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
			return nil
		}

		// Mount appears to exist but is not accessible - recover
		klog.Warningf("Mount point %s appears mounted but is not accessible (err: %v), attempting recovery", targetPath, err)

		if err := ns.mounter.Unmount(targetPath); err != nil {
			klog.Errorf("Failed to unmount corrupted mount point %s: %v", targetPath, err)
			return status.Errorf(codes.Internal, "corrupted mount could not be cleaned up: %v", err)
		}

		klog.V(2).Infof("Successfully unmounted corrupted mount point %s, will remount", targetPath)
	}

	return nil
}

// generateConfigData generates rclone config from parameters if needed
func generateConfigData(pvp *publishVolumeParams) error {
	if pvp.configData == "" && pvp.remoteType != "" {
		klog.V(2).Infof("Generating dynmaic rcone config for remote type: %s", pvp.remoteType)

		// Extract remote params
		remoteParams := extractRemoteTypeParams(pvp.params, pvp.remoteType)

		if len(remoteParams) > 0 {
			pvp.configData = generateRecloneConfigFromParams(remoteParams, pvp.remoteType, pvp.remoteName)
			klog.V(4).Infof("Generated configData: %d bytes", len(pvp.configData))
		}
	}

	if pvp.configData == "" {
		return status.Error(codes.InvalidArgument, "failed to parse configData")
	}

	return nil
}

// loadRcloneConfig loads config into rclone's in-memory storage
func (ns *NodeServer) loadRcloneConfig(ctx context.Context, pvp *publishVolumeParams) ([]string, error) {
	if pvp.configData == "" {
		return nil, nil
	}

	// Parse ALL remotes from configData to support nested remotes
	allRemotes, err := parseAllConfigRemotes(pvp.configData)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse configData: %v", err)
	}

	klog.V(4).Infof("Parsed %d remotes from configData", len(allRemotes))

	// Pre-allocate slice with known capacity
	remotes := make([]string, 0, len(allRemotes))

	updateRemoteOpts := config.UpdateRemoteOpt{
		NonInteractive: true,
		NoObscure:      false,
	}

	// Load all remotes into rclone's in-memory config storage
	ns.configMu.Lock()
	defer ns.configMu.Unlock()

	for remoteName, remoteData := range allRemotes {
		for key, value := range remoteData {
			// Set remote config
			config.LoadedData().SetValue(remoteName, key, value)

			// Get params for a given remote type
			if key == paramBackendTypeKey && len(pvp.params) > 0 {
				remoteParams := extractRemoteTypeParams(pvp.params, value)

				if len(remoteParams) > 0 {
					// Set the remaining values (params)
					if _, err := config.UpdateRemote(ctx, remoteName, remoteParams, updateRemoteOpts); err != nil {
						return nil, status.Errorf(codes.Internal, "failed to update remote: %v", err)
					}
				}
			}
		}
		remotes = append(remotes, remoteName)
		klog.V(4).Infof("Loaded config remote: %s with %d keys", remoteName, len(remoteData))
	}

	return remotes, nil
}

// buildFsPath constructs the filesystem path for rclone
func buildFsPath(remoteName, remotePath string) string {
	if remotePath != "" {
		return fmt.Sprintf("%s:%s", remoteName, remotePath)
	}
	return fmt.Sprintf("%s:", remoteName)
}

// cleanupConfigRemotes removes loaded remotes from rclone
func (ns *NodeServer) cleanupConfigRemotes(remotes []string) {
	if len(remotes) == 0 {
		return
	}

	ns.configMu.Lock()
	defer ns.configMu.Unlock()

	for _, remoteName := range remotes {
		config.LoadedData().DeleteSection(remoteName)
	}
	klog.V(4).Infof("Cleaned up %d remotes", len(remotes))
}

// createAndMountFilesystem initializes and mounts the rclone filesystem
func (ns *NodeServer) createAndMountFilesystem(
	ctx context.Context,
	fsPath, targetPath string,
	mountOptions []string,
	params map[string]string,
) (*mountlib.MountPoint, context.Context, context.CancelFunc, error) {
	// Create per-mount context with isolated config
	ctx, ci := fs.AddConfig(ctx)

	// TODO: REVISIT - Per-mount accounting.Start(ctx) - Is this needed or is global accounting sufficient?
	// Start accounting (bandwidth limiting, stats, TPS limiting)
	// accounting.Start(ctx)

	// Extract volume mount options
	volumeMountOpts, err := extractVolumeMountOptions(mountOptions)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to parse volume mount options: %v", err)
	}

	// Merge both params and mount options
	opts := mergeCopy(params, volumeMountOpts)

	// Set rclone configuration flags
	if err := setRcloneConfigFlags(ctx, ci, opts); err != nil {
		return nil, nil, nil, err
	}

	// Initialize filesystem
	rcloneFs, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to initialize filesystem: %v", err)
	}

	// Extract Rclone mount options
	mountOpts, err := extractMountOptions(opts)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to parse mount options: %v", err)
	}

	// Extract Rclone VFS options
	vfsOpts, err := extractVFSOptions(opts)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to parse VFS options: %v", err)
	}

	// Set device name if not already set
	if mountOpts.DeviceName == "" {
		mountOpts.DeviceName = fsPath
	}

	// Get mount function with enhanced resolution
	mountType, mountFn, err := resolveMountMethod(opts)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.InvalidArgument, "mount method resolution failed: %v", err)
	}

	klog.V(4).Infof("Using mount method: %s", mountType)

	// Create mount point
	mountPoint := mountlib.NewMountPoint(mountFn, targetPath, rcloneFs, mountOpts, vfsOpts)

	// Create context with cancellation for VFS goroutines
	ctx, cancel := context.WithCancel(ctx)

	// Mount the filesystem
	mountDaemon, err := mountPoint.Mount()
	if err != nil {
		cancel()
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to mount: %v", err)
	}

	// Handle rclone daemon if needed
	if err := handleRcloneDaemon(mountPoint, mountDaemon, mountOpts, cancel); err != nil {
		return nil, nil, nil, err
	}

	return mountPoint, ctx, cancel, nil
}

// handleRcloneDaemon manages rclone daemon lifecycle for mount operations.
// It handles daemon process management, timeout waiting, and proper cleanup
// when daemon mode is enabled in mount options.
func handleRcloneDaemon(
	mountPoint *mountlib.MountPoint,
	mountDaemon *os.Process,
	mountOpts *mountlib.Options,
	cancel context.CancelFunc,
) error {
	if mountOpts.Daemon {
		config.PassConfigKeyForDaemonization = true
	}

	if mountOpts.DaemonWait <= 0 {
		// No daemon wait configured, nothing to do
		return nil
	}

	// Wait for mountDaemon, if any...
	killOnce := sync.Once{}
	killDaemon := func(reason string) {
		killOnce.Do(func() {
			if err := mountDaemon.Signal(os.Interrupt); err != nil {
				klog.Errorf("%s. Failed to terminate daemon pid %d: %v", reason, mountDaemon.Pid, err)
				return
			}
			klog.V(2).Infof("%s. Terminating daemon pid %d", reason, mountDaemon.Pid)
		})
	}

	handle := atexit.Register(func() {
		killDaemon("Got interrupt")
	})

	defer atexit.Unregister(handle)

	if err := mountlib.WaitMountReady(
		mountPoint.MountPoint, time.Duration(mountOpts.DaemonWait), mountDaemon,
	); err != nil {
		killDaemon("Daemon timed out")
		cancel()
		return status.Errorf(codes.Internal, "failed to wait for mount: %v", err)
	}

	return nil
}

// waitForVFSCacheSync waits for VFS cache uploads to complete before unmount
func waitForVFSCacheSync(mc *mountContext) {
	if mc == nil || mc.mountPoint == nil {
		return
	}

	// Get VFS stats to check if cache is enabled
	stats := mc.mountPoint.VFS.Stats()

	// Check if diskCache is present (only when cache mode > off)
	_, hasDiskCache := stats["diskCache"].(rc.Params)
	if !hasDiskCache {
		klog.V(4).Infof("VFS cache mode is off, no cache sync needed")
		return
	}

	klog.V(2).Infof("Waiting for VFS cache sync (remote: %s)", mc.remoteName)

	timeout := time.Now().Add(2 * time.Minute)
	retryCount := 0
	maxRetries := 5

	for time.Now().Before(timeout) && retryCount < maxRetries {
		allClear := true

		stats := mc.mountPoint.VFS.Stats()
		if diskCache, ok := stats["diskCache"].(rc.Params); ok {
			uploadsInProgress, _ := diskCache["uploadsInProgress"].(int)
			uploadsQueued, _ := diskCache["uploadsQueued"].(int)

			if uploadsInProgress > 0 || uploadsQueued > 0 {
				klog.V(4).Infof(
					"Waiting for VFS cache uploads (in progress: %d, queued: %d, retry: %d/%d)",
					uploadsInProgress, uploadsQueued, retryCount+1, maxRetries,
				)
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
}

// extractVFSOptions extracts and configures VFS (Virtual File System) options from parameters.
// It loads the default VFS options from rclone's configuration system and then applies
// any overrides provided in the params map. This allows the CSI driver to customize
// VFS behavior such as caching, read-ahead, and file permissions based on volume
// configuration parameters.
func extractVFSOptions(params map[string]string) (*vfscommon.Options, error) {
	vfsOpts := new(vfscommon.Options)

	// Load VFS options from parsed flags
	configMap := fs.ConfigMap("", vfscommon.OptionsInfo, "", nil)
	if err := configstruct.Set(configMap, vfsOpts); err != nil {
		return nil, fmt.Errorf("failed to load VFS options: %v", err)
	}

	// Create a mutable config map and update it
	mutableMap := configmap.Simple{}

	// Copy existing values from the read-only config map
	for _, opt := range vfscommon.OptionsInfo {
		// Set defaults
		if value, ok := configMap.Get(opt.Name); ok {
			mutableMap.Set(opt.Name, value)
		}

		// Override with vfs options in the params
		if value, ok := params[opt.Name]; ok {
			mutableMap.Set(opt.Name, value)
		}
	}

	// update the mutable config
	if err := configstruct.Set(mutableMap, vfsOpts); err != nil {
		return nil, fmt.Errorf("failed to update VFS options: %v", err)
	}

	return vfsOpts, nil
}

// resolveMountMethod resolves the mount method for the current platform.
// It supports user-specified mount methods via the "mount_type" parameter and falls back
// to rclone's default mount method resolution.
//
// If user specifies a mount type, it tries that first. If not available, it returns an error.
// If not specified, it falls back to rclone's default mount method resolution.
func resolveMountMethod(params map[string]string) (string, mountlib.MountFn, error) {
	// Check if user specified a mount type
	if mountType, ok := params[paramMountType]; ok {
		klog.V(4).Infof("Specified mount type: %s", mountType)
		// Try the specified mount type first
		if resolvedType, mountFn := mountlib.ResolveMountMethod(mountType); mountFn != nil {
			return resolvedType, mountFn, nil
		}
		return "", nil, fmt.Errorf("specified mount type '%s' not available", mountType)
	}

	// Fallback to rclone's default resolution
	mountType, mountFn := mountlib.ResolveMountMethod("")
	if mountFn != nil {
		klog.V(4).Infof("Using rclone default mount method: %s", mountType)
		return mountType, mountFn, nil
	}

	return "", nil, fmt.Errorf("no mount methods available")
}

// extractMountOptions extracts and configures mount options from parameters.
// It loads the default mount options from rclone's configuration system and then applies
// any overrides provided in the params map. This allows the CSI driver to customize
// mount behavior such as FUSE options, permissions, and performance settings based on
// volume configuration parameters.
func extractMountOptions(params map[string]string) (*mountlib.Options, error) {
	mountOpts := new(mountlib.Options)

	// Load mount options from parsed flags
	configMap := fs.ConfigMap("", mountlib.OptionsInfo, "", nil)
	if err := configstruct.Set(configMap, mountOpts); err != nil {
		return nil, fmt.Errorf("failed to load mount options: %v", err)
	}

	// Create a mutable config map and update it
	mutableMap := configmap.Simple{}

	// Copy existing values from the read-only config map
	for _, opt := range mountlib.OptionsInfo {
		// Set defaults
		if value, ok := configMap.Get(opt.Name); ok {
			mutableMap.Set(opt.Name, value)
		}

		// Override with mount options in the params
		if value, ok := params[opt.Name]; ok {
			mutableMap.Set(opt.Name, value)
		}
	}

	// update the mutable config
	if err := configstruct.Set(mutableMap, mountOpts); err != nil {
		return nil, fmt.Errorf("failed to update mount options: %v", err)
	}

	return mountOpts, nil
}

// extractVolumeMountOptions parses CSI mount options into a key-value map.
// It handles both key=value format options and boolean flags (without values).
// Boolean flags are automatically set to "true" when no value is provided.
//
// This function is used to convert mount options from the CSI NodePublishVolume
// request into a format that can be used with rclone's configuration system.
//
// Supported formats:
//   - "key=value" -> map["key"] = "value"
//   - "boolean_flag" -> map["boolean_flag"] = "true"
//
// Example:
//
//	Input:  ["ro", "noatime", "uid=1000", "gid=1000"]
//	Output: map[string]string{
//	          "ro": "true",
//	          "noatime": "true",
//	          "uid": "1000",
//	          "gid": "1000"
//	        }
func extractVolumeMountOptions(mountOptions []string) (map[string]string, error) {
	volumeMountOptions := make(map[string]string)

	for _, option := range mountOptions {
		if strings.Contains(option, "=") {
			parts := strings.SplitN(option, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid mount option format: %s", option)
			}

			rcloneKey := normalizeRcloneFlag(parts[0])
			volumeMountOptions[rcloneKey] = parts[1]
		} else {
			rcloneKey := normalizeRcloneFlag(option)
			// Default a boolean value
			volumeMountOptions[rcloneKey] = "true"
		}
	}

	return volumeMountOptions, nil
}

// unmountVolume unmounts the volume and performs cleanup
func (ns *NodeServer) unmountVolume(mc *mountContext, targetPath string) error {
	if mc != nil && mc.mountPoint != nil {
		// Wait for cache sync
		waitForVFSCacheSync(mc)

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

		// Clean up loaded remotes
		if len(mc.remotes) > 0 {
			ns.configMu.Lock()
			for _, remoteName := range mc.remotes {
				config.LoadedData().DeleteSection(remoteName)
			}
			ns.configMu.Unlock()
			klog.V(4).Infof("Deleted %d remotes from config", len(mc.remotes))
		}
	}

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

	return err
}

// NodePublishVolume mounts the rclone volume using direct rclone library integration
//
//nolint:lll
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// Validate request
	if err := validatePublishVolumeRequest(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Acquire lock for this volume operation
	lockKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	if acquired := ns.Driver.volumeLocks.TryAcquire(lockKey); !acquired {
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.Driver.volumeLocks.Release(lockKey)

	// Get mount options from VolumeCapability (CSI standard)
	readOnly := req.GetReadonly()
	mountOptions := req.GetVolumeCapability().GetMount().GetMountFlags()
	if readOnly {
		mountOptions = append(mountOptions, "read-only")
	}

	// Merge parameters from secrets and volume context
	params := ns.mergeVolumeParameters(req)

	// Extract and validate required parameters
	pvp, err := extractPublishParams(params)
	if err != nil {
		return nil, err
	}

	// Prepare target directory and check if already mounted
	if err := ns.prepareTargetDirectory(targetPath, volumeID); err != nil {
		if err.Error() == "" {
			// Already mounted and accessible
			return &csi.NodePublishVolumeResponse{}, nil
		}
		return nil, err
	}

	klog.V(2).Infof("NodePublishVolume: mounting %s:%s at %s", pvp.remoteName, pvp.remotePath, targetPath)

	// Generate config data if needed
	if err := generateConfigData(pvp); err != nil {
		return nil, err
	}

	// Load rclone config
	remotes, err := ns.loadRcloneConfig(ctx, pvp)
	if err != nil {
		return nil, err
	}

	// Build filesystem path
	fsPath := buildFsPath(pvp.remoteName, pvp.remotePath)
	klog.V(2).Infof("Using configData with %d remotes, resolving remote: %s", len(remotes), fsPath)

	// Ensure cleanup on failure
	var mountSuccess bool
	defer func() {
		if !mountSuccess {
			ns.cleanupConfigRemotes(remotes)
		}
	}()

	// Create and mount the filesystem
	mountPoint, ctx, cancel, err := ns.createAndMountFilesystem(ctx, fsPath, targetPath, mountOptions, pvp.params)
	if err != nil {
		return nil, err
	}

	mountSuccess = true

	// Store mount context
	ns.setMountContext(targetPath, &mountContext{
		mountPoint: mountPoint,
		remoteName: pvp.remoteName,
		remotes:    remotes,
		cancel:     cancel,
		ctx:        ctx,
	})

	klog.V(2).Infof("Successfully mounted volume %s to %s (remote: %s)", volumeID, targetPath, pvp.remoteName)
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the rclone volume using direct stats access
//
//nolint:lll
func (ns *NodeServer) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	// Validate request
	if err := validateUnpublishVolumeRequest(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Acquire lock for this volume operation
	lockKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	if acquired := ns.Driver.volumeLocks.TryAcquire(lockKey); !acquired {
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.Driver.volumeLocks.Release(lockKey)

	klog.V(2).Infof("NodeUnpublishVolume: unmounting volume %s from %s", volumeID, targetPath)

	// Get mount context and unmount
	mc := ns.getMountContext(targetPath)
	if err := ns.unmountVolume(mc, targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount target %q: %v", targetPath, err)
	}

	// Remove mount context
	ns.deleteMountContext(targetPath)

	klog.V(2).Infof("Successfully unmounted volume %s from %s", volumeID, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// isMountHealthy checks if a mount is healthy and returns a detailed error message
func (ns *NodeServer) isMountHealthy(targetPath string) (bool, string) {
	// Check if mount point is accessible
	if _, err := os.ReadDir(targetPath); err != nil {
		return false, fmt.Sprintf("mount point not accessible: %v", err)
	}

	// Check VFS stats for errors if mount context is available
	if mc := ns.getMountContext(targetPath); mc != nil {
		if mc.mountPoint != nil && mc.mountPoint.VFS != nil {
			stats := mc.mountPoint.VFS.Stats()
			if errors, ok := stats["errors"]; ok && errors.(int) > 0 {
				return false, fmt.Sprintf("VFS errors detected: %d", errors.(int))
			}
		}
	}

	// If we can read the directory, consider it healthy even if we don't have mount context
	// (this handles edge cases where mount context might be missing but mount is working)
	return true, ""
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

// NodeGetVolumeStats returns volume stats and health condition
//
//nolint:lll
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	volumePath := req.GetVolumePath()
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume path is required")
	}

	klog.V(4).Infof("NodeGetVolumeStats called for volume path: %s", volumePath)

	// Get filesystem statistics
	var statfs unix.Statfs_t
	if err := unix.Statfs(volumePath, &statfs); err != nil {
		klog.Errorf("Failed to get filesystem stats for %s: %v", volumePath, err)
		// Return abnormal condition if we can't read stats
		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{},
			VolumeCondition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  fmt.Sprintf("Failed to get filesystem statistics: %v", err),
			},
		}, nil
	}

	// Calculate volume usage in bytes
	// Note: Bsize might be different on different platforms, so we use int64 to ensure compatibility
	blockSize := int64(statfs.Bsize)
	totalBytes := int64(statfs.Blocks) * blockSize
	availableBytes := int64(statfs.Bavail) * blockSize
	freeBytes := int64(statfs.Bfree) * blockSize
	usedBytes := totalBytes - freeBytes

	// Get inode statistics
	totalInodes := int64(statfs.Files)
	freeInodes := int64(statfs.Ffree)
	usedInodes := totalInodes - freeInodes

	// Build volume usage for bytes
	usage := []*csi.VolumeUsage{
		{
			Available: availableBytes,
			Total:     totalBytes,
			Used:      usedBytes,
			Unit:      csi.VolumeUsage_BYTES,
		},
	}

	// Add inode usage if available
	if totalInodes > 0 {
		usage = append(usage, &csi.VolumeUsage{
			Available: freeInodes,
			Total:     totalInodes,
			Used:      usedInodes,
			Unit:      csi.VolumeUsage_INODES,
		})
	}

	// Check mount health
	healthy, healthMessage := ns.isMountHealthy(volumePath)
	volumeCondition := &csi.VolumeCondition{
		Abnormal: !healthy,
		Message:  healthMessage,
	}

	if healthy {
		volumeCondition.Message = "Volume is healthy and accessible"
	}

	klog.V(4).Infof("Volume stats for %s: Total=%d bytes, Available=%d bytes, Used=%d bytes, Healthy=%v",
		volumePath, totalBytes, availableBytes, usedBytes, healthy)

	return &csi.NodeGetVolumeStatsResponse{
		Usage:          usage,
		VolumeCondition: volumeCondition,
	}, nil
}

// NodeExpandVolume is not implemented
//
//nolint:lll
func (ns *NodeServer) NodeExpandVolume(_ context.Context, _ *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
