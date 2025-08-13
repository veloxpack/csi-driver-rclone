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
	"fmt"
	"strings"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
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
