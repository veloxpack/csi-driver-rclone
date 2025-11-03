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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rclone/rclone/fs/accounting"
	"github.com/rclone/rclone/fs/rc"
	"k8s.io/klog/v2"
)

const (
	unknownValue = "unknown"
)

var (
	// csiCollector is the global CSI metrics collector instance.
	csiCollector *csiRcloneCollector
)

// volumeMetrics aggregates VFS statistics for a single volume across all its mounts
type volumeMetrics struct {
	volumeID            string
	remoteName          string
	inUse               int
	metadataCacheDirs   int
	metadataCacheFiles  int
	diskCacheBytesUsed  int64
	diskCacheFiles      int
	diskCacheErrors     int
	uploadsInProgress   int
	uploadsQueued       int
	diskCacheOutOfSpace bool
}

// csiRcloneCollector implements the Prometheus Collector interface
// for CSI-specific metrics.
type csiRcloneCollector struct {
	ctx         context.Context
	nodeID      string
	driverName  string
	endpoint    string
	versionInfo VersionInfo
	nodeServer  *NodeServer
	nodeInfo    *prometheus.Desc
	// VFS metrics
	vfsInUse               *prometheus.Desc
	vfsMetadataCacheDirs   *prometheus.Desc
	vfsMetadataCacheFiles  *prometheus.Desc
	vfsDiskCacheBytesUsed  *prometheus.Desc
	vfsDiskCacheFiles      *prometheus.Desc
	vfsDiskCacheErrors     *prometheus.Desc
	vfsUploadsInProgress   *prometheus.Desc
	vfsUploadsQueued       *prometheus.Desc
	vfsDiskCacheOutOfSpace *prometheus.Desc
	mountHealthy           *prometheus.Desc
	// Remote statistics metrics
	remoteTransferSpeed    *prometheus.Desc
	remoteTransferEta      *prometheus.Desc
	remoteChecksTotal      *prometheus.Desc
	remoteDeletesTotal     *prometheus.Desc
	remoteServerSideCopies *prometheus.Desc
	remoteServerSideMoves  *prometheus.Desc
	remoteTransferring     *prometheus.Desc
	remoteChecking         *prometheus.Desc
}

