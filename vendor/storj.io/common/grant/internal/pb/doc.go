// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

// Package pb contains protobuf definitions for Storj peers.
// Run protolock in the storj.io/common module root to update the lock file.
// E.g. `protolock commit`, no other parameters necessary.
package pb

//go:generate protoc --lint_out=. --pico_out=paths=source_relative:. -I=../../../pb ../../../pb/encryption.proto ../../../pb/encryption_access.proto ../../../pb/scope.proto
