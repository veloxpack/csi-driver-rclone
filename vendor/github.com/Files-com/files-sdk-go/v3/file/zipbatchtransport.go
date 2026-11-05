package file

// ZIP batch transport: ZIP create requests, byte movement, extraction, retry, fallback, and entry mapping.
//
// Cross-cutting ZIP batch invariants:
//   - zipBatchDownloader fields are walk-goroutine-only unless the field is explicitly atomic or mutex-protected.
//   - Each DownloadStatus is signaled exactly once; a duplicate signal can panic and a missing signal can hang the job.
//   - Every file admission Wait/WaitWithContext has exactly one Done, including retry, fallback, and cancel paths.
//   - Batch-routed files prepare at dispatch; per-file downloads prepare at download.

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

type zipBatchFallbackReason string

const (
	zipBatchFallbackCreateError      zipBatchFallbackReason = "create-error"
	zipBatchFallbackTripwire         zipBatchFallbackReason = "tripwire"
	zipBatchFallbackRetriesExhausted zipBatchFallbackReason = "retries-exhausted"
	zipBatchFallbackMissingEntry     zipBatchFallbackReason = "missing-entry"
)

type zipBatchFallback struct {
	status *DownloadStatus
	reason zipBatchFallbackReason
}

type zipDownloadCreateParams struct {
	// Paths is required by the API even when all paths are sent URI-encoded via
	// EncodedPaths; it must serialize as an empty array, never null.
	Paths        []string `json:"paths" url:"paths"`
	EncodedPaths []string `json:"encoded_paths" url:"encoded_paths"`
}

type zipDownloadCreateResponse struct {
	DownloadURI string `json:"download_uri"`
}

type zipBatchCreateResult struct {
	downloadURI string
	statuses    []*DownloadStatus
	fallback    []*DownloadStatus
}

type zipBatchCorruptArchiveError struct {
	error
}

type zipBatchTripwireError struct {
	error
}

var zipBatchIncrementPattern = regexp.MustCompile(`(.+?)(\.[^.\\:*?"<>|\r\n]+$)`)

func zipBatchFallbacks(reason zipBatchFallbackReason, statuses []*DownloadStatus) []zipBatchFallback {
	fallbacks := make([]zipBatchFallback, 0, len(statuses))
	for _, downloadStatus := range statuses {
		fallbacks = append(fallbacks, zipBatchFallback{status: downloadStatus, reason: reason})
	}
	return fallbacks
}

func zipBatchErrorClass(err error) string {
	var corrupt zipBatchCorruptArchiveError
	if errors.As(err, &corrupt) {
		return "corrupt"
	}
	return "stream"
}

