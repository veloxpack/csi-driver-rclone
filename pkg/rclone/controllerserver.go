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

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const (
	// Volume context parameters
	paramRemote     = "remote"
	paramRemotePath = "remotePath"
	paramConfigData = "configData"
)

// ControllerServer implements the CSI Controller service
type ControllerServer struct {
	Driver *Driver
	csi.UnimplementedControllerServer
}

// CreateVolume validates volume parameters
//
//nolint:lll
func (cs *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	name := req.GetName()
	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume name must be provided")
	}

	if err := isValidVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	parameters := req.GetParameters()
	if parameters == nil {
		parameters = make(map[string]string)
	}

	// Debug: Log all received parameters
	klog.V(2).Infof("CreateVolume: received %d parameters", len(parameters))
	for k, v := range parameters {
		// Mask sensitive parameters in logs
		if strings.Contains(strings.ToLower(k), "key") || strings.Contains(strings.ToLower(k), "secret") || strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "token") {
			klog.V(4).Infof("CreateVolume parameter: %q = [MASKED]", k)
		} else {
			klog.V(4).Infof("CreateVolume parameter: %q = %q", k, v)
		}
	}

	// NOTE: remote and remotePath validation relaxed to support secret-based configuration
	// Fix for issue #51: Multiple Mounts with Different Secrets
	// When using node-publish-secret, these values can be provided via secrets instead of StorageClass
	// Validation will happen at NodePublishVolume time where secrets are available
	// if parameters[paramRemote] == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "remote parameter is required")
	// }
	// if parameters[paramRemotePath] == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "remotePath parameter is required")
	// }

	// Validate parameters (case-insensitive)
	var remote, remotePath string
	remotePathReplaceMap := map[string]string{}

	// Log parameter validation for debugging
	klog.V(4).Infof("Validating parameters for volume %s", name)

	// Track unknown parameters for better debugging
	var unknownParams []string

	for k, v := range parameters {
		// Trim whitespace from parameter values
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		switch strings.ToLower(k) {
		case strings.ToLower(paramRemote):
			remote = v
		case strings.ToLower(paramRemotePath):
			remotePath = v
		case strings.ToLower(paramConfigData):
			// optional - will be validated at mount time
		case pvcNamespaceKey:
			remotePathReplaceMap[pvcNamespaceMetadata] = v
		case pvcNameKey:
			remotePathReplaceMap[pvcNameMetadata] = v
		case pvNameKey:
			remotePathReplaceMap[pvNameMetadata] = v
		default:
			unknownParams = append(unknownParams, k)
		}
	}

	// Log unknown parameters for debugging
	if len(unknownParams) > 0 {
		klog.V(2).Infof("Unknown parameters in storage class: %v", unknownParams)
	}

	// Apply dynamic path substitution to remotePath if template variables are present
	// Supports: ${pvc.metadata.name}, ${pvc.metadata.namespace}, ${pv.metadata.name}
	if remotePath != "" {
		remotePath = replaceWithMap(remotePath, remotePathReplaceMap)
	}

	// Generate volume ID
	// When remote is provided in StorageClass: use format "remote#name" for better debugging
	// When remote comes from secrets: use just the name (already unique, secrets provide remote at mount time)
	var volumeID string
	if remote != "" {
		volumeID = fmt.Sprintf("%s%s%s", remote, separator, name)
		klog.V(2).Infof("CreateVolume: name=%s, remote=%s, remotePath=%s, volumeID=%s", name, remote, remotePath, volumeID)
	} else {
		volumeID = name
		klog.V(2).Infof("CreateVolume: name=%s, remote from secret, remotePath=%s, volumeID=%s", name, remotePath, volumeID)
	}

	// Build volumeContext from parameters
	volumeContext := make(map[string]string)
	for k, v := range parameters {
		volumeContext[k] = v
	}

	// Update with the resolved remotePath only if it was provided in parameters
	// If remotePath comes from secrets, don't override with empty value
	if remotePath != "" {
		volumeContext[paramRemotePath] = remotePath
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: 0, // rclone doesn't enforce capacity
			VolumeContext: volumeContext,
		},
	}, nil
}

// DeleteVolume is a no-op for rclone (no cleanup needed)
//
//nolint:lll
func (cs *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}

	klog.V(2).Infof("DeleteVolume: volumeID=%s (no-op for rclone)", volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume is not implemented
//
//nolint:lll
func (cs *ControllerServer) ControllerPublishVolume(_ context.Context, _ *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume is not implemented
//
//nolint:lll
func (cs *ControllerServer) ControllerUnpublishVolume(_ context.Context, _ *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetVolume is not implemented
//
//nolint:lll
func (cs *ControllerServer) ControllerGetVolume(_ context.Context, _ *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ValidateVolumeCapabilities validates volume capabilities
//
//nolint:lll
func (cs *ControllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if err := isValidVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
		Message: "",
	}, nil
}

// ListVolumes is not implemented
//
//nolint:lll
func (cs *ControllerServer) ListVolumes(_ context.Context, _ *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity is not implemented
//
//nolint:lll
func (cs *ControllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities returns the capabilities of the controller
//
//nolint:lll
func (cs *ControllerServer) ControllerGetCapabilities(_ context.Context, _ *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

// ControllerExpandVolume is not implemented
//
//nolint:lll
func (cs *ControllerServer) ControllerExpandVolume(_ context.Context, _ *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerModifyVolume is not implemented
//
//nolint:lll
func (cs *ControllerServer) ControllerModifyVolume(_ context.Context, _ *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// isValidVolumeCapabilities validates the given VolumeCapability array is valid
func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) error {
	if len(volCaps) == 0 {
		return fmt.Errorf("volume capabilities missing in request")
	}
	for _, c := range volCaps {
		if c.GetBlock() != nil {
			return fmt.Errorf("block volume capability not supported")
		}
	}
	return nil
}
