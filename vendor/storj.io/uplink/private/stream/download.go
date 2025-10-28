// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package stream

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"io"

	"storj.io/common/encryption"
	"storj.io/common/paths"
	"storj.io/common/storj"
	"storj.io/eventkit"
	"storj.io/picobuf"
	"storj.io/uplink/private/eestream"
	"storj.io/uplink/private/metaclient"
	"storj.io/uplink/private/storage/streams"
)

const (
	maxDownloadRetries = 6
)

var (
	evs = eventkit.Package()
)

// Download implements Reader, Seeker and Closer for reading from stream.
type Download struct {
	ctx     context.Context
	info    metaclient.DownloadInfo
	streams *streams.Store
	reader  io.ReadCloser
	offset  int64
	length  int64
	closed  bool

	decryptionRetries int
	quiescenceRetries int
}

// NewDownload creates new stream download.
func NewDownload(ctx context.Context, info metaclient.DownloadInfo, streams *streams.Store) *Download {
	return NewDownloadRange(ctx, info, streams, 0, -1)
}

// NewDownloadRange creates new stream range download with range from start to start+length.
func NewDownloadRange(ctx context.Context, info metaclient.DownloadInfo, streams *streams.Store, start, length int64) *Download {
	size := info.Object.Size
	if start > size {
		start = size
	}
	if length < 0 || length+start > size {
		length = size - start
	}
	return &Download{
		ctx:     ctx,
		info:    info,
		streams: streams,
		offset:  start,
		length:  length,
	}
}

// Read reads up to len(data) bytes into data.
//
// If this is the first call it will read from the beginning of the stream.
// Use Seek to change the current offset for the next Read call.
//
// See io.Reader for more details.
func (download *Download) Read(data []byte) (n int, err error) {
	if download.closed {
		return 0, Error.New("already closed")
	}

	if download.reader == nil {
		err = download.resetReader(false)
		if err != nil {
			return 0, err
		}
	}

	if download.length <= 0 {
		return 0, io.EOF
	}
	if download.length < int64(len(data)) {
		data = data[:download.length]
	}
	n, err = download.reader.Read(data)
	download.length -= int64(n)
	download.offset += int64(n)

	if err == nil && n > 0 {
		download.decryptionRetries = 0

	} else if encryption.ErrDecryptFailed.Has(err) {
		evs.Event("decryption-failure",
			eventkit.Int64("decryption-retries", int64(download.decryptionRetries)),
			eventkit.Int64("quiescence-retries", int64(download.quiescenceRetries)),
			eventkit.Int64("offset", download.offset),
			eventkit.Int64("length", download.length),
			eventkit.Bytes("path-checksum", pathChecksum(download.info.EncPath)),
			eventkit.String("cipher-suite", download.info.Object.CipherSuite.String()),
			eventkit.Bytes("stream-id", maybeSatStreamID(download.info.Object.Stream.ID)),
		)

		if download.decryptionRetries+download.quiescenceRetries < maxDownloadRetries {
			download.decryptionRetries++

			// force us to get new a new collection of limits.
			download.info.DownloadedSegments = nil

			err = download.resetReader(true)
		}
	} else if eestream.ErrInactive.Has(err) {
		evs.Event("quiescence-failure",
			eventkit.Int64("decryption-retries", int64(download.decryptionRetries)),
			eventkit.Int64("quiescence-retries", int64(download.quiescenceRetries)),
			eventkit.Int64("offset", download.offset),
			eventkit.Int64("length", download.length),
			eventkit.Bytes("path-checksum", pathChecksum(download.info.EncPath)),
			eventkit.String("cipher-suite", download.info.Object.CipherSuite.String()),
			eventkit.Bytes("stream-id", maybeSatStreamID(download.info.Object.Stream.ID)),
		)

		if download.decryptionRetries+download.quiescenceRetries < maxDownloadRetries {
			download.quiescenceRetries++

			download.info.DownloadedSegments = nil

			err = download.resetReader(false)
		}
	}

	return n, err
}

// Close closes the stream and releases the underlying resources.
func (download *Download) Close() error {
	if download.closed {
		return Error.New("already closed")
	}

	download.closed = true

	if download.reader == nil {
		return nil
	}

	return download.reader.Close()
}

func (download *Download) resetReader(nextSegmentErrorDetection bool) error {
	if download.reader != nil {
		err := download.reader.Close()
		if err != nil {
			return err
		}
	}

	obj := download.info.Object

	rr, err := download.streams.Get(download.ctx, obj.Bucket.Name, obj.Path, download.info, nextSegmentErrorDetection)
	if err != nil {
		return err
	}

	download.reader, err = rr.Range(download.ctx, download.offset, download.length)
	if err != nil {
		return err
	}

	return nil
}

// pathChecksum matches uplink.pathChecksum.
func pathChecksum(encPath paths.Encrypted) []byte {
	mac := hmac.New(sha1.New, []byte(encPath.Raw()))
	_, err := mac.Write([]byte("event"))
	if err != nil {
		panic(err)
	}
	return mac.Sum(nil)[:16]
}

// maybeSatStreamID returns the satellite-internal stream id for a given
// uplink stream id, without needing access to the internalpb package.
// it relies on the stream id being a protocol buffer with the internal id
// as a bytes field at position 10. if this ever changes, then this will
// just return nil, so callers should expect nil here.
func maybeSatStreamID(streamID storj.StreamID) (rv []byte) {
	const satStreamIDField = 10

	decoder := picobuf.NewDecoder(streamID.Bytes())
	decoder.Loop(func(d *picobuf.Decoder) { d.Bytes(satStreamIDField, &rv) })
	if decoder.Err() != nil {
		rv = nil
	}

	return rv
}
