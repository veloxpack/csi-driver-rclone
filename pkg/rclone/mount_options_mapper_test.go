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
	"testing"
	"time"

	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/vfs/vfscommon"
)

func TestNewMountOptionsMapper(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	if mapper == nil {
		t.Fatal("Expected mapper to be created, got nil")
	}

	if mapper.defaultMountOpts != defaultMountOpts {
		t.Error("Expected defaultMountOpts to be set correctly")
	}

	if mapper.defaultVFSOpts != defaultVFSOpts {
		t.Error("Expected defaultVFSOpts to be set correctly")
	}
}

func TestParseMountOptions_NilDefaults(t *testing.T) {
	// Test with nil defaultMountOpts - should initialize with zero values
	mapper1 := NewMountOptionsMapper(nil, &vfscommon.Options{})

	mountOpts, vfsOpts, err := mapper1.ParseMountOptions([]string{"debug-fuse"})
	if err != nil {
		t.Errorf("Expected no error for nil defaultMountOpts, got: %v", err)
		return
	}
	if mountOpts == nil {
		t.Error("Expected mountOpts to be initialized")
		return
	}
	if vfsOpts == nil {
		t.Error("Expected vfsOpts to be initialized")
		return
	}
	// Should have debug-fuse set to true despite nil defaults
	if !mountOpts.DebugFUSE {
		t.Error("Expected DebugFUSE to be true from mount option")
	}

	// Test with nil defaultVFSOpts - should initialize with zero values
	mapper2 := NewMountOptionsMapper(&mountlib.Options{}, nil)

	mountOpts, vfsOpts, err = mapper2.ParseMountOptions([]string{"debug-fuse"})
	if err != nil {
		t.Errorf("Expected no error for nil defaultVFSOpts, got: %v", err)
		return
	}
	if mountOpts == nil {
		t.Error("Expected mountOpts to be initialized")
		return
	}
	if vfsOpts == nil {
		t.Error("Expected vfsOpts to be initialized")
		return
	}
	// Should have debug-fuse set to true despite nil defaults
	if !mountOpts.DebugFUSE {
		t.Error("Expected DebugFUSE to be true from mount option")
	}

	// Test with both nil - should initialize with zero values
	mapper3 := NewMountOptionsMapper(nil, nil)

	mountOpts, vfsOpts, err = mapper3.ParseMountOptions([]string{"debug-fuse"})
	if err != nil {
		t.Errorf("Expected no error for both nil defaults, got: %v", err)
		return
	}
	if mountOpts == nil {
		t.Error("Expected mountOpts to be initialized")
		return
	}
	if vfsOpts == nil {
		t.Error("Expected vfsOpts to be initialized")
		return
	}
	// Should have debug-fuse set to true despite nil defaults
	if !mountOpts.DebugFUSE {
		t.Error("Expected DebugFUSE to be true from mount option")
	}
}

func TestParseMountOptions_EmptyOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{
		AllowNonEmpty: true,
		DebugFUSE:     false,
	}
	defaultVFSOpts := &vfscommon.Options{
		ReadOnly:  false,
		CacheMode: vfscommon.CacheModeOff,
	}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	mountOpts, vfsOpts, err := mapper.ParseMountOptions([]string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should return copies of defaults
	if mountOpts.AllowNonEmpty != defaultMountOpts.AllowNonEmpty {
		t.Error("Expected AllowNonEmpty to match default")
	}
	if vfsOpts.ReadOnly != defaultVFSOpts.ReadOnly {
		t.Error("Expected ReadOnly to match default")
	}
}

func TestParseMountOptions_BooleanFlags(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	tests := []struct {
		name       string
		options    []string
		checkMount func(*mountlib.Options) bool
		checkVFS   func(*vfscommon.Options) bool
	}{
		{
			name:    "allow-non-empty",
			options: []string{"allow-non-empty"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.AllowNonEmpty == true
			},
		},
		{
			name:    "allow-other",
			options: []string{"allow-other"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.AllowOther == true
			},
		},
		{
			name:    "debug-fuse",
			options: []string{"debug-fuse"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.DebugFUSE == true
			},
		},
		{
			name:    "read-only",
			options: []string{"read-only"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.ReadOnly == true
			},
		},
		{
			name:    "ro",
			options: []string{"ro"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.ReadOnly == true
			},
		},
		{
			name:    "vfs-case-insensitive",
			options: []string{"vfs-case-insensitive"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CaseInsensitive == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountOpts, vfsOpts, err := mapper.ParseMountOptions(tt.options)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.checkMount != nil && !tt.checkMount(mountOpts) {
				t.Errorf("Mount option %s not set correctly", tt.name)
			}

			if tt.checkVFS != nil && !tt.checkVFS(vfsOpts) {
				t.Errorf("VFS option %s not set correctly", tt.name)
			}
		})
	}
}

