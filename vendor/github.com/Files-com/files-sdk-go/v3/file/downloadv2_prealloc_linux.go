package file

import (
	"errors"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func downloadV2PreallocateFile(file *os.File, size int64) error {
	if err := file.Truncate(size); err != nil {
		return err
	}
	for {
		err := unix.Fallocate(int(file.Fd()), 0, 0, size)
		if err == nil {
			return nil
		}
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if errors.Is(err, syscall.EOPNOTSUPP) || errors.Is(err, syscall.ENOSYS) {
			return nil
		}
		return err
	}
}
