// Copyright 2017 The go-darwin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package apfs

/*
#include <stdlib.h>
#include <copyfile.h>
*/
import "C"
import (
	"fmt"
	"syscall"
	"unsafe"
)

// COPYFILE_FLAG provides the copyfile flag.
type COPYFILE_FLAG int

var (
	// COPYFILE_ACL copy the source file's access control lists.
	COPYFILE_ACL COPYFILE_FLAG = 1 << 0
	// COPYFILE_STAT copy the source file's POSIX information (mode, modification time, etc.).
	COPYFILE_STAT COPYFILE_FLAG = 1 << 1
	// COPYFILE_XATTR copy the source file's extended attributes.
	COPYFILE_XATTR COPYFILE_FLAG = 1 << 2
	// COPYFILE_DATA copy the source file's data.
	COPYFILE_DATA COPYFILE_FLAG = 1 << 3

	// COPYFILE_SECURITY copy the source file's POSIX and ACL information; equivalent to (COPYFILE_STAT|COPYFILE_ACL).
	COPYFILE_SECURITY = COPYFILE_STAT | COPYFILE_ACL
	// COPYFILE_METADATA copy the metadata; equivalent to (COPYFILE_SECURITY|COPYFILE_XATTR).
	COPYFILE_METADATA = COPYFILE_SECURITY | COPYFILE_XATTR
	// COPYFILE_ALL copy the entire file; equivalent to (COPYFILE_METADATA|COPYFILE_DATA).
	COPYFILE_ALL = COPYFILE_METADATA | COPYFILE_DATA

	// COPYFILE_RECURSIVE causes CopyFile to recursively copy a hierarchy.
	//
	// This flag is not used by FcopyFile; see below for more information.
	COPYFILE_RECURSIVE COPYFILE_FLAG = 1 << 15 // Descend into hierarchies
	// COPYFILE_CHECK return a bitmask (corresponding to the flags argument) indicating which contents would be copied
	// no data are actually copied.
	//
	// (E.g., if flags was set to COPYFILE_CHECK|COPYFILE_METADATA, and the from
	// file had extended attributes but no ACLs, the return value would be COPYFILE_XATTR .)
	COPYFILE_CHECK COPYFILE_FLAG = 1 << 16 // return flags for xattr or acls if set
	// COPYFILE_EXCL fail if the to file already exists.  (This is only applicable for the CopyFile function.)
	COPYFILE_EXCL COPYFILE_FLAG = 1 << 17 // fail if destination exists
	// COPYFILE_NOFOLLOW_SRC do not follow the from file, if it is a symbolic link.  (This is only applicable for the CopyFile function.)
	COPYFILE_NOFOLLOW_SRC COPYFILE_FLAG = 1 << 18 // don't follow if source is a symlink
	// COPYFILE_NOFOLLOW_DST do not follow the to file, if it is a symbolic link.  (This is only applicable for the CopyFile function.)
	COPYFILE_NOFOLLOW_DST COPYFILE_FLAG = 1 << 19 // don't follow if dst is a symlink
	// COPYFILE_MOVE unlink (using remove(3)) the from file.  (This is only applicable for the copyfile() function.) No error is returned if remove(3) fails.
	//
	// Note that remove(3) removes a symbolic link itself, not the target of the link.
	COPYFILE_MOVE COPYFILE_FLAG = 1 << 20 // unlink src after copy
	// COPYFILE_UNLINK unlink the to file before starting.  (This is only applicable for the copyfile() function.)
	COPYFILE_UNLINK COPYFILE_FLAG = 1 << 21 // unlink dst before copy
	// COPYFILE_NOFOLLOW this is a convenience macro, equivalent to (COPYFILE_NOFOLLOW_DST|COPYFILE_NOFOLLOW_SRC).
	COPYFILE_NOFOLLOW = COPYFILE_NOFOLLOW_SRC | COPYFILE_NOFOLLOW_DST

	// COPYFILE_PACK serialize the from file. The to file is an AppleDouble-format file.
	COPYFILE_PACK COPYFILE_FLAG = 1 << 22
	// COPYFILE_UNPACK unserialize the from file.
	//
	// The from file is an AppleDouble-format file; the to file will have the extended attributes, ACLs, resource fork, and FinderInfo data from the to file, regardless of the flags argument passed in.
	COPYFILE_UNPACK COPYFILE_FLAG = 1 << 23

	// COPYFILE_CLONE try to clone the file/directory instead.
	//
	// This is a best try flag i.e. if cloning fails, fallback to copying the file.
	//
	// This flag is equivalent to (COPYFILE_EXCL | COPYFILE_ACL | COPYFILE_STAT | COPYFILE_XATTR | COPYFILE_DATA).
	//
	// Note that if cloning is successful, callbacks will not be invoked.
	COPYFILE_CLONE COPYFILE_FLAG = 1 << 24
	// COPYFILE_CLONE_FORCE clone the file/directory instead.  This is a force flag i.e. if cloning fails, an error is returned.
	//
	// This flag is equivalent to (COPYFILE_EXCL | COPYFILE_ACL | COPYFILE_STAT | COPYFILE_XATTR | COPYFILE_DATA).
	//
	// Note that if cloning is successful, callbacks will not be invoked.
	COPYFILE_CLONE_FORCE COPYFILE_FLAG = 1 << 25

	// COPYFILE_RUN_IN_PLACE if the src file has quarantine information, add the QTN_FLAG_DO_NOT_TRANSLOCATE flag to the quarantine information of the dst file.
	//
	// This allows a bundle to run in place instead of being translocated.
	COPYFILE_RUN_IN_PLACE COPYFILE_FLAG = 1 << 26

	COPYFILE_VERBOSE COPYFILE_FLAG = 1 << 30
)

