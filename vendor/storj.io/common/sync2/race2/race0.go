// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information

//go:build !race

package race2

import "unsafe"

// Read marks addr as being read for the race detector.
func Read(addr unsafe.Pointer) {}

// Write marks addr as being written to for the race detector.
func Write(addr unsafe.Pointer) {}

// ReadRange marks [addr:addr+len] as being read for the race detector.
func ReadRange(addr unsafe.Pointer, len int) {}

// WriteRange marks [addr:addr+len] as being written to for the race detector.
func WriteRange(addr unsafe.Pointer, len int) {}

// ReadSlice marks data slice as being read for the race detector.
func ReadSlice[T any](data []T) {}

// WriteSlice marks [addr:addr+len] as being written to for the race detector.
func WriteSlice[T any](data []T) {}
