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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/vfs/vfscommon"
	"github.com/veloxpack/csi-driver-rclone/pkg/rclone"
	"k8s.io/klog/v2"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "", "node id")
	driverName = flag.String("drivername", rclone.DefaultDriverName, "name of the driver")
	uid        = flag.Uint("uid", 1000, "Override the uid field set by the filesystem (not supported on Windows)")
	gid        = flag.Uint("gid", 1000, "Override the gid field set by the filesystem (not supported on Windows)")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	mountOpts := &mountlib.Options{}
	vfsOpts := &vfscommon.Options{}

	// Mount Options
	flag.BoolVar(&mountOpts.AllowNonEmpty, "allow-non-empty", true, "Allow mounting over a non-empty directory (not supported on Windows)")
	flag.BoolVar(&mountOpts.AllowOther, "allow-other", true, "Allow access to other users (not supported on Windows)")
	// TODO: comment to fix /usr/bin/fusermount3: unknown option 'allow_root
	// flag.BoolVar(&mountOpts.AllowRoot, "allow-root", true, "Allow access to root user (not supported on Windows)")
	flag.BoolVar(&mountOpts.AsyncRead, "async-read", true, "Use asynchronous reads (not supported on Windows)")
	flag.DurationVar((*time.Duration)(&mountOpts.AttrTimeout), "attr-timeout", time.Second, "Time for which file/directory attributes are cached")
	flag.BoolVar(&mountOpts.Daemon, "daemon", false, "Run mount in background and exit parent process (not supported on Windows)")
	flag.DurationVar((*time.Duration)(&mountOpts.DaemonTimeout), "daemon-timeout", 0, "Time limit for rclone to respond to kernel (not supported on Windows)")
	flag.DurationVar((*time.Duration)(&mountOpts.DaemonWait), "daemon-wait", time.Minute, "Time to wait for ready mount from daemon (not supported on Windows)")
	flag.BoolVar(&mountOpts.DebugFUSE, "debug-fuse", false, "Debug the FUSE internals - needs -v")
	flag.BoolVar(&mountOpts.DefaultPermissions, "default-permissions", false, "Makes kernel enforce access control based on file mode (not supported on Windows)")
	flag.StringVar(&mountOpts.DeviceName, "devname", "", "Set the device name - default is remote:path")
	flag.BoolVar(&mountOpts.DirectIO, "direct-io", false, "Use Direct IO, disables caching of data")
	flag.Var((*stringArrayValue)(&mountOpts.ExtraFlags), "fuse-flag", "Flags or arguments to be passed direct to libfuse/WinFsp (repeat if required)")
	flag.Var(&mountOpts.MaxReadAhead, "max-read-ahead", "The number of bytes that can be prefetched for sequential reads (not supported on Windows)")
	flag.Var(&mountOpts.CaseInsensitive, "mount-case-insensitive", "Tell the OS the mount is case insensitive (true) or sensitive (false) (default unset)")
	flag.BoolVar(&mountOpts.NetworkMode, "network-mode", false, "Mount as remote network drive, instead of fixed disk drive (supported on Windows only)")
	flag.BoolVar(&mountOpts.NoAppleDouble, "noappledouble", true, "Ignore Apple Double (._) and .DS_Store files (supported on OSX only)")
	flag.BoolVar(&mountOpts.NoAppleXattr, "noapplexattr", false, "Ignore all 'com.apple.*' extended attributes (supported on OSX only)")
	flag.Var((*stringArrayValue)(&mountOpts.ExtraOptions), "option", "Option for libfuse/WinFsp (repeat if required)")
	flag.StringVar(&mountOpts.VolumeName, "volname", "", "Set the volume name (supported on Windows and OSX only)")
	flag.BoolVar(&mountOpts.WritebackCache, "write-back-cache", false, "Makes kernel buffer writes before sending them to rclone (not supported on Windows)")

	// VFS Options
	flag.BoolVar(&vfsOpts.BlockNormDupes, "vfs-block-norm-dupes", false, "Hide duplicate filenames after normalization (may have performance cost)")
	flag.DurationVar((*time.Duration)(&vfsOpts.CacheMaxAge), "vfs-cache-max-age", time.Hour, "Max time since last access of objects in the cache")
	flag.Var(&vfsOpts.CacheMaxSize, "vfs-cache-max-size", "Max total size of objects in the cache")
	flag.Var(&vfsOpts.CacheMinFreeSpace, "vfs-cache-min-free-space", "Target minimum free space on disk containing cache")
	flag.Var(&vfsOpts.CacheMode, "vfs-cache-mode", "Cache mode off|minimal|writes|full")
	flag.DurationVar((*time.Duration)(&vfsOpts.CachePollInterval), "vfs-cache-poll-interval", time.Minute, "Interval to poll the cache for stale objects")
	flag.BoolVar(&vfsOpts.CaseInsensitive, "vfs-case-insensitive", false, "If a file name not found, find a case insensitive match")
	flag.Var(&vfsOpts.DiskSpaceTotalSize, "vfs-disk-space-total-size", "Specify the total space of disk (default off)")
	flag.BoolVar(&vfsOpts.FastFingerprint, "vfs-fast-fingerprint", false, "Use fast (less accurate) fingerprints for change detection")
	flag.BoolVar(&vfsOpts.Links, "vfs-links", false, "Translate symlinks to/from .rclonelink extension files for the VFS")
	flag.StringVar(&vfsOpts.MetadataExtension, "vfs-metadata-extension", "", "Set the extension to read metadata from")
	flag.Var(&vfsOpts.ReadAhead, "vfs-read-ahead", "Extra read ahead over --buffer-size when using cache-mode full")
	flag.DurationVar((*time.Duration)(&vfsOpts.ReadWait), "vfs-read-wait", 20*time.Millisecond, "Time to wait for in-sequence read before seeking")
	flag.BoolVar(&vfsOpts.Refresh, "vfs-refresh", false, "Refreshes the directory cache recursively in the background on start")
	flag.BoolVar(&vfsOpts.UsedIsSize, "vfs-used-is-size", false, "Use rclone size algorithm for Used size")
	flag.DurationVar((*time.Duration)(&vfsOpts.WriteBack), "vfs-write-back", 5*time.Second, "Time to writeback files after last use when using cache")
	flag.DurationVar((*time.Duration)(&vfsOpts.WriteWait), "vfs-write-wait", time.Second, "Time to wait for in-sequence write before giving error")
	flag.Var(&vfsOpts.Umask, "umask", "Override the permission bits set by the filesystem (not supported on Windows)")
	flag.DurationVar((*time.Duration)(&vfsOpts.PollInterval), "poll-interval", time.Minute, "Time to wait between polling for changes")
	flag.BoolVar(&vfsOpts.ReadOnly, "read-only", false, "Only allow read-only access")
	flag.BoolVar(&vfsOpts.NoSeek, "no-seek", false, "Don't allow seeking in files")
	flag.BoolVar(&vfsOpts.NoModTime, "no-modtime", false, "Don't read/write the modification time (can speed things up)")
	flag.BoolVar(&vfsOpts.NoChecksum, "no-checksum", false, "Don't compare checksums on up/download")
	flag.Var(&vfsOpts.LinkPerms, "link-perms", "Link permissions")
	flag.DurationVar((*time.Duration)(&vfsOpts.DirCacheTime), "dir-cache-time", 5*time.Second, "Time to cache directory entries for")
	flag.Var(&vfsOpts.DirPerms, "dir-perms", "Directory permissions")
	flag.Var(&vfsOpts.FilePerms, "file-perms", "File permissions")

	// Default param values (can be overridden by CLI flags)
	defaultParams := map[string]string{
		"cache-info-age":             "48h", // Reduced from 72h for better memory usage
		"cache-chunk-clean-interval": "10m", // More frequent cleanup
		"cache-dir":                  "/tmp/rclone-vfs-cache/",
		"vfs-cache-max-age":          "12h", // Reduced cache retention for better performance
		"dir-cache-time":             "3s",  // Faster directory listing
		"vfs-cache-max-size":         "1G",  // Limit cache size
		"vfs-cache-min-free-space":   "2G",  // Ensure free space
	}

	// Apply defaults *before* parsing CLI flags, so they can be overridden
	for k, v := range defaultParams {
		_ = flag.Set(k, v)
	}

	flag.Parse()

	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	// Assign flag values to vfsOpts
	vfsOpts.UID = uint32(*uid)
	vfsOpts.GID = uint32(*gid)

	// Cache mode
	if vfsOpts.CacheMode.String() == "" {
		vfsOpts.CacheMode = vfscommon.CacheModeWrites
	}

	driverOptions := rclone.DriverOptions{
		NodeID:             *nodeID,
		DriverName:         *driverName,
		Endpoint:           *endpoint,
		RcloneVFSOptions:   vfsOpts,
		RcloneMountOptions: mountOpts,
		RcloneOtherParams:  defaultParams,
	}

	driver := rclone.NewDriver(&driverOptions)
	driver.Run(false)
	os.Exit(0)
}

type stringArrayValue []string

func (s *stringArrayValue) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringArrayValue) Set(val string) error {
	*s = append(*s, val)
	return nil
}
