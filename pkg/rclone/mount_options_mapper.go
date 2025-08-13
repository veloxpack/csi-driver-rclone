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
	"strconv"
	"strings"
	"time"

	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/vfs/vfscommon"
	"k8s.io/klog/v2"
)

// MountOptionsMapper handles conversion of Kubernetes mount options to rclone options
type MountOptionsMapper struct {
	defaultMountOpts *mountlib.Options
	defaultVFSOpts   *vfscommon.Options
	optionParsers    map[string]optionParser
}

// optionParser defines how to parse a specific key-value option
type optionParser func(value string, mountOpts *mountlib.Options, vfsOpts *vfscommon.Options) error

// NewMountOptionsMapper creates a new mapper with default options
func NewMountOptionsMapper(defaultMountOpts *mountlib.Options, defaultVFSOpts *vfscommon.Options) *MountOptionsMapper {
	mapper := &MountOptionsMapper{
		defaultMountOpts: defaultMountOpts,
		defaultVFSOpts:   defaultVFSOpts,
	}
	mapper.initOptionParsers()
	return mapper
}

// initOptionParsers initializes the option parsing table
func (m *MountOptionsMapper) initOptionParsers() {
	m.optionParsers = map[string]optionParser{
		// Mount boolean options
		"allow-non-empty":     parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.AllowNonEmpty = v }),
		"allow-other":         parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.AllowOther = v }),
		"allow-root":          parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.AllowRoot = v }),
		"async-read":          parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.AsyncRead = v }),
		"daemon":              parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.Daemon = v }),
		"debug-fuse":          parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.DebugFUSE = v }),
		"default-permissions": parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.DefaultPermissions = v }),
		"direct-io":           parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.DirectIO = v }),
		"network-mode":        parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.NetworkMode = v }),
		"noappledouble":       parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.NoAppleDouble = v }),
		"noapplexattr":        parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.NoAppleXattr = v }),
		"write-back-cache":    parseBoolMountOpt(func(mo *mountlib.Options, v bool) { mo.WritebackCache = v }),

		// Mount duration options
		"attr-timeout": parseDurationMountOpt(func(mo *mountlib.Options, d fs.Duration) { mo.AttrTimeout = d }),

		// Mount string options
		"devname": parseStringMountOpt(func(mo *mountlib.Options, s string) { mo.DeviceName = s }),
		"volname": parseStringMountOpt(func(mo *mountlib.Options, s string) { mo.VolumeName = s }),

		// Mount size options
		"max-read-ahead": parseSizeMountOpt(func(mo *mountlib.Options, sz fs.SizeSuffix) { mo.MaxReadAhead = sz }),

		// Mount tristate options
		"mount-case-insensitive": parseTristateMountOpt(func(mo *mountlib.Options, v bool) {
			mo.CaseInsensitive = fs.Tristate{Value: v, Valid: true}
		}),

		// VFS boolean options
		"vfs-block-norm-dupes": parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.BlockNormDupes = v }),
		"vfs-case-insensitive": parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.CaseInsensitive = v }),
		"vfs-fast-fingerprint": parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.FastFingerprint = v }),
		"vfs-links":            parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.Links = v }),
		"vfs-refresh":          parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.Refresh = v }),
		"vfs-used-is-size":     parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.UsedIsSize = v }),
		"read-only":            parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.ReadOnly = v }),
		"ro":                   parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.ReadOnly = v }),
		"no-seek":              parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.NoSeek = v }),
		"no-modtime":           parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.NoModTime = v }),
		"no-checksum":          parseBoolVFSOpt(func(vo *vfscommon.Options, v bool) { vo.NoChecksum = v }),

		// VFS duration options
		"vfs-cache-max-age":       parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.CacheMaxAge = d }),
		"vfs-cache-poll-interval": parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.CachePollInterval = d }),
		"vfs-read-wait":           parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.ReadWait = d }),
		"vfs-write-back":          parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.WriteBack = d }),
		"vfs-write-wait":          parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.WriteWait = d }),
		"poll-interval":           parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.PollInterval = d }),
		"dir-cache-time":          parseDurationVFSOpt(func(vo *vfscommon.Options, d fs.Duration) { vo.DirCacheTime = d }),

		// VFS size options
		"vfs-cache-max-size":        parseSizeVFSOpt(func(vo *vfscommon.Options, sz fs.SizeSuffix) { vo.CacheMaxSize = sz }),
		"vfs-cache-min-free-space":  parseSizeVFSOpt(func(vo *vfscommon.Options, sz fs.SizeSuffix) { vo.CacheMinFreeSpace = sz }),
		"vfs-disk-space-total-size": parseSizeVFSOpt(func(vo *vfscommon.Options, sz fs.SizeSuffix) { vo.DiskSpaceTotalSize = sz }),
		"vfs-read-ahead":            parseSizeVFSOpt(func(vo *vfscommon.Options, sz fs.SizeSuffix) { vo.ReadAhead = sz }),

		// VFS string options
		"vfs-metadata-extension": parseStringVFSOpt(func(vo *vfscommon.Options, s string) { vo.MetadataExtension = s }),

		// VFS mode options
		"umask":      parseModeVFSOpt(func(vo *vfscommon.Options, m vfscommon.FileMode) { vo.Umask = m }),
		"link-perms": parseModeVFSOpt(func(vo *vfscommon.Options, m vfscommon.FileMode) { vo.LinkPerms = m }),
		"dir-perms":  parseModeVFSOpt(func(vo *vfscommon.Options, m vfscommon.FileMode) { vo.DirPerms = m }),
		"file-perms": parseModeVFSOpt(func(vo *vfscommon.Options, m vfscommon.FileMode) { vo.FilePerms = m }),

		// VFS cache mode (special case)
		"vfs-cache-mode": parseVFSCacheMode,
	}
}

