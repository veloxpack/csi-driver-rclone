package file

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/directory"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/ignore"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
	"github.com/Files-com/files-sdk-go/v3/lib/keyvalue"
	gitignore "github.com/sabhiram/go-gitignore"
)

func downloader(ctx context.Context, fileSys fs.FS, params DownloaderParams) *Job {
	job := (&Job{}).Init()
	SetJobParams(job, direction.DownloadType, params, params.config.Logger, fileSys)
	job.Config = params.config
	jobCtx := job.WithContext(ctx)
	remoteFs, ok := fileSys.(lib.FSWithContext)
	if ok {
		fileSys = remoteFs.WithContext(jobCtx)
	}
	if params.RemoteFile.Path != "" {
		job.LocalPath = params.LocalPath
		job.RemotePath = params.RemoteFile.Path

		if params.RemoteFile.Type == "directory" {
			job.Type = directory.Dir
		} else {
			job.Type = directory.File
		}
	} else {
		job.LocalPath = params.LocalPath
		job.RemotePath = lib.NewUrlPath(params.RemotePath).PruneStartingSlash().String()
		if job.RemotePath == "" {
			job.RemotePath = "."
		}
		var remoteType directory.Type
		remoteFile, err := fileSys.Open(params.RemotePath)
		remoteType = directory.Dir // default to Dir not found error will have to be dealt with downstream
		if err == nil {
			remoteStat, err := remoteFile.Stat()
			if err == nil {
				if remoteStat.IsDir() {
					remoteType = directory.Dir
				} else {
					remoteType = directory.File
				}
			}
		}
		job.LocalPath = lib.ExpandTilde(job.LocalPath)
		var localType directory.Type
		stats, err := os.Stat(job.LocalPath)
		if os.IsNotExist(err) {
			if (lib.Path{Path: job.LocalPath}).EndingSlash() { // explicit directory
				localType = directory.Dir
			} else if remoteType == directory.File {
				localType = directory.File
			} else {
				localType = directory.Dir // implicit directory
			}
		} else if err == nil {
			if stats.IsDir() {
				localType = directory.Dir
			} else {
				localType = directory.File
			}
		} else {
			// Propagating this error is difficult, but this error will happen again in CodeStart.
		}
		if (!lib.NewUrlPath(params.RemotePath).EndingSlash() && localType == directory.Dir) || remoteType == directory.File && localType == directory.Dir {
			job.LocalPath = filepath.Join(job.LocalPath, lib.NewUrlPath(job.RemotePath).SwitchPathSeparator(string(os.PathSeparator)).Pop())
			if remoteType == directory.File {
				localType = directory.File
			}
		}

		// Use relative path
		if job.LocalPath == "" {
			job.LocalPath = lib.NewUrlPath(job.RemotePath).SwitchPathSeparator(string(os.PathSeparator)).Pop()
		}

		job.Type = localType
		job.Logger.Printf(keyvalue.New(map[string]interface{}{
			"LocalPath":  job.LocalPath,
			"RemotePath": job.RemotePath,
		}))
	}
	onComplete := make(chan *DownloadStatus)
	job.CodeStart = func() {
		job.Scan()
		go enqueueIndexedDownloads(job, jobCtx, onComplete)
		WaitTellFinished(job, onComplete, func() { RetryByPolicy(jobCtx, job, job.RetryPolicy.(RetryPolicy), false) })

		// ignore.New only returns an error if run on an unsupported OS
		job.Ignore, _ = ignore.New(params.Ignore...)
		if len(params.Include) > 0 {
			job.Include, _ = ignore.New(params.Include...)
		}

		it := (&lib.Walk[lib.DirEntry]{
			FS:                 fileSys,
			Root:               lib.UrlJoinNoEscape(job.RemotePath),
			ConcurrencyManager: job.Manager.FilePartsManager,
			WalkFile:           lib.DirEntryWalkFile,
			ListDirectories:    true,
		}).Walk(jobCtx)

		for it.Next() {
			if it.Resource().Err() != nil {
				createIndexedStatus(Entity{error: it.Resource().Err()}, params, job)
			} else {
				f, err := fileSys.Open(it.Resource().Path())
				createIndexedStatus(Entity{error: err, File: f, FS: fileSys}, params, job)
			}
		}

		if it.Err() != nil {
			metaFile := &DownloadStatus{
				job:         job,
				status:      status.Errored,
				localPath:   params.LocalPath,
				remotePath:  params.RemotePath,
				tempPath:    params.TempPath,
				Sync:        params.Sync,
				NoOverwrite: params.NoOverwrite,
				Mutex:       &sync.RWMutex{},
			}
			metaFile.file = files_sdk.File{
				DisplayName: filepath.Base(params.LocalPath),
				Type:        job.Direction.Name(),
				Path:        params.RemotePath,
			}
			job.Add(metaFile)
			job.UpdateStatus(status.Errored, metaFile, it.Err())
			onComplete <- metaFile
		}

		job.EndScan()
	}

	return job
}

