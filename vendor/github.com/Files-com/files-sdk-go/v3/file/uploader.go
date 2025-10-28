package file

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/directory"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/ignore"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
)

type Uploader interface {
	UploadWithResume(...UploadOption) (UploadResumable, error)
	Find(files_sdk.FileFindParams, ...files_sdk.RequestResponseOption) (files_sdk.File, error)
	CreateFolder(files_sdk.FolderCreateParams, ...files_sdk.RequestResponseOption) (files_sdk.File, error)
}

func uploader(parentCtx context.Context, c Uploader, params UploaderParams) *Job {
	var job *Job
	if params.Job == nil {
		job = (&Job{}).Init()
	} else {
		job = params.Job
	}
	SetJobParams(job, direction.UploadType, params, params.config.Logger, (&FS{}).Init(params.config, true))
	job.Config = params.config
	jobCtx := job.WithContext(parentCtx)

	fileInfoLocalPath, statErr := os.Stat(params.LocalPath)

	if statErr == nil && fileInfoLocalPath.IsDir() {
		job.Type = directory.Dir
	} else {
		job.Type = directory.File
	}

	if len(params.LocalPaths) > 0 {
		job.Type = directory.Files
	}
	job.Client = c
	onComplete := make(chan *UploadStatus)
	job.CodeStart = func() {
		job.Scan()

		go enqueueIndexedUploads(job, jobCtx, onComplete)
		WaitTellFinished(job, onComplete, func() { RetryByPolicy(jobCtx, job, job.RetryPolicy.(RetryPolicy), false) })

		metaFile := UploadStatus{
			job:         job,
			status:      status.Errored,
			localPath:   params.LocalPath,
			Sync:        params.Sync,
			NoOverwrite: params.NoOverwrite,
			Uploader:    c,
			Mutex:       &sync.RWMutex{},
		}
		if errorJob(job, metaFile, statErr) {
			return
		}
		var err error
		if job.Ignore, err = ignore.New(params.Ignore...); errorJob(job, metaFile, err) {
			return
		}

		if len(params.Include) > 0 {
			if job.Include, err = ignore.New(params.Include...); errorJob(job, metaFile, err) {
				return
			}
		}
		// Move everything into the path loop
		if job.Type == directory.File {
			params.LocalPaths = []string{params.LocalPath}
		} else if job.Type == directory.Dir {
			params.LocalPaths = []string{params.LocalPath}
		}

		for _, path := range params.LocalPaths {
			var fi os.FileInfo
			var err error
			var isDir bool
			statusFile := metaFile
			statusFile.localPath = path
			statusFile.status = status.Indexed

			// Optimization: Don't Stat if we know it's a directory already
			if strings.HasSuffix(path, string(os.PathSeparator)) {
				isDir = true
			} else {
				// Don't call os.Stat again
				if path == params.LocalPath {
					fi = fileInfoLocalPath
					err = nil
				} else {
					// Lazy call Stat but also make available for if it's a file.
					fi, err = os.Stat(path)
				}
				// Fallback to checking stat if heuristic fails
				if err == nil && fi.IsDir() {
					isDir = true
				}
			}

			if isDir {
				it := processDirectory(path, params, job, jobCtx, c)
				if it.Err() != nil {
					statusFile.error = it.Err()
					job.Add(&statusFile)
				}
			} else if err != nil {
				statusFile.error = err
				job.Add(&statusFile)
			} else {
				statusFile.remotePath = params.RemotePath
				statusFile.file = files_sdk.File{
					DisplayName: filepath.Base(path),
					Type:        "file",
					Mtime:       lib.Time(fi.ModTime()),
					Size:        fi.Size(),
				}
				job.Add(&statusFile)
			}
		}
		job.EndScan()
	}

	return job
}

func errorJob(job *Job, metaFile UploadStatus, err error) bool {
	if err != nil {
		job.Add(&metaFile)
		job.UpdateStatus(status.Errored, &metaFile, err)
		job.EndScan()
		job.Finish()
		return true
	}
	return false
}

func processDirectory(localPath string, params UploaderParams, job *Job, jobCtx context.Context, c Uploader) *lib.IterChan[lib.DirEntry] {
	root := ""
	remotePath := params.RemotePath

	if !strings.HasSuffix(localPath, string(os.PathSeparator)) {
		root = "."
		_, lastDir := filepath.Split(localPath)
		remotePath = filepath.Join(remotePath, lastDir)
	}

	it := (&lib.Walk[lib.DirEntry]{
		FS:                 os.DirFS(localPath),
		ConcurrencyManager: job.Manager.DirectoryListingManager,
		WalkFile:           lib.DirEntryWalkFile,
		ListDirectories:    true,
		Root:               root,
	}).Walk(jobCtx)

	for it.Next() {
		uploadStatus, ok := buildUploadStatus(filepath.Join(localPath, it.Resource().Path()), localPath, remotePath, c, job, params)
		if !ok {
			continue
		}

		if it.Resource().Err() != nil {
			uploadStatus.missingStat = true
			uploadStatus.error = it.Resource().Err()
		} else {
			if it.Resource().DirEntry.IsDir() {
				uploadStatus.file.Type = "directory"
			} else {
				uploadStatus.file.Type = "file"
				uploadStatus.file.Size = it.Resource().FileInfo.Size()
				uploadStatus.file.Mtime = lib.Time(it.Resource().FileInfo.ModTime())
			}
		}
		job.Add(&uploadStatus)
	}
	return it
}