// COPYFILE_STATE_FLAG provides the copyfile state flag.
type COPYFILE_STATE_FLAG uint32

var (
	// COPYFILE_STATE_SRC_FD get or set the file descriptor associated with the source (or destination) file.
	//
	// If this has not been initialized yet, the value will be -2.
	//
	// The dst (for CopyFileStateGet) and src (for CopyFileStateSet) parameters are pointers to int.
	COPYFILE_STATE_SRC_FD COPYFILE_STATE_FLAG = C.COPYFILE_STATE_SRC_FD
	// COPYFILE_STATE_DST_FD get or set the file descriptor associated with the source (or destination) file.
	//
	// If this has not been initialized yet, the value will be -2.
	//
	// The dst (for CopyFileStateGet) and src (for CopyFileStateSet) parameters are pointers to int.
	COPYFILE_STATE_DST_FD COPYFILE_STATE_FLAG = C.COPYFILE_STATE_DST_FD
	// COPYFILE_STATE_SRC_FILENAME get or set the filename associated with the source (or destination) file.
	//
	// If it has not been initialized yet, the value will be NULL.
	//
	// For copyfile_state_set(), the src parameter is a pointer to a C string (i.e., char* ); copyfile_state_set() makes a private copy of this string.
	//
	// For copyfile_state_get() function, the dst parameter is a pointer to a pointer to a C string (i.e., char** ); the returned value is a pointer to the state 's copy, and must not be modified or released.
	COPYFILE_STATE_SRC_FILENAME COPYFILE_STATE_FLAG = C.COPYFILE_STATE_SRC_FILENAME
	// COPYFILE_STATE_DST_FILENAME get or set the filename associated with the source (or destination) file.
	//
	// If it has not been initialized yet, the value will be NULL.
	//
	// For copyfile_state_set(), the src parameter is a pointer to a C string (i.e., char* ); copyfile_state_set() makes a private copy of this string.
	//
	// For copyfile_state_get() function, the dst parameter is a pointer to a pointer to a C string (i.e., char** ); the returned value is a pointer to the state 's copy, and must not be modified or released.
	COPYFILE_STATE_DST_FILENAME COPYFILE_STATE_FLAG = C.COPYFILE_STATE_DST_FILENAME
	// COPYFILE_STATE_STATUS_CB get or set the callback status function (currently only used for recursive copies; see below for details).
	//
	// The src parameter is a pointer to a function of type copyfile_callback_t (see above).
	COPYFILE_STATE_STATUS_CB COPYFILE_STATE_FLAG = C.COPYFILE_STATE_STATUS_CB
	// COPYFILE_STATE_STATUS_CTX get or set the context parameter for the status call-back function (see below for details).
	//
	// The src parameter is a void *.
	COPYFILE_STATE_STATUS_CTX COPYFILE_STATE_FLAG = C.COPYFILE_STATE_STATUS_CTX
	// COPYFILE_STATE_QUARANTINE get or set the quarantine information with the source file.
	//
	// The src parameter is a pointer to an opaque object (type void * ).
	COPYFILE_STATE_QUARANTINE COPYFILE_STATE_FLAG = C.COPYFILE_STATE_QUARANTINE
	// COPYFILE_STATE_COPIED get the number of data bytes copied so far.  (Only valid for copyfile_state_get(); see below for
	// more details about callbacks.)
	//
	// If a COPYFILE_CLONE or COPYFILE_CLONE_FORCE operation successfully cloned the requested objects, then this value will be 0.
	//
	// The dst parameter is a pointer to off_t (type off_t * ).
	COPYFILE_STATE_COPIED COPYFILE_STATE_FLAG = C.COPYFILE_STATE_COPIED
	// COPYFILE_STATE_XATTRNAME get the name of the extended attribute during a callback for COPYFILE_COPY_XATTR (see below for details).
	//
	// This field cannot be set, and may be NULL.
	COPYFILE_STATE_XATTRNAME COPYFILE_STATE_FLAG = C.COPYFILE_STATE_XATTRNAME
	// COPYFILE_STATE_WAS_CLONED true if a COPYFILE_CLONE or COPYFILE_CLONE_FORCE operation successfully cloned the requested objects.
	//
	// The dst parameter is a pointer to bool (type bool * ).
	COPYFILE_STATE_WAS_CLONED COPYFILE_STATE_FLAG = C.COPYFILE_STATE_WAS_CLONED
)

