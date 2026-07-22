// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package streams

import (
	"context"
	"io"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"

	"storj.io/common/encryption"
	"storj.io/common/paths"
	"storj.io/common/ranger"
	"storj.io/common/storj"
	"storj.io/uplink/private/ecclient"
	"storj.io/uplink/private/eestream"
	"storj.io/uplink/private/metaclient"
	"storj.io/uplink/private/stalldetection"
)

// DisableDeleteOnCancel is now a no-op.
func DisableDeleteOnCancel(ctx context.Context) context.Context {
	return ctx
}

var mon = monkit.Package()

// Meta info about a stream.
type Meta struct {
	Modified    time.Time
	Expiration  time.Time
	Size        int64
	Data        []byte
	Version     []byte
	IsVersioned bool
	Retention   *metaclient.Retention
	LegalHold   *bool
}

// Part info about a part.
type Part struct {
	PartNumber uint32
	Size       int64
	Modified   time.Time
	ETag       []byte
}

// Metadata interface returns the latest metadata for an object.
type Metadata interface {
	Metadata() ([]byte, error)
	ETag() ([]byte, error)
}

// Store is a store for streams. It implements typedStore as part of an ongoing migration
// to use typed paths. See the shim for the store that the rest of the world interacts with.
type Store struct {
	*Uploader

	metainfo             *metaclient.Client
	ec                   ecclient.Client
	segmentSize          int64
	encStore             *encryption.Store
	encryptionParameters storj.EncryptionParameters
	inlineThreshold      int
}

// NewStreamStore constructs a stream store.
func NewStreamStore(metainfo *metaclient.Client, ec ecclient.Client, segmentSize int64, encStore *encryption.Store, encryptionParameters storj.EncryptionParameters, inlineThreshold int) (*Store, error) {
	if segmentSize <= 0 {
		return nil, errs.New("segment size must be larger than 0")
	}
	if encryptionParameters.BlockSize <= 0 {
		return nil, errs.New("encryption block size must be larger than 0")
	}
	// Load stall detection env variables
	stallDetectionConfig := stalldetection.ConfigFromEnv()
	// TODO: this is a hack for now. Once the new upload codepath is enabled
	// by default, we can clean this up and stop embedding the uploader in
	// the streams store.
	uploader, err := NewUploader(metainfo, ec, segmentSize, encStore, encryptionParameters, inlineThreshold, stallDetectionConfig)
	if err != nil {
		return nil, err
	}

	return &Store{
		Uploader:             uploader,
		metainfo:             metainfo,
		ec:                   ec,
		segmentSize:          segmentSize,
		encStore:             encStore,
		encryptionParameters: encryptionParameters,
		inlineThreshold:      inlineThreshold,
	}, nil
}

// Close closes the underlying resources passed to the metainfo DB.
func (s *Store) Close() error {
	return s.metainfo.Close()
}

// ErrorDetection is a struct that contains information about whether error detection is enabled
type ErrorDetection struct {
	Enabled bool
	Offset  int64
}

// IsEnabled checks if error detection is enabled for the given offset and length.
func (ed ErrorDetection) IsEnabled(offset, length int64) bool {
	if !ed.Enabled {
		return false
	}
	return ed.Offset >= offset && ed.Offset < offset+length
}

