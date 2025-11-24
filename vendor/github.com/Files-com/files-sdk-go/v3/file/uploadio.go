package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/downloadurl"
	"github.com/Files-com/files-sdk-go/v3/file/manager"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

type Progress func(int64)

type Len interface {
	Len() int
}

type UploadResumable struct {
	files_sdk.FileUploadPart
	Parts
	files_sdk.File
}

type uploadIO struct {
	ByteOffset
	Path     string
	reader   io.Reader
	readerAt io.ReaderAt
	Size     *int64
	Progress
	Manager       lib.ConcurrencyManagerWithSubWorker
	ProvidedMtime *time.Time
	Parts
	files_sdk.FileUploadPart
	MkdirParents    *bool
	passedInContext context.Context

	onComplete    chan *Part
	partsFinished chan struct{}
	bytesWritten  int64
	*Client
	etags                      []files_sdk.EtagsParam
	file                       files_sdk.File
	RewindAllProgressOnFailure bool
	notResumable               *atomic.Bool
	actionAttributes           map[string]any
	startedCallback            func(files_sdk.FileUploadPart)
	renamedCallback            func() (string, string)
}

func (u *uploadIO) Run(ctx context.Context) (UploadResumable, error) {
	u.notResumable = &atomic.Bool{}
	u.notResumable.Store(false)
	if u.Path == "" {
		return u.UploadResumable(), errors.New("UploadWithDestinationPath is required")
	}

	if u.reader == nil && u.readerAt == nil {
		return u.UploadResumable(), errors.New("UploadWithReader or UploadWithReaderAt required")
	}

	u.onComplete = make(chan *Part)
	u.partsFinished = make(chan struct{})
	u.etags = make([]files_sdk.EtagsParam, 0)
	if u.Manager == nil {
		u.Manager = manager.Sync().FilePartsManager
	}
	if u.Progress == nil {
		u.Progress = func(i int64) {}
	}

	var defaultSize int64
	if u.Size != nil {
		defaultSize = *u.Size
	}

	if u.MkdirParents == nil {
		u.MkdirParents = lib.Bool(true)
	}
	var err error
	if time.Now().After(u.FileUploadPart.UploadExpires()) || u.isParallelParts(u.FileUploadPart) {
		if len(u.Parts) > 0 {
			u.LogPath(u.Path, map[string]any{
				"timestamp":      time.Now(),
				"event":          "check resumability",
				"parallel_parts": u.isParallelParts(u.FileUploadPart),
				"expired":        time.Now().After(u.FileUploadPart.UploadExpires()),
				"message":        "parts are invalidated must start over",
			})
		}
		u.Parts = Parts{} // parts are invalidated
		if u.Manager.WaitWithContext(ctx) {
			u.FileUploadPart, err = u.startUpload(
				ctx,
				files_sdk.FileBeginUploadParams{
					Size:         defaultSize,
					Path:         u.Path,
					MkdirParents: u.MkdirParents,
				},
			)
			u.Manager.Done()
		} else {
			err = ctx.Err()
		}
		if err != nil {
			return u.UploadResumable(), err
		}
	}

	partCtx, cancel := context.WithCancelCause(ctx)

	u.FileUploadPart.Path = u.Path
	go u.uploadParts(partCtx)
	return u.waitOnParts(ctx, cancel)
}

func (u *uploadIO) UploadResumable() UploadResumable {
	if u.notResumable.Load() {
		return UploadResumable{File: u.file}
	}
	return UploadResumable{Parts: u.Parts, FileUploadPart: u.FileUploadPart, File: u.file}
}

func (u *uploadIO) rewindSuccessfulParts() {
	if !u.RewindAllProgressOnFailure {
		return
	}
	for _, part := range u.Parts {
		if part.Successful() {
			u.Progress(-part.bytes)
		}
	}
}