func (b *zipBatchDownloader) run(statuses []*DownloadStatus) []zipBatchFallback {
	statuses = b.prepareStatuses(statuses)
	if len(statuses) == 0 {
		return nil
	}
	remaining := statuses
	var fallback []zipBatchFallback
	retryCount := 0
	if policy, ok := b.job.RetryPolicy.(RetryPolicy); ok {
		retryCount = policy.RetryCount
	}

	for attempt := 0; len(remaining) > 0 && attempt <= retryCount; attempt++ {
		if attempt > 0 {
			zipBatchResetProgress(b.job, remaining)
		}
		done := make(map[*DownloadStatus]struct{})
		createResult, err := b.createZipDownload(remaining)
		fallback = append(fallback, zipBatchFallbacks(zipBatchFallbackCreateError, createResult.fallback)...)
		if err != nil {
			reason := zipBatchFallbackCreateError
			var tripwire zipBatchTripwireError
			if errors.As(err, &tripwire) {
				reason = zipBatchFallbackTripwire
			}
			fallback = append(fallback, zipBatchFallbacks(reason, createResult.statuses)...)
			break
		}
		if len(createResult.statuses) == 0 {
			break
		}

		finalized := func(downloadStatus *DownloadStatus) {
			// Ended statuses are already signaled here; enqueueFallbacks depends on
			// that invariant to skip them without double-signaling the job.
			done[downloadStatus] = struct{}{}
			b.probeZipFinalized(downloadStatus)
			b.signal <- downloadStatus
		}
		if stats := b.stats(); stats != nil {
			stats.streamAttempts.Add(1)
		}
		b.log("stream-attempt", map[string]interface{}{"attempt": attempt + 1, "files": len(createResult.statuses)})
		spoolPath := ""
		var missing []*DownloadStatus
		if b.params.Extraction == ZipBatchExtractionStream {
			missing, err = b.extractZipStream(createResult.downloadURI, createResult.statuses, finalized)
		} else {
			spoolPath, err = b.spoolZip(createResult.downloadURI, createResult.statuses)
			if err == nil {
				missing, err = b.extractZipSpool(spoolPath, createResult.statuses, finalized)
				if err == nil {
					_ = os.Remove(spoolPath)
				}
			}
		}
		if err == nil {
			fallback = append(fallback, zipBatchFallbacks(zipBatchFallbackMissingEntry, missing)...)
			break
		}

		var tripwire zipBatchTripwireError
		if errors.As(err, &tripwire) {
			if spoolPath != "" {
				_ = os.Remove(spoolPath)
			}
			remaining = zipBatchRemaining(createResult.statuses, done)
			fallback = append(fallback, zipBatchFallbacks(zipBatchFallbackTripwire, remaining)...)
			b.log("tripwire", map[string]interface{}{
				"attempt": attempt + 1,
				"error":   err.Error(),
				"files":   len(remaining),
			})
			break
		}

		if err != nil {
			if stats := b.stats(); stats != nil {
				stats.streamFailures.Add(1)
			}
			b.log("stream-failed", map[string]interface{}{
				"attempt": attempt + 1,
				"class":   zipBatchErrorClass(err),
				"error":   err.Error(),
			})
		}
		if spoolPath != "" {
			b.salvageZipSpool(spoolPath, createResult.statuses, finalized)
			_ = os.Remove(spoolPath)
		}
		remaining = zipBatchRemaining(createResult.statuses, done)
		if attempt == retryCount {
			fallback = append(fallback, zipBatchFallbacks(zipBatchFallbackRetriesExhausted, remaining)...)
			break
		}
	}

	return fallback
}

func (b *zipBatchDownloader) prepareStatuses(statuses []*DownloadStatus) []*DownloadStatus {
	prepared := make([]*DownloadStatus, 0, len(statuses))
	for _, downloadStatus := range statuses {
		if downloadStatus == nil || downloadStatus.Status().Is(status.Ended...) {
			continue
		}
		if _, ok := prepareDownloadFolderItem(downloadStatus); !ok {
			b.signal <- downloadStatus
			continue
		}
		prepared = append(prepared, downloadStatus)
	}
	return prepared
}

func (b *zipBatchDownloader) enqueueFallbacks(fallbacks []zipBatchFallback) {
	// Ended statuses must already have been signaled by run's finalized closure.
	// Sending them again here would double-signal or hang WaitTellFinished.
	seen := make(map[*DownloadStatus]struct{}, len(fallbacks))
	reasons := make(map[zipBatchFallbackReason]int)
	var enqueue []*DownloadStatus
	for _, fallback := range fallbacks {
		downloadStatus := fallback.status
		if downloadStatus == nil {
			continue
		}
		if _, ok := seen[downloadStatus]; ok {
			continue
		}
		seen[downloadStatus] = struct{}{}
		if downloadStatus.Status().Is(status.Ended...) {
			continue
		}
		reasons[fallback.reason]++
		enqueue = append(enqueue, downloadStatus)
	}
	for reason, count := range reasons {
		b.recordFallback(reason, count)
		b.log("fallback", map[string]interface{}{"reason": string(reason), "files": count})
	}
	zipBatchResetProgress(b.job, enqueue)
	for _, downloadStatus := range enqueue {
		enqueueIndexedDownloadDirect(b.job, b.ctx, downloadStatus, b.signal)
	}
}

