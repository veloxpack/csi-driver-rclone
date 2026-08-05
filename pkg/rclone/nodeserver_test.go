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
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getTestNodeServer() (NodeServer, error) {
	d := NewEmptyDriver("")
	d.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_UNKNOWN,
	})
	mounter, err := NewFakeMounter()
	if err != nil {
		return NodeServer{}, errors.New("failed to get fake mounter")
	}
	return NodeServer{
		Driver:  d,
		mounter: mounter,
	}, nil
}

func TestNodeGetInfo(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	req := &csi.NodeGetInfoRequest{}
	resp, err := ns.NodeGetInfo(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fakeNodeID, resp.NodeId)
}

func TestNodeGetCapabilities(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	req := &csi.NodeGetCapabilitiesRequest{}
	resp, err := ns.NodeGetCapabilities(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Capabilities)
	assert.Equal(t, 1, len(resp.Capabilities))
	assert.Equal(t, csi.NodeServiceCapability_RPC_UNKNOWN, resp.Capabilities[0].GetRpc().Type)
}

func TestNodePublishVolumeValidation(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	tests := []struct {
		desc        string
		req         *csi.NodePublishVolumeRequest
		expectedErr error
	}{
		{
			desc:        "Volume ID missing",
			req:         &csi.NodePublishVolumeRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID missing in request"),
		},
		{
			desc: "Target path missing",
			req: &csi.NodePublishVolumeRequest{
				VolumeId: testVolumeID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "Target path not provided"),
		},
		{
			desc: "Volume capability missing",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   testVolumeID,
				TargetPath: "/mnt/test",
			},
			expectedErr: status.Error(codes.InvalidArgument, "Volume capability missing in request"),
		},
		{
			desc: "Remote parameter missing in volume context",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   testVolumeID,
				TargetPath: "/mnt/test",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "remote is required (provide via volumeAttributes or secrets)"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := ns.NodePublishVolume(context.Background(), test.req)
			if err == nil {
				t.Errorf("Expected error but got nil")
			}
			if status.Code(err) != status.Code(test.expectedErr) {
				t.Errorf("Expected error code %v, got %v (error: %v)", status.Code(test.expectedErr), status.Code(err), err)
			}
		})
	}
}

func TestNodeUnpublishVolumeValidation(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	tests := []struct {
		desc        string
		req         *csi.NodeUnpublishVolumeRequest
		expectedErr error
	}{
		{
			desc:        "Volume ID missing",
			req:         &csi.NodeUnpublishVolumeRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID missing in request"),
		},
		{
			desc: "Target path missing",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId: testVolumeID,
			},
			expectedErr: status.Error(codes.InvalidArgument, "Target path missing in request"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := ns.NodeUnpublishVolume(context.Background(), test.req)
			if err == nil {
				t.Errorf("Expected error but got nil")
			}
			if status.Code(err) != status.Code(test.expectedErr) {
				t.Errorf("Expected error code %v, got %v", status.Code(test.expectedErr), status.Code(err))
			}
		})
	}
}

func TestNodeServerMountContext(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	targetPath := "/mnt/test-volume"

	// Initially, no mount context should exist
	mc := ns.getMountContext(targetPath)
	assert.Nil(t, mc)

	// Create a test mount context
	testMC := &mountContext{
		remoteName: "test-remote",
	}
	ns.setMountContext(targetPath, testMC)

	// Verify it was set
	mc = ns.getMountContext(targetPath)
	assert.NotNil(t, mc)
	assert.Equal(t, "test-remote", mc.remoteName)

	// Delete the mount context
	ns.deleteMountContext(targetPath)

	// Verify it was deleted
	mc = ns.getMountContext(targetPath)
	assert.Nil(t, mc)
}

// TestPrepareTargetDirectoryAlreadyMounted verifies the idempotency fix: an
// already-mounted, accessible target reports alreadyMounted=true so NodePublishVolume
// can return a no-op success instead of trying to mount again (issue #86).
func TestPrepareTargetDirectoryAlreadyMounted(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	// The fake mounter treats any path containing "false_is_likely" as already mounted.
	base := t.TempDir()
	targetPath := filepath.Join(base, "false_is_likely_mount")
	assert.NoError(t, os.MkdirAll(targetPath, 0755))

	alreadyMounted, err := ns.prepareTargetDirectory(targetPath, testVolumeID)
	assert.NoError(t, err)
	assert.True(t, alreadyMounted, "an accessible already-mounted target must report alreadyMounted=true")
}

// TestPrepareTargetDirectoryFreshPath verifies a not-yet-mounted target reports
// alreadyMounted=false so the caller proceeds to mount.
func TestPrepareTargetDirectoryFreshPath(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	// Default fake mounter behavior: the path is not a mount point.
	targetPath := t.TempDir()

	alreadyMounted, err := ns.prepareTargetDirectory(targetPath, testVolumeID)
	assert.NoError(t, err)
	assert.False(t, alreadyMounted, "a fresh/unmounted target must report alreadyMounted=false")
}

