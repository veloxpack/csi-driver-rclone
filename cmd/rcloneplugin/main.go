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
	"context"
	"flag"
	"os"

	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configflags"
	"github.com/rclone/rclone/fs/config/flags"
	"github.com/rclone/rclone/vfs/vfsflags"
	"github.com/spf13/pflag"
	"github.com/veloxpack/csi-driver-rclone/pkg/rclone"
	"k8s.io/klog/v2"
)

var (
	endpoint   = pflag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = pflag.String("nodeid", "", "node id")
	driverName = pflag.String("drivername", rclone.DefaultDriverName, "name of the driver")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")

	// Register Rclone CLI args
	// Initialize the missing flag groups that VFS and mount options need
	flags.All.NewGroup("VFS", "Virtual File System options")
	flags.All.NewGroup("Mount", "Mount options")
	rcloneCi := fs.GetConfig(context.Background())

	// Add global flags
	configflags.AddFlags(rcloneCi, pflag.CommandLine)

	// Add VFS flags
	vfsflags.AddFlags(pflag.CommandLine)

	// Add mount flags
	mountlib.AddFlags(pflag.CommandLine)

	// Add backend flags (this adds ALL backend-specific flags)
	cmd.AddBackendFlags()

	// Merge standard flags into pflag, handling conflicts
	// We need to manually add flags and skip those that conflict
	flag.CommandLine.VisitAll(func(goflag *flag.Flag) {
		// Check if this flag already exists in pflag
		if pflag.CommandLine.Lookup(goflag.Name) != nil {
			// Flag already exists, skip it
			return
		}

		// Check if there's a shorthand conflict
		// We need to add the flag but potentially without the shorthand
		pflagFlag := pflag.PFlagFromGoFlag(goflag)

		// Check if the shorthand is already in use
		if pflagFlag.Shorthand != "" && pflag.CommandLine.ShorthandLookup(pflagFlag.Shorthand) != nil {
			// Remove the shorthand to avoid conflict
			pflagFlag.Shorthand = ""
		}

		pflag.CommandLine.AddFlag(pflagFlag)
	})

	// Parse all flags using pflag
	pflag.Parse()

	// Apply the flags to the config
	configflags.SetFlags(rcloneCi)

	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	driverOptions := rclone.DriverOptions{
		NodeID:     *nodeID,
		DriverName: *driverName,
		Endpoint:   *endpoint,
	}

	driver := rclone.NewDriver(&driverOptions)
	driver.Run(false)
	os.Exit(0)
}