func zipBatchHasFallbackReason(fallbacks []zipBatchFallback, reason zipBatchFallbackReason) bool {
	for _, fallback := range fallbacks {
		if fallback.reason == reason {
			return true
		}
	}
	return false
}

func zipBatchRemaining(statuses []*DownloadStatus, done map[*DownloadStatus]struct{}) []*DownloadStatus {
	statuses = zipBatchNotEnded(statuses)
	var remaining []*DownloadStatus
	for _, downloadStatus := range statuses {
		if _, ok := done[downloadStatus]; !ok && !downloadStatus.Status().Is(status.Ended...) {
			remaining = append(remaining, downloadStatus)
		}
	}
	return remaining
}

func zipBatchNotEnded(statuses []*DownloadStatus) []*DownloadStatus {
	// Compacts in place; callers must not reuse the original slice length.
	active := statuses[:0]
	for _, downloadStatus := range statuses {
		if !downloadStatus.Status().Is(status.Ended...) {
			active = append(active, downloadStatus)
		}
	}
	return active
}

func zipBatchResetProgress(job *Job, statuses []*DownloadStatus) {
	for _, downloadStatus := range statuses {
		if downloadStatus != nil && !downloadStatus.Status().Is(status.Ended...) && downloadStatus.TransferBytes() > 0 {
			job.UpdateStatus(status.Retrying, downloadStatus, nil)
		}
	}
}

func (b *zipBatchDownloader) createZipDownload(statuses []*DownloadStatus) (zipBatchCreateResult, error) {
	sortedStatuses, err := zipBatchSortedStatuses(statuses)
	if err != nil {
		return zipBatchCreateResult{statuses: statuses}, zipBatchTripwireError{error: err}
	}

	downloadURI, err := b.createZipDownloadOnce(sortedStatuses)
	if err == nil {
		return zipBatchCreateResult{downloadURI: downloadURI, statuses: sortedStatuses}, nil
	}

	var pathErr zipBatchPathError
	if !errors.As(err, &pathErr) || len(pathErr.paths) == 0 {
		return zipBatchCreateResult{statuses: sortedStatuses}, err
	}

	remainder, fallback := zipBatchSplitPathErrors(sortedStatuses, pathErr.paths)
	if len(remainder) == 0 {
		return zipBatchCreateResult{fallback: fallback}, nil
	}

	downloadURI, err = b.createZipDownloadOnce(remainder)
	if err != nil {
		return zipBatchCreateResult{statuses: append(fallback, remainder...)}, err
	}
	return zipBatchCreateResult{downloadURI: downloadURI, statuses: remainder, fallback: fallback}, nil
}

func (b *zipBatchDownloader) createZipDownloadOnce(statuses []*DownloadStatus) (string, error) {
	if stats := b.stats(); stats != nil {
		stats.createRequests.Add(1)
	}
	b.log("create-request", map[string]interface{}{"files": len(statuses)})
	paths := make([]string, 0, len(statuses))
	for _, downloadStatus := range statuses {
		paths = append(paths, url.PathEscape(downloadStatus.RemotePath()))
	}
	params := lib.Params{Params: zipDownloadCreateParams{Paths: []string{}, EncodedPaths: paths}}

	headers := &http.Header{}
	b.job.Config.SetHeaders(headers)
	response, err := files_sdk.CallRaw(&files_sdk.CallParams{
		Method:  http.MethodPost,
		Config:  b.job.Config,
		Uri:     b.job.Config.RootPath() + "/zip_downloads",
		Params:  params,
		Headers: headers,
		Context: b.ctx,
		RetryableBody: func() (io.Reader, error) {
			return params.ToJSON()
		},
	})
	if err != nil {
		return "", err
	}
	defer lib.CloseBody(response)

	if response.StatusCode == http.StatusUnprocessableEntity {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return "", readErr
		}
		return "", zipBatchPathError{paths: zipBatchParse422Paths(body), body: string(body)}
	}

	data, _, err := files_sdk.ParseResponse(response, "/zip_downloads")
	if err != nil {
		return "", err
	}
	var createResponse zipDownloadCreateResponse
	if err := json.Unmarshal(*data, &createResponse); err != nil {
		return "", err
	}
	if createResponse.DownloadURI == "" {
		return "", fmt.Errorf("zip batch: zip_downloads response missing download_uri")
	}
	return createResponse.DownloadURI, nil
}

