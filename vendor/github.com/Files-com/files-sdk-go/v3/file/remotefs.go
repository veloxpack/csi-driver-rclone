package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	goFs "io/fs"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/folder"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/samber/lo"
)

type SizeTrust int

const (
	NullSizeTrust SizeTrust = iota
	UntrustedSizeValue
	TrustedSizeValue
)

type FS struct {
	files_sdk.Config
	context.Context
	Root       string
	cache      *sync.Map
	cacheDir   *sync.Map
	useCache   bool
	cacheMutex *lib.KeyedMutex
}

func (f *FS) Init(config files_sdk.Config, cache bool) *FS {
	f.Config = config
	f.ClearCache()
	f.useCache = cache
	return f
}

func (f *FS) WithContext(ctx context.Context) interface{} {
	return &FS{Context: ctx, Config: f.Config, cache: f.cache, useCache: f.useCache, cacheDir: f.cacheDir, cacheMutex: f.cacheMutex}
}

func (f *FS) ClearCache() {
	f.cache = &sync.Map{}
	f.cacheDir = &sync.Map{}
	m := lib.NewKeyedMutex()
	f.cacheMutex = &m
}

type File struct {
	*files_sdk.File
	*FS
	ReadCloser        lib.AtomicValue[io.ReadCloser]
	downloadRequestId string
	MaxConnections    int
	stat              bool
	fileMutex         *sync.Mutex
	SizeTrust
	serverBytesSent int64
}

type ReadDirFile struct {
	*File
	count int
}

func (f *File) safeFile() files_sdk.File {
	f.fileMutex.Lock()
	defer f.fileMutex.Unlock()
	return *f.File
}

func (f *File) Init() *File {
	f.fileMutex = &sync.Mutex{}
	f.SizeTrust = NullSizeTrust
	return f
}

func (f *File) Name() string {
	return f.safeFile().DisplayName
}

func (f *File) IsDir() bool {
	return f.safeFile().Type == "directory"
}

func (f *File) Type() goFs.FileMode {
	return goFs.ModePerm
}

func (f *File) Info() (goFs.FileInfo, error) {
	return f.Stat()
}

type Info struct {
	files_sdk.File
	sizeTrust SizeTrust
}

func (i Info) Name() string {
	return i.File.DisplayName
}

func (i Info) Size() int64 {
	return i.File.Size
}

type UntrustedSize interface {
	UntrustedSize() bool
	SizeTrust() SizeTrust
	goFs.FileInfo
}

func (i Info) UntrustedSize() bool {
	return i.sizeTrust == UntrustedSizeValue || i.sizeTrust == NullSizeTrust
}

func (i Info) SizeTrust() SizeTrust {
	return i.sizeTrust
}

type PossibleSize interface {
	PossibleSize() int64
}

func (i Info) PossibleSize() int64 {
	return i.File.Size
}

func (i Info) Mode() goFs.FileMode {
	return goFs.ModePerm
}

func (i Info) IsDir() bool {
	return i.File.Type == "directory"
}

func (i Info) Sys() interface{} {
	return i.File
}

func (i Info) RemoteMount() bool {
	if i.Crc32 != "" { // Detect if is Files.com native file.
		return false
	}

	return true
}

func (f *File) Stat() (goFs.FileInfo, error) {
	f.fileMutex.Lock()
	defer f.fileMutex.Unlock()
	return Info{File: *f.File, sizeTrust: f.SizeTrust}, nil
}