func (u *uploadIO) waitOnParts(ctx context.Context, cancelParts context.CancelCauseFunc) (UploadResumable, error) {
	var allErrors error
	for {
		select {
		case <-u.partsFinished:
			close(u.onComplete)
			if allErrors != nil {
				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     allErrors.Error(),
					"event":     "partsFinished",
					"message":   "rewindSuccessfulParts",
				})
				u.rewindSuccessfulParts()
				return u.UploadResumable(), allErrors
			}
			// Rate limit all outgoing connections
			if u.Manager.WaitWithContext(ctx) {
				var err error
				path, ref := u.Path, u.FileUploadPart.Ref
				if u.renamedCallback != nil {
					path, ref = u.renamedCallback()
				}
				u.file, err = u.completeUpload(ctx, u.ProvidedMtime, u.etags, u.bytesWritten, path, ref)
				u.Manager.Done()
				if err != nil {
					u.LogPath(u.Path, map[string]any{
						"timestamp": time.Now(),
						"error":     err.Error(),
						"event":     "complete upload",
						"message":   "rewindSuccessfulParts",
					})
					u.rewindSuccessfulParts()
				}
				return u.UploadResumable(), err
			} else {
				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     ctx.Err(),
					"event":     "complete upload",
					"message":   "rewindSuccessfulParts",
				})
				u.rewindSuccessfulParts()
				return u.UploadResumable(), ctx.Err()
			}
		case part := <-u.onComplete:
			if part.error == nil {
				u.etags = append(u.etags, part.EtagsParam)
				u.bytesWritten += part.bytes
			} else {
				u.Progress(-part.bytes)
				allErrors = errors.Join(allErrors, part.error)

				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     part.Error(),
					"part":      part.PartNumber,
					"event":     "onComplete",
				})

				if strings.Contains(part.Error(), "File Upload Not Found") {
					cancelParts(part.error)

					u.notResumable.Store(true)
				}
			}
		}
	}
}

func (u *uploadIO) Reader() (io.Reader, bool) {
	return u.reader, u.reader != nil
}

func (u *uploadIO) ReaderAt() (io.ReaderAt, bool) {
	if u.readerAt != nil {
		return u.readerAt, true
	}
	readerAt, ok := u.reader.(io.ReaderAt)
	return readerAt, ok
}

func (u *uploadIO) partBuilder(offset OffSet, final bool, number int) *Part {
	proxyReader, err := u.buildReader(offset)
	part := &Part{
		OffSet: offset, number: number, final: final, ProxyReader: proxyReader, error: err,
	}
	if number == 1 {
		part.FileUploadPart = u.FileUploadPart
	} else {
		part.FileUploadPart = files_sdk.FileUploadPart{
			HttpMethod:    u.FileUploadPart.HttpMethod,
			Path:          u.FileUploadPart.Path,
			Ref:           u.FileUploadPart.Ref,
			PartNumber:    int64(number),
			ParallelParts: u.ParallelParts,
		}

		if u.usesSameUrl(u.FileUploadPart) {
			part.FileUploadPart.UploadUri = u.FileUploadPart.UploadUri
			part.Expires = u.FileUploadPart.Expires
			if part.PartNumber > 1 {
				// When using the same URL, after each use, the expiry time is extended by 3 minutes.
				part.Expires = u.FileUploadPart.ExpiresTime().Add(3 * time.Minute).Format(time.RFC3339)
			}
			u.ensurePartNumber(part)
		}
	}

	// Parts are stored so a retry can pick up failed parts. Since io.Reader is a stream is better to just retry the whole file
	if _, readerAtOk := u.ReaderAt(); readerAtOk && u.Size != nil {
		u.Parts = append(u.Parts, part)
	}

	return part
}

func (u *uploadIO) ensurePartNumber(part *Part) {
	uri, err := url.Parse(part.UploadUri)
	if err == nil {
		q := uri.Query()
		q.Del("part_number")
		q.Del("partNumber")
		q.Add("part_number", strconv.FormatInt(part.PartNumber, 10))
		uri.RawQuery = q.Encode()
		part.UploadUri = uri.String()
	}
}

