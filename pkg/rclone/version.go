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
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"
)

// These are set during build time via -ldflags
var (
	driverVersion = "N/A"
	rcloneVersion = "N/A"
	gitCommit     = "N/A"
	buildDate     = "N/A"
)

// VersionInfo holds the version information of the driver
type VersionInfo struct {
	DriverName    string `json:"Driver Name"`
	DriverVersion string `json:"Driver Version"`
	RcloneVersion string `json:"Rclone Version"`
	GitCommit     string `json:"Git Commit"`
	BuildDate     string `json:"Build Date"`
	GoVersion     string `json:"Go Version"`
	Compiler      string `json:"Compiler"`
	Platform      string `json:"Platform"`
}

// GetVersion returns the version information of the driver
func GetVersion(driverName string) VersionInfo {
	return VersionInfo{
		DriverName:    driverName,
		DriverVersion: driverVersion,
		RcloneVersion: rcloneVersion,
		GitCommit:     gitCommit,
		BuildDate:     buildDate,
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersionYAML returns the version information of the driver
// in YAML format
func GetVersionYAML(driverName string) (string, error) {
	info := GetVersion(driverName)
	marshalled, err := yaml.Marshal(&info)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(marshalled)), nil
}

// GetVersionString returns a formatted version string
func GetVersionString(driverName string) string {
	return fmt.Sprintf("%s version %s (rclone: %s, commit: %s, built: %s)",
		driverName, driverVersion, rcloneVersion, gitCommit, buildDate)
}