// Get returns a ranger that knows what the overall size is (from l/<key>)
// and then returns the appropriate data from segments s0/<key>, s1/<key>,
// ..., l/<key>.
func (s *Store) Get(ctx context.Context, bucket, unencryptedKey string, info metaclient.DownloadInfo, errorDetection ErrorDetection) (rr ranger.Ranger, err error) {
	defer mon.Task()(&ctx)(&err)

	object := info.Object
	if object.Size == 0 {
		return ranger.ByteRanger(nil), nil
	}

	derivedKey, err := encryption.DeriveContentKey(bucket, paths.NewUnencrypted(unencryptedKey), s.encStore)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	// make copies of these slices so we aren't mutating data that was passed in
	// to Get. even though info was passed by copy, the slices it contains weren't
	// deep copied, so we'll copy them here and only use the copies below.
	downloaded := slices.Clone(info.DownloadedSegments)
	listed := slices.Clone(info.ListSegments.Items)

	// calculate plain offset and plain size for migrated objects.
	for i := 0; i < len(downloaded); i++ {
		seg := &downloaded[i].Info
		seg.PlainOffset, seg.PlainSize = calculatePlain(*seg.Position, seg.PlainOffset, seg.PlainSize, object)
	}
	for i := 0; i < len(listed); i++ {
		seg := &listed[i]
		seg.PlainOffset, seg.PlainSize = calculatePlain(seg.Position, seg.PlainOffset, seg.PlainSize, object)
	}

	// ensure that the items are correctly sorted
	sort.Slice(downloaded, func(i, k int) bool {
		return downloaded[i].Info.PlainOffset < downloaded[k].Info.PlainOffset
	})
	sort.Slice(listed, func(i, k int) bool {
		return listed[i].PlainOffset < listed[k].PlainOffset
	})

	// calculate the offset for the range listed / downloaded
	var offset int64
	switch {
	case len(downloaded) > 0 && len(listed) > 0:
		offset = min(listed[0].PlainOffset, downloaded[0].Info.PlainOffset)
	case len(downloaded) > 0:
		offset = downloaded[0].Info.PlainOffset
	case len(listed) > 0:
		offset = listed[0].PlainOffset
	}

	rangers := make([]ranger.Ranger, 0, len(downloaded)+len(listed)+2)

	if offset > 0 {
		rangers = append(rangers, &invalidRanger{size: offset})
	}

	for len(downloaded) > 0 || len(listed) > 0 {
		switch {
		case len(downloaded) > 0 && downloaded[0].Info.PlainOffset == offset:
			segment := downloaded[0]
			downloaded = downloaded[1:]

			// drop any duplicate segment info in listing
			for len(listed) > 0 && listed[0].PlainOffset == offset {
				if listed[0].Position != *segment.Info.Position {
					return nil, errs.New("segment info for download and list does not match: %v != %v", listed[0].Position, *segment.Info.Position)
				}
				listed = listed[1:]
			}

			errorDetection := errorDetection.IsEnabled(segment.Info.PlainOffset, segment.Info.PlainSize)
			encryptedRanger, err := s.Ranger(ctx, segment, errorDetection)
			if err != nil {
				return nil, errs.Wrap(err)
			}

			contentNonce, err := deriveContentNonce(*segment.Info.Position)
			if err != nil {
				return nil, errs.Wrap(err)
			}

			enc := segment.Info.SegmentEncryption
			decrypted, err := decryptRanger(ctx, encryptedRanger, segment.Info.PlainSize, object.EncryptionParameters, derivedKey, enc.EncryptedKey, &enc.EncryptedKeyNonce, &contentNonce)
			if err != nil {
				return nil, errs.Wrap(err)
			}

			rangers = append(rangers, decrypted)
			offset += segment.Info.PlainSize

		case len(listed) > 0 && listed[0].PlainOffset == offset:
			segment := listed[0]
			listed = listed[1:]

			contentNonce, err := deriveContentNonce(segment.Position)
			if err != nil {
				return nil, errs.Wrap(err)
			}

			errorDetection := errorDetection.IsEnabled(segment.PlainOffset, segment.PlainSize)
			rangers = append(rangers, &lazySegmentRanger{
				metainfo:             s.metainfo,
				streams:              s,
				streamID:             object.ID,
				position:             segment.Position,
				plainSize:            segment.PlainSize,
				derivedKey:           derivedKey,
				startingNonce:        &contentNonce,
				encryptionParameters: object.EncryptionParameters,
				errorDetection:       errorDetection,
			})
			offset += segment.PlainSize

		default:
			return nil, errs.New("missing segment for offset %d", offset)
		}
	}

	if offset < object.Size {
		rangers = append(rangers, &invalidRanger{size: object.Size - offset})
	}
	if offset > object.Size {
		return nil, errs.New("invalid final offset %d; expected %d", offset, object.Size)
	}

	return ranger.ConcatWithOpts(ranger.ConcatOpts{
		Prefetch:                   true,
		ForceReads:                 prefetchForceReads,
		PrefetchWhenBytesRemaining: prefetchBytesRemaining,
	}, rangers...), nil
}

var (
	// EXPERIMENTAL VALUES
	// TODO: once we understand the usefulness of these, we should expose useful
	// values as real options.
	prefetchForceReads, _     = strconv.ParseBool(os.Getenv("STORJ_EXP_UPLINK_DOWNLOAD_PREFETCH_FORCE_READS"))
	prefetchBytesRemaining, _ = strconv.ParseInt(os.Getenv("STORJ_EXP_UPLINK_DOWNLOAD_PREFETCH_BYTES_REMAINING"), 0, 64)
)

func deriveContentNonce(pos metaclient.SegmentPosition) (storj.Nonce, error) {
	// The increment by 1 is to avoid nonce reuse with the metadata encryption,
	// which is encrypted with the zero nonce.
	var n storj.Nonce
	_, err := encryption.Increment(&n, int64(pos.PartNumber)<<32|(int64(pos.Index)+1))
	return n, err
}

// calculatePlain calculates segment plain size, taking into account migrated objects.
func calculatePlain(pos metaclient.SegmentPosition, rawOffset, rawSize int64, object metaclient.Object) (plainOffset, plainSize int64) {
	switch {
	case object.FixedSegmentSize <= 0:
		// this is a multipart object and has correct offset and size.
		return rawOffset, rawSize
	case pos.PartNumber > 0:
		// this case should be impossible, however let's return the initial values.
		return rawOffset, rawSize
	case pos.Index == int32(object.SegmentCount-1):
		// this is a last segment
		return int64(pos.Index) * object.FixedSegmentSize, object.LastSegment.Size
	default:
		// this is a fixed size segment
		return int64(pos.Index) * object.FixedSegmentSize, object.FixedSegmentSize
	}
}

