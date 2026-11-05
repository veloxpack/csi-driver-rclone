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
	"github.com/hashicorp/go-retryablehttp"
)

const maxUploadPartAttempts = 3

type Progress func(int64)

type Len interface {
	Len() int
}

type UploadResumable struct {
	files_sdk.FileUploadPart
	Parts
	files_sdk.File
}

// JobUploadCheckpoint holds folder-level resume state for a paused upload job.
type JobUploadCheckpoint struct {
	CompletedPaths []string
	PendingParts   map[string]UploadResumable // local path → partial upload
}

// UploadCheckpoint builds a JobUploadCheckpoint from the job's settled file statuses.
// Call at terminal time (Canceled or Finished) instead of tracking state incrementally.
func (j *Job) UploadCheckpoint() *JobUploadCheckpoint {
	completed := make(map[string]struct{})
	for p := range j.CompletedPaths {
		completed[p] = struct{}{}
	}
	pendingParts := make(map[string]UploadResumable)

	j.statusesMutex.RLock()
	for _, f := range j.Statuses {
		if f.Status().Is(status.Complete) {
			completed[f.LocalPath()] = struct{}{}
		} else if cr, ok := f.(checkpointResumableProvider); ok {
			if r := cr.CheckpointResumable(); r.FileUploadPart.Ref != "" {
				pendingParts[f.LocalPath()] = r
			}
		}
	}
	j.statusesMutex.RUnlock()

	paths := make([]string, 0, len(completed))
	for p := range completed {
		paths = append(paths, p)
	}
	return &JobUploadCheckpoint{
		CompletedPaths: paths,
		PendingParts:   pendingParts,
	}
}

type uploadIO struct {
	ByteOffset
	Path     string
	reader   io.Reader
	readerAt io.ReaderAt
	Size     *int64
	Progress
	Manager       lib.ConcurrencyManagerWithSubWorker
	managerSet    bool
	ProvidedMtime *time.Time
	Parts
	files_sdk.FileUploadPart
	MkdirParents               *bool
	passedInContext            context.Context
	uploadV2                   bool
	uploadV2UseSDKDefaultCaps  bool
	uploadV2TargetClassifier   UploadV2TargetClassifier
	uploadV2ManagerProvider    uploadV2AdaptiveManagerProvider
	uploadV2HTTPClientProvider uploadV2HTTPClientProvider
	uploadV2ReadyRunway        uploadV2ReadyRunwayConfig
	uploadV2Tuning             UploadV2Tuning
	uploadV1Stats              uploadV1SchedulerStats
	directTransferDisabled     atomic.Bool

	onComplete      chan *Part
	partsFinished   chan struct{}
	transferStarted *atomic.Bool
	bytesWritten    int64
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
	ctx, closeDirectClients := files_sdk.WithDirectTransferClientCache(ctx)
	defer closeDirectClients()

	u.notResumable = &atomic.Bool{}
	u.notResumable.Store(false)
	u.transferStarted = &atomic.Bool{}
	u.uploadV1Stats.enable(u.Client != nil && u.Config.InDebug())

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
	if u.FileUploadPart.Ref != "" {
		// Propagate session info to non-successful restored parts so uploadPart()
		// can request fresh upload URLs using the existing Ref (without starting a new session).
		for _, p := range u.Parts {
			if !p.Successful() {
				p.FileUploadPart = files_sdk.FileUploadPart{
					Ref:           u.FileUploadPart.Ref,
					Path:          u.Path,
					HttpMethod:    u.FileUploadPart.HttpMethod,
					ParallelParts: u.FileUploadPart.ParallelParts,
					PartNumber:    int64(p.number),
				}
			}
		}
	} else if time.Now().After(u.FileUploadPart.UploadExpires()) || u.isParallelParts(u.FileUploadPart) {
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

	u.FileUploadPart.Path = u.Path
	if u.uploadV2Enabled() {
		if resumable, err, handled := u.runUploadV2(ctx); handled {
			return resumable, err
		}
	}

	partCtx, cancel := context.WithCancelCause(ctx)
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
				u.logUploadV1SchedulerSummary(allErrors)
				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     allErrors.Error(),
					"event":     "partsFinished",
					"message":   "rewindSuccessfulParts",
				})
				u.rewindSuccessfulParts()
				u.logUploadV2Complete(allErrors, u.bytesWritten)
				return u.UploadResumable(), allErrors
			}
			// Use a detached context so a job cancellation doesn't prevent
			// committing a file whose bytes are already on the server.
			path, ref := u.Path, u.FileUploadPart.Ref
			if u.renamedCallback != nil {
				path, ref = u.renamedCallback()
			}
			u.Manager.WaitWithContext(context.Background())
			var err error
			finalizeStart := time.Now()
			u.file, err = u.completeUpload(context.Background(), u.ProvidedMtime, u.etags, u.bytesWritten, path, ref)
			u.uploadV1Stats.recordFinalize(time.Since(finalizeStart), err)
			u.Manager.Done()
			u.logUploadV1SchedulerSummary(err)
			if err != nil {
				u.LogPath(u.Path, map[string]any{
					"timestamp": time.Now(),
					"error":     err.Error(),
					"event":     "complete upload",
					"message":   "rewindSuccessfulParts",
				})
				u.rewindSuccessfulParts()
			}
			u.logUploadV2Complete(err, u.bytesWritten)
			return u.UploadResumable(), err
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
			part.FileUploadPart.DirectConnectionInfo = u.FileUploadPart.DirectConnectionInfo
			part.Expires = u.FileUploadPart.Expires
			if part.PartNumber > 1 {
				// When using the same URL, after each use, the expiry time is extended by 3 minutes.
				part.Expires = u.FileUploadPart.ExpiresTime().Add(3 * time.Minute).Format(time.RFC3339)
			}
			u.ensurePartNumber(part)
		}
	}
	part.FileUploadPart.PartNumber = int64(number)

	// Parts are stored so a retry can pick up failed parts. Since io.Reader is a stream is better to just retry the whole file
	if _, readerAtOk := u.ReaderAt(); readerAtOk && u.Size != nil {
		u.Parts = append(u.Parts, part)
	}

	return part
}

