// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package files

import (
	"fmt"
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/contenthash"
)

type autoContentHashDisabled interface {
	autoContentHashDisabled()
}

type autoContentHashDisabledReader struct {
	io.Reader
}

func (autoContentHashDisabledReader) autoContentHashDisabled() {}

type autoContentHashDisabledReadSeeker struct {
	io.ReadSeeker
}

func (autoContentHashDisabledReadSeeker) autoContentHashDisabled() {}

type autoContentHashDisabledWriterTo struct {
	io.Reader
	io.WriterTo
}

func (autoContentHashDisabledWriterTo) autoContentHashDisabled() {}

type autoContentHashDisabledReadSeekerWriterTo struct {
	io.ReadSeeker
	io.WriterTo
}

func (autoContentHashDisabledReadSeekerWriterTo) autoContentHashDisabled() {}

// WithoutAutoContentHash returns a reader that disables automatic SDK
// content_hash population for upload requests. A manually supplied
// ContentHash value on the upload argument is still sent.
//
// By default, the SDK automatically computes a Dropbox content_hash for file
// upload calls whose argument includes ContentHash, when the content reader
// implements io.ReadSeeker. The hash is sent to the server which rejects the
// request on mismatch, providing end-to-end integrity protection.
//
// Because the hash must be known before the HTTP body is sent, the SDK reads
// the content once to compute the hash, seeks back to the original position,
// then reads again for the upload. For local files this is typically served
// from the OS page cache; for network-backed readers (NFS, S3) this doubles
// I/O. Use this function to disable auto-hashing when the extra read is
// unacceptable.
//
// A reader is considered seekable when it implements io.ReadSeeker.
// Non-seekable readers are never hashed automatically because they cannot be
// rewound for the upload. Callers may still set ContentHash manually on the
// upload argument when they already know the Dropbox content hash.
// The returned reader preserves io.ReadSeeker and io.WriterTo when the original
// reader implements them.
func WithoutAutoContentHash(content io.Reader) io.Reader {
	if content == nil {
		return nil
	}
	readSeeker, hasReadSeeker := content.(io.ReadSeeker)
	writerTo, hasWriterTo := content.(io.WriterTo)
	switch {
	case hasReadSeeker && hasWriterTo:
		return autoContentHashDisabledReadSeekerWriterTo{
			ReadSeeker: readSeeker,
			WriterTo:   writerTo,
		}
	case hasReadSeeker:
		return autoContentHashDisabledReadSeeker{ReadSeeker: readSeeker}
	case hasWriterTo:
		return autoContentHashDisabledWriterTo{
			Reader:   content,
			WriterTo: writerTo,
		}
	}
	return autoContentHashDisabledReader{Reader: content}
}

func addAutoContentHashToUploadArg(arg *UploadArg, content io.Reader) (*UploadArg, io.Reader, error) {
	if arg == nil {
		return arg, content, nil
	}
	hash, content, ok, err := autoContentHash(arg.ContentHash, content)
	if err != nil || !ok {
		return arg, content, err
	}
	argCopy := *arg
	argCopy.ContentHash = hash
	return &argCopy, content, nil
}

func addAutoContentHashToUploadSessionStartArg(arg *UploadSessionStartArg, content io.Reader) (*UploadSessionStartArg, io.Reader, error) {
	if arg == nil {
		return arg, content, nil
	}
	hash, content, ok, err := autoContentHash(arg.ContentHash, content)
	if err != nil || !ok {
		return arg, content, err
	}
	argCopy := *arg
	argCopy.ContentHash = hash
	return &argCopy, content, nil
}

func addAutoContentHashToUploadSessionAppendArg(arg *UploadSessionAppendArg, content io.Reader) (*UploadSessionAppendArg, io.Reader, error) {
	if arg == nil {
		return arg, content, nil
	}
	hash, content, ok, err := autoContentHash(arg.ContentHash, content)
	if err != nil || !ok {
		return arg, content, err
	}
	argCopy := *arg
	argCopy.ContentHash = hash
	return &argCopy, content, nil
}

func addAutoContentHashToUploadSessionAppendBatchArg(arg *UploadSessionAppendBatchArg, content io.Reader) (*UploadSessionAppendBatchArg, io.Reader, error) {
	if arg == nil {
		return arg, content, nil
	}
	hash, content, ok, err := autoContentHash(arg.ContentHash, content)
	if err != nil || !ok {
		return arg, content, err
	}
	argCopy := *arg
	argCopy.ContentHash = hash
	return &argCopy, content, nil
}

func addAutoContentHashToUploadSessionFinishArg(arg *UploadSessionFinishArg, content io.Reader) (*UploadSessionFinishArg, io.Reader, error) {
	if arg == nil {
		return arg, content, nil
	}
	hash, content, ok, err := autoContentHash(arg.ContentHash, content)
	if err != nil || !ok {
		return arg, content, err
	}
	argCopy := *arg
	argCopy.ContentHash = hash
	return &argCopy, content, nil
}

func autoContentHash(existing string, content io.Reader) (string, io.Reader, bool, error) {
	if existing != "" || content == nil {
		return "", content, false, nil
	}
	if _, ok := content.(autoContentHashDisabled); ok {
		return "", content, false, nil
	}
	seeker, ok := content.(io.ReadSeeker)
	if !ok {
		return "", content, false, nil
	}

	start, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", content, false, nil
	}

	hash, err := contenthash.Compute(seeker)
	_, seekErr := seeker.Seek(start, io.SeekStart)
	if err != nil {
		return "", content, false, fmt.Errorf("files: auto content_hash: compute: %w", err)
	}
	if seekErr != nil {
		return "", content, false, fmt.Errorf("files: auto content_hash: restore reader offset: %w", seekErr)
	}
	return hash, content, true, nil
}
