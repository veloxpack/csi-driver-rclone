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

	EncryptedUserData

	EncryptionParameters storj.EncryptionParameters
	RedundancyScheme     storj.RedundancyScheme

	LegalHold *bool
	Retention *Retention
}

// EncryptedUserData contains all fields about encrypted user data.
type EncryptedUserData struct {
	EncryptedMetadata             []byte
	EncryptedMetadataEncryptedKey []byte
	EncryptedMetadataNonce        storj.Nonce
	EncryptedETag                 []byte
}

// Retention represents an object's Object Lock retention information.
type Retention struct {
	Mode        storj.RetentionMode
	RetainUntil time.Time
}

// IsVersioned returns true if the item is an Object Versioning-versioned item.
func (r RawObjectItem) IsVersioned() bool {
	return version(r.Status, r.Version) != nil
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
	IsLatest           bool

	CreatedAt time.Time
	ExpiresAt time.Time

	PlainSize int64

	EncryptedUserData

	IsPrefix bool
}

// IsVersioned returns true if the listed item is an Object Versioning-versioned item.
func (r RawObjectListItem) IsVersioned() bool {
	return version(r.Status, r.Version) != nil
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

	// ErrProjectNoLock is an error class for cases when object lock is not enabled for a project.
	ErrProjectNoLock = errs.Class("no project object lock")

	// ErrLockNotEnabled is an error class for cases when object lock is not enabled as a feature.
	ErrLockNotEnabled = errs.Class("object lock not enabled")

	// ErrMethodNotAllowed is returned when method is not allowed against specified entity (e.g. object).
	ErrMethodNotAllowed = errs.Class("method not allowed")

	// ErrObjectProtected is an error class for cases when object is protected by Object Lock settings.
	ErrObjectProtected = errs.Class("object protected")

	// ErrObjectLockInvalidObjectState is an error class for cases where an object is in an invalid state for Object Lock operations.
	ErrObjectLockInvalidObjectState = errs.Class("invalid object state for object lock")

	// ErrUnsupportedDelimiter is an error class for when an unsupported delimiter is provided in a listing request.
	ErrUnsupportedDelimiter = errs.Class("unsupported delimiter")

	// ErrUnimplemented is an error class for unimplemented methods.
	ErrUnimplemented = errs.Class("unimplemented")

	// ErrObjectMetadata is an error class for object metadata-related issues.
	ErrObjectMetadata = errs.Class("object metadata")
)

// Object contains information about a specific object.
type Object struct {
	Version        []byte
	Bucket         Bucket
	Path           string
	IsPrefix       bool
	IsVersioned    bool
	IsDeleteMarker bool
	IsLatest       bool

	Metadata map[string]string
	ETag     []byte

	ContentType string
	Created     time.Time
	Modified    time.Time
	Expires     time.Time

	LegalHold *bool
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

	// ErrBucketTagsNotFound is returned when no tags were found on a bucket.
	ErrBucketTagsNotFound = errs.Class("tags not found on bucket")

	// ErrTooManyBucketTags is returned when a set of bucket tags is too long.
	ErrTooManyBucketTags = errs.Class("too many bucket tags")

	// ErrBucketTagKeyInvalid is returned when the key of a bucket tag is invalid.
	ErrBucketTagKeyInvalid = errs.Class("bucket tag key invalid")

	// ErrBucketTagKeyDuplicate is returned when a set of bucket tags contains multiple tags with the same key.
	ErrBucketTagKeyDuplicate = errs.Class("bucket tag key is duplicated")

	// ErrBucketTagValueInvalid is returned when the value of a bucket tag is invalid.
	ErrBucketTagValueInvalid = errs.Class("bucket tag value invalid")

	// ErrBucketNoLock is an error class for cases when object lock is not enabled for a bucket.
	ErrBucketNoLock = errs.Class("no bucket object lock")

	// ErrBucketInvalidObjectLockConfig is an error class for cases when provided an invalid bucket object lock config.
	ErrBucketInvalidObjectLockConfig = errs.Class("invalid bucket object lock config")

	// ErrBucketInvalidStateObjectLock is an error class for when a bucket's state conflicts with Object Lock settings.
	ErrBucketInvalidStateObjectLock = errs.Class("bucket state incompatible with object lock")

	// ErrInvalidPlacement is an error class for invalid placement.
	ErrInvalidPlacement = errs.Class("invalid placement")
	// ErrConflictingPlacement is an error class for conflicting placement
	// between a project default and a new bucket.
	ErrConflictingPlacement = errs.Class("conflicting placement")

	// ErrDeleteObjectsNoItems is returned when attempting to delete an empty list of objects from a bucket.
	ErrDeleteObjectsNoItems = errs.Class("no objects to delete")
)

// Bucket contains information about a specific bucket.
type Bucket struct {
	Name        string
	Created     time.Time
	Attribution string
}

// BucketTag represents a bucket tag.
type BucketTag struct {
	Key   string
	Value string
}