func (b *zipBatchDownloader) spoolZip(downloadURI string, statuses []*DownloadStatus) (string, error) {
	streamURI, err := zipBatchStreamURI(b.job.Config, downloadURI)
	if err != nil {
		return "", err
	}
	spoolPath, err := tmpDownloadPath(zipBatchSpoolBasePath(statuses), statuses[0].tempPath)
	if err != nil {
		return "", err
	}
	out, err := os.Create(spoolPath)
	if err != nil {
		return spoolPath, err
	}
	writer := lib.ProgressWriter{WriterAndAt: out}
	writer.ProgressWatcher = func(bytes int64) {
		if bytes > 0 {
			b.job.Meter.Record(time.Now(), uint64(bytes))
		}
	}
	zipFile := newZipBatchFile(b.job.Config, b.ctx, streamURI)
	defer zipFile.Close()
	zipInfo, _ := zipFile.Stat()
	downloadParts := (&DownloadParts{}).Init(
		zipFile,
		zipInfo,
		b.job.Manager.FilePartsManager,
		writer,
		b.job.Config,
		0,
	)

	var runErr error
	lib.AnyError(func(err error) {
		runErr = err
	},
		func() error { return downloadParts.Run(b.ctx) },
		func() error { return downloadParts.CloseError },
	)
	return spoolPath, runErr
}

func zipBatchSpoolBasePath(statuses []*DownloadStatus) string {
	return statuses[0].LocalPath() + ".zip-batch"
}

func (b *zipBatchDownloader) extractZipSpool(spoolPath string, statuses []*DownloadStatus, finalized func(*DownloadStatus)) ([]*DownloadStatus, error) {
	expected, err := zipBatchEntryMap(statuses)
	if err != nil {
		return nil, err
	}
	reader, err := zip.OpenReader(spoolPath)
	if err != nil {
		return nil, zipBatchCorruptArchiveError{error: err}
	}
	defer reader.Close()

	seen := make(map[*DownloadStatus]struct{}, len(expected))
	for _, zipFile := range reader.File {
		downloadStatus, ok := expected[zipFile.Name]
		if !ok {
			return nil, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q was not requested", zipFile.Name)}
		}
		if downloadStatus.Status().Is(status.Ended...) {
			seen[downloadStatus] = struct{}{}
			continue
		}
		if _, ok := seen[downloadStatus]; ok {
			return nil, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q was duplicated", zipFile.Name)}
		}
		if err := extractZipArchiveEntry(b.ctx, zipFile, downloadStatus); err != nil {
			return nil, err
		}
		if !downloadStatus.Status().Is(status.Errored) {
			downloadStatus.Job().UpdateStatus(status.Complete, downloadStatus, nil)
		}
		if stats := b.stats(); stats != nil && !downloadStatus.Status().Is(status.Errored) {
			stats.cleanFinalized.Add(1)
		}
		seen[downloadStatus] = struct{}{}
		finalized(downloadStatus)
	}

	return zipBatchMissing(expected, seen), nil
}

func (b *zipBatchDownloader) extractZipStream(downloadURI string, statuses []*DownloadStatus, finalized func(*DownloadStatus)) ([]*DownloadStatus, error) {
	streamURI, err := zipBatchStreamURI(b.job.Config, downloadURI)
	if err != nil {
		return nil, err
	}
	expected, err := zipBatchEntryMap(statuses)
	if err != nil {
		return nil, zipBatchTripwireError{error: err}
	}
	zipFile := newZipBatchFile(b.job.Config, b.ctx, streamURI)
	defer zipFile.Close()
	return extractZipBatchStream(
		b.ctx,
		zipFile,
		expected,
		func(downloadStatus *DownloadStatus) {
			if stats := b.stats(); stats != nil && !downloadStatus.Status().Is(status.Errored) {
				stats.cleanFinalized.Add(1)
			}
			finalized(downloadStatus)
		},
		func(downloadStatus *DownloadStatus, bytes int64) {
			downloadStatus.Job().UpdateStatusWithBytes(status.Downloading, downloadStatus, bytes)
		},
	)
}