// ParseMountOptions converts Kubernetes mount options to rclone mount and VFS options
func (m *MountOptionsMapper) ParseMountOptions(mountOptions []string) (*mountlib.Options, *vfscommon.Options, error) {
	mountOpts := m.copyMountOpts()
	vfsOpts := m.copyVFSOpts()

	for _, opt := range mountOptions {
		if err := m.parseOption(opt, mountOpts, vfsOpts); err != nil {
			klog.Warningf("Failed to parse mount option '%s': %v", opt, err)
			return nil, nil, err
		}
	}

	return mountOpts, vfsOpts, nil
}

// copyMountOpts creates a copy of default mount options
func (m *MountOptionsMapper) copyMountOpts() *mountlib.Options {
	if m.defaultMountOpts != nil {
		opts := *m.defaultMountOpts
		return &opts
	}
	return &mountlib.Options{}
}

// copyVFSOpts creates a copy of default VFS options
func (m *MountOptionsMapper) copyVFSOpts() *vfscommon.Options {
	if m.defaultVFSOpts != nil {
		opts := *m.defaultVFSOpts
		return &opts
	}
	return &vfscommon.Options{}
}

// parseOption parses a single mount option
func (m *MountOptionsMapper) parseOption(opt string, mountOpts *mountlib.Options, vfsOpts *vfscommon.Options) error {
	if strings.Contains(opt, "=") {
		return m.parseKeyValueOption(opt, mountOpts, vfsOpts)
	}
	return m.parseBooleanOption(opt, mountOpts, vfsOpts)
}

// parseKeyValueOption handles key=value mount options
func (m *MountOptionsMapper) parseKeyValueOption(opt string, mountOpts *mountlib.Options, vfsOpts *vfscommon.Options) error {
	parts := strings.SplitN(opt, "=", 2)
	key, value := parts[0], parts[1]

	parser, exists := m.optionParsers[key]
	if !exists {
		klog.V(4).Infof("Unknown mount option: %s", key)
		return nil
	}

	return parser(value, mountOpts, vfsOpts)
}