func (f *File) Read(b []byte) (n int, err error) {
	f.ReadCloser.CompareAndUpdate(nil, func() io.ReadCloser {
		f.fileMutex.Lock()
		defer f.fileMutex.Unlock()

		var readCloser io.ReadCloser
		readCloser, err = f.readCloserInit()
		if downloadRequestExpired(err) {
			f.Config.LogPath(f.File.Path, map[string]interface{}{"message": "downloadRequestExpired", "error": err})
			f.File.DownloadUri = "" // force a new query
			*f.File, err = (&Client{Config: f.Config}).DownloadUri(files_sdk.FileDownloadParams{File: *f.File}, files_sdk.WithContext(f.Context))
			if err == nil {
				readCloser, err = f.readCloserInit()
			}
		}

		if err != nil {
			readCloser = nil
			status, statusErr := (&Client{Config: f.Config}).DownloadRequestStatus(f.File.DownloadUri, f.downloadRequestId, files_sdk.WithContext(f.Context))
			if statusErr != nil {
				err = statusErr
			}
			if !status.IsNil() {
				err = status
			}
		}
		return readCloser
	})

	if err != nil {
		return
	}

	return f.ReadCloser.Load().Read(b)
}

func parseSize(response *http.Response) (size int64, sizeTrust SizeTrust) {
	var err error

	if response.StatusCode == http.StatusPartialContent {
		if contentRange := response.Header.Get("Content-Range"); contentRange != "" {
			rangeParts := strings.SplitN(contentRange, "/", 2)
			if len(rangeParts) == 2 {
				size, err = strconv.ParseInt(rangeParts[1], 10, 64)
				if err == nil {
					sizeTrust = TrustedSizeValue
					return
				}
			}
		}
	} else if response.ContentLength > -1 {
		sizeTrust = TrustedSizeValue
		size = response.ContentLength

		return
	}

	// For some remote mounts file size information cannot be trusted and will not be returned.
	// In order to ensure the total file was received after a download `Client{}.DownloadRequestStatus` should be called.
	sizeTrust = UntrustedSizeValue

	return
}

func parseMaxConnections(response *http.Response) int {
	maxConnections, _ := strconv.Atoi(response.Header.Get("X-Files-Max-Connections"))
	return maxConnections
}

func (f *File) readCloserInit() (readCloser io.ReadCloser, err error) {
	*f.File, err = (&Client{Config: f.Config}).Download(
		files_sdk.FileDownloadParams{File: *f.File},
		files_sdk.WithContext(f.Context),
		files_sdk.ResponseOption(func(response *http.Response) error {
			f.MaxConnections = parseMaxConnections(response)
			f.downloadRequestId = response.Header.Get("X-Files-Download-Request-Id")
			f.Size, f.SizeTrust = parseSize(response)
			if err := lib.ResponseErrors(response, files_sdk.APIError(), lib.NotStatus(http.StatusOK)); err != nil {
				return &goFs.PathError{Path: f.File.Path, Err: err, Op: "read"}
			}

			readCloser = &ReadWrapper{ReadCloser: response.Body}
			return nil
		}),
	)
	return readCloser, err
}

type ReaderRange interface {
	ReaderRange(off int64, end int64) (io.ReadCloser, error)
	goFs.File
}

type ReadAtLeastWrapper struct {
	io.ReadCloser
	io.Reader
}

func (r ReadAtLeastWrapper) Close() error {
	return r.ReadCloser.Close()
}

func (f ReadAtLeastWrapper) Read(b []byte) (n int, err error) {
	return f.Reader.Read(b)
}

