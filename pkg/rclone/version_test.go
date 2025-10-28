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
	"reflect"
	"runtime"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion(DefaultDriverName)

	expected := VersionInfo{
		DriverName:    DefaultDriverName,
		DriverVersion: "N/A",
		RcloneVersion: "N/A",
		GitCommit:     "N/A",
		BuildDate:     "N/A",
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if !reflect.DeepEqual(version, expected) {
		t.Errorf("Unexpected error. \n Expected: %v \n Found: %v", expected, version)
	}
}

func TestGetVersionWithCustomDriverName(t *testing.T) {
	customName := "custom.rclone.csi"
	version := GetVersion(customName)

	if version.DriverName != customName {
		t.Errorf("Expected driver name %s, got %s", customName, version.DriverName)
	}

	if version.GoVersion != runtime.Version() {
		t.Errorf("Expected Go version %s, got %s", runtime.Version(), version.GoVersion)
	}

	if version.Platform != fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
		t.Errorf("Expected platform %s/%s, got %s", runtime.GOOS, runtime.GOARCH, version.Platform)
	}
}

func TestGetVersionYAML(t *testing.T) {
	resp, err := GetVersionYAML("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	versionInfo := GetVersion("")
	marshalled, _ := yaml.Marshal(&versionInfo)
	expected := strings.TrimSpace(string(marshalled))

	if resp != expected {
		t.Fatalf("Unexpected error. \n Expected:%v\nFound:%v", expected, resp)
	}
}

func TestGetVersionYAMLWithDriverName(t *testing.T) {
	driverName := "test.driver.name"
	resp, err := GetVersionYAML(driverName)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the response contains the driver name
	if !strings.Contains(resp, driverName) {
		t.Errorf("Expected YAML to contain driver name %s, got: %s", driverName, resp)
	}

	// Verify it's valid YAML
	var versionInfo VersionInfo
	err = yaml.Unmarshal([]byte(resp), &versionInfo)
	if err != nil {
		t.Errorf("Failed to unmarshal YAML: %v", err)
	}

	if versionInfo.DriverName != driverName {
		t.Errorf("Expected driver name %s in unmarshaled YAML, got %s", driverName, versionInfo.DriverName)
	}
}
