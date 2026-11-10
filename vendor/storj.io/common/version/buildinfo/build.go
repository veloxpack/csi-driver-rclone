// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package buildinfo

import (
	"runtime/debug"
	"sync"

	"github.com/zeebo/errs"
)

// Error is common error for version package.
var Error = errs.Class("version")

// moduleVersions returns a lookup table of module path to version.
//
// debug.ReadBuildInfo reparses and reallocates the whole build info on every
// call, which is wasteful on hot paths, and the result cannot change during the
// lifetime of the process, so compute it once.
var moduleVersions = sync.OnceValue(func() map[string]string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}

	versions := make(map[string]string, len(info.Deps)+1)
	for _, mod := range info.Deps {
		versions[mod.Path] = mod.Version
	}
	// the main module takes precedence over a dependency with the same path.
	versions[info.Main.Path] = info.Main.Version

	return versions
})

// FromBuild returns version string for a module.
//
// This does not work inside tests.
func FromBuild(modname string) (string, error) {
	versions := moduleVersions()
	if versions == nil {
		return "", Error.New("unable to read build info")
	}

	version, ok := versions[modname]
	if !ok {
		return "", Error.New("unable to find module %q", modname)
	}

	return version, nil
}