type lazySegmentRanger struct {
	ranger               ranger.Ranger
	metainfo             *metaclient.Client
	streams              *Store
	streamID             storj.StreamID
	position             metaclient.SegmentPosition
	plainSize            int64
	derivedKey           *storj.Key
	startingNonce        *storj.Nonce
	encryptionParameters storj.EncryptionParameters
	errorDetection       bool
}

// Size implements Ranger.Size.
func (lr *lazySegmentRanger) Size() int64 {
	return lr.plainSize
}

var all = int32(math.MaxInt32)

// Range implements Ranger.Range to be lazily connected.
func (lr *lazySegmentRanger) Range(ctx context.Context, offset, length int64) (_ io.ReadCloser, err error) {
	defer mon.Task()(&ctx)(&err)

	if lr.ranger == nil {
		var desiredNodes int32
		if lr.errorDetection {
			desiredNodes = all
		}

		downloadResponse, err := lr.metainfo.DownloadSegmentWithRS(ctx, metaclient.DownloadSegmentParams{
			StreamID: lr.streamID,
			Position: metaclient.SegmentPosition{
				PartNumber: lr.position.PartNumber,
				Index:      lr.position.Index,
			},
			DesiredNodes: desiredNodes,
		})
		if err != nil {
			return nil, err
		}

		rr, err := lr.streams.Ranger(ctx, downloadResponse, lr.errorDetection)
		if err != nil {
			return nil, err
		}

		encryptedKey, keyNonce := downloadResponse.Info.SegmentEncryption.EncryptedKey, downloadResponse.Info.SegmentEncryption.EncryptedKeyNonce
		lr.ranger, err = decryptRanger(ctx, rr, lr.plainSize, lr.encryptionParameters, lr.derivedKey, encryptedKey, &keyNonce, lr.startingNonce)
		if err != nil {
			return nil, err
		}
	}
	return lr.ranger.Range(ctx, offset, length)
}

// decryptRanger returns a decrypted ranger of the given rr ranger.
func decryptRanger(ctx context.Context, rr ranger.Ranger, plainSize int64, encryptionParameters storj.EncryptionParameters, derivedKey *storj.Key, encryptedKey storj.EncryptedPrivateKey, encryptedKeyNonce, startingNonce *storj.Nonce) (decrypted ranger.Ranger, err error) {
	defer mon.Task()(&ctx)(&err)
	contentKey, err := encryption.DecryptKey(encryptedKey, encryptionParameters.CipherSuite, derivedKey, encryptedKeyNonce)
	if err != nil {
		return nil, err
	}

	decrypter, err := encryption.NewDecrypter(encryptionParameters.CipherSuite, contentKey, startingNonce, int(encryptionParameters.BlockSize))
	if err != nil {
		return nil, err
	}

	var rd ranger.Ranger
	if rr.Size()%int64(decrypter.InBlockSize()) != 0 {
		reader, err := rr.Range(ctx, 0, rr.Size())
		if err != nil {
			return nil, err
		}
		defer func() { err = errs.Combine(err, reader.Close()) }()
		cipherData, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		data, err := encryption.Decrypt(cipherData, encryptionParameters.CipherSuite, contentKey, startingNonce)
		if err != nil {
			return nil, err
		}
		return ranger.ByteRanger(data), nil
	}

	rd, err = encryption.Transform(rr, decrypter)
	if err != nil {
		return nil, err
	}
	return encryption.Unpad(rd, int(rd.Size()-plainSize))
}

// Ranger creates a ranger for downloading erasure codes from piece store nodes.
func (s *Store) Ranger(ctx context.Context, response metaclient.DownloadSegmentWithRSResponse, errorDetection bool) (rr ranger.Ranger, err error) {
	info := response.Info
	limits := response.Limits

	defer mon.Task()(&ctx, info, limits, info.RedundancyScheme)(&err)

	// no order limits also means its inline segment
	if len(info.EncryptedInlineData) != 0 || len(limits) == 0 {
		return ranger.ByteRanger(info.EncryptedInlineData), nil
	}

	redundancy, err := eestream.NewRedundancyStrategyFromStorj(info.RedundancyScheme)
	if err != nil {
		return nil, err
	}

	rr, err = s.ec.GetWithOptions(ctx, limits, info.PiecePrivateKey, redundancy, info.EncryptedSize, ecclient.GetOptions{ErrorDetection: errorDetection})
	return rr, err
}

// invalidRanger is used to mark a range as invalid.
type invalidRanger struct {
	size int64
}

func (d *invalidRanger) Size() int64 { return d.size }

func (d *invalidRanger) Range(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	if offset < 0 {
		return nil, errs.New("negative offset")
	}
	if length < 0 {
		return nil, errs.New("negative length")
	}
	// allow reading zero bytes from an invalid range.
	if 0 <= offset && offset <= d.size && length == 0 {
		return emptyReader{}, nil
	}
	return nil, errs.New("invalid range %d:%d (size:%d)", offset, length, d.size)
}

// emptyReader is used to read no data.
type emptyReader struct{}

func (emptyReader) Read(data []byte) (n int, err error) { return 0, io.EOF }

func (emptyReader) Close() error { return nil }