func remotePath(ctx context.Context, localPath, remotePath string, c Uploader, job *Job) (string, error) {
	destination := remotePath
	_, localFileName := filepath.Split(localPath)
	if remotePath == "" {
		destination = localFileName
	} else {
		var err error
		var remoteFile files_sdk.File
		if job.FilePartsManager.WaitWithContext(ctx) {
			remoteFile, err = c.Find(files_sdk.FileFindParams{Path: lib.NewUrlPath(remotePath).PruneEndingSlash().String()}, files_sdk.WithContext(ctx))
			job.FilePartsManager.Done()
		} else {
			return "", ctx.Err()
		}
		var responseError files_sdk.ResponseError
		ok := errors.As(err, &responseError)
		if remoteFile.Type == "directory" {
			destination = lib.UrlJoinNoEscape(remotePath, localFileName)
		} else if ok && responseError.Type == "not-found" {
			if destination[len(destination)-1:] == "/" {
				destination = lib.UrlJoinNoEscape(remotePath, localFileName)
			}
		} else if err != nil {
			return "", err
		}
	}
	return destination, nil
}

func enqueueIndexedUploads(job *Job, jobCtx context.Context, onComplete chan *UploadStatus) {
	for !job.EndScanning.Called() || job.Count(status.Indexed) > 0 {
		if f, ok := job.EnqueueNext(); ok {
			if job.FilesManager.WaitWithContext(jobCtx) {
				go enqueueUpload(jobCtx, job, f.(*UploadStatus), onComplete)
			} else {
				job.UpdateStatus(status.Canceled, f.(*UploadStatus), nil)
				onComplete <- f.(*UploadStatus)
			}
		}
	}
}

func enqueueUpload(ctx context.Context, job *Job, uploadStatus *UploadStatus, onComplete chan *UploadStatus) {
	finish := func() {
		job.FilesManager.Done()
		onComplete <- uploadStatus
	}
	if uploadStatus.error != nil || uploadStatus.missingStat {
		job.UpdateStatus(status.Errored, uploadStatus, uploadStatus.RecentError())
		finish()
		return
	}
	func() {
		var localFile *os.File
		var err error
		defer func() {
			if localFile != nil {
				localFile.Close()
			}
			finish()
		}()
		if uploadStatus.file.Path == "" {
			uploadStatus.file.Path, err = remotePath(ctx, uploadStatus.LocalPath(), uploadStatus.RemotePath(), uploadStatus, job)
			if err != nil {
				uploadStatus.Job().UpdateStatus(status.Errored, uploadStatus, err)
				return
			}
			uploadStatus.remotePath = uploadStatus.file.Path
		}
		if excludeFile(uploadStatus, job.Config.FeatureFlag("incremental-updates")) {
			return
		}
		if uploadStatus.dryRun {
			uploadStatus.Job().UpdateStatus(status.Complete, uploadStatus, nil)
			return
		}
		if uploadStatus.File().IsDir() {
			_, err = uploadStatus.CreateFolder(files_sdk.FolderCreateParams{Path: uploadStatus.RemotePath(), MkdirParents: lib.Bool(true)}, files_sdk.WithContext(ctx))
			if err == nil {
				uploadStatus.Job().UpdateStatus(status.FolderCreated, uploadStatus, nil)
				return
			}

			if files_sdk.IsExist(err) {
				remoteFile, err := uploadStatus.Find(files_sdk.FileFindParams{Path: uploadStatus.RemotePath()}, files_sdk.WithContext(ctx))
				if err == nil && remoteFile.IsDir() {
					uploadStatus.Job().UpdateStatus(status.FolderCreated, uploadStatus, nil)
					return
				}
			}
			uploadStatus.Job().UpdateStatus(status.Errored, uploadStatus, err)
			return
		}
		localFile, err = os.Open(uploadStatus.LocalPath())
		if err != nil {
			uploadStatus.Job().UpdateStatus(status.Errored, uploadStatus, err)
			return
		}
		opts := []UploadOption{
			UploadWithContext(ctx),
			UploadWithManager(job.FilePartsManager),
			UploadWithReaderAt(localFile),
			UploadWithSize(uploadStatus.File().Size),
			UploadWithResume(uploadStatus.UploadResumable),
			UploadWithProgress(uploadProgress(uploadStatus)),
			UploadWithDestinationPath(uploadStatus.RemotePath()),
		}

		if params, ok := job.Params.(UploaderParams); ok && params.PreserveTimes {
			opts = append(opts, UploadWithProvidedMtime(*uploadStatus.File().Mtime))
		}

		uploadStatus.UploadResumable, err = uploadStatus.UploadWithResume(opts...)
		if err != nil {
			uploadStatus.Job().StatusFromError(uploadStatus, err)
		} else {
			uploadStatus.SetUploadedBytes(uploadStatus.Parts.SuccessfulBytes())
			if localFile != nil {
				localFile.Close()
			}
			uploadStatus.Job().UpdateStatus(status.Complete, uploadStatus, nil)
		}
	}()
}