func enqueueIndexedDownloads(job *Job, jobCtx context.Context, onComplete chan *DownloadStatus) {
	for !job.EndScanning.Called() || job.Count(status.Indexed) > 0 {
		if f, ok := job.EnqueueNext(); ok {
			if job.FilesManager.WaitWithContext(jobCtx) {
				go enqueueDownload(jobCtx, job, f.(*DownloadStatus), onComplete)
			} else {
				job.UpdateStatus(status.Canceled, f.(*DownloadStatus), nil)
				onComplete <- f.(*DownloadStatus)
			}
		}
	}
}

func normalizePath(rootDestination string) string {
	if rootDestination != "" && rootDestination[len(rootDestination)-1:] == string(os.PathSeparator) {
	} else {
		rootDestination, _ = filepath.Abs(rootDestination)
	}
	return rootDestination
}

func createIndexedStatus(f Entity, params DownloaderParams, job *Job) {
	s := &DownloadStatus{
		error:         f.error,
		fsFile:        f.File,
		FS:            f.FS,
		job:           job,
		localPath:     params.LocalPath,
		remotePath:    params.RemotePath,
		tempPath:      params.TempPath,
		Sync:          params.Sync,
		NoOverwrite:   params.NoOverwrite,
		status:        status.Indexed,
		Mutex:         &sync.RWMutex{},
		PreserveTimes: params.PreserveTimes,
		dryRun:        params.DryRun,
	}
	var err error
	if f.error == nil {
		s.FileInfo, err = f.File.Stat()
		if err == nil {
			s.file = s.FileInfo.Sys().(files_sdk.File)
			s.localPath = localPath(s.file, *job)
			s.remotePath = s.file.Path
		} else {
			s.SetStatus(status.Errored, err)
		}
	}

	job.Add(s)
}

func enqueueDownload(ctx context.Context, job *Job, downloadStatus *DownloadStatus, signal chan *DownloadStatus) {
	if downloadStatus.error != nil || downloadStatus.fsFile == nil {
		job.UpdateStatus(status.Errored, downloadStatus, downloadStatus.RecentError())
		job.FilesManager.Done()
		signal <- downloadStatus
		return
	}
	if ignoreDownloadJob(job, downloadStatus) {
		job.UpdateStatus(status.Ignored, downloadStatus, nil)
		job.FilesManager.Done()
		signal <- downloadStatus
		return
	}

	downloadFolderItem(ctx, signal, downloadStatus)
}

func ignoreDownloadJob(job *Job, downloadStatus *DownloadStatus) bool {
	return ignorePath(downloadStatus.RemotePath(), job.Ignore, job.Include)
}

func ignorePath(path string, ignored, included *gitignore.GitIgnore) bool {
	// if the ignore matches, or the include doesn't match, we skip the file
	if (ignored != nil && ignored.MatchesPath(path)) || (included != nil && !included.MatchesPath(path)) {
		return true
	}
	return false
}

