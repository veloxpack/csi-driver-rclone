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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
)

const (
	// Secret naming and labeling constants
	secretNamePrefix = "rclone-mount-state-"
	secretHashLength = 16

	// Kubernetes labels
	labelAppName      = "app.kubernetes.io/name"
	labelComponent    = "app.kubernetes.io/component"
	labelVolumeID     = "rclone.csi.veloxpack.io/volume-id"
	labelValueAppName = "csi-driver-rclone"
	labelValueComp    = "mount-state"

	// Secret data keys
	keyVolumeID     = "volumeId"
	keyTargetPath   = "targetPath"
	keyTimestamp    = "timestamp"
	keyConfigData   = "configData"
	keyRemoteName   = "remoteName"
	keyRemotePath   = "remotePath"
	keyRemoteType   = "remoteType"
	keyMountParams  = "mountParams"
	keyMountOptions = "mountOptions"
	keyReadOnly     = "readonly"

	// Default namespace
	defaultNamespace = "default"
)

// MountState represents the complete state needed to remount a volume.
// All fields are persisted to Kubernetes Secrets for recovery scenarios.
type MountState struct {
	// Core volume identification
	VolumeID   string    `json:"volumeId"`
	TargetPath string    `json:"targetPath"`
	Timestamp  time.Time `json:"timestamp"`

	// Rclone configuration (includes credentials)
	ConfigData string `json:"configData"`
	RemoteName string `json:"remoteName"`
	RemotePath string `json:"remotePath"`
	RemoteType string `json:"remoteType"`

	// Mount configuration
	MountParams  map[string]string `json:"mountParams"`
	MountOptions []string          `json:"mountOptions"`
	ReadOnly     bool              `json:"readonly"`
}

// Validate checks if the MountState contains required fields.
func (ms *MountState) Validate() error {
	if ms == nil {
		return fmt.Errorf("mount state is nil")
	}
	if ms.VolumeID == "" {
		return fmt.Errorf("volumeID is required")
	}
	if ms.TargetPath == "" {
		return fmt.Errorf("targetPath is required")
	}
	return nil
}

// MountStateManager handles persistence of volume mount states in Kubernetes Secrets.
// It provides thread-safe operations for storing and retrieving mount state information.
type MountStateManager struct {
	namespace string
	secrets   typedv1.SecretInterface
	mu        sync.RWMutex
}

// newMountStateManager creates a new MountStateManager instance.
// It initializes the Kubernetes client and sets up the secret interface.
func NewMountStateManager(namespace string) (*MountStateManager, error) {
	if namespace == "" {
		namespace = defaultNamespace
	}

	clientset, err := getK8sClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize k8s client: %w", err)
	}

	klog.V(2).Infof("Initialized mount state manager (namespace: %s)", namespace)

	return &MountStateManager{
		namespace: namespace,
		secrets:   clientset.CoreV1().Secrets(namespace),
	}, nil
}

// makeSecretName creates a deterministic secret name from volume ID.
// Uses SHA-256 hash to ensure consistent naming while keeping names readable.
func (sm *MountStateManager) makeSecretName(volumeID string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(volumeID)))[:secretHashLength]
	return secretNamePrefix + hash
}

// GetState retrieves the complete mount state for a specific volume.
// Returns nil without error if no state exists for the volume.
func (sm *MountStateManager) GetState(ctx context.Context, volumeID, targetPath string) (*MountState, error) {
	if volumeID == "" {
		return nil, fmt.Errorf("volumeID is required")
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secretName := sm.makeSecretName(volumeID)
	secret, err := sm.secrets.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("No state found for volume %s", volumeID)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get state secret %s: %w", secretName, err)
	}

	state, err := sm.deserializeSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize secret %s: %w", secretName, err)
	}

	klog.V(4).Infof("Retrieved state for volume %s from secret %s", volumeID, secretName)
	return state, nil
}

// deserializeSecret converts a Kubernetes Secret into a MountState struct.
func (sm *MountStateManager) deserializeSecret(secret *v1.Secret) (*MountState, error) {
	state := &MountState{
		VolumeID:   byteToString(secret.Data[keyVolumeID]),
		TargetPath: byteToString(secret.Data[keyTargetPath]),
		ConfigData: byteToString(secret.Data[keyConfigData]),
		RemoteName: byteToString(secret.Data[keyRemoteName]),
		RemotePath: byteToString(secret.Data[keyRemotePath]),
		RemoteType: byteToString(secret.Data[keyRemoteType]),
		ReadOnly:   byteToString(secret.Data[keyReadOnly]) == "true",
	}

	// Parse timestamp
	if ts := string(secret.Data[keyTimestamp]); ts != "" {
		var err error
		state.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			klog.Warningf("Failed to parse timestamp '%s': %v", ts, err)
		}
	}

	// Parse mount params JSON
	if paramsJSON := secret.Data[keyMountParams]; len(paramsJSON) > 0 {
		if err := json.Unmarshal(paramsJSON, &state.MountParams); err != nil {
			klog.Warningf("Failed to parse mount params: %v", err)
			state.MountParams = make(map[string]string)
		}
	} else {
		state.MountParams = make(map[string]string)
	}

	// Parse mount options JSON
	if optsJSON := secret.Data[keyMountOptions]; len(optsJSON) > 0 {
		if err := json.Unmarshal(optsJSON, &state.MountOptions); err != nil {
			klog.Warningf("Failed to parse mount options: %v", err)
			state.MountOptions = make([]string, 0)
		}
	} else {
		state.MountOptions = make([]string, 0)
	}

	return state, nil
}

