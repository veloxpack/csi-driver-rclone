// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package uplink

import (
	"context"

	"github.com/zeebo/errs"

	"storj.io/uplink/private/metaclient"
)

// CopyObjectOptions options for CopyObject method.
type CopyObjectOptions struct {
	// may contain additional options in the future
}

// CopyObject atomically copies object to a different bucket or/and key.
func (project *Project) CopyObject(ctx context.Context, oldBucket, oldKey, newBucket, newKey string, options *CopyObjectOptions) (_ *Object, err error) {
	defer mon.Task()(&ctx)(&err)

	db, err := dialMetainfoDB(ctx, project)
	if err != nil {
		return nil, packageError.Wrap(err)
	}
	defer func() { err = errs.Combine(err, db.Close()) }()

	obj, err := db.CopyObject(ctx, oldBucket, oldKey, nil, newBucket, newKey, metaclient.CopyObjectOptions{})
	if err != nil {
		return nil, convertKnownErrors(err, oldBucket, oldKey)
	}

	return convertObject(obj), nil
}
