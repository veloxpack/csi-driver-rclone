// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

// DeleteObjectsStatus represents the success or failure status of an individual DeleteObjects deletion.
type DeleteObjectsStatus int

const (
	// DeleteObjectsStatusInternalError indicates that the deletion was not processed due to an internal error.
	DeleteObjectsStatusInternalError = DeleteObjectsStatus(0)

	// DeleteObjectsStatusUnauthorized indicates that the deletion was not processed due to insufficient privileges.
	DeleteObjectsStatusUnauthorized = DeleteObjectsStatus(1)

	// DeleteObjectsStatusNotFound indicates that the object could not be deleted because it did not exist.
	DeleteObjectsStatusNotFound = DeleteObjectsStatus(2)

	// DeleteObjectsStatusOK indicates that the object was successfully deleted.
	DeleteObjectsStatusOK = DeleteObjectsStatus(3)

	// DeleteObjectsStatusLocked indicates that the object's Object Lock configuration prevented its deletion.
	DeleteObjectsStatusLocked = DeleteObjectsStatus(4)
)