func downloadFolderItem(ctx context.Context, signal chan *DownloadStatus, s *DownloadStatus) {
	func(ctx context.Context, reportStatus *DownloadStatus) {
		defer func() {
			s.job.FilesManager.Done()
			signal <- reportStatus
		}()
		dir, _ := filepath.Split(reportStatus.LocalPath())
		if dir != "" {
			_, err := os.Stat(dir)
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
					return
				}
			}
		}

		remoteStat, remoteStatErr := reportStatus.fsFile.Stat()
		if remoteStatErr != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, remoteStatErr)
			return
		}

		if reportStatus.NoOverwrite {
			_, localStatErr := os.Stat(reportStatus.LocalPath())
			if localStatErr == nil {
				reportStatus.Job().UpdateStatus(status.FileExists, reportStatus, localStatErr)
				return
			}
			if !os.IsNotExist(localStatErr) {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, localStatErr)
				return
			}
		}

		if reportStatus.Job().Sync {
			localStat, localStatErr := os.Stat(reportStatus.LocalPath())
			if localStatErr != nil && !os.IsNotExist(localStatErr) {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, localStatErr)
				return
			}
			// server is not after local
			if !os.IsNotExist(localStatErr) && remoteStat.Size() == localStat.Size() {
				// Local version is the same or newer
				reportStatus.Job().UpdateStatus(status.Skipped, reportStatus, nil)
				return
			}
			reportStatus.Job().UpdateStatus(status.Compared, reportStatus, nil)
		}

		if reportStatus.dryRun {
			reportStatus.Job().UpdateStatus(status.Complete, reportStatus, nil)
			return
		}

		if reportStatus.File().IsDir() {
			err := os.MkdirAll(reportStatus.LocalPath(), 0755)
			if err != nil {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			} else {
				reportStatus.Job().UpdateStatus(status.FolderCreated, reportStatus, nil)
			}
			return
		}

		tmpName, err := tmpDownloadPath(reportStatus.LocalPath(), reportStatus.tempPath)
		if err != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			return
		}
		reportStatus.Job().Config.LogPath(
			reportStatus.RemotePath(),
			map[string]interface{}{
				"LocalTempPath": tmpName,
			},
		)
		writer := openFile(tmpName, reportStatus)
		downloadParts := (&DownloadParts{}).Init(
			reportStatus.fsFile,
			remoteStat,
			reportStatus.Job().Manager.FilePartsManager,
			writer,
			reportStatus.Job().Config,
		)

		lib.AnyError(func(err error) {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
		},
			func() error { return downloadParts.Run(ctx) },
			func() error { return downloadParts.CloseError },
		)

		if reportStatus.Status().Is(status.Valid...) {
			reportStatus.SetFinalSize(downloadParts.FinalSize())
			reportStatus.Job().Config.LogPath(
				reportStatus.RemotePath(),
				map[string]interface{}{
					"LocalTempPath": tmpName,
					"FinalSize":     downloadParts.FinalSize(),
				},
			)
			err := finalizeTmpDownload(tmpName, reportStatus.LocalPath())
			if err != nil {
				removeTmpDownload(tmpName)
			}

			if err == nil && reportStatus.PreserveTimes {
				var t time.Time
				if s.file.ProvidedMtime != nil {
					t = *s.file.ProvidedMtime
				} else if s.file.Mtime != nil {
					t = *s.file.Mtime
				}
				if !t.IsZero() {
					err = os.Chtimes(reportStatus.LocalPath(), t.Local(), t.Local())
				}
			}

			if err != nil {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			} else if reportStatus.Status().Is(status.Downloading) {
				reportStatus.Job().UpdateStatus(status.Complete, reportStatus, nil)
			}
		} else {
			err := removeTmpDownload(tmpName) // Clean up on invalid download
			if err != nil {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			}
		}
	}(ctx, s)
}

func openFile(partName string, reportStatus *DownloadStatus) lib.ProgressWriter {
	out, createErr := os.Create(partName)
	if createErr != nil {
		reportStatus.Job().UpdateStatus(status.Errored, reportStatus, createErr)
	}
	writer := lib.ProgressWriter{WriterAndAt: out}
	writer.ProgressWatcher = func(incDownloadedBytes int64) {
		reportStatus.Job().UpdateStatusWithBytes(status.Downloading, reportStatus, incDownloadedBytes)
	}
	return writer
}

func localPath(file files_sdk.File, job Job) string {
	var path string
	if job.Type == directory.File {
		path = job.LocalPath
	} else {
		path = filepath.Join(normalizePath(job.LocalPath), relativePath(job, file))
	}

	return path
}

func relativePath(job Job, file files_sdk.File) string {
	relativePath, err := filepath.Rel(job.RemotePath, file.Path)
	if err != nil {
		panic(err)
	}
	return relativePath
}