// newMetricsCollector creates and returns a new CSI metrics collector instance.
func newMetricsCollector(ctx context.Context, nodeID, driverName, endpoint string, ns *NodeServer) *csiRcloneCollector {
	namespace := "csi_driver_"
	versionInfo := GetVersion(driverName)

	return &csiRcloneCollector{
		ctx:         ctx,
		nodeID:      nodeID,
		driverName:  driverName,
		endpoint:    endpoint,
		versionInfo: versionInfo,
		nodeServer:  ns,
		nodeInfo: prometheus.NewDesc(
			namespace+"info",
			"Information about the CSI driver",
			[]string{"node_id", "driver_name", "endpoint", "rclone_version", "driver_version"},
			nil,
		),
		vfsInUse: prometheus.NewDesc(
			namespace+"vfs_file_handles_in_use",
			"Number of file handles currently in use for this mount",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsMetadataCacheDirs: prometheus.NewDesc(
			namespace+"vfs_metadata_cache_dirs_total",
			"Number of directories in the VFS metadata cache",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsMetadataCacheFiles: prometheus.NewDesc(
			namespace+"vfs_metadata_cache_files_total",
			"Number of files in the VFS metadata cache",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsDiskCacheBytesUsed: prometheus.NewDesc(
			namespace+"vfs_disk_cache_bytes_used",
			"Bytes used by the VFS disk cache",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsDiskCacheFiles: prometheus.NewDesc(
			namespace+"vfs_disk_cache_files_total",
			"Number of files in the VFS disk cache",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsDiskCacheErrors: prometheus.NewDesc(
			namespace+"vfs_disk_cache_errored_files_total",
			"Number of files with errors in the VFS disk cache",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsUploadsInProgress: prometheus.NewDesc(
			namespace+"vfs_uploads_in_progress_total",
			"Number of uploads currently in progress",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsUploadsQueued: prometheus.NewDesc(
			namespace+"vfs_uploads_queued_total",
			"Number of uploads queued for processing",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		vfsDiskCacheOutOfSpace: prometheus.NewDesc(
			namespace+"vfs_disk_cache_out_of_space",
			"Whether the VFS disk cache is out of space (1=yes, 0=no)",
			[]string{"volume_id", "remote_name"},
			nil,
		),
		mountHealthy: prometheus.NewDesc(
			namespace+"mount_healthy",
			"Mount health status with mount details (1=healthy, 0=unhealthy)",
			[]string{
				"volume_id", "pod_id", "target_path", "remote_name", "mount_type",
				"device_name", "volume_name", "read_only", "mount_duration_seconds",
			},
			nil,
		),
		// Remote statistics metrics
		remoteTransferSpeed: prometheus.NewDesc(
			namespace+"remote_transfer_speed_bytes_per_second",
			"Current transfer speed in bytes per second",
			nil,
			nil,
		),
		remoteTransferEta: prometheus.NewDesc(
			namespace+"remote_transfer_eta_seconds",
			"Estimated time to completion in seconds",
			nil,
			nil,
		),
		remoteChecksTotal: prometheus.NewDesc(
			namespace+"remote_checks_total",
			"Total number of file checks completed",
			nil,
			nil,
		),
		remoteDeletesTotal: prometheus.NewDesc(
			namespace+"remote_deletes_total",
			"Total number of files deleted",
			nil,
			nil,
		),
		remoteServerSideCopies: prometheus.NewDesc(
			namespace+"remote_server_side_copies_total",
			"Total number of server-side copies",
			nil,
			nil,
		),
		remoteServerSideMoves: prometheus.NewDesc(
			namespace+"remote_server_side_moves_total",
			"Total number of server-side moves",
			nil,
			nil,
		),
		remoteTransferring: prometheus.NewDesc(
			namespace+"remote_transferring_files",
			"Number of files currently being transferred",
			nil,
			nil,
		),
		remoteChecking: prometheus.NewDesc(
			namespace+"remote_checking_files",
			"Number of files currently being checked",
			nil,
			nil,
		),
	}
}

// initMetricsCollector initializes and registers both the rclone and CSI Prometheus collectors.
// It ensures that collectors are only initialized once.
func initMetricsCollector(ctx context.Context, nodeID, driverName, endpoint string, ns *NodeServer) error {
	if csiCollector != nil {
		klog.V(4).Info("CSI collector already initialized; skipping re-initialization")
		return nil
	}

	// Register rclone collector
	rcloneCollector := accounting.NewRcloneCollector(ctx)
	if err := prometheus.Register(rcloneCollector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			klog.V(4).Info("rclone Prometheus collector already registered")
		} else {
			klog.Warningf("failed to register rclone Prometheus collector: %v", err)
		}
	}

	// Create and register CSI collector
	csiCollector = newMetricsCollector(ctx, nodeID, driverName, endpoint, ns)
	if err := prometheus.Register(csiCollector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			klog.V(4).Info("CSI Prometheus collector already registered")
		} else {
			return fmt.Errorf("failed to register CSI Prometheus collector: %w", err)
		}
	}

	return nil
}

// Describe implements prometheus.Collector.
func (c *csiRcloneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeInfo
	ch <- c.vfsInUse
	ch <- c.vfsMetadataCacheDirs
	ch <- c.vfsMetadataCacheFiles
	ch <- c.vfsDiskCacheBytesUsed
	ch <- c.vfsDiskCacheFiles
	ch <- c.vfsDiskCacheErrors
	ch <- c.vfsUploadsInProgress
	ch <- c.vfsUploadsQueued
	ch <- c.vfsDiskCacheOutOfSpace
	ch <- c.mountHealthy
	ch <- c.remoteTransferSpeed
	ch <- c.remoteTransferEta
	ch <- c.remoteChecksTotal
	ch <- c.remoteDeletesTotal
	ch <- c.remoteServerSideCopies
	ch <- c.remoteServerSideMoves
	ch <- c.remoteTransferring
	ch <- c.remoteChecking
}

// Collect implements prometheus.Collector.
func (c *csiRcloneCollector) Collect(ch chan<- prometheus.Metric) {
	if c == nil {
		klog.Warning("CSI collector is nil; skipping collection")
		return
	}

	c.collectNodeInfo(ch)
	c.collectVFSMetrics(ch)
	c.collectRemoteStats(ch)

	klog.V(2).Infof("CSI metrics collection completed node_id=%s", c.nodeID)
}

// collectNodeInfo emits the node info metric
func (c *csiRcloneCollector) collectNodeInfo(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.nodeInfo,
		prometheus.GaugeValue,
		1,
		c.nodeID,
		c.driverName,
		c.endpoint,
		c.versionInfo.RcloneVersion,
		c.versionInfo.DriverVersion,
	)
}

// collectVFSMetrics aggregates and emits VFS metrics for all mounted volumes
func (c *csiRcloneCollector) collectVFSMetrics(ch chan<- prometheus.Metric) {
	if c.nodeServer == nil {
		return
	}

	c.nodeServer.mu.RLock()
	defer c.nodeServer.mu.RUnlock()

	// Aggregate metrics by volume_id
	volumeStats := c.aggregateVolumeStats()

	// Emit aggregated VFS metrics
	c.emitVolumeMetrics(ch, volumeStats)

	// Emit mount health metrics
	c.collectMountHealthMetrics(ch)
}

// aggregateVolumeStats aggregates VFS statistics across all mount points by volume ID
func (c *csiRcloneCollector) aggregateVolumeStats() map[string]*volumeMetrics {
	volumeStats := make(map[string]*volumeMetrics)

	for targetPath, mc := range c.nodeServer.mountContext {
		if mc == nil || mc.mountPoint == nil || mc.mountPoint.VFS == nil {
			continue
		}

		volumeID := extractVolumeID(targetPath)

		// Initialize volume stats if not exists
		if volumeStats[volumeID] == nil {
			volumeStats[volumeID] = &volumeMetrics{
				volumeID:   volumeID,
				remoteName: mc.remoteName,
			}
		}

		// Aggregate VFS statistics
		c.aggregateVFSStats(mc.mountPoint.VFS.Stats(), volumeStats[volumeID])
	}

	return volumeStats
}

// aggregateVFSStats aggregates individual VFS stats into volume metrics
func (c *csiRcloneCollector) aggregateVFSStats(stats map[string]interface{}, vm *volumeMetrics) {
	// Aggregate inUse metric
	if inUse, ok := stats["inUse"].(int32); ok {
		vm.inUse += int(inUse)
	}

	// Aggregate metadata cache metrics
	if metadataCache, ok := stats["metadataCache"].(rc.Params); ok {
		if dirs, ok := metadataCache["dirs"].(int); ok {
			vm.metadataCacheDirs += dirs
		}
		if files, ok := metadataCache["files"].(int); ok {
			vm.metadataCacheFiles += files
		}
	}

	// Aggregate disk cache metrics
	c.aggregateDiskCacheStats(stats, vm)
}

// aggregateDiskCacheStats aggregates disk cache statistics into volume metrics
func (c *csiRcloneCollector) aggregateDiskCacheStats(stats map[string]interface{}, vm *volumeMetrics) {
	diskCache, ok := stats["diskCache"].(rc.Params)
	if !ok {
		return
	}

	if bytesUsed, ok := diskCache["bytesUsed"].(int64); ok {
		vm.diskCacheBytesUsed += bytesUsed
	}
	if files, ok := diskCache["files"].(int); ok {
		vm.diskCacheFiles += files
	}
	if erroredFiles, ok := diskCache["erroredFiles"].(int); ok {
		vm.diskCacheErrors += erroredFiles
	}
	if uploadsInProgress, ok := diskCache["uploadsInProgress"].(int); ok {
		vm.uploadsInProgress += uploadsInProgress
	}
	if uploadsQueued, ok := diskCache["uploadsQueued"].(int); ok {
		vm.uploadsQueued += uploadsQueued
	}
	if outOfSpace, ok := diskCache["outOfSpace"].(bool); ok && outOfSpace {
		vm.diskCacheOutOfSpace = true
	}
}

// emitVolumeMetrics emits all aggregated volume metrics
func (c *csiRcloneCollector) emitVolumeMetrics(ch chan<- prometheus.Metric, volumeStats map[string]*volumeMetrics) {
	for _, vs := range volumeStats {
		c.emitSingleVolumeMetrics(ch, vs)
	}
}

// emitSingleVolumeMetrics emits metrics for a single volume
func (c *csiRcloneCollector) emitSingleVolumeMetrics(ch chan<- prometheus.Metric, vs *volumeMetrics) {
	ch <- prometheus.MustNewConstMetric(
		c.vfsInUse, prometheus.GaugeValue, float64(vs.inUse), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsMetadataCacheDirs, prometheus.GaugeValue, float64(vs.metadataCacheDirs), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsMetadataCacheFiles, prometheus.GaugeValue, float64(vs.metadataCacheFiles), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsDiskCacheBytesUsed, prometheus.GaugeValue, float64(vs.diskCacheBytesUsed), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsDiskCacheFiles, prometheus.GaugeValue, float64(vs.diskCacheFiles), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsDiskCacheErrors, prometheus.CounterValue, float64(vs.diskCacheErrors), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsUploadsInProgress, prometheus.GaugeValue, float64(vs.uploadsInProgress), vs.volumeID, vs.remoteName,
	)
	ch <- prometheus.MustNewConstMetric(
		c.vfsUploadsQueued, prometheus.GaugeValue, float64(vs.uploadsQueued), vs.volumeID, vs.remoteName,
	)

	outOfSpaceValue := float64(0)
	if vs.diskCacheOutOfSpace {
		outOfSpaceValue = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.vfsDiskCacheOutOfSpace, prometheus.GaugeValue, outOfSpaceValue, vs.volumeID, vs.remoteName,
	)
}

// collectMountHealthMetrics emits mount health status for each mount point
func (c *csiRcloneCollector) collectMountHealthMetrics(ch chan<- prometheus.Metric) {
	for targetPath, mc := range c.nodeServer.mountContext {
		if mc == nil || mc.mountPoint == nil {
			continue
		}

		healthValue := float64(0)
		healthy, _ := c.nodeServer.isMountHealthy(targetPath)
		if healthy {
			healthValue = 1
		}

		ch <- prometheus.MustNewConstMetric(
			c.mountHealthy,
			prometheus.GaugeValue,
			healthValue,
			extractVolumeID(targetPath),
			extractPodID(targetPath),
			targetPath,
			mc.remoteName,
			extractMountType(mc),
			getDeviceName(mc),
			getVolumeName(mc),
			isReadOnly(mc),
			getMountDuration(mc),
		)
	}
}

// collectRemoteStats collects remote transfer statistics from rclone's global accounting
func (c *csiRcloneCollector) collectRemoteStats(ch chan<- prometheus.Metric) {
	globalStats := accounting.GlobalStats()
	remoteStats, err := globalStats.RemoteStats(false)
	if err != nil {
		return
	}

	c.emitTransferMetrics(ch, remoteStats)
	c.emitOperationMetrics(ch, remoteStats)
}

// emitTransferMetrics emits transfer-related metrics
func (c *csiRcloneCollector) emitTransferMetrics(ch chan<- prometheus.Metric, remoteStats rc.Params) {
	if speed, ok := remoteStats["speed"].(float64); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteTransferSpeed, prometheus.GaugeValue, speed)
	}

	if eta := remoteStats["eta"]; eta != nil {
		if etaSeconds, ok := eta.(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.remoteTransferEta, prometheus.GaugeValue, etaSeconds)
		}
	}
}

// emitOperationMetrics emits operation-related metrics (checks, deletes, copies, moves, etc.)
func (c *csiRcloneCollector) emitOperationMetrics(ch chan<- prometheus.Metric, remoteStats rc.Params) {
	if checks, ok := remoteStats["checks"].(int64); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteChecksTotal, prometheus.CounterValue, float64(checks))
	}

	if deletes, ok := remoteStats["deletes"].(int64); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteDeletesTotal, prometheus.CounterValue, float64(deletes))
	}

	if serverSideCopies, ok := remoteStats["serverSideCopies"].(int64); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteServerSideCopies, prometheus.CounterValue, float64(serverSideCopies))
	}

	if serverSideMoves, ok := remoteStats["serverSideMoves"].(int64); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteServerSideMoves, prometheus.CounterValue, float64(serverSideMoves))
	}

	if transferring, ok := remoteStats["transferring"].([]rc.Params); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteTransferring, prometheus.GaugeValue, float64(len(transferring)))
	}

	if checking, ok := remoteStats["checking"].([]string); ok {
		ch <- prometheus.MustNewConstMetric(c.remoteChecking, prometheus.GaugeValue, float64(len(checking)))
	}
}

