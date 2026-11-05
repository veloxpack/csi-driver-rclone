//go:build !linux

package file

import "os"

func downloadV2PreallocateFile(file *os.File, size int64) error {
	return file.Truncate(size)
}
