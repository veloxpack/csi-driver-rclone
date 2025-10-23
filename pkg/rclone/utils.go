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
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/rclone/rclone/fs/rc"
	"github.com/unknwon/goconfig"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

const (
	// Separator is the separator used in volume IDs
	separator                       = "#"
	volumeOperationAlreadyExistsFmt = "An operation with the given Volume ID %s already exists"
)

// VolumeLocks manages locks for volume operations
type VolumeLocks struct {
	locks sets.Set[string] //nolint:staticcheck
	mux   sync.Mutex
}

// NewVolumeLocks creates a new VolumeLocks instance
func NewVolumeLocks() *VolumeLocks {
	return &VolumeLocks{
		locks: make(sets.Set[string]),
	}
}

// TryAcquire attempts to acquire a lock for the given volumeID
func (vl *VolumeLocks) TryAcquire(volumeID string) bool {
	vl.mux.Lock()
	defer vl.mux.Unlock()
	if vl.locks.Has(volumeID) {
		return false
	}
	vl.locks.Insert(volumeID)
	return true
}

// Release releases the lock for the given volumeID
func (vl *VolumeLocks) Release(volumeID string) {
	vl.mux.Lock()
	defer vl.mux.Unlock()
	vl.locks.Delete(volumeID)
}

// ParseEndpoint parses the CSI endpoint
func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}

	// Support unix:// without protocol prefix for backward compatibility
	if strings.HasPrefix(ep, "/") {
		return "unix", ep, nil
	}

	return "", "", fmt.Errorf("invalid endpoint: %v", ep)
}

// NewDefaultIdentityServer creates a new default IdentityServer
func NewDefaultIdentityServer(d *Driver) *IdentityServer {
	return &IdentityServer{
		Driver: d,
	}
}

// NewControllerServer creates a new ControllerServer
func NewControllerServer(d *Driver) *ControllerServer {
	return &ControllerServer{
		Driver: d,
	}
}

// NewControllerServiceCapability creates a new controller service capability
func NewControllerServiceCapability(c csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: c,
			},
		},
	}
}

// NewNodeServiceCapability creates a new node service capability
func NewNodeServiceCapability(c csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: c,
			},
		},
	}
}

// getLogLevel returns the appropriate log level for different gRPC methods
func getLogLevel(method string) int32 {
	if method == "/csi.v1.Identity/Probe" ||
		method == "/csi.v1.Node/NodeGetCapabilities" ||
		method == "/csi.v1.Node/NodeGetVolumeStats" {
		return 8
	}
	return 2
}

// logGRPC is a gRPC unary interceptor for logging
//
//nolint:lll
func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	level := klog.Level(getLogLevel(info.FullMethod))
	klog.V(level).Infof("GRPC call: %s", info.FullMethod)
	klog.V(level).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))

	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.V(level).Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

const (
	// flagPrefixSeparator is the character that separates remote name from flag name
	flagPrefixSeparator = "-"
	// flagLongPrefix is the prefix used for long command-line options
	flagLongPrefix = "--"
	// flagHyphen is the character to be replaced with underscore
	flagHyphen = "-"
	// flagUnderscore is the replacement character for hyphens
	flagUnderscore = "_"
)

// sanitizeFlag normalizes a flag key by removing the remote prefix and standardizing format.
// The remote prefix comparison is case-insensitive, so both "s3-endpoint" and "S3-endpoint"
// will be treated the same way when remote is "s3".
//
// Transformations applied:
//  1. Remove remote prefix (case-insensitive): "s3-endpoint" or "S3-endpoint" -> "endpoint"
//  2. Remove leading "--" prefix: "--endpoint" -> "endpoint"
//  3. Replace hyphens with underscores: "cache-mode" -> "cache_mode"
//  4. Convert to lowercase: "EndPoint" -> "endpoint"
//
// Example: sanitizeFlag("s3", "S3-cache-mode") -> "cache_mode"
func sanitizeFlag(remote, key string) string {
	if key == "" {
		return key
	}

	// Only remove remote prefix if remote is not empty
	if remote != "" {
		// Create the prefix pattern (remote + separator) and remove it case-insensitively
		prefix := fmt.Sprintf("%s%s", remote, flagPrefixSeparator)
		key = removePrefixCaseInsensitive(key, prefix)
	}

	return normalizeRcloneFlag(key)
}