// Helper function to extract volume ID from target path
func extractVolumeID(targetPath string) string {
	// Target path format: /var/lib/kubelet/pods/{pod-uid}/volumes/kubernetes.io~csi/{volumeID}/mount
	parts := strings.Split(targetPath, "/")
	for i, part := range parts {
		if part == "kubernetes.io~csi" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	// Fallback: use last component
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return unknownValue
}

// Helper function to extract pod ID from target path
func extractPodID(targetPath string) string {
	// Target path format: /var/lib/kubelet/pods/{pod-uid}/volumes/kubernetes.io~csi/{volumeID}/mount
	parts := strings.Split(targetPath, "/")
	for i, part := range parts {
		if part == "pods" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return unknownValue
}

// Helper function to extract mount type from mount function
func extractMountType(mc *mountContext) string {
	if mc == nil || mc.mountPoint == nil {
		return unknownValue
	}
	// Determine mount type based on mount function or other identifiers
	// This may require additional context from mountlib
	return "mount" // default
}

// Helper function to get mount duration in seconds
func getMountDuration(mc *mountContext) string {
	if mc == nil || mc.mountPoint == nil {
		return "0"
	}
	duration := time.Since(mc.mountPoint.MountedOn).Seconds()
	return fmt.Sprintf("%.0f", duration)
}

// Helper function to get device name
func getDeviceName(mc *mountContext) string {
	if mc == nil || mc.mountPoint == nil {
		return unknownValue
	}
	deviceName := mc.mountPoint.MountOpt.DeviceName
	if deviceName == "" {
		return unknownValue
	}
	return deviceName
}

// Helper function to get volume name
func getVolumeName(mc *mountContext) string {
	if mc == nil || mc.mountPoint == nil {
		return unknownValue
	}
	volumeName := mc.mountPoint.MountOpt.VolumeName
	if volumeName == "" {
		return unknownValue
	}
	return volumeName
}

// Helper function to determine read-only status
func isReadOnly(mc *mountContext) string {
	if mc == nil || mc.mountPoint == nil {
		return "false"
	}
	// Check VFS options for read-only mode
	if mc.mountPoint.VFSOpt.ReadOnly {
		return "true"
	}
	return "false"
}
