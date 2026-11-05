package file

import (
	"context"
	"errors"
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

func downloader(ctx context.Context, fileSys fs.FS, params DownloaderParams, opts ...files_sdk.RequestResponseOption) *Job {
	job := (&Job{}).Init()
	SetJobParams(job, direction.DownloadType, params, params.config.Logger, fileSys)
	job.Config = params.config
	if params.PriorJobCheckpoint != nil {
		job.CompletedPaths = make(map[string]struct{}, len(params.PriorJobCheckpoint.CompletedPaths))
		for _, p := range params.PriorJobCheckpoint.CompletedPaths {
			job.CompletedPaths[p] = struct{}{}
		}
	}
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
		WaitTellFinished(job, onComplete, func() {
			RetryByPolicy(jobCtx, job, job.RetryPolicy.(RetryPolicy), false)
			runSyncAfterEmptyFolders(job, params.SyncAfterActions, params.DryRun, params.config, opts...)
		})

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
	params, _ := job.Params.(DownloaderParams)
	if zipBatchDisabled(params, job) {
		enqueueIndexedDownloadsDirect(job, jobCtx, onComplete)
		return
	}

	batcher := newZipBatchDownloader(job, jobCtx, params, onComplete)
	for !job.EndScanning.Called() || job.Count(status.Indexed) > 0 {
		if f, ok := job.EnqueueNext(); ok {
			downloadStatus := f.(*DownloadStatus)
			if batcher.offer(downloadStatus) {
				continue
			}
			enqueueIndexedDownloadDirect(job, jobCtx, downloadStatus, onComplete)
		}
	}
	batcher.flushEnd()
}

func enqueueIndexedDownloadsDirect(job *Job, jobCtx context.Context, onComplete chan *DownloadStatus) {
	for !job.EndScanning.Called() || job.Count(status.Indexed) > 0 {
		if f, ok := job.EnqueueNext(); ok {
			enqueueIndexedDownloadDirect(job, jobCtx, f.(*DownloadStatus), onComplete)
		}
	}
}

func enqueueIndexedDownloadDirect(job *Job, jobCtx context.Context, downloadStatus *DownloadStatus, onComplete chan *DownloadStatus) {
	params, _ := job.Params.(DownloaderParams)
	adaptiveAdmission := params.AdaptiveConcurrency && (params.AdaptiveConcurrencyUseSDKDefaultCaps || params.Manager == nil)
	adaptiveTargets := downloadV2JobAdmissionTargets{job: job}
	fileAdmissionManager := job.fileAdmissionManager()
	gateDownload := shouldGateAdaptiveDownloadAdmission(downloadStatus, adaptiveAdmission)
	admitted := !gateDownload ||
		waitForAdaptiveFileAdmission(jobCtx, fileAdmissionManager, adaptiveTargets, adaptiveFileAdmissionInitialTarget())
	if admitted && fileAdmissionManager.WaitWithContext(jobCtx) {
		go enqueueDownload(jobCtx, job, downloadStatus, onComplete)
	} else {
		job.UpdateStatus(status.Canceled, downloadStatus, nil)
		onComplete <- downloadStatus
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
			s.TmpPath = params.ResumeTmpPath
		} else {
			s.SetStatus(status.Errored, err)
		}
	}

	job.Add(s)
}