// normalizeRcloneFlag normalizes rclone flag names by standardizing their format.
// It performs several transformations to ensure consistent flag naming across
// different input formats and makes them compatible with rclone's internal
// configuration system.
//
// Transformations applied:
//  1. Remove leading "--" prefix (used for long command-line options)
//  2. Replace all hyphens with underscores for consistency
//  3. Convert to lowercase for case-insensitive matching
//
// This function is used to ensure that flag names from various sources
// (command line, config files, volume parameters) are normalized to a
// consistent format that rclone can understand.
//
// Example transformations:
//
//	"--cache-mode" -> "cache_mode"
//	"Cache-Mode" -> "cache_mode"
//	"vfs-read-ahead" -> "vfs_read_ahead"
func normalizeRcloneFlag(key string) string {
	if key == "" {
		return key
	}

	// Remove any leading "--" prefix (used for long command-line options)
	// even though we don't normally pass prefixed args, we keep this for correctness
	key = strings.TrimPrefix(key, flagLongPrefix)

	// Replace all remaining hyphens with underscores
	// this makes the key more consistent and easier to work with
	key = strings.ReplaceAll(key, flagHyphen, flagUnderscore)

	// Convert the key to lowercase for consistency
	key = strings.ToLower(key)

	return key
}

// removePrefixCaseInsensitive removes the prefix from the string in a case-insensitive manner.
// Returns the original string if the prefix is not found.
func removePrefixCaseInsensitive(s, prefix string) string {
	if len(s) < len(prefix) {
		return s
	}

	// Check if the prefix matches case-insensitively
	if strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):]
	}

	return s
}

// parseAllConfigRemotes parses rclone config data (INI format) and extracts all sections
// This supports nested remotes (crypt, alias, chunker, union, etc.) by loading the entire config
func parseAllConfigRemotes(configData string) (map[string]map[string]string, error) {
	// Parse INI-style config data using goconfig
	gc, err := goconfig.LoadFromReader(bytes.NewReader([]byte(configData)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config data: %w", err)
	}

	allSections := make(map[string]map[string]string)

	// Iterate through all sections in the config
	for _, sectionName := range gc.GetSectionList() {
		// Skip the default section
		if sectionName == "DEFAULT" {
			continue
		}

		sectionData := make(map[string]string)
		keys := gc.GetKeyList(sectionName)

		// Extract all key-value pairs for this section
		for _, key := range keys {
			value, err := gc.GetValue(sectionName, key)
			if err == nil {
				sectionData[key] = value
			}
		}

		allSections[sectionName] = sectionData
	}

	return allSections, nil
}

// extractRemoteTypeParams filters and returns parameters for a given remote type.
//
// It scans through the given `params` map and collects only the key–value pairs
// where the key starts with the specified `remoteType` prefix (e.g., "s3", "b2").
// The matching keys are sanitized with `sanitizeFlag` before being added to the
// resulting `rc.Params` map.
//
// Example:
//
//	params := map[string]string{
//		"s3_bucket": "my-bucket",
//		"s3_region": "us-east-1",
//		"b2_key":    "abcd1234",
//	}
//
//	rcParams := getRemoteTypeParams(params, "s3")
//
//	// rcParams will contain:
//	// {
//	//   "bucket": "mybucket",
//	//   "region": "us-east-1",
//	// }
//
// Note: The `sanitizeFlag` function is expected to remove the prefix and clean
// the key name (e.g., turning "s3_bucket" into "bucket").
func extractRemoteTypeParams(params map[string]string, remoteType string) rc.Params {
	rcParams := make(rc.Params)

	for k, v := range params {
		if strings.HasPrefix(strings.ToLower(k), strings.ToLower(remoteType)) {
			delete(params, k)
			// Remove the remote type prefix
			k = sanitizeFlag(remoteType, k)
			rcParams[k] = v
		}
	}

	return rcParams
}

// generateRecloneConfigFromParams builds and returns an INI-formatted string
// given a section header and key-value pairs.
//
// Example:
//
// input:
//
//	header = "minio-sample"
//	params = map[string]string{
//	    "type": "s3",
//	    "provider": "Minio",
//	    "endpoint": "http://localhost:9000",
//	}
//
// output:
//
//	[minio-sample]
//	type = s3
//	provider = Minio
//	endpoint = http://localhost:9000
func generateRecloneConfigFromParams(params rc.Params, remoteType, remoteName string) string {
	var sb strings.Builder

	// Write section header
	sb.WriteString(fmt.Sprintf("[%s]\n", remoteName))

	// Add backend type
	sb.WriteString(fmt.Sprintf("%s = %s\n", "type", remoteType))

	// Write key-value pairs
	for key, value := range params {
		sb.WriteString(fmt.Sprintf("%s = %s\n", key, value))
	}

	return sb.String()
}

// mergeCopy returns a new map containing all key-value pairs
// from both m1 and m2. Neither input map is modified.
//
// If the same key exists in both maps, the value from m2
// overwrites the value from m1.
//
// Example:
//
//	m1 := map[string]int{"a": 1, "b": 2}
//	m2 := map[string]int{"b": 3, "c": 4}
//	merged := MergeCopy(m1, m2)
//	// merged == map[string]int{"a":1, "b":3, "c":4}
func mergeCopy[K comparable, V any](m1, m2 map[K]V) map[K]V {
	merged := make(map[K]V)
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		merged[k] = v
	}
	return merged
}