func (f *File) ReaderRange(off int64, end int64) (r io.ReadCloser, err error) {
	if err = f.downloadURI(); err != nil {
		return
	}
	f.fileMutex.Lock()
	rangerReaderCloser := ReaderCloserDownloadStatus{file: f, expectedSize: (end + 1) - off, rangeRequest: true, ReadWrapper: &ReadWrapper{}}

	headers := &http.Header{}
	headers.Set("Range", fmt.Sprintf("bytes=%v-%v", off, end))
	_, err = (&Client{Config: f.Config}).Download(
		files_sdk.FileDownloadParams{File: *f.File},
		files_sdk.WithContext(f.Context),
		files_sdk.RequestHeadersOption(headers),
		files_sdk.ResponseOption(func(response *http.Response) error {
			f.downloadRequestId = response.Header.Get("X-Files-Download-Request-Id")
			rangerReaderCloser.file.downloadRequestId = response.Header.Get("X-Files-Download-Request-Id")
			f.MaxConnections = parseMaxConnections(response)
			f.Size, f.SizeTrust = parseSize(response)
			if err := lib.ResponseErrors(response, lib.IsStatus(http.StatusForbidden), files_sdk.APIError(), lib.NotStatus(http.StatusPartialContent)); err != nil {
				return &goFs.PathError{Path: f.File.Path, Err: err, Op: "ReaderRange"}
			}
			rangerReaderCloser.ReadCloser = &ReadWrapper{ReadCloser: response.Body}
			return nil
		}),
	)
	f.fileMutex.Unlock()
	if downloadRequestExpired(err) {
		f.Config.LogPath(f.File.Path, map[string]interface{}{"message": "downloadRequestExpired", "error": err})
		f.File.DownloadUri = "" // force a new query
		err = f.downloadURI()
		if err != nil {
			return r, err
		}

		return f.ReaderRange(off, end)
	}
	return rangerReaderCloser, err
}

type ReadWrapper struct {
	io.ReadCloser
	read int
}

func (r *ReadWrapper) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.read += n
	return
}

type ReaderCloserDownloadStatus struct {
	*ReadWrapper
	file         *File
	expectedSize int64
	rangeRequest bool
	UntrustedSizeRangeRequestSize
}

type UntrustedSizeRangeRequestSize struct {
	ExpectedSize int64
	SentSize     int64
	ReceivedSize int64
	Status       string
}

func (u UntrustedSizeRangeRequestSize) VerifyReceived() error {
	if u.Status == "started" { // Race condition where server does not record download status. Trust what we asked for and got is correct.
		if u.ReceivedSize != u.ExpectedSize {
			errors.Join(UntrustedSizeRangeRequestSizeExpectedReceived, fmt.Errorf("expected %v bytes %v received", u.ExpectedSize, u.ReceivedSize))
		}
	} else if u.ReceivedSize != u.SentSize {
		return errors.Join(UntrustedSizeRangeRequestSizeSentReceived, fmt.Errorf("expected %v bytes sent %v received", u.SentSize, u.ReceivedSize))
	}
	return nil
}

func (u UntrustedSizeRangeRequestSize) Log() map[string]interface{} {
	return map[string]interface{}{
		"expected_size":  u.ExpectedSize,
		"sent_size":      u.SentSize,
		"received_size":  u.ReceivedSize,
		"VerifyReceived": u.VerifyReceived(),
		"Mismatch":       u.Mismatch(),
		"Status":         u.Status,
	}
}

var UntrustedSizeRangeRequestSizeExpectedReceived = fmt.Errorf("received size did not match server expected size")
var UntrustedSizeRangeRequestSizeSentReceived = fmt.Errorf("received size did not match server send size")

func (u UntrustedSizeRangeRequestSize) Mismatch() error {
	if u.Status == "started" {
		return nil
	}
	if u.ExpectedSize > u.SentSize {
		return UntrustedSizeRangeRequestSizeSentLessThanExpected
	}
	if u.ExpectedSize < u.SentSize {
		return UntrustedSizeRangeRequestSizeSentMoreThanExpected
	}
	return nil
}

var UntrustedSizeRangeRequestSizeSentMoreThanExpected = fmt.Errorf("server send more than expected")

var UntrustedSizeRangeRequestSizeSentLessThanExpected = fmt.Errorf("server send less than expected")

