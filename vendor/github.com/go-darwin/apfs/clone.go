// Copyright 2017 The go-darwin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package apfs

/*
#include <fcntl.h>  // for AT_FDCWD
#include <stdint.h> // for uint32_t
#include <sys/attr.h>
#include <sys/clonefile.h>
*/
import "C"
import (
	"fmt"
	"path/filepath"
	"syscall"
)

// CLONEFILE_FLAG provides the clonefile(2) flag.
type CLONEFILE_FLAG uint32

var (
	// CLONE_NOFOLLOW don't follow the src file if it is a symbolic link (applicable only if the source is not a directory).
	//
	// The symbolic link is itself cloned if src names a symbolic link.
	CLONE_NOFOLLOW CLONEFILE_FLAG = 0x0001
)

// CloneFile function causes the named file src to be cloned to the named file dst.
//
// The cloned file dst shares its data blocks with the src file but has its own copy of attributes,
// extended attributes and ACL's which are identical to those of the named file src with the exceptions listed below
//
// 1. ownership information is set as it would be if dst was created by openat(2) or mkdirat(2) or symlinkat(2) if the current
//      user does not have privileges to change ownership.
//
// 2. setuid and setgid bits are turned off in the mode bits for regular files.
//
// Subsequent writes to either the original or cloned file are private to the file being modified (copy-on-write).
//
// The named file dst must not exist for the call to be successful. Since the clonefile() system call might not allocate new storage for
// data blocks, it is possible for a subsequent overwrite of an existing data block to return ENOSPC.
//
// If src names a directory, the directory hierarchy is cloned as if each item was cloned individually.
//
// However, the use of copyfile(3) is more appropriate for copying large directory hierarchies instead of clonefile(2).
//
//  int
//  clonefile(const char * src, const char * dst, int flags);
func CloneFile(src, dst string, flag CLONEFILE_FLAG) error {
	if err := C.clonefile(C.CString(src), C.CString(dst), C.uint32_t(flag)); err != 0 {
		return fmt.Errorf("error: C.clonefile: %v", syscall.Errno(err))
	}

	return nil
}

// CloneFileAt is equivalent to clonefile() except in the case where either src or dst specifies a relative path.
//
// If src is a relative path, the file to be cloned is located relative to the directory associated with the file descriptor
// src_dirfd instead of the current working directory.
//
// If dst is a relative path, the same happens only relative to the directory associated with dst_dirfd.
//
// If clonefileat() is passed the special value AT_FDCWD in either the src_dirfd or dst_dirfd parameters,
// the current working directory is used in the determination of the file for the respective path parameter.
//
//  int
//  clonefileat(int src_dirfd, const char * src, int dst_dirfd, const char * dst, int flags);
func CloneFileAt(src, dst string, flag CLONEFILE_FLAG) error {
	var srcDirFd C.int
	if !filepath.IsAbs(src) {
		srcDirFd = C.AT_FDCWD
	}
	var dstDirFd C.int
	if !filepath.IsAbs(dst) {
		dstDirFd = C.AT_FDCWD
	}
	if err := C.clonefileat(srcDirFd, C.CString(src), dstDirFd, C.CString(dst), C.uint32_t(flag)); err != 0 {
		return fmt.Errorf("error: C.clonefileat: %v", syscall.Errno(err))
	}

	return nil
}

// FcloneFileAt function is similar to clonefileat() except that the source is identified by file descriptor srcfd rather
// than a path (as in clonefile() or clonefileat())
//
// The flags parameter specifies the options that can be passed.
//
//  int
//  fclonefileat(int srcfd, int dst_dirfd, const char * dst, int flags);
func FcloneFileAt(srcFd uintptr, dst string, flag CLONEFILE_FLAG) error {
	var dstDirFd C.int
	if !filepath.IsAbs(dst) {
		dstDirFd = C.AT_FDCWD
	}
	if err := C.fclonefileat(C.int(srcFd), dstDirFd, C.CString(dst), C.uint32_t(flag)); err != 0 {
		return fmt.Errorf("error: C.fclonefileat: %v", syscall.Errno(err))
	}

	return nil
}
