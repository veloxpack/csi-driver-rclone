//go:build windows

package io

import (
	"os"
	"syscall"
	"time"
)

// GetCreationTime returns the creation time of the file.
// For non-Windows platforms, this is the same as the modification time.
func GetCreationTime(fileStat os.FileInfo) time.Time {
	return time.Unix(0, fileStat.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds())
}
