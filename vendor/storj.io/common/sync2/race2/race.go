// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information

//go:build race

package race2

import (
	"runtime"
	"unsafe"
)

// Read marks addr as being read for the race detector.
func Read(addr unsafe.Pointer) {
	runtime.RaceRead(addr)
}

// Write marks addr as being written to for the race detector.
func Write(addr unsafe.Pointer) {
	runtime.RaceWrite(addr)
}

// ReadRange marks [addr:addr+len] as being read for the race detector.
func ReadRange(addr unsafe.Pointer, len int) {
	runtime.RaceReadRange(addr, len)
}

// WriteRange marks [addr:addr+len] as being written to for the race detector.
func WriteRange(addr unsafe.Pointer, len int) {
	runtime.RaceWriteRange(addr, len)
}

// ReadSlice marks data slice as being read for the race detector.
func ReadSlice[T any](data []T) {
	if len(data) == 0 {
		return
	}
	runtime.RaceReadRange(unsafe.Pointer(&data[0]), len(data)*int(unsafe.Sizeof(data[0])))
}

// WriteSlice marks [addr:addr+len] as being written to for the race detector.
func WriteSlice[T any](data []T) {
	if len(data) == 0 {
		return
	}
	runtime.RaceWriteRange(unsafe.Pointer(&data[0]), len(data)*int(unsafe.Sizeof(data[0])))
}