// COPYFILE_RECURSE_CALLBACK provides the copyfile callbacks.
type COPYFILE_RECURSE_CALLBACK uint32

var (
	// COPYFILE_RECURSE_ERROR there was an error in processing an element of the source hierarchy.
	//
	// This happens when fts(3) returns an error or unknown file type.  (Currently, the second argument to the call-back function will always be COPYFILE_ERR in this case.)
	COPYFILE_RECURSE_ERROR COPYFILE_RECURSE_CALLBACK = C.COPYFILE_RECURSE_ERROR // 0
	// COPYFILE_RECURSE_FILE the object being copied is a file (or, rather, something other than a directory).
	COPYFILE_RECURSE_FILE COPYFILE_RECURSE_CALLBACK = C.COPYFILE_RECURSE_FILE // 1
	// COPYFILE_RECURSE_DIR the object being copied is a directory, and is being entered. (That is, none of the filesystem objects contained within the directory have been copied yet.)
	COPYFILE_RECURSE_DIR COPYFILE_RECURSE_CALLBACK = C.COPYFILE_RECURSE_DIR // 2
	// COPYFILE_RECURSE_DIR_CLEANUP the object being copied is a directory, and all of the objects contained have been copied.
	//
	// At this stage, the destination directory being copied will have any extra permissions that were added to allow the copying will be removed.
	COPYFILE_RECURSE_DIR_CLEANUP COPYFILE_RECURSE_CALLBACK = C.COPYFILE_RECURSE_DIR_CLEANUP // 3
	COPYFILE_COPY_DATA           COPYFILE_RECURSE_CALLBACK = C.COPYFILE_COPY_DATA           // 4
	COPYFILE_COPY_XATTR          COPYFILE_RECURSE_CALLBACK = C.COPYFILE_COPY_XATTR          // 5
)

var (
	// COPYFILE_START before copying has begun. The third parameter will be a newly-created copyfile_state_t object with the callback function and context pre-loaded.
	COPYFILE_START COPYFILE_RECURSE_CALLBACK = C.COPYFILE_START // 1
	// COPYFILE_FINISH after copying has successfully finished.
	COPYFILE_FINISH COPYFILE_RECURSE_CALLBACK = C.COPYFILE_FINISH // 2
	// COPYFILE_ERR indicates an error has happened at some stage.
	//
	// If the first argument to the call-back function is COPYFILE_RECURSE_ERROR, then an error occurred while processing the source hierarchy;
	//
	// otherwise, it will indicate what type of object was being copied, and errno will be set to indicate the error.
	COPYFILE_ERR      COPYFILE_RECURSE_CALLBACK = C.COPYFILE_ERR      // 3
	COPYFILE_PROGRESS COPYFILE_RECURSE_CALLBACK = C.COPYFILE_PROGRESS // 4
)

