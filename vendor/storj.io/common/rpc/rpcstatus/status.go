// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// Package rpcstatus contains status code definitions for rpc.
package rpcstatus

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeebo/errs"

	"storj.io/drpc/drpcerr"
)

//go:generate stringer -type StatusCode .

// StatusCode is an enumeration of rpc status codes.
type StatusCode uint64

// These constants are all the rpc error codes. It is important that
// their numerical values do not change.
const (
	Unknown            StatusCode = 0
	OK                 StatusCode = 1
	Canceled           StatusCode = 2
	InvalidArgument    StatusCode = 3
	DeadlineExceeded   StatusCode = 4
	NotFound           StatusCode = 5
	AlreadyExists      StatusCode = 6
	PermissionDenied   StatusCode = 7
	ResourceExhausted  StatusCode = 8
	FailedPrecondition StatusCode = 9
	Aborted            StatusCode = 10
	OutOfRange         StatusCode = 11
	Unimplemented      StatusCode = 12
	Internal           StatusCode = 13
	Unavailable        StatusCode = 14
	DataLoss           StatusCode = 15
	Unauthenticated    StatusCode = 16
	MethodNotAllowed   StatusCode = 17

	BucketNameMissing    StatusCode = 10014
	BucketNameInvalid    StatusCode = 10015
	ObjectKeyMissing     StatusCode = 10016
	ObjectKeyTooLong     StatusCode = 10017
	ObjectVersionInvalid StatusCode = 10018

	BucketNotFound StatusCode = 10019

	ObjectLockEndpointsDisabled                      StatusCode = 10000
	ObjectLockDisabledForProject                     StatusCode = 10001
	ObjectLockInvalidBucketState                     StatusCode = 10002
	ObjectLockBucketRetentionConfigurationMissing    StatusCode = 10003
	ObjectLockObjectRetentionConfigurationMissing    StatusCode = 10004
	ObjectLockObjectProtected                        StatusCode = 10005
	ObjectLockInvalidObjectState                     StatusCode = 10006
	ObjectLockInvalidBucketRetentionConfiguration    StatusCode = 10007
	ObjectLockUploadWithTTL                          StatusCode = 10008
	ObjectLockUploadWithTTLAPIKey                    StatusCode = 10009
	ObjectLockUploadWithTTLAndDefaultRetention       StatusCode = 10010
	ObjectLockUploadWithTTLAPIKeyAndDefaultRetention StatusCode = 10011

	PlacementInvalidValue      StatusCode = 10012
	PlacementConflictingValues StatusCode = 10013

	DeleteObjectsNoItems      StatusCode = 10020
	DeleteObjectsTooManyItems StatusCode = 10021

	TooManyTags     StatusCode = 10022
	TagKeyInvalid   StatusCode = 10023
	TagKeyDuplicate StatusCode = 10024
	TagValueInvalid StatusCode = 10025
	TagsNotFound    StatusCode = 10026
)

// Code returns the status code associated with the error.
func Code(err error) StatusCode {
	// special case: if the error is context canceled or deadline exceeded, the code
	// must be those. additionally, grpc returns OK for a nil error, so we will, too.
	switch {
	case err == nil:
		return OK
	case errors.Is(err, context.Canceled):
		return Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return DeadlineExceeded
	default:
		return StatusCode(drpcerr.Code(err))
	}
}

// NamedWrap returns the err error wrapped with a defined monkit class name and status code.
func NamedWrap(name string, code StatusCode, err error) error {
	if err == nil {
		return nil
	}

	ce := &codeErr{
		name: name,
		code: code,
	}

	if !errors.As(err, &ce.errsError) {
		werr := errs.Wrap(err)
		errors.As(werr, &ce.errsError)
	}

	return ce
}

// NamedError creates and returns an error wrapped with a defined monkit class name and status code.
func NamedError(name string, code StatusCode, msg string) error {
	return NamedWrap(name, code, errs.New("%s", msg))
}

// NamedErrorf creates and returns an error wrapped with a defined monkit class name and status code.
func NamedErrorf(name string, code StatusCode, format string, a ...any) error {
	return NamedWrap(name, code, errs.New(format, a...))
}

// Wrap wraps the error with the provided status code.
func Wrap(code StatusCode, err error) error {
	return NamedWrap("", code, err)
}

// Error wraps the message with a status code into an error.
func Error(code StatusCode, msg string) error {
	return NamedError("", code, msg)
}

// Errorf : Error :: fmt.Sprintf : fmt.Sprint.
func Errorf(code StatusCode, format string, a ...any) error {
	return NamedErrorf("", code, format, a...)
}

type errsError interface {
	error
	fmt.Formatter
	Name() (string, bool)
}

// codeErr implements error that can work both in grpc and drpc.
type codeErr struct {
	name string
	errsError
	code StatusCode
}

func (c *codeErr) Unwrap() error { return c.errsError }
func (c *codeErr) Cause() error  { return c.errsError }

func (c *codeErr) Code() uint64         { return uint64(c.code) }
func (c *codeErr) Name() (string, bool) { return c.name, c.name != "" }
