// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package internal

import "errors"

// RootError unwraps all layers of an error to get at the core error; the
// one not wrapping any other error. If it encounters a grouped error while
// unwrapping, it will treat the first error in the group as the most
// important; the one which will be unwrapped further in search of the core
// error.
//
// This method exists because errs.Unwrap was deprecated in the version 1.4 and
// errors.Unwrap doesn't produce the same result.
func RootError(err error) error {
	for {
		if multiUnwrappingErr, ok := err.(interface {
			Unwrap() []error
		}); ok {
			unwrappedGroup := multiUnwrappingErr.Unwrap()
			if len(unwrappedGroup) == 0 {
				return err
			}
			err = unwrappedGroup[0]
			continue
		}
		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			return err
		}
		err = unwrappedErr
	}
}