func buildUploadStatus(path string, localFolderPath string, destinationRootPath string, c Uploader, job *Job, params UploaderParams) (UploadStatus, bool) {
	dir, filename := filepath.Split(path)

	if localFolderPath == dir && filename == "" {
		return UploadStatus{}, false
	}

	destination := buildDestination(path, localFolderPath, destinationRootPath, filename)
	uploadStatus := UploadStatus{
		Uploader:    c,
		job:         job,
		remotePath:  destination,
		localPath:   path,
		Sync:        params.Sync,
		status:      status.Indexed,
		Mutex:       &sync.RWMutex{},
		file:        files_sdk.File{Path: destination, DisplayName: filepath.Base(destination)},
		dryRun:      params.DryRun,
		NoOverwrite: params.NoOverwrite,
	}
	return uploadStatus, true
}

func buildDestination(path string, localFolderPath string, destinationRootPath string, filename string) string {
	var destination string
	var baseDestination string
	if localFolderPath != "." {
		baseDestination = strings.TrimPrefix(path, localFolderPath)
	} else if path != "." {
		baseDestination = path
	}
	baseDestination = strings.TrimLeft(baseDestination, string(os.PathSeparator))
	baseDestination = strings.TrimPrefix(baseDestination, string(os.PathSeparator))
	if destinationRootPath == "" {
		destination = baseDestination
	} else {
		destination = filepath.Join(destinationRootPath, baseDestination)
	}

	if destination == "." || destination == "" {
		destination = filename
	}
	return lib.Path{Path: destination}.NormalizePathSystemForAPI().String()
}

func excludeFile(uploadStatus *UploadStatus, incrementalUpdates bool) bool {
	if uploadStatus.Job().Ignore.MatchesPath(uploadStatus.LocalPath()) {
		uploadStatus.Job().UpdateStatus(status.Ignored, uploadStatus, nil)
		return true
	}

	if uploadStatus.Job().Include != nil && !uploadStatus.Job().Include.MatchesPath(uploadStatus.LocalPath()) {
		uploadStatus.Job().UpdateStatus(status.Ignored, uploadStatus, nil)
		return true
	}

	if uploadStatus.NoOverwrite {
		_, found, err := uploadStatus.Job().FindRemoteFile(uploadStatus)
		if found {
			uploadStatus.Job().UpdateStatus(status.FileExists, uploadStatus, err)
			return true
		}
		return false
	}

	if uploadStatus.Sync {
		file, found, err := uploadStatus.Job().FindRemoteFile(uploadStatus)
		var responseError files_sdk.ResponseError
		ok := errors.As(err, &responseError)
		if !found || (ok && responseError.Type == "not-found") {
			uploadStatus.Job().UpdateStatus(status.Compared, uploadStatus, nil)
			uploadStatus.Job().Logger.Printf("sync %v not found on destination", uploadStatus.RemotePath())
			return false
		}
		// local is not after server
		if uploadStatus.File().Size == file.Size {
			// Server version is the same or newer
			uploadStatus.Job().UpdateStatus(status.Skipped, uploadStatus, nil)
			uploadStatus.Job().Logger.Printf("sync %v size match", uploadStatus.RemotePath())
			return true
		}
		if incrementalUpdates && !file.ModTime().IsZero() {
			if file.ModTime().After(*uploadStatus.File().Mtime) {
				uploadStatus.Job().UpdateStatus(status.Skipped, uploadStatus, nil)
				uploadStatus.Job().Logger.Printf("sync incremental-updates %v server has a newer version", uploadStatus.RemotePath())
				return true
			}
			if file.ModTime().Truncate(time.Minute).Equal(uploadStatus.File().Mtime.Truncate(time.Minute)) {
				uploadStatus.Job().UpdateStatus(status.Skipped, uploadStatus, nil)
				uploadStatus.Job().Logger.Printf("sync incremental-updates %v both times are within the same minute", uploadStatus.RemotePath())
				return true
			}
		}
		uploadStatus.Job().UpdateStatus(status.Compared, uploadStatus, nil)
		uploadStatus.Job().Logger.Printf("sync %v found on destination with non matching sizes: local: %v, remote: %v", uploadStatus.RemotePath(), uploadStatus.File().Size, file.Size)
	}
	return false
}