func enqueueDownload(ctx context.Context, job *Job, downloadStatus *DownloadStatus, signal chan *DownloadStatus) {
	if downloadStatus.error != nil || downloadStatus.fsFile == nil {
		job.UpdateStatus(status.Errored, downloadStatus, downloadStatus.RecentError())
		job.fileAdmissionManager().Done()
		signal <- downloadStatus
		return
	}
	if _, ok := job.CompletedPaths[downloadStatus.LocalPath()]; ok {
		job.UpdateStatus(status.Skipped, downloadStatus, nil)
		job.fileAdmissionManager().Done()
		signal <- downloadStatus
		return
	}
	if ignoreDownloadJob(job, downloadStatus) {
		job.UpdateStatus(status.Ignored, downloadStatus, nil)
		job.fileAdmissionManager().Done()
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
	defer func() {
		s.job.fileAdmissionManager().Done()
		signal <- s
	}()
	runDownloadFolderItem(ctx, s)
}

func runDownloadFolderItem(ctx context.Context, reportStatus *DownloadStatus) {
	remoteStat, ok := prepareDownloadFolderItem(reportStatus)
	if !ok {
		return
	}

	var startOffset int64
	tmpName := reportStatus.TmpPath
	if tmpName != "" {
		if _, err := os.Stat(tmpName); os.IsNotExist(err) {
			reportStatus.Job().Logger.Printf("tmp download file not found, starting over: %v", tmpName)
			tmpName = ""
			reportStatus.TmpPath = ""
		}
	}
	if tmpName == "" {
		tmpName = existingTmpDownloadPath(reportStatus.LocalPath(), reportStatus.tempPath)
	}
	if tmpName == "" {
		var err error
		tmpName, err = tmpDownloadPath(reportStatus.LocalPath(), reportStatus.tempPath)
		if err != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			return
		}
	} else {
		fi, err := os.Stat(tmpName)
		if err == nil {
			startOffset = fi.Size()
		}
		if startOffset > remoteStat.Size() {
			startOffset = 0
		}
		if startOffset == remoteStat.Size() {
			// Temp file is already fully downloaded - skip the network request and finalize directly.
			if err := finalizeTmpDownload(tmpName, reportStatus.LocalPath()); err != nil {
				removeTmpDownload(tmpName)
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			} else {
				reportStatus.SetFinalSize(startOffset)
				reportStatus.Job().UpdateStatus(status.Complete, reportStatus, nil)
			}
			return
		}
	}
	reportStatus.TmpPath = tmpName
	if startOffset > 0 {
		reportStatus.IncrementTransferBytes(startOffset)
	}
	var finalSize int64
	downloadV2Used, downloadV2FinalSize, downloadV2Err := runDownloadV2IfSupported(ctx, reportStatus, remoteStat, tmpName, startOffset)
	if downloadV2Used {
		finalSize = downloadV2FinalSize
		if downloadV2Err != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, downloadV2Err)
		}
	} else {
		writer := openFile(tmpName, reportStatus, startOffset)
		downloadParts := (&DownloadParts{}).Init(
			reportStatus.fsFile,
			remoteStat,
			reportStatus.Job().Manager.FilePartsManager,
			writer,
			reportStatus.Job().Config,
			startOffset,
		)

		lib.AnyError(func(err error) {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
		},
			func() error { return downloadParts.Run(ctx) },
			func() error { return downloadParts.CloseError },
		)
		finalSize = downloadParts.FinalSize()
	}

	cause := context.Cause(ctx)

	if reportStatus.Status().Is(status.Valid...) {
		err := completeTmpDownload(reportStatus, tmpName, finalSize)
		if err != nil {
			removeTmpDownload(tmpName)
		}

		if err != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
		} else if reportStatus.Status().Is(status.Downloading) {
			reportStatus.Job().UpdateStatus(status.Complete, reportStatus, nil)
		}
	} else {
		if !errors.Is(cause, ErrJobPaused) {
			err := removeTmpDownload(tmpName) // Clean up on invalid download
			if err != nil {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
			}
		}
	}
}

func prepareDownloadFolderItem(reportStatus *DownloadStatus) (fs.FileInfo, bool) {
	dir, _ := filepath.Split(reportStatus.LocalPath())
	if dir != "" {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
				return nil, false
			}
		}
	}

	remoteStat, remoteStatErr := reportStatus.fsFile.Stat()
	if remoteStatErr != nil {
		reportStatus.Job().UpdateStatus(status.Errored, reportStatus, remoteStatErr)
		return nil, false
	}

	if reportStatus.NoOverwrite {
		_, localStatErr := os.Stat(reportStatus.LocalPath())
		if localStatErr == nil {
			reportStatus.Job().UpdateStatus(status.FileExists, reportStatus, localStatErr)
			return nil, false
		}
		if !os.IsNotExist(localStatErr) {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, localStatErr)
			return nil, false
		}
	}

	if reportStatus.Job().Sync {
		localStat, localStatErr := os.Stat(reportStatus.LocalPath())
		if localStatErr != nil && !os.IsNotExist(localStatErr) {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, localStatErr)
			return nil, false
		}
		// server is not after local
		if !os.IsNotExist(localStatErr) && remoteStat.Size() == localStat.Size() {
			// Local version is the same or newer
			reportStatus.Job().UpdateStatus(status.Skipped, reportStatus, nil)
			return nil, false
		}
		reportStatus.Job().UpdateStatus(status.Compared, reportStatus, nil)
	}

	if reportStatus.dryRun {
		reportStatus.Job().UpdateStatus(status.Complete, reportStatus, nil)
		return nil, false
	}

	if reportStatus.File().IsDir() {
		err := os.MkdirAll(reportStatus.LocalPath(), 0755)
		if err != nil {
			reportStatus.Job().UpdateStatus(status.Errored, reportStatus, err)
		} else {
			reportStatus.Job().UpdateStatus(status.FolderCreated, reportStatus, nil)
		}
		return nil, false
	}

	return remoteStat, true
}

func completeTmpDownload(reportStatus *DownloadStatus, tmpName string, finalSize int64) error {
	reportStatus.SetFinalSize(finalSize)
	if err := finalizeTmpDownload(tmpName, reportStatus.LocalPath()); err != nil {
		return err
	}

	if reportStatus.PreserveTimes {
		var t time.Time
		if reportStatus.file.ProvidedMtime != nil {
			t = *reportStatus.file.ProvidedMtime
		} else if reportStatus.file.Mtime != nil {
			t = *reportStatus.file.Mtime
		}
		if !t.IsZero() {
			return os.Chtimes(reportStatus.LocalPath(), t.Local(), t.Local())
		}
	}

	return nil
}

func openFile(partName string, reportStatus *DownloadStatus, startOffset int64) lib.ProgressWriter {
	var out *os.File
	var createErr error
	if startOffset > 0 {
		out, createErr = os.OpenFile(partName, os.O_WRONLY|os.O_CREATE, 0644)
	} else {
		out, createErr = os.Create(partName)
	}
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