func (r ReaderCloserDownloadStatus) Close() error {
	if r.ReadCloser == nil {
		return nil
	}
	err := r.ReadCloser.Close()
	defer func() { r.ReadCloser = nil }()
	if err != nil {
		return err
	}

	if r.file.downloadRequestId == "" {
		return nil
	}

	info, err := r.file.Info()
	if err != nil {
		return err
	}

	if untrustedInfo, ok := info.(UntrustedSize); ok && (untrustedInfo.UntrustedSize() || untrustedInfo.SizeTrust() == NullSizeTrust) {
		r.file.fileMutex.Lock()

		status, err := (&Client{Config: r.file.Config}).DownloadRequestStatus(r.file.DownloadUri, r.file.downloadRequestId, files_sdk.WithContext(r.file.Context))
		r.file.fileMutex.Unlock()
		if err != nil {
			return err
		}
		if !status.IsNil() && (status.Data.Status == "failed" || status.Data.Status == "error") {
			return status
		}
		r.UntrustedSizeRangeRequestSize = UntrustedSizeRangeRequestSize{
			r.expectedSize,
			status.Data.BytesTransferred,
			int64(r.ReadWrapper.read),
			status.Data.Status,
		}

		if err := r.UntrustedSizeRangeRequestSize.VerifyReceived(); err != nil {
			r.file.Config.LogPath(info.Name(), r.UntrustedSizeRangeRequestSize.Log())
			return err
		}

		// The true size can only be known after the server determines that the full file has been sent without any errors.
		if r.rangeRequest {
			if err := r.UntrustedSizeRangeRequestSize.Mismatch(); err != nil {
				r.file.Config.LogPath(info.Name(), r.UntrustedSizeRangeRequestSize.Log())
				return err
			}

			if r.file.SizeTrust == UntrustedSizeValue {
				r.file.serverBytesSent += status.Data.BytesTransferred
			}
		} else {
			r.file.SizeTrust = TrustedSizeValue
			r.file.Size = status.Data.BytesTransferred
		}

		if dataBytes, err := json.Marshal(status.Data); err == nil {
			dataMap := make(map[string]interface{})
			if err = json.Unmarshal(dataBytes, &dataMap); err == nil {
				r.file.Config.LogPath(info.Name(), lo.Assign(dataMap, map[string]interface{}{"message": "download request server status"}))
			}
		}
	}
	return nil
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	err = f.downloadURI()
	if err != nil {
		return
	}
	headers := &http.Header{}
	headers.Set("Range", fmt.Sprintf("bytes=%v-%v", off, int64(len(p))+off-1))
	_, err = (&Client{Config: f.Config}).Download(
		files_sdk.FileDownloadParams{
			File: *f.File,
		},
		files_sdk.WithContext(f.Context),
		files_sdk.RequestHeadersOption(headers),
		files_sdk.ResponseOption(func(response *http.Response) error {
			if err := lib.ResponseErrors(response, lib.IsStatus(http.StatusForbidden), lib.NotStatus(http.StatusPartialContent), files_sdk.APIError()); err != nil {
				return &goFs.PathError{Path: f.File.Path, Err: err, Op: "ReadAt"}
			}
			n, err = io.ReadFull(response.Body, p)
			if err != nil && err != io.EOF {
				return err
			}
			if int64(len(p)) >= response.ContentLength && int64(n) != response.ContentLength {
				return &goFs.PathError{Path: f.File.Path, Err: fmt.Errorf("content-length did not match body"), Op: "ReadAt"}
			}
			return nil
		}),
	)

	if downloadRequestExpired(err) {
		f.Config.LogPath(f.File.Path, map[string]interface{}{"message": "downloadRequestExpired", "error": err})
		f.File.DownloadUri = "" // force a new query
		err = f.downloadURI()
		if err != nil {
			return n, err
		}

		return f.ReadAt(p, off)
	}

	return n, err
}

func downloadRequestExpired(err error) bool {
	if err == nil {
		return false
	}
	var responseErr lib.ResponseError
	ok := errors.As(errors.Unwrap(err), &responseErr)
	return ok && responseErr.StatusCode == http.StatusForbidden
}

func (f *File) downloadURI() (err error) {
	f.fileMutex.Lock()
	*f.File, err = (&Client{Config: f.Config}).DownloadUri(files_sdk.FileDownloadParams{File: *f.File}, files_sdk.WithContext(f.Context))
	f.fileMutex.Unlock()
	return
}