// TestMountTimeoutDefault verifies the mount timeout falls back to the default when unset
// and honors an explicit configuration.
func TestMountTimeoutDefault(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	assert.Equal(t, defaultMountTimeout, ns.mountTimeout(), "zero MountTimeout must fall back to the default")

	ns.Driver.mountTimeout = 5 * time.Second
	assert.Equal(t, 5*time.Second, ns.mountTimeout(), "configured MountTimeout must be honored")
}

// TestReapOrphanMountReleasesLockOnFailure verifies that when a handed-off mount ultimately
// fails, the reaper releases the volume lock so future retries can proceed (issue #86).
func TestReapOrphanMountReleasesLockOnFailure(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	lockKey := testVolumeID + "-/mnt/target"
	assert.True(t, ns.Driver.volumeLocks.TryAcquire(lockKey))

	resultCh := make(chan mountResult, 1)
	resultCh <- mountResult{err: errors.New("mount failed")}

	ns.reapOrphanMount(resultCh, filepath.Join(t.TempDir(), "unmounted"), nil, lockKey)

	// After reaping, the lock must be free again.
	assert.True(t, ns.Driver.volumeLocks.TryAcquire(lockKey), "reaper must release the lock after a failed orphan mount")
}

// TestReapOrphanMountReleasesLockOnSuccess verifies that when a handed-off mount actually
// came up, the reaper tears it down (bounded) and still releases the lock.
func TestReapOrphanMountReleasesLockOnSuccess(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	lockKey := testVolumeID + "-/mnt/target-ok"
	assert.True(t, ns.Driver.volumeLocks.TryAcquire(lockKey))

	targetPath := t.TempDir()
	cancelled := false
	resultCh := make(chan mountResult, 1)
	// mountPoint is nil (forceUnmountBounded does not dereference it); cancel records invocation.
	resultCh <- mountResult{cancel: func() { cancelled = true }}

	done := make(chan struct{})
	go func() {
		ns.reapOrphanMount(resultCh, targetPath, nil, lockKey)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(gracefulUnmountTimeout + 5*time.Second):
		t.Fatal("reapOrphanMount did not complete within the bounded time")
	}

	assert.True(t, cancelled, "reaper must cancel the mount context before detaching")
	assert.True(t, ns.Driver.volumeLocks.TryAcquire(lockKey),
		"reaper must release the lock after tearing down an orphan mount")
}

func TestPrepareTargetDirectory(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	t.Run("not a mount point returns not-mounted", func(t *testing.T) {
		// fakeMounter.IsLikelyNotMountPoint returns (true, nil) for a plain path.
		targetPath := filepath.Join(t.TempDir(), "target")
		assert.NoError(t, os.MkdirAll(targetPath, 0755))

		alreadyMounted, err := ns.prepareTargetDirectory(targetPath, testVolumeID)
		assert.NoError(t, err)
		assert.False(t, alreadyMounted)
	})

	t.Run("healthy existing mount is idempotent", func(t *testing.T) {
		// "false_is_likely" makes the fake mounter report an existing mount point,
		// and a real, readable directory makes isMountHealthy report it healthy.
		targetPath := filepath.Join(t.TempDir(), "false_is_likely")
		assert.NoError(t, os.MkdirAll(targetPath, 0755))

		alreadyMounted, err := ns.prepareTargetDirectory(targetPath, testVolumeID)
		assert.NoError(t, err)
		assert.True(t, alreadyMounted)
	})

	t.Run("IsLikelyNotMountPoint error is surfaced", func(t *testing.T) {
		// "error_is_likely" makes the fake mounter return a generic error.
		targetPath := filepath.Join(t.TempDir(), "error_is_likely")
		assert.NoError(t, os.MkdirAll(targetPath, 0755))

		alreadyMounted, err := ns.prepareTargetDirectory(targetPath, testVolumeID)
		assert.Error(t, err)
		assert.False(t, alreadyMounted)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestNodePublishVolumeIdempotent(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	// A healthy, already-mounted target ("false_is_likely" + real readable dir) must
	// short-circuit to success without attempting to remount. Regression test for #74.
	targetPath := filepath.Join(t.TempDir(), "false_is_likely")
	assert.NoError(t, os.MkdirAll(targetPath, 0755))

	req := &csi.NodePublishVolumeRequest{
		VolumeId:   testVolumeID,
		TargetPath: targetPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		VolumeContext: map[string]string{
			paramRemote:      "test-remote",
			paramBackendType: "s3",
		},
	}

	resp, err := ns.NodePublishVolume(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// A second publish for the same volume+target must also succeed (idempotent),
	// not return codes.Aborted.
	resp, err = ns.NodePublishVolume(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUnimplementedNodeMethods(t *testing.T) {
	ns, err := getTestNodeServer()
	assert.NoError(t, err)

	ctx := context.Background()

	// Test NodeStageVolume
	_, err = ns.NodeStageVolume(ctx, &csi.NodeStageVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test NodeUnstageVolume
	_, err = ns.NodeUnstageVolume(ctx, &csi.NodeUnstageVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Test NodeGetVolumeStats
	_, err = ns.NodeGetVolumeStats(ctx, &csi.NodeGetVolumeStatsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Test NodeExpandVolume
	_, err = ns.NodeExpandVolume(ctx, &csi.NodeExpandVolumeRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
}