var (
	// COPYFILE_CONTINUE the copy will continue as expected.
	COPYFILE_CONTINUE COPYFILE_RECURSE_CALLBACK = C.COPYFILE_CONTINUE // 0
	// COPYFILE_SKIP this object will be skipped, and the next object will be processed.
	//
	// (Note that, when entering a directory. returning COPYFILE_SKIP from the call-back function will prevent the contents of the directory from being copied.)
	COPYFILE_SKIP COPYFILE_RECURSE_CALLBACK = C.COPYFILE_SKIP // 1
	// COPYFILE_QUIT the entire copy is aborted at this stage.  Any filesystem objects created up to this point will remain.
	//
	// CopyFile will return -1, but errno will be unmodified.
	COPYFILE_QUIT COPYFILE_RECURSE_CALLBACK = C.COPYFILE_QUIT // 2
)

// COPYFILE_STATE provides the copyfile state.
type COPYFILE_STATE C.copyfile_state_t

// CopyFile function can copy the named from file to the named to file.
//
// If the state parameter is the return value from CopyFileStateAlloc, then CopyFile will use the information from the state object.
//
// If it is NULL, then both functions will work normally, but less control will be available to the caller.
//
//  int
//  copyfile(const char *from, const char *to, copyfile_state_t state, copyfile_flags_t flags);
func CopyFile(src, dst string, state COPYFILE_STATE, flag COPYFILE_FLAG) (bool, error) {
	if err := C.copyfile(C.CString(src), C.CString(dst), state, C.copyfile_flags_t(flag)); err != 0 {
		return false, fmt.Errorf("couldn't copy from %s to %s: %v", src, dst, syscall.Errno(err))
	}

	var isCloned int
	if err := CopyFileStateGet(state, COPYFILE_STATE_WAS_CLONED, &isCloned); err != nil {
		return false, fmt.Errorf("couldn't get copyfile_state_get: %v", err)
	}

	return isCloned != 0, nil
}

// FcopyFile function does the same CopyFile, but using the file descriptors of already-opened files.
//
// If the state parameter is the return value from CopyFileStateAlloc, then FcopyFile will use the information from the state object.
//
// If it is NULL, then both functions will work normally, but less control will be available to the caller.
//
//  int
//  fcopyfile(int from, int to, copyfile_state_t state, copyfile_flags_t flags);
func FcopyFile(src, dst uintptr, state COPYFILE_STATE, flag COPYFILE_FLAG) error {
	if err := C.fcopyfile(C.int(src), C.int(dst), state, C.copyfile_flags_t(flag)); err != 0 {
		return fmt.Errorf("couldn't fcopy from %d to %d: %v", src, dst, syscall.Errno(err))
	}

	return nil
}

// CopyFileStateAlloc function initializes a copyfile_state_t object (which is an opaque data type).
//
// This object can be passed to CopyFile and FcopyFile; CopyfileStateGet and CopyFileStateSet can be used to manipulate the state (see below).
//
//  copyfile_state_t
//  copyfile_state_alloc(void);
func CopyFileStateAlloc() COPYFILE_STATE {
	return COPYFILE_STATE(C.copyfile_state_alloc())
}

// CopyFileStateFree function is used to deallocate the object and its contents.
//
//  int
//  copyfile_state_free(copyfile_state_t state);
func CopyFileStateFree(state COPYFILE_STATE) error {
	if err := C.copyfile_state_free(C.copyfile_state_t(state)); err != 0 {
		return syscall.Errno(err)
	}

	return nil
}

// CopyFileStateGet functions can be used to manipulate the COPYFILE_STATE object returned by CopyFileStateAlloc.
//
// The dst parameter's type depends on the flag parameter that is passed in.
//
//  int
//  copyfile_state_get(copyfile_state_t state, uint32_t flag, void * dst);
func CopyFileStateGet(state COPYFILE_STATE, flag COPYFILE_STATE_FLAG, result *int) error {
	if err := C.copyfile_state_get(state, C.uint32_t(flag), unsafe.Pointer(result)); err != 0 {
		return syscall.Errno(err)
	}

	return nil
}

// CopyFileStateSet functions can be used to manipulate the COPYFILE_STATE object returned by CopyFileStateAlloc.
//
// The dst parameter's type depends on the flag parameter that is passed in.
//
//  int
//  copyfile_state_set(copyfile_state_t state, uint32_t flag, const void * src);
func CopyFileStateSet(state COPYFILE_STATE, flag COPYFILE_STATE_FLAG, result *int) error {
	if err := C.copyfile_state_set(state, C.uint32_t(flag), unsafe.Pointer(result)); err != 0 {
		return syscall.Errno(err)
	}

	return nil
}