func TestParseMountOptions_KeyValueOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	tests := []struct {
		name        string
		options     []string
		checkMount  func(*mountlib.Options) bool
		checkVFS    func(*vfscommon.Options) bool
		expectError bool
	}{
		{
			name:    "devname",
			options: []string{"devname=my-device"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.DeviceName == "my-device"
			},
		},
		{
			name:    "volname",
			options: []string{"volname=my-volume"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.VolumeName == "my-volume"
			},
		},
		{
			name:    "vfs-cache-mode writes",
			options: []string{"vfs-cache-mode=writes"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CacheMode == vfscommon.CacheModeWrites
			},
		},
		{
			name:    "vfs-cache-mode full",
			options: []string{"vfs-cache-mode=full"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CacheMode == vfscommon.CacheModeFull
			},
		},
		{
			name:    "vfs-cache-mode minimal",
			options: []string{"vfs-cache-mode=minimal"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CacheMode == vfscommon.CacheModeMinimal
			},
		},
		{
			name:    "vfs-cache-mode off",
			options: []string{"vfs-cache-mode=off"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CacheMode == vfscommon.CacheModeOff
			},
		},
		{
			name:        "invalid vfs-cache-mode",
			options:     []string{"vfs-cache-mode=invalid"},
			expectError: true,
		},
		{
			name:    "read-only true",
			options: []string{"read-only=true"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.ReadOnly == true
			},
		},
		{
			name:    "read-only false",
			options: []string{"read-only=false"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.ReadOnly == false
			},
		},
		{
			name:    "debug-fuse with 1",
			options: []string{"debug-fuse=1"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.DebugFUSE == true
			},
		},
		{
			name:    "allow-non-empty with 0",
			options: []string{"allow-non-empty=0"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.AllowNonEmpty == false
			},
		},
		{
			name:    "vfs-case-insensitive with t",
			options: []string{"vfs-case-insensitive=t"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CaseInsensitive == true
			},
		},
		{
			name:    "vfs-case-insensitive with f",
			options: []string{"vfs-case-insensitive=f"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CaseInsensitive == false
			},
		},
		{
			name:        "invalid boolean value",
			options:     []string{"debug-fuse=maybe"},
			expectError: true,
		},
		{
			name:    "debug-fuse with empty value (should default to true)",
			options: []string{"debug-fuse="},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.DebugFUSE == true
			},
		},
		{
			name:    "allow-non-empty with empty value (should default to true)",
			options: []string{"allow-non-empty="},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.AllowNonEmpty == true
			},
		},
		{
			name:    "vfs-case-insensitive with empty value (should default to true)",
			options: []string{"vfs-case-insensitive="},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CaseInsensitive == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountOpts, vfsOpts, err := mapper.ParseMountOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.checkMount != nil && !tt.checkMount(mountOpts) {
				t.Errorf("Mount option %s not set correctly", tt.name)
			}

			if tt.checkVFS != nil && !tt.checkVFS(vfsOpts) {
				t.Errorf("VFS option %s not set correctly", tt.name)
			}
		})
	}
}

func TestParseMountOptions_DurationOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	tests := []struct {
		name        string
		options     []string
		checkMount  func(*mountlib.Options) bool
		checkVFS    func(*vfscommon.Options) bool
		expectError bool
	}{
		{
			name:    "attr-timeout",
			options: []string{"attr-timeout=30s"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.AttrTimeout == fs.Duration(30*time.Second)
			},
		},
		{
			name:    "vfs-cache-max-age",
			options: []string{"vfs-cache-max-age=2h"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.CacheMaxAge == fs.Duration(2*time.Hour)
			},
		},
		{
			name:    "dir-cache-time",
			options: []string{"dir-cache-time=1m"},
			checkVFS: func(opts *vfscommon.Options) bool {
				return opts.DirCacheTime == fs.Duration(1*time.Minute)
			},
		},
		{
			name:        "invalid duration",
			options:     []string{"attr-timeout=invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountOpts, vfsOpts, err := mapper.ParseMountOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.checkMount != nil && !tt.checkMount(mountOpts) {
				t.Errorf("Mount option %s not set correctly", tt.name)
			}

			if tt.checkVFS != nil && !tt.checkVFS(vfsOpts) {
				t.Errorf("VFS option %s not set correctly", tt.name)
			}
		})
	}
}

func TestParseMountOptions_SizeOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	tests := []struct {
		name        string
		options     []string
		checkMount  func(*mountlib.Options) bool
		checkVFS    func(*vfscommon.Options) bool
		expectError bool
	}{
		{
			name:    "max-read-ahead",
			options: []string{"max-read-ahead=1M"},
			checkMount: func(opts *mountlib.Options) bool {
				expected := fs.SizeSuffix(1024 * 1024) // 1M in bytes
				return opts.MaxReadAhead == expected
			},
		},
		{
			name:    "vfs-cache-max-size",
			options: []string{"vfs-cache-max-size=10G"},
			checkVFS: func(opts *vfscommon.Options) bool {
				expected := fs.SizeSuffix(10 * 1024 * 1024 * 1024) // 10G in bytes
				return opts.CacheMaxSize == expected
			},
		},
		{
			name:    "vfs-read-ahead",
			options: []string{"vfs-read-ahead=512K"},
			checkVFS: func(opts *vfscommon.Options) bool {
				expected := fs.SizeSuffix(512 * 1024) // 512K in bytes
				return opts.ReadAhead == expected
			},
		},
		{
			name:        "invalid size",
			options:     []string{"max-read-ahead=invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountOpts, vfsOpts, err := mapper.ParseMountOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.checkMount != nil && !tt.checkMount(mountOpts) {
				t.Errorf("Mount option %s not set correctly", tt.name)
			}

			if tt.checkVFS != nil && !tt.checkVFS(vfsOpts) {
				t.Errorf("VFS option %s not set correctly", tt.name)
			}
		})
	}
}