func extractZipArchiveEntry(ctx context.Context, zipFile *zip.File, downloadStatus *DownloadStatus) error {
	in, err := zipFile.Open()
	if err != nil {
		return zipBatchCorruptArchiveError{error: err}
	}
	defer in.Close()
	if err := extractZipEntryReader(ctx, zipFile.Name, in, downloadStatus); err != nil {
		var tripwire zipBatchTripwireError
		if errors.As(err, &tripwire) {
			return err
		}
		return zipBatchCorruptArchiveError{error: err}
	}
	return nil
}

func (b *zipBatchDownloader) salvageZipSpool(spoolPath string, statuses []*DownloadStatus, finalized func(*DownloadStatus)) {
	expected, err := zipBatchEntryMap(statuses)
	if err != nil {
		return
	}
	spool, err := os.Open(spoolPath)
	if err != nil {
		return
	}
	defer spool.Close()
	recovered := 0
	_, _ = extractZipBatchStream(b.ctx, spool, expected, func(downloadStatus *DownloadStatus) {
		if downloadStatus.Status().Is(status.Errored) {
			finalized(downloadStatus)
			return
		}
		recovered++
		if stats := b.stats(); stats != nil {
			stats.salvageFinalized.Add(1)
		}
		finalized(downloadStatus)
	}, nil)
	b.log("salvage-result", map[string]interface{}{"entries": recovered})
}

func extractZipBatchStream(ctx context.Context, r io.Reader, expected map[string]*DownloadStatus, finalized func(*DownloadStatus), onBytes func(*DownloadStatus, int64)) ([]*DownloadStatus, error) {
	stream := newZipStream(r)
	seen := make(map[*DownloadStatus]struct{}, len(expected))
	for {
		header, ok, err := stream.nextHeader()
		if err != nil {
			return nil, err
		}
		if !ok {
			return zipBatchMissing(expected, seen), nil
		}

		downloadStatus, ok := expected[header.name]
		if !ok {
			return nil, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q was not requested", header.name)}
		}
		if downloadStatus.Status().Is(status.Ended...) {
			result, err := stream.extractEntry(ctx, header, io.Discard, nil)
			if err != nil {
				return nil, err
			}
			if int64(result.uncompressedSize) != downloadStatus.Size() {
				return nil, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q size %d did not match listed size %d", header.name, result.uncompressedSize, downloadStatus.Size())}
			}
			seen[downloadStatus] = struct{}{}
			continue
		}
		if _, ok := seen[downloadStatus]; ok {
			return nil, zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q was duplicated", header.name)}
		}

		if err := extractZipStreamEntry(ctx, stream, header, downloadStatus, onBytes); err != nil {
			return nil, err
		}
		if !downloadStatus.Status().Is(status.Errored) {
			downloadStatus.Job().UpdateStatus(status.Complete, downloadStatus, nil)
		}
		seen[downloadStatus] = struct{}{}
		finalized(downloadStatus)
	}
}

func extractZipStreamEntry(ctx context.Context, stream *zipStream, header zipStreamEntryHeader, downloadStatus *DownloadStatus, onBytes func(*DownloadStatus, int64)) error {
	tmpName, err := tmpDownloadPath(downloadStatus.LocalPath(), downloadStatus.tempPath)
	if err != nil {
		return err
	}
	downloadStatus.TmpPath = tmpName
	out, err := os.Create(tmpName)
	if err != nil {
		return err
	}
	result, extractErr := stream.extractEntry(ctx, header, out, func(bytes int64) {
		if onBytes != nil {
			onBytes(downloadStatus, bytes)
		}
	})
	closeErr := out.Close()
	if extractErr != nil {
		removeTmpDownload(tmpName)
		return extractErr
	}
	if closeErr != nil {
		removeTmpDownload(tmpName)
		return closeErr
	}
	if int64(result.uncompressedSize) != downloadStatus.Size() {
		removeTmpDownload(tmpName)
		return zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q size %d did not match listed size %d", header.name, result.uncompressedSize, downloadStatus.Size())}
	}
	if err := completeTmpDownload(downloadStatus, tmpName, int64(result.uncompressedSize)); err != nil {
		removeTmpDownload(tmpName)
		downloadStatus.Job().UpdateStatus(status.Errored, downloadStatus, err)
	}
	return nil
}

