// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package metaclient

import (
	"time"

	"github.com/zeebo/errs"

	"storj.io/common/pb"
	"storj.io/common/storj"
)

// RawObjectItem represents raw object item from get request.
type RawObjectItem struct {
	Bucket             string
	EncryptedObjectKey []byte
	Version            []byte
	StreamID           storj.StreamID
	Status             int32

	Created time.Time
	Expires time.Time

	PlainSize int64

	EncryptedMetadataNonce        storj.Nonce
	EncryptedMetadataEncryptedKey []byte
	EncryptedMetadata             []byte

	EncryptionParameters storj.EncryptionParameters
	RedundancyScheme     storj.RedundancyScheme

	Retention *Retention
}

// Retention represents an object's Object Lock retention information.
type Retention struct {
	Mode        storj.RetentionMode
	RetainUntil time.Time
}

// IsDeleteMarker returns true if object is a delete marker.
func (r RawObjectItem) IsDeleteMarker() bool {
	return r.Status == int32(pb.Object_DELETE_MARKER_UNVERSIONED) || r.Status == int32(pb.Object_DELETE_MARKER_VERSIONED)
}

// IsPrefix returns true if object is a prefix.
func (r RawObjectItem) IsPrefix() bool {
	return r.Status == int32(pb.Object_PREFIX)
}

// RawObjectListItem represents raw object item from list objects request.
type RawObjectListItem struct {
	Bucket             string
	EncryptedObjectKey []byte
	Version            []byte
	StreamID           storj.StreamID
	Status             int32

	CreatedAt time.Time
	ExpiresAt time.Time

	PlainSize int64

	EncryptedMetadataNonce        storj.Nonce
	EncryptedMetadataEncryptedKey []byte
	EncryptedMetadata             []byte

	IsPrefix bool
}

// IsDeleteMarker returns true if listed object item is a delete marker.
func (r RawObjectListItem) IsDeleteMarker() bool {
	return r.Status == int32(pb.Object_DELETE_MARKER_UNVERSIONED) || r.Status == int32(pb.Object_DELETE_MARKER_VERSIONED)
}

// SegmentPosition the segment position within its parent object.
// It is an identifier for the segment.
type SegmentPosition struct {
	// PartNumber indicates the ordinal of the part within an object.
	// A part contains one or more segments.
	// PartNumber is defined by the user.
	// This is only relevant for multipart objects.
	// A non-multipart object only has one Part, and its number is 0.
	PartNumber int32
	// Index indicates the ordinal of this segment within a part.
	// Index is managed by uplink.
	// It is zero-indexed within each part.
	Index int32
}

// SegmentDownloadResponseInfo represents segment download information inline/remote.
type SegmentDownloadResponseInfo struct {
	SegmentID           storj.SegmentID
	EncryptedSize       int64
	EncryptedInlineData []byte
	Next                SegmentPosition
	Position            SegmentPosition
	PiecePrivateKey     storj.PiecePrivateKey

	SegmentEncryption SegmentEncryption
}

// SegmentEncryption represents segment encryption key and nonce.
type SegmentEncryption struct {
	EncryptedKeyNonce storj.Nonce
	EncryptedKey      storj.EncryptedPrivateKey
}

var (
	// ErrNoPath is an error class for using empty path.
	ErrNoPath = errs.Class("no path specified")

	// ErrObjectNotFound is an error class for non-existing object.
	ErrObjectNotFound = errs.Class("object not found")

	// ErrRetentionNotFound is an error class for non-existing object retention.
	ErrRetentionNotFound = errs.Class("object retention not found")

	// ErrUploadIDInvalid is an error class for invalid upload ID.
	ErrUploadIDInvalid = errs.Class("upload ID invalid")
)

// Object contains information about a specific object.
type Object struct {
	Version        []byte
	Bucket         Bucket
	Path           string
	IsPrefix       bool
	IsDeleteMarker bool

	Metadata map[string]string

	ContentType string
	Created     time.Time
	Modified    time.Time
	Expires     time.Time

	Retention *Retention

	Stream
}

// Stream is information about an object stream.
type Stream struct {
	ID storj.StreamID

	// Size is the total size of the stream in bytes
	Size int64

	// SegmentCount is the number of segments
	SegmentCount int64
	// FixedSegmentSize is the size of each segment,
	// when all segments have the same size. It is -1 otherwise.
	FixedSegmentSize int64

	// RedundancyScheme specifies redundancy strategy used for this stream
	storj.RedundancyScheme
	// EncryptionParameters specifies encryption strategy used for this stream
	storj.EncryptionParameters

	LastSegment LastSegment // TODO: remove
}

// LastSegment contains info about last segment.
type LastSegment struct {
	Size              int64
	EncryptedKeyNonce storj.Nonce
	EncryptedKey      storj.EncryptedPrivateKey
}

var (
	// ErrBucket is an error class for general bucket errors.
	ErrBucket = errs.Class("bucket")

	// ErrNoBucket is an error class for using empty bucket name.
	ErrNoBucket = errs.Class("no bucket specified")

	// ErrBucketNotFound is an error class for non-existing bucket.
	ErrBucketNotFound = errs.Class("bucket not found")
)

// Bucket contains information about a specific bucket.
type Bucket struct {
	Name        string
	Created     time.Time
	Attribution string
}

// ListDirection specifies listing direction.
type ListDirection = pb.ListDirection

const (
	// Forward lists forwards from cursor, including cursor.
	Forward = pb.ListDirection_FORWARD
	// After lists forwards from cursor, without cursor.
	After = pb.ListDirection_AFTER
)

// ListOptions lists objects.
type ListOptions struct {
	Prefix                storj.Path
	Cursor                storj.Path // Cursor is relative to Prefix, full path is Prefix + Cursor
	CursorEnc             []byte
	VersionCursor         []byte
	Delimiter             rune
	Recursive             bool
	Direction             ListDirection
	Limit                 int
	IncludeCustomMetadata bool
	IncludeSystemMetadata bool
	Status                int32
	IncludeAllVersions    bool
}

// NextPage returns options for listing the next page.
func (opts ListOptions) NextPage(list ObjectList) ListOptions {
	if !list.More || len(list.Items) == 0 {
		return ListOptions{}
	}

	return ListOptions{
		Prefix:                opts.Prefix,
		CursorEnc:             list.Cursor,
		VersionCursor:         list.VersionCursor,
		Delimiter:             opts.Delimiter,
		Recursive:             opts.Recursive,
		IncludeAllVersions:    opts.IncludeAllVersions,
		IncludeSystemMetadata: opts.IncludeSystemMetadata,
		IncludeCustomMetadata: opts.IncludeCustomMetadata,
		Direction:             After,
		Limit:                 opts.Limit,
		Status:                opts.Status,
	}
}

// ObjectList is a list of objects.
type ObjectList struct {
	Bucket        string
	Prefix        string
	More          bool
	Cursor        []byte
	VersionCursor []byte

	// Items paths are relative to Prefix
	// To get the full path use list.Prefix + list.Items[0].Path
	Items []Object
}

// BucketList is a list of buckets.
type BucketList struct {
	More  bool
	Items []Bucket
}

// BucketListOptions lists objects.
type BucketListOptions struct {
	Cursor    string
	Direction ListDirection
	Limit     int
}

// NextPage returns options for listing the next page.
func (opts BucketListOptions) NextPage(list BucketList) BucketListOptions {
	if !list.More || len(list.Items) == 0 {
		return BucketListOptions{}
	}

	return BucketListOptions{
		Cursor:    list.Items[len(list.Items)-1].Name,
		Direction: After,
		Limit:     opts.Limit,
	}
}