// DefaultRetention contains information about a bucket's default retention.
type DefaultRetention struct {
	Mode  storj.RetentionMode
	Years int32
	Days  int32
}

// BucketObjectLockConfiguration contains information about a bucket's
// object lock configuration.
type BucketObjectLockConfiguration struct {
	Enabled          bool
	DefaultRetention *DefaultRetention
}

// FilterRule represents a filter rule for bucket notification configurations.
type FilterRule struct {
	Prefix string
	Suffix string
}

// BucketNotificationConfiguration represents a bucket notification configuration.
type BucketNotificationConfiguration struct {
	ID         string
	TopicName  string
	Events     []string
	FilterRule FilterRule
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
	Prefix        storj.Path
	Cursor        storj.Path // Cursor is relative to Prefix, full path is Prefix + Cursor
	CursorEnc     []byte
	VersionCursor []byte
	Delimiter     string
	Recursive     bool
	Direction     ListDirection
	Limit         int

	IncludeCustomMetadata       bool
	IncludeSystemMetadata       bool
	IncludeETag                 bool
	IncludeETagOrCustomMetadata bool

	Status             int32
	IncludeAllVersions bool
}

// NextPage returns options for listing the next page.
func (opts ListOptions) NextPage(list ObjectList) ListOptions {
	if !list.More || len(list.Items) == 0 {
		return ListOptions{}
	}

	return ListOptions{
		Prefix:             opts.Prefix,
		CursorEnc:          list.Cursor,
		VersionCursor:      list.VersionCursor,
		Delimiter:          opts.Delimiter,
		Recursive:          opts.Recursive,
		IncludeAllVersions: opts.IncludeAllVersions,

		IncludeSystemMetadata:       opts.IncludeSystemMetadata,
		IncludeCustomMetadata:       opts.IncludeCustomMetadata,
		IncludeETag:                 opts.IncludeETag,
		IncludeETagOrCustomMetadata: opts.IncludeETagOrCustomMetadata,

		Direction: After,
		Limit:     opts.Limit,
		Status:    opts.Status,
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

// StreamRange contains range specification.
type StreamRange struct {
	Mode   StreamRangeMode
	Start  int64
	Limit  int64
	Suffix int64
}

// StreamRangeMode contains different modes for range.
type StreamRangeMode byte

const (
	// StreamRangeAll selects all.
	StreamRangeAll StreamRangeMode = iota
	// StreamRangeStart selects starting from range.Start.
	StreamRangeStart
	// StreamRangeStartLimit selects starting from range.Start to range.End (inclusive).
	StreamRangeStartLimit
	// StreamRangeSuffix selects last range.Suffix bytes.
	StreamRangeSuffix
)

func (streamRange StreamRange) toProto() *pb.Range {
	switch streamRange.Mode {
	case StreamRangeAll:
	case StreamRangeStart:
		return &pb.Range{
			Range: &pb.Range_Start{
				Start: &pb.RangeStart{
					PlainStart: streamRange.Start,
				},
			},
		}
	case StreamRangeStartLimit:
		return &pb.Range{
			Range: &pb.Range_StartLimit{
				StartLimit: &pb.RangeStartLimit{
					PlainStart: streamRange.Start,
					PlainLimit: streamRange.Limit,
				},
			},
		}
	case StreamRangeSuffix:
		return &pb.Range{
			Range: &pb.Range_Suffix{
				Suffix: &pb.RangeSuffix{
					PlainSuffix: streamRange.Suffix,
				},
			},
		}
	}
	return nil
}

// Normalize converts the range to a StreamRangeStartLimit or StreamRangeAll.
func (streamRange StreamRange) Normalize(plainSize int64) StreamRange {
	switch streamRange.Mode {
	case StreamRangeAll:
		streamRange.Start = 0
		streamRange.Limit = plainSize
	case StreamRangeStart:
		streamRange.Mode = StreamRangeStartLimit
		streamRange.Limit = plainSize
	case StreamRangeStartLimit:
	case StreamRangeSuffix:
		streamRange.Mode = StreamRangeStartLimit
		streamRange.Start = plainSize - streamRange.Suffix
		streamRange.Limit = plainSize
	}

	if streamRange.Start < 0 {
		streamRange.Start = 0
	}
	if streamRange.Limit > plainSize {
		streamRange.Limit = plainSize
	}
	streamRange.Suffix = 0

	return streamRange
}

// NewStreamRange creates a new StreamRange based on the offset and length.
func NewStreamRange(offset, length int64) (streamRange StreamRange, err error) {
	switch {
	case offset < 0:
		if length >= 0 {
			return StreamRange{}, errs.New("suffix requires length to be negative, got %v", length)
		}
		streamRange = StreamRange{
			Mode:   StreamRangeSuffix,
			Suffix: -offset,
		}
	case length < 0:
		streamRange = StreamRange{
			Mode:  StreamRangeStart,
			Start: offset,
		}

	default:
		streamRange = StreamRange{
			Mode:  StreamRangeStartLimit,
			Start: offset,
			Limit: offset + length,
		}
	}
	return streamRange, nil
}