func TestParseMountOptions_TristateOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	tests := []struct {
		name        string
		options     []string
		checkMount  func(*mountlib.Options) bool
		expectError bool
	}{
		{
			name:    "mount-case-insensitive true",
			options: []string{"mount-case-insensitive=true"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.CaseInsensitive.Value == true && opts.CaseInsensitive.Valid == true
			},
		},
		{
			name:    "mount-case-insensitive false",
			options: []string{"mount-case-insensitive=false"},
			checkMount: func(opts *mountlib.Options) bool {
				return opts.CaseInsensitive.Value == false && opts.CaseInsensitive.Valid == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountOpts, _, err := mapper.ParseMountOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.checkMount != nil && !tt.checkMount(mountOpts) {
				t.Errorf("Mount option %s not set correctly", tt.name)
			}
		})
	}
}

func TestParseMountOptions_MultipleOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	options := []string{
		"allow-non-empty",
		"debug-fuse",
		"devname=my-device",
		"vfs-cache-mode=writes",
		"vfs-cache-max-size=5G",
		"read-only",
		"dir-cache-time=30s",
	}

	mountOpts, vfsOpts, err := mapper.ParseMountOptions(options)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check mount options
	if !mountOpts.AllowNonEmpty {
		t.Error("Expected AllowNonEmpty to be true")
	}
	if !mountOpts.DebugFUSE {
		t.Error("Expected DebugFUSE to be true")
	}
	if mountOpts.DeviceName != "my-device" {
		t.Errorf("Expected DeviceName to be 'my-device', got '%s'", mountOpts.DeviceName)
	}

	// Check VFS options
	if vfsOpts.CacheMode != vfscommon.CacheModeWrites {
		t.Error("Expected CacheMode to be CacheModeWrites")
	}
	expectedSize := fs.SizeSuffix(5 * 1024 * 1024 * 1024) // 5G
	if vfsOpts.CacheMaxSize != expectedSize {
		t.Errorf("Expected CacheMaxSize to be %d, got %d", expectedSize, vfsOpts.CacheMaxSize)
	}
	if !vfsOpts.ReadOnly {
		t.Error("Expected ReadOnly to be true")
	}
	expectedDuration := fs.Duration(30 * time.Second)
	if vfsOpts.DirCacheTime != expectedDuration {
		t.Errorf("Expected DirCacheTime to be %v, got %v", expectedDuration, vfsOpts.DirCacheTime)
	}
}

func TestParseMountOptions_UnknownOptions(t *testing.T) {
	defaultMountOpts := &mountlib.Options{}
	defaultVFSOpts := &vfscommon.Options{}

	mapper := NewMountOptionsMapper(defaultMountOpts, defaultVFSOpts)

	// Unknown options should not cause errors, just be ignored
	options := []string{
		"unknown-option",
		"another-unknown=value",
		"allow-non-empty", // This should work
	}

	mountOpts, _, err := mapper.ParseMountOptions(options)
	if err != nil {
		t.Fatalf("Expected no error for unknown options, got: %v", err)
	}

	// Should still apply known options
	if !mountOpts.AllowNonEmpty {
		t.Error("Expected AllowNonEmpty to be true despite unknown options")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		hasError bool
	}{
		{"", true, false},        // Empty value defaults to true
		{"true", true, false},    // Standard boolean true
		{"false", false, false},  // Standard boolean false
		{"1", true, false},       // Numeric true
		{"0", false, false},      // Numeric false
		{"t", true, false},       // Single character true
		{"f", false, false},      // Single character false
		{"TRUE", true, false},    // Uppercase true
		{"FALSE", false, false},  // Uppercase false
		{"True", true, false},    // Mixed case true
		{"False", false, false},  // Mixed case false
		{"maybe", false, true},   // Invalid boolean
		{"invalid", false, true}, // Invalid boolean
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseBool(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error for input '%s', got: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v for input '%s', got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected fs.SizeSuffix
		hasError bool
	}{
		{"1K", fs.SizeSuffix(1024), false},
		{"1M", fs.SizeSuffix(1024 * 1024), false},
		{"1G", fs.SizeSuffix(1024 * 1024 * 1024), false},
		{"512K", fs.SizeSuffix(512 * 1024), false},
		{"10G", fs.SizeSuffix(10 * 1024 * 1024 * 1024), false},
		{"invalid", fs.SizeSuffix(0), true},
		{"", fs.SizeSuffix(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error for input '%s', got: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d for input '%s', got %d", tt.expected, tt.input, result)
			}
		})
	}
}