// parseBooleanOption handles boolean mount options (flags without values)
func (m *MountOptionsMapper) parseBooleanOption(opt string, mountOpts *mountlib.Options, vfsOpts *vfscommon.Options) error {
	parser, exists := m.optionParsers[opt]
	if !exists {
		klog.V(4).Infof("Unknown boolean mount option: %s", opt)
		return nil
	}

	// For boolean flags, pass empty string which will be interpreted as true
	return parser("", mountOpts, vfsOpts)
}

// Parser factory functions for mount options
func parseBoolMountOpt(setter func(*mountlib.Options, bool)) optionParser {
	return func(value string, mountOpts *mountlib.Options, _ *vfscommon.Options) error {
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		setter(mountOpts, b)
		return nil
	}
}

func parseDurationMountOpt(setter func(*mountlib.Options, fs.Duration)) optionParser {
	return func(value string, mountOpts *mountlib.Options, _ *vfscommon.Options) error {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration: %s", value)
		}
		setter(mountOpts, fs.Duration(duration))
		return nil
	}
}

func parseStringMountOpt(setter func(*mountlib.Options, string)) optionParser {
	return func(value string, mountOpts *mountlib.Options, _ *vfscommon.Options) error {
		setter(mountOpts, value)
		return nil
	}
}

func parseSizeMountOpt(setter func(*mountlib.Options, fs.SizeSuffix)) optionParser {
	return func(value string, mountOpts *mountlib.Options, _ *vfscommon.Options) error {
		size, err := parseSize(value)
		if err != nil {
			return fmt.Errorf("invalid size: %s", value)
		}
		setter(mountOpts, size)
		return nil
	}
}

func parseTristateMountOpt(setter func(*mountlib.Options, bool)) optionParser {
	return func(value string, mountOpts *mountlib.Options, _ *vfscommon.Options) error {
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		setter(mountOpts, b)
		return nil
	}
}

// Parser factory functions for VFS options
func parseBoolVFSOpt(setter func(*vfscommon.Options, bool)) optionParser {
	return func(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		setter(vfsOpts, b)
		return nil
	}
}

func parseDurationVFSOpt(setter func(*vfscommon.Options, fs.Duration)) optionParser {
	return func(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration: %s", value)
		}
		setter(vfsOpts, fs.Duration(duration))
		return nil
	}
}

func parseStringVFSOpt(setter func(*vfscommon.Options, string)) optionParser {
	return func(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
		setter(vfsOpts, value)
		return nil
	}
}

func parseSizeVFSOpt(setter func(*vfscommon.Options, fs.SizeSuffix)) optionParser {
	return func(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
		size, err := parseSize(value)
		if err != nil {
			return fmt.Errorf("invalid size: %s", value)
		}
		setter(vfsOpts, size)
		return nil
	}
}

func parseModeVFSOpt(setter func(*vfscommon.Options, vfscommon.FileMode)) optionParser {
	return func(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
		mode, err := strconv.ParseUint(value, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode: %s", value)
		}
		setter(vfsOpts, vfscommon.FileMode(mode))
		return nil
	}
}

// parseVFSCacheMode handles the special case of VFS cache mode parsing
func parseVFSCacheMode(value string, _ *mountlib.Options, vfsOpts *vfscommon.Options) error {
	switch value {
	case "off":
		vfsOpts.CacheMode = vfscommon.CacheModeOff
	case "minimal":
		vfsOpts.CacheMode = vfscommon.CacheModeMinimal
	case "writes":
		vfsOpts.CacheMode = vfscommon.CacheModeWrites
	case "full":
		vfsOpts.CacheMode = vfscommon.CacheModeFull
	default:
		return fmt.Errorf("invalid vfs-cache-mode: %s (valid: off, minimal, writes, full)", value)
	}
	return nil
}

// parseBool parses boolean values with special handling for empty values
func parseBool(value string) (bool, error) {
	if value == "" {
		return true, nil
	}
	return strconv.ParseBool(value)
}

// parseSize parses size strings like "10G", "1M", "512K", etc.
func parseSize(sizeStr string) (fs.SizeSuffix, error) {
	var size fs.SizeSuffix
	err := size.Set(sizeStr)
	return size, err
}