func extractZipEntryReader(ctx context.Context, name string, in io.Reader, downloadStatus *DownloadStatus) error {
	tmpName, err := tmpDownloadPath(downloadStatus.LocalPath(), downloadStatus.tempPath)
	if err != nil {
		return err
	}
	downloadStatus.TmpPath = tmpName
	out, err := os.Create(tmpName)
	if err != nil {
		return err
	}
	written, copyErr := copyWithContext(ctx, out, in)
	closeErr := out.Close()
	if copyErr != nil {
		removeTmpDownload(tmpName)
		return copyErr
	}
	if closeErr != nil {
		removeTmpDownload(tmpName)
		return closeErr
	}
	if written != downloadStatus.Size() {
		removeTmpDownload(tmpName)
		return zipBatchTripwireError{error: fmt.Errorf("zip batch: zip entry %q size %d did not match listed size %d", name, written, downloadStatus.Size())}
	}
	if err := completeTmpDownload(downloadStatus, tmpName, written); err != nil {
		removeTmpDownload(tmpName)
		downloadStatus.Job().UpdateStatus(status.Errored, downloadStatus, err)
	}
	return nil
}

func copyWithContext(ctx context.Context, out io.Writer, in io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, readErr := in.Read(buf)
		if n > 0 {
			wn, writeErr := out.Write(buf[:n])
			written += int64(wn)
			if writeErr != nil {
				return written, writeErr
			}
			if wn != n {
				return written, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
}

func zipBatchStreamURI(config files_sdk.Config, downloadURI string) (string, error) {
	parsed, err := url.Parse(downloadURI)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	base, err := url.Parse(config.Endpoint())
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}

type zipBatchFile struct {
	files_sdk.File
	config files_sdk.Config
	ctx    context.Context
	body   io.ReadCloser
	read   int64
}

func newZipBatchFile(config files_sdk.Config, ctx context.Context, downloadURI string) *zipBatchFile {
	return &zipBatchFile{
		File: files_sdk.File{
			DisplayName: "zip-batch.zip",
			Path:        "zip-batch.zip",
			Type:        "file",
			DownloadUri: downloadURI,
		},
		config: config,
		ctx:    ctx,
	}
}

func (f *zipBatchFile) Stat() (fs.FileInfo, error) {
	file := f.File
	// Returning size-so-far keeps DownloadParts on its single-stream path; a
	// fixed unknown size would route to ReaderRange and panic on zipBatchFile.
	file.Size = f.read
	return Info{File: file, sizeTrust: TrustedSizeValue}, nil
}

func (f *zipBatchFile) Read(p []byte) (int, error) {
	if f.body == nil {
		if err := f.open(); err != nil {
			return 0, err
		}
	}
	n, err := f.body.Read(p)
	f.read += int64(n)
	return n, err
}

func (f *zipBatchFile) Close() error {
	if f.body == nil {
		return nil
	}
	body := f.body
	f.body = nil
	return body.Close()
}

func (f *zipBatchFile) open() error {
	_, err := (&Client{Config: f.config}).Download(
		files_sdk.FileDownloadParams{File: f.File},
		files_sdk.WithContext(f.ctx),
		files_sdk.ResponseOption(func(response *http.Response) error {
			if err := lib.ResponseErrors(response, files_sdk.APIError(), lib.NotStatus(http.StatusOK)); err != nil {
				return err
			}
			f.body = response.Body
			return nil
		}),
	)
	return err
}

func zipBatchSortedStatuses(statuses []*DownloadStatus) ([]*DownloadStatus, error) {
	sortedStatuses := append([]*DownloadStatus(nil), statuses...)
	sort.Slice(sortedStatuses, func(i, j int) bool {
		return sortedStatuses[i].RemotePath() < sortedStatuses[j].RemotePath()
	})
	if _, err := zipBatchEntryMap(sortedStatuses); err != nil {
		return nil, err
	}
	return sortedStatuses, nil
}

func zipBatchEntryMap(statuses []*DownloadStatus) (map[string]*DownloadStatus, error) {
	sortedStatuses := append([]*DownloadStatus(nil), statuses...)
	sort.Slice(sortedStatuses, func(i, j int) bool {
		return sortedStatuses[i].RemotePath() < sortedStatuses[j].RemotePath()
	})

	counts := make(map[string]int, len(sortedStatuses))
	entries := make(map[string]*DownloadStatus, len(sortedStatuses))
	for _, downloadStatus := range sortedStatuses {
		name := filepath.Base(downloadStatus.RemotePath())
		if count, ok := counts[name]; ok {
			count++
			counts[name] = count
			incrementedName, err := zipBatchIncrementFilename(name, count)
			if err != nil {
				return nil, err
			}
			name = incrementedName
		}
		if existing := entries[name]; existing != nil {
			return nil, fmt.Errorf("zip batch: zip entry name %q maps to both %q and %q", name, existing.RemotePath(), downloadStatus.RemotePath())
		}
		entries[name] = downloadStatus
		counts[name] = 0
	}
	return entries, nil
}

func zipBatchIncrementFilename(filename string, count int) (string, error) {
	matches := zipBatchIncrementPattern.FindStringSubmatch(filename)
	if len(matches) == 3 {
		return fmt.Sprintf("%s (%d)%s", matches[1], count, matches[2]), nil
	}
	if !strings.Contains(filename, ".") {
		return fmt.Sprintf("%s (%d)", filename, count), nil
	}
	return "", fmt.Errorf("zip batch: could not match filename/extension")
}

func zipBatchMissing(expected map[string]*DownloadStatus, seen map[*DownloadStatus]struct{}) []*DownloadStatus {
	var missing []*DownloadStatus
	for _, downloadStatus := range expected {
		if _, ok := seen[downloadStatus]; !ok {
			missing = append(missing, downloadStatus)
		}
	}
	sort.Slice(missing, func(i, j int) bool {
		return missing[i].RemotePath() < missing[j].RemotePath()
	})
	return missing
}

type zipBatchPathError struct {
	paths map[string]struct{}
	body  string
}

func (e zipBatchPathError) Error() string {
	return e.body
}

func zipBatchParse422Paths(body []byte) map[string]struct{} {
	var response files_sdk.ResponseError
	if err := json.Unmarshal(body, &response); err == nil {
		paths := make(map[string]struct{})
		zipBatchCollect422Paths(paths, response.ErrorMessage)
		for _, nested := range response.Errors {
			zipBatchCollect422Paths(paths, nested.ErrorMessage)
		}
		if len(paths) > 0 {
			return paths
		}
	}

	paths := make(map[string]struct{})
	zipBatchCollect422Paths(paths, string(body))
	return paths
}

func zipBatchCollect422Paths(paths map[string]struct{}, message string) {
	const missingPrefix = "Paths does not match existing file at: "
	const lockedPrefix = "Paths cannot be downloaded at: "
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{missingPrefix, lockedPrefix} {
			if strings.HasPrefix(line, prefix) {
				paths[strings.TrimSpace(strings.TrimPrefix(line, prefix))] = struct{}{}
			}
		}
	}
}

func zipBatchSplitPathErrors(statuses []*DownloadStatus, paths map[string]struct{}) ([]*DownloadStatus, []*DownloadStatus) {
	var remainder, fallback []*DownloadStatus
	for _, downloadStatus := range statuses {
		if _, ok := paths[downloadStatus.RemotePath()]; ok {
			fallback = append(fallback, downloadStatus)
		} else {
			remainder = append(remainder, downloadStatus)
		}
	}
	return remainder, fallback
}