func (f *File) Close() error {
	f.fileMutex.Lock()
	f.fileMutex.Unlock()
	defer func() { f.ReadCloser.Store(nil) }()
	switch t := f.ReadCloser.Load().(type) {
	case *ReadWrapper:
		return ReaderCloserDownloadStatus{ReadWrapper: t, file: f}.Close()
	default:
		return ReaderCloserDownloadStatus{ReadWrapper: &ReadWrapper{ReadCloser: t}, file: f}.Close()
	}
}

func (f *File) WithContext(ctx context.Context) goFs.File {
	newF := *f
	fs := *newF.FS
	newF.FS = fs.WithContext(ctx).(*FS)
	return &newF
}

func (f *FS) Open(name string) (goFs.File, error) {
	if name == "." {
		name = ""
	}
	result, ok := f.cache.Load(lib.NormalizeForComparison(name))
	if ok {
		file := result.(*File)
		if file.IsDir() {
			return &ReadDirFile{File: file}, nil
		}
		return file, nil
	}
	path := lib.UrlJoinNoEscape(f.Root, name)
	var err error
	var fileInfo files_sdk.File
	if path == "" { // skip call on root path
		fileInfo = files_sdk.File{Type: "directory"}
	} else {
		fileInfo, err = (&Client{Config: f.Config}).Find(files_sdk.FileFindParams{Path: path}, files_sdk.WithContext(f.Context))
		if err != nil {
			return &File{}, &goFs.PathError{Path: path, Err: err, Op: "open"}
		}
	}

	file := (&File{File: &fileInfo, FS: f}).Init()
	if f.useCache {
		f.cache.Store(lib.NormalizeForComparison(path), file)
	}
	if fileInfo.Type == "directory" {
		return &ReadDirFile{File: file}, nil
	} else {
		return file, nil
	}
}

type DirEntryError struct {
	DirEntries []goFs.DirEntry
	error
}

func (f *FS) ReadDir(name string) ([]goFs.DirEntry, error) {
	if name == "." {
		name = ""
	}
	cacheName := lib.NormalizeForComparison(name)
	if f.useCache {
		f.cacheMutex.Lock(cacheName)
		defer f.cacheMutex.Unlock(cacheName)

		dirs, ok := f.cacheDir.Load(cacheName)
		if ok {
			dirEntryError := dirs.(DirEntryError)
			return dirEntryError.DirEntries, dirEntryError.error
		}
	}

	dirs, err := ReadDirFile{File: (&File{File: &files_sdk.File{Path: name}, FS: f}).Init()}.ReadDir(0)
	if f.useCache && errors.Is(err, files_sdk.ResponseError{}) {
		f.cacheDir.Store(cacheName, DirEntryError{dirs, err})
	}
	return dirs, err
}

func (f ReadDirFile) ReadDir(n int) ([]goFs.DirEntry, error) {
	var files []goFs.DirEntry
	if f.Context != nil && f.Context.Err() != nil {
		return files, &goFs.PathError{Path: f.Path, Err: f.Context.Err(), Op: "readdir"}
	}
	folderClient := folder.Client{Config: f.Config}
	it, err := folderClient.ListFor(files_sdk.FolderListForParams{Path: f.Path}, files_sdk.WithContext(f.Context))
	if err != nil {
		return files, &goFs.PathError{Path: f.Path, Err: err, Op: "readdir"}
	}
	if f.count > 0 {
		return files, io.EOF
	}
	for it.Next() && (n <= 0 || n > 0 && n >= f.count) {
		fi := it.File()
		if err != nil {
			return files, &goFs.PathError{Path: f.Path, Err: err, Op: "readdir"}
		}
		parts := strings.Split(fi.Path, "/")
		dir := strings.Join(parts[0:len(parts)-1], "/")
		if lib.NormalizeForComparison(dir) == lib.NormalizeForComparison(f.Path) {
			// There is a bug in the API that it could return a nested file not in the current directory.
			file := (&File{File: &fi, FS: f.FS}).Init()
			if f.useCache {
				f.cache.Store(lib.NormalizeForComparison(fi.Path), file)
			}
			files = append(files, file)
		}

		f.count += 1
	}

	if it.Err() != nil {
		return files, &goFs.PathError{Path: f.Path, Err: it.Err(), Op: "readdir"}
	}
	return files, nil
}