func (u *uploadIO) buildReader(offset OffSet) (ProxyReader, error) {
	readerAt, readerAtOk := u.ReaderAt()

	if u.Size != nil && readerAtOk {
		return &ProxyReaderAt{
			ReaderAt: readerAt,
			off:      offset.off,
			len:      offset.len,
			onRead:   u.Progress,
		}, nil
	}

	if u.Size == nil || *u.FileUploadPart.ParallelParts {
		reader, _ := u.Reader()
		if readerAtOk {
			reader = io.NewSectionReader(readerAt, offset.off, offset.len)
		}

		buf := new(bytes.Buffer)
		n, err := io.CopyN(buf, reader, offset.len)
		if err != nil && err != io.EOF {
			return nil, err
		}

		return &ProxyRead{
			Reader: buf,
			len:    n,
			onRead: u.Progress,
		}, nil
	}

	reader, _ := u.Reader()
	return &ProxyRead{
		Reader: reader,
		len:    offset.len,
		onRead: u.Progress,
	}, nil
}

type PartRunnerReturn int

func (u *uploadIO) manageUpdatePart(ctx context.Context, part *Part, wait lib.ConcurrencyManager) bool {
	if part.error != nil {
		u.onComplete <- part
		return true
	}

	if wait.WaitWithContext(ctx) {
		if *u.FileUploadPart.ParallelParts {
			go func() {
				u.runUploadPart(ctx, part)
				wait.Done()
			}()
		} else {
			u.runUploadPart(ctx, part)
			wait.Done()
		}
		if part.ProxyReader.Len() != int(part.len) {
			return true
		}
	} else {
		return true
	}

	return false
}

func (u *uploadIO) runUploadPart(ctx context.Context, part *Part) {
	runCount := 0
	for {
		runCount++
		part.EtagsParam, part.error = u.uploadPart(ctx, part)
		part.bytes = int64(part.ProxyReader.BytesRead())
		part.Touch()
		if part.error == nil && runCount < 3 {
			break
		} else if lib.S3ErrorIsRequestHasExpired(part.error) || files_sdk.IsExpired(part.error) {
			part.FileUploadPart.Expires = ""
			part.FileUploadPart.UploadUri = ""
			if part.ProxyReader.Rewind() {
				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     part.error,
					"part":      part.PartNumber,
					"run_count": runCount,
					"message":   "clearing upload_uri and fetching new one",
				})
			} else {
				break
			}
		} else {
			break
		}
	}

	u.onComplete <- part
}

func (u *uploadIO) uploadParts(ctx context.Context) {
	wait := u.Manager.NewSubWorker()
	defer func() {
		wait.WaitAllDone()
		u.partsFinished <- struct{}{}
	}()

	for _, part := range u.Parts {
		if part.Successful() {
			u.Progress(part.bytes)
			u.onComplete <- part
		} else {
			part.Clear()
			part.ProxyReader, part.error = u.buildReader(part.OffSet)

			if u.manageUpdatePart(ctx, part, wait) {
				break
			}
		}
	}
	var (
		offset   OffSet
		iterator Iterator
		index    int
	)
	if !u.isParallelParts(u.FileUploadPart) {
		// Server optimization for remotes that don't support parallel parts.
		// There is no maximum part count of 10,000, so we can use the default chunk size of 5MB
		u.ByteOffset.OverrideChunkSize = lib.BasePart
	}
	if len(u.Parts) > 0 {
		lastPart := u.Parts[len(u.Parts)-1]
		iterator = u.ByteOffset.Resume(u.Size, lastPart.off, lastPart.number-1)
		offset, iterator, index = iterator()
	} else {
		iterator = u.ByteOffset.BySize(u.Size)
	}

	for {
		if iterator == nil {
			break
		}

		offset, iterator, index = iterator()

		if u.manageUpdatePart(
			ctx,
			u.partBuilder(
				offset,
				iterator == nil && u.Size != nil,
				index+1,
			),
			wait,
		) {
			break
		}
	}
}