func (u *uploadIO) ensurePartNumber(part *Part) {
	part.UploadUri = uploadURLWithPartNumber(part.UploadUri, part.PartNumber)
	if part.DirectConnectionInfo.DirectUri != "" {
		part.DirectConnectionInfo.DirectUri = uploadURLWithPartNumber(part.DirectConnectionInfo.DirectUri, part.PartNumber)
	}
}

func uploadURLWithPartNumber(rawURL string, partNumber int64) string {
	uri, err := url.Parse(rawURL)
	if err == nil {
		q := uri.Query()
		q.Del("part_number")
		q.Del("partNumber")
		q.Add("part_number", strconv.FormatInt(partNumber, 10))
		uri.RawQuery = q.Encode()
		return uri.String()
	}
	return rawURL
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

	waitStart := time.Now()
	if wait.WaitWithContext(ctx) {
		u.uploadV1Stats.recordScheduled(time.Since(waitStart), true, wait.RunningCount())
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
		u.uploadV1Stats.recordScheduled(time.Since(waitStart), false, wait.RunningCount())
		part.error = ctx.Err()
		u.onComplete <- part
		return true
	}

	return false
}

func (u *uploadIO) runUploadPart(ctx context.Context, part *Part) {
	runCount := 0
	start := time.Now()
	for {
		runCount++
		part.EtagsParam, part.error = u.uploadPart(ctx, part)
		part.bytes = int64(part.ProxyReader.BytesRead())
		part.Touch()
		if part.error == nil {
			break
		} else if lib.S3ErrorIsRequestHasExpired(part.error) || files_sdk.IsExpired(part.error) {
			if runCount >= maxUploadPartAttempts {
				break
			}
			part.FileUploadPart.Expires = ""
			part.FileUploadPart.UploadUri = ""
			if !u.rewindPartForRetry(part, runCount, "clearing upload_uri and fetching new one") {
				break
			}
		} else {
			break
		}
	}

	u.uploadV1Stats.recordPartComplete(part.bytes, time.Since(start), part.error)
	u.onComplete <- part
}

func (u *uploadIO) rewindPartForRetry(part *Part, runCount int, message string) bool {
	if !part.ProxyReader.Rewind() {
		return false
	}

	logAttrs := map[string]any{
		"timestamp": time.Now(),
		"error":     uploadRetryLogError(part.error),
		"part":      part.PartNumber,
		"run_count": runCount,
		"message":   message,
	}

	u.LogPath(u.Path, logAttrs)
	return true
}

func uploadRetryLogError(err error) string {
	if err == nil {
		return ""
	}
	if classified, ok := lib.ClassifyS3Error(err); ok {
		return classified.Message
	}
	return strings.Join(strings.Fields(err.Error()), " ")
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
		iterator = u.ByteOffset.Resume(u.Size, 0, 0)
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
	if !u.Config.DisableDirectTransfers && beginUpload.WithDirectConnectionInfo == nil {
		beginUpload.WithDirectConnectionInfo = lib.Bool(true)
	}
	uploads, err := u.BeginUpload(beginUpload, files_sdk.WithContext(ctx))
	if err != nil {
		return files_sdk.FileUploadPart{}, err
	}
	if !u.uploadV2 {
		u.markTransferStarted()
	}
	part := uploads[0]
	if u.startedCallback != nil {
		u.startedCallback(part)
	}
	return part, err
}