// SaveState persists complete volume state as a Kubernetes Secret.
// Creates a new secret or updates an existing one.
func (sm *MountStateManager) SaveState(ctx context.Context, state *MountState) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	secret, err := sm.buildSecret(state)
	if err != nil {
		return fmt.Errorf("failed to build secret: %w", err)
	}

	// Try create, fall back to update if already exists
	_, err = sm.secrets.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = sm.secrets.Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update state secret %s: %w", secret.Name, err)
			}
			klog.V(4).Infof("Updated mount state secret %s/%s", sm.namespace, secret.Name)
		} else {
			return fmt.Errorf("failed to create state secret %s: %w", secret.Name, err)
		}
	} else {
		klog.V(4).Infof("Created mount state secret %s/%s", sm.namespace, secret.Name)
	}

	return nil
}

// buildSecret constructs a Kubernetes Secret from a MountState.
func (sm *MountStateManager) buildSecret(state *MountState) (*v1.Secret, error) {
	secretName := sm.makeSecretName(state.VolumeID)

	// Serialize mount params to JSON
	paramsJSON, err := json.Marshal(state.MountParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mount params: %w", err)
	}

	// Serialize mount options to JSON
	mountOptsJSON, err := json.Marshal(state.MountOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mount options: %w", err)
	}

	// Set timestamp if not already set
	timestamp := state.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: sm.namespace,
			Labels: map[string]string{
				labelAppName:   labelValueAppName,
				labelComponent: labelValueComp,
				labelVolumeID:  state.VolumeID,
			},
		},
		Type: v1.SecretTypeOpaque,
		StringData: map[string]string{
			keyVolumeID:     state.VolumeID,
			keyTargetPath:   state.TargetPath,
			keyConfigData:   state.ConfigData,
			keyRemoteName:   state.RemoteName,
			keyRemotePath:   state.RemotePath,
			keyRemoteType:   state.RemoteType,
			keyMountParams:  string(paramsJSON),
			keyMountOptions: string(mountOptsJSON),
			keyTimestamp:    timestamp.Format(time.RFC3339),
			keyReadOnly:     fmt.Sprintf("%v", state.ReadOnly),
		},
	}

	return secret, nil
}

// DeleteState removes a volume state Secret from Kubernetes.
// Does not return an error if the secret doesn't exist.
func (sm *MountStateManager) DeleteState(ctx context.Context, volumeID, targetPath string) error {
	if volumeID == "" {
		return fmt.Errorf("volumeID is required")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	secretName := sm.makeSecretName(volumeID)

	err := sm.secrets.Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete state secret %s: %w", secretName, err)
	}

	klog.V(4).Infof("Deleted mount state secret %s/%s", sm.namespace, secretName)
	return nil
}

// LoadState loads all mount state secrets for this driver.
// Useful for debugging, recovery, and administrative purposes.
func (sm *MountStateManager) LoadState(ctx context.Context) ([]*MountState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	labelSelector := fmt.Sprintf("%s=%s,%s=%s",
		labelAppName, labelValueAppName,
		labelComponent, labelValueComp,
	)

	secretList, err := sm.secrets.List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list state secrets: %w", err)
	}

	states := make([]*MountState, 0, len(secretList.Items))
	for _, secret := range secretList.Items {
		state, err := sm.deserializeSecret(&secret)
		if err != nil {
			klog.Warningf("Failed to deserialize secret %s: %v", secret.Name, err)
			continue
		}
		states = append(states, state)
	}

	klog.V(4).Infof("Loaded %d mount states from namespace %s", len(states), sm.namespace)
	return states, nil
}

// CleanupStaleStates removes mount state secrets older than the specified duration.
// Useful for cleaning up orphaned secrets from failed mounts.
func (sm *MountStateManager) CleanupStaleStates(ctx context.Context, olderThan time.Duration) (int, error) {
	states, err := sm.LoadState(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to load states: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	deleted := 0

	for _, state := range states {
		if state.Timestamp.Before(cutoff) {
			if err := sm.DeleteState(ctx, state.VolumeID, state.TargetPath); err != nil {
				klog.Warningf("Failed to delete stale state for volume %s: %v", state.VolumeID, err)
				continue
			}
			deleted++
			klog.V(4).Infof("Deleted stale state for volume %s (age: %v)", state.VolumeID, time.Since(state.Timestamp))
		}
	}

	return deleted, nil
}

func byteToString(value []byte) string {
	return string(value)
}