func (u *uploadIO) startUpload(ctx context.Context, beginUpload files_sdk.FileBeginUploadParams) (files_sdk.FileUploadPart, error) {
	uploads, err := u.BeginUpload(beginUpload, files_sdk.WithContext(ctx))
	if err != nil {
		return files_sdk.FileUploadPart{}, err
	}
	u.Progress(0)
	part := uploads[0]
	if u.startedCallback != nil {
		u.startedCallback(part)
	}
	return part, err
}

func (u *uploadIO) completeUpload(ctx context.Context, providedMtime *time.Time, etags []files_sdk.EtagsParam, bytesWritten int64, path string, ref string) (files_sdk.File, error) {
	if providedMtime != nil && providedMtime.IsZero() {
		providedMtime = nil
	}

	return u.Create(files_sdk.FileCreateParams{
		ProvidedMtime:    providedMtime,
		EtagsParam:       etags,
		Action:           "end",
		Path:             path,
		Ref:              ref,
		Size:             bytesWritten,
		MkdirParents:     lib.Bool(true),
		ActionAttributes: u.actionAttributes,
	}, files_sdk.WithContext(ctx))
}

func (u *uploadIO) uploadPart(ctx context.Context, part *Part) (files_sdk.EtagsParam, error) {
	var err error
	// Stub for test fixtures being expired
	if part.PartNumber != 1 || part.UploadUri == "" {
		if time.Now().After(part.ExpiresTime()) {
			if !part.ExpiresTime().IsZero() {
				u.LogPath(u.Path, map[string]any{
					"timestamp":           time.Now(),
					"part":                part.PartNumber,
					"event":               "uploadPart",
					"previous_upload_uri": part.UploadUri != "",
					"expired":             time.Now().After(part.ExpiresTime()),
				})
			}
			params := files_sdk.FileBeginUploadParams{
				Path:         part.Path,
				Ref:          part.Ref,
				Part:         part.PartNumber,
				MkdirParents: lib.Bool(true),
			}

			part.FileUploadPart, err = u.startUpload(ctx, params)
			part.FileUploadPart.PartNumber = int64(part.number) // Ensure it didn't change PartNumber

			if err != nil {
				return files_sdk.EtagsParam{}, err
			}
		}
	}
	headers := http.Header{}
	headers.Add("Content-Length", strconv.FormatInt(int64(part.ProxyReader.Len()), 10))
	res, err := files_sdk.CallRaw(
		&files_sdk.CallParams{
			Method:  part.HttpMethod,
			Config:  u.Config,
			Uri:     part.UploadUri,
			BodyIo:  part.ProxyReader,
			Headers: &headers,
			Context: ctx,
		},
	)
	if err != nil {
		return files_sdk.EtagsParam{}, err
	}
	if err := lib.ResponseErrors(res, files_sdk.APIError(), lib.S3XMLError, lib.NonOkError); err != nil {
		return files_sdk.EtagsParam{}, err
	}
	etag := strings.Trim(res.Header.Get("Etag"), "\"")
	if etag == "" {
		// With remote mounts this has no value, but the code strip the value causing a validation error.
		etag = "null"
	}
	if res != nil {
		err = res.Body.Close()
		if err != nil {
			return files_sdk.EtagsParam{}, err
		}
	}
	return files_sdk.EtagsParam{
		Etag: etag,
		Part: strconv.FormatInt(part.PartNumber, 10),
	}, nil
}

func (u *uploadIO) usesSameUrl(file files_sdk.FileUploadPart) bool {
	downloadURL, err := downloadurl.New(file.UploadUri)
	if err != nil {
		return false
	}
	return downloadURL.Type == downloadurl.Files
}

func (u *uploadIO) isParallelParts(file files_sdk.FileUploadPart) bool {
	return lib.UnWrapBool(file.ParallelParts)
}

func uploadProgress(uploadStatus *UploadStatus) func(bytesCount int64) {
	return func(bytesCount int64) {
		uploadStatus.Job().UpdateStatusWithBytes(status.Uploading, uploadStatus, bytesCount)
	}
}
