// Copyright 2017 The go-darwin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package apfs

/*
#include <fcntl.h> // for AT_FDCWD
#include <stdio.h>
*/
import "C"
import (
	"fmt"
	"path/filepath"
	"syscall"
)

// RENAME_FALG provides the rename flag.
type RENAME_FALG uint

var (
	// RENAME_SWAP on file systems that support it (see getattrlist(2) VOL_CAP_INT_RENAME_SWAP), it will cause the source and target to be atomically swapped.
	//
	// Source and target need not be of the same type, i.e. it is possible to swap a file with a directory.
	//
	// EINVAL is returned in case of bitwise-inclusive OR with RENAME_EXCL.
	RENAME_SWAP RENAME_FALG = 0x00000002
	// RENAME_EXCL on file systems that support it (see getattrlist(2) VOL_CAP_INT_RENAME_EXCL), it will cause EEXIST to be returned if the destination already exists.
	//
	// EINVAL is returned in case of bitwise-inclusive OR with RENAME_SWAP.
	RENAME_EXCL RENAME_FALG = 0x00000004
)

// RenamexNp system calls are similar to rename() and renameat() counterparts except that they take a flags argument.
//  int
//  renamex_np(const char *from, const char *to, unsigned int flags);
func RenamexNp(src, dst string, flag RENAME_FALG) error {
	if err := C.renamex_np(C.CString(src), C.CString(dst), C.unsigned(flag)); err != 0 {
		return fmt.Errorf("error: C.renamex_np: %v", (syscall.Errno(err)))
	}

	return nil
}

// RenameatxNp system calls are similar to rename() and renameat() counterparts except that they take a flags argument.
//  int
//  renameatx_np(int fromfd, const char *from, int tofd, const char *to, unsigned int flags);
func RenameatxNp(src, dst string, flag RENAME_FALG) error {
	var srcFd C.int
	if !filepath.IsAbs(src) {
		srcFd = C.AT_FDCWD
	}
	var dstFd C.int
	if !filepath.IsAbs(dst) {
		dstFd = C.AT_FDCWD
	}
	if err := C.renameatx_np(srcFd, C.CString(src), dstFd, C.CString(dst), C.unsigned(flag)); err != 0 {
		return fmt.Errorf("error: C.renameatx_np: %v", syscall.Errno(err))
	}

	return nil
}