func (f *FS) MkdirAll(dir string, _ goFs.FileMode) error {
	var parentPath string
	for _, dirPath := range strings.Split(dir, "/") {
		if dirPath == "" {
			break
		}
		folderClient := folder.Client{Config: f.Config}
		_, err := folderClient.Create(files_sdk.FolderCreateParams{Path: lib.UrlJoinNoEscape(parentPath, dirPath)}, files_sdk.WithContext(f.Context))
		rErr, ok := err.(files_sdk.ResponseError)
		if err != nil && ok && rErr.Type != "processing-failure/destination-exists" {
			return err
		}

		parentPath = lib.UrlJoinNoEscape(parentPath, dirPath)
	}
	return nil
}

func (f *FS) PathSeparator() string {
	return "/"
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (f *FS) MkdirTemp(dir, pattern string) (string, error) {
	if dir == "" {
		dir = filepath.Join(f.TempDir(), randSeq(10))
	}
	path := f.PathJoin(dir, pattern)
	return path, f.MkdirAll(path, 0750)
}

type WritableFile struct {
	*Client
	*FS
	path string
	*bytes.Buffer
}

func (w WritableFile) init() WritableFile {
	w.Buffer = bytes.NewBuffer([]byte{})
	return w
}

func (w WritableFile) Write(p []byte) (int, error) {
	return w.Buffer.Write(p)
}

func (w WritableFile) Close() (err error) {
	return w.Client.Upload(
		UploadWithContext(w.Context),
		UploadWithReader(bytes.NewReader(w.Buffer.Bytes())),
		UploadWithDestinationPath(w.path),
		UploadWithSize(int64(w.Buffer.Len())),
	)
}

// Create Not for performant use cases.
func (f *FS) Create(path string) (io.WriteCloser, error) {
	return WritableFile{FS: f, Client: &Client{Config: f.Config}, path: path}.init(), nil
}

func (f *FS) RemoveAll(path string) error {
	return (&Client{Config: f.Config}).Delete(files_sdk.FileDeleteParams{Path: path, Recursive: lib.Bool(true)}, files_sdk.WithContext(f.Context))
}

func (f *FS) Remove(path string) error {
	return (&Client{Config: f.Config}).Delete(files_sdk.FileDeleteParams{Path: path}, files_sdk.WithContext(f.Context))
}

func (f *FS) PathJoin(s ...string) string {
	return lib.UrlJoinNoEscape(s...)
}

func (f *FS) RelPath(parent, child string) (string, error) {
	path := strings.Replace(child, parent, "", 1)
	if path == "" {
		return ".", nil
	}
	path = strings.TrimSuffix(path, f.PathSeparator())
	path = strings.TrimPrefix(path, f.PathSeparator())
	return path, nil
}

func (f *FS) SplitPath(path string) (string, string) {
	if path == "" {
		return "", ""
	}

	parts := strings.Split(path, f.PathSeparator())

	return f.PathJoin(parts[:int(math.Min(float64(len(parts)-2), float64(len(parts))))]...), parts[len(parts)-1]
}

func (f *FS) TempDir() string {
	return "tmp"
}

func (f *FS) Chtimes(name string, _atime time.Time, mtime time.Time) error {
	_, err := (&Client{Config: f.Config}).Update(files_sdk.FileUpdateParams{Path: name, ProvidedMtime: &mtime}, files_sdk.WithContext(f.Context))
	if err != nil {
		return &goFs.PathError{Path: name, Err: err, Op: "chtimes"}
	}
	return nil
}