func (u *uploadIO) markTransferStarted() {
	// uploadIO is scoped to one file. Upload v2 acquires capacity once per part,
	// so keep the zero-byte start signal idempotent across parts.
	if !u.transferStarted.CompareAndSwap(false, true) {
		return
	}
	u.Progress(0)
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

			start := time.Now()
			part.FileUploadPart, err = u.startUpload(ctx, params)
			u.uploadV1Stats.recordUploadURLRefresh(time.Since(start), err)
			if err != nil && files_sdk.IsNotExist(err) && params.Ref != "" {
				// Stale upload session (server cleaned it up) — retry without the ref
				// to start a fresh session instead of failing the part.
				params.Ref = ""
				start = time.Now()
				part.FileUploadPart, err = u.startUpload(ctx, params)
				u.uploadV1Stats.recordUploadURLRefresh(time.Since(start), err)
			}
			part.FileUploadPart.PartNumber = int64(part.number) // Ensure it didn't change PartNumber

			if err != nil {
				return files_sdk.EtagsParam{}, err
			}
		}
	}
	headers := http.Header{}
	headers.Add("Content-Length", strconv.FormatInt(int64(part.ProxyReader.Len()), 10))
	params := &files_sdk.CallParams{
		Method:  part.HttpMethod,
		Config:  u.Config,
		Uri:     part.UploadUri,
		BodyIo:  part.ProxyReader,
		Headers: &headers,
		Context: ctx,
	}
	canReplay := func() bool {
		if part.ProxyReader.Len() == 0 {
			return true
		}
		return part.ProxyReader.Rewind()
	}
	params.Client = lib.UploadRetryableHttp(u.Config.Client, canReplay)
	if part.ProxyReader.Len() > 0 {
		// Rewind is the only replay-safety signal this reader exposes. It also resets
		// the reader, so the retry body factory rewinds again before each retry read.
		params.RetryableBody = retryablehttp.ReaderFunc(func() (io.Reader, error) {
			if !part.ProxyReader.Rewind() {
				return nil, errors.New("upload part body not rewindable")
			}
			return uploadPartRetryReader{Reader: part.ProxyReader, len: part.ProxyReader.Len()}, nil
		})
	}
	callStart := time.Now()
	res, err := u.callUploadPart(ctx, part, params, canReplay)
	if err != nil {
		u.uploadV1Stats.recordHTTPCall(time.Since(callStart), int64(part.ProxyReader.Len()), err)
		return files_sdk.EtagsParam{}, err
	}
	if err := lib.ResponseErrors(res, files_sdk.APIError(), lib.S3XMLError, lib.NonOkError); err != nil {
		u.uploadV1Stats.recordHTTPCall(time.Since(callStart), int64(part.ProxyReader.Len()), err)
		return files_sdk.EtagsParam{}, err
	}
	u.uploadV1Stats.recordHTTPCall(time.Since(callStart), int64(part.ProxyReader.Len()), nil)
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

func (u *uploadIO) callUploadPart(ctx context.Context, part *Part, params *files_sdk.CallParams, canReplay func() bool) (*http.Response, error) {
	if res, ok := u.callRawWithDirectTransfer(ctx, part.DirectConnectionInfo, params, directTransferUploadOptions{
		partNumber: part.PartNumber,
		canReplay:  canReplay,
		stats:      &u.uploadV1Stats,
	}); ok {
		return res, nil
	}

	return files_sdk.CallRaw(params)
}

func (u *uploadIO) directTransferUploadAvailable(info files_sdk.DirectConnectionInfo) bool {
	return !u.Config.DisableDirectTransfers &&
		!u.directTransferDisabled.Load() &&
		files_sdk.DirectConnectionInfoPresent(info)
}

func (u *uploadIO) disableDirectTransfersForUpload(reason string, err error, partNumber int64) bool {
	firstDisable := u.directTransferDisabled.CompareAndSwap(false, true)
	if firstDisable {
		attrs := map[string]any{
			"message":   "direct upload disabled; using proxy URLs for remaining upload parts",
			"direction": "upload",
			"reason":    reason,
			"part":      partNumber,
		}
		if err != nil {
			attrs["error"] = err.Error()
		}
		u.LogPath(u.Path, attrs)
	}
	return firstDisable
}

type uploadPartRetryReader struct {
	io.Reader
	len int
}

func (r uploadPartRetryReader) Len() int {
	return r.len
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
