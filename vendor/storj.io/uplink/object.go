// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package uplink

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	_ "unsafe" // for go:linkname

	"github.com/zeebo/errs"

	"storj.io/uplink/private/metaclient"
)

// ErrObjectKeyInvalid is returned when the object key is invalid.
var ErrObjectKeyInvalid = errors.New("object key invalid")

// ErrObjectNotFound is returned when the object is not found.
var ErrObjectNotFound = errors.New("object not found")

// Object contains information about an object.
type Object struct {
	Key string
	// IsPrefix indicates whether the Key is a prefix for other objects.
	IsPrefix bool

	System SystemMetadata
	Custom CustomMetadata

	version []byte
}

// SystemMetadata contains information about the object that cannot be changed directly.
type SystemMetadata struct {
	Created       time.Time
	Expires       time.Time
	ContentLength int64
}

// CustomMetadata contains custom user metadata about the object.
//
// The keys and values in custom metadata are expected to be valid UTF-8.
//
// When choosing a custom key for your application start it with a prefix "app:key",
// as an example application named "Image Board" might use a key "image-board:title".
type CustomMetadata map[string]string

// Clone makes a deep clone.
func (meta CustomMetadata) Clone() CustomMetadata {
	r := CustomMetadata{}
	for k, v := range meta {
		r[k] = v
	}
	return r
}

// Verify verifies whether CustomMetadata contains only "utf-8".
func (meta CustomMetadata) Verify() error {
	var invalid []string
	for k, v := range meta {
		if !utf8.ValidString(k) || !utf8.ValidString(v) {
			invalid = append(invalid, fmt.Sprintf("not utf-8 %q=%q", k, v))
		}
		if strings.IndexByte(k, 0) >= 0 || strings.IndexByte(v, 0) >= 0 {
			invalid = append(invalid, fmt.Sprintf("contains 0 byte: %q=%q", k, v))
		}
		if k == "" {
			invalid = append(invalid, "empty key")
		}
	}

	if len(invalid) > 0 {
		return errs.New("invalid pairs %v", invalid)
	}

	return nil
}

// StatObject returns information about an object at the specific key.
func (project *Project) StatObject(ctx context.Context, bucket, key string) (info *Object, err error) {
	defer mon.Task()(&ctx)(&err)

	db, err := project.dialMetainfoDB(ctx)
	if err != nil {
		return nil, convertKnownErrors(err, bucket, key)
	}
	defer func() { err = errs.Combine(err, db.Close()) }()

	obj, err := db.GetObject(ctx, bucket, key, nil)
	if err != nil {
		return nil, convertKnownErrors(err, bucket, key)
	}

	return convertObject(&obj), nil
}

// DeleteObject deletes the object at the specific key.
// Returned deleted is not nil when the access grant has read permissions and
// the object was deleted.
func (project *Project) DeleteObject(ctx context.Context, bucket, key string) (deleted *Object, err error) {
	defer mon.Task()(&ctx)(&err)

	db, err := project.dialMetainfoDB(ctx)
	if err != nil {
		return nil, convertKnownErrors(err, bucket, key)
	}
	defer func() { err = errs.Combine(err, db.Close()) }()

	obj, err := db.DeleteObject(ctx, bucket, key, nil)
	if err != nil {
		return nil, convertKnownErrors(err, bucket, key)
	}
	return convertObject(&obj), nil
}

// UploadObjectMetadataOptions contains additional options for updating object's metadata.
// Reserved for future use.
type UploadObjectMetadataOptions struct {
}

// UpdateObjectMetadata replaces the custom metadata for the object at the specific key with newMetadata.
// Any existing custom metadata will be deleted.
func (project *Project) UpdateObjectMetadata(ctx context.Context, bucket, key string, newMetadata CustomMetadata, options *UploadObjectMetadataOptions) (err error) {
	defer mon.Task()(&ctx)(&err)

	db, err := project.dialMetainfoDB(ctx)
	if err != nil {
		return convertKnownErrors(err, bucket, key)
	}
	defer func() { err = errs.Combine(err, db.Close()) }()

	err = db.UpdateObjectMetadata(ctx, bucket, key, newMetadata.Clone())
	if err != nil {
		return convertKnownErrors(err, bucket, key)
	}

	return nil
}

// convertObject converts metaclient.Object to uplink.Object.
func convertObject(obj *metaclient.Object) *Object {
	if obj == nil || obj.Bucket.Name == "" { // nil or zero object
		return nil
	}

	object := &Object{
		Key:      obj.Path,
		IsPrefix: obj.IsPrefix,
		System: SystemMetadata{
			Created:       obj.Created,
			Expires:       obj.Expires,
			ContentLength: obj.Size,
		},
		Custom: obj.Metadata,

		version: obj.Version,
	}

	if object.Custom == nil {
		object.Custom = CustomMetadata{}
	}

	return object
}

// objectVersion is exposing object version field.
//
// NB: this is used with linkname in private/object.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:deadcode,unused
//go:linkname objectVersion
func objectVersion(object *Object) []byte {
	return object.version
}
