package file

import (
	"io"
	"io/fs"
	"os"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/manager"
)

func (c *Client) DownloadRetry(job Job, opts ...files_sdk.RequestResponseOption) *Job {
	newJob := job.ClearStatuses()
	return c.Downloader(
		DownloaderParams{
			RemotePath:                           newJob.RemotePath,
			Sync:                                 newJob.Sync,
			Manager:                              newJob.Manager,
			LocalPath:                            newJob.LocalPath,
			RetryPolicy:                          newJob.RetryPolicy.(RetryPolicy),
			EventsReporter:                       newJob.EventsReporter,
			Ignore:                               newJob.Params.(DownloaderParams).Ignore,
			Include:                              newJob.Params.(DownloaderParams).Include,
			SyncAfterActions:                     newJob.Params.(DownloaderParams).SyncAfterActions,
			AdaptiveConcurrency:                  newJob.Params.(DownloaderParams).AdaptiveConcurrency,
			AdaptiveConcurrencyUseSDKDefaultCaps: newJob.Params.(DownloaderParams).AdaptiveConcurrencyUseSDKDefaultCaps,
			AdaptiveDownloadV2TargetClassifier:   newJob.Params.(DownloaderParams).AdaptiveDownloadV2TargetClassifier,
			AdaptiveDownloadV2TuningSet:          newJob.Params.(DownloaderParams).AdaptiveDownloadV2TuningSet,
			AdaptiveDownloadV2Tuning:             newJob.Params.(DownloaderParams).AdaptiveDownloadV2Tuning,
			ZipBatch:                             newJob.Params.(DownloaderParams).ZipBatch,
		},
		opts...)
}

func (c *Client) DownloadToFile(params files_sdk.FileDownloadParams, filePath string, opts ...files_sdk.RequestResponseOption) (files_sdk.File, error) {
	out, err := os.Create(filePath)
	if err != nil {
		return files_sdk.File{}, err
	}
	return c.Download(params, append(opts, files_sdk.ResponseBodyOption(func(closer io.ReadCloser) error {
		_, err := io.Copy(out, closer)
		return err
	}))...)
}

type DownloaderParams struct {
	// Ignore gitignore formatted pattern
	Ignore []string
	// Include gitignore formatted pattern
	Include       []string
	RemotePath    string
	RemoteFile    files_sdk.File
	LocalPath     string
	TempPath      string // Folder path where the file(s) will be downloaded to before being moved to LocalPath. If not set, the file(s) will be downloaded directly to LocalPath.
	Sync          bool
	PreserveTimes bool
	NoOverwrite   bool
	// SyncAfterActions optionally run source cleanup/move actions after sync comparison.
	SyncAfterActions SyncAfterActions
	RetryPolicy
	*manager.Manager
	EventsReporter
	config             files_sdk.Config
	DryRun             bool
	PriorJobCheckpoint *JobDownloadCheckpoint
	ResumeTmpPath      string // Full path to the temp file from a prior paused download session.
	// AdaptiveConcurrency enables opt-in download V2 adaptive range concurrency.
	AdaptiveConcurrency bool
	// AdaptiveConcurrencyUseSDKDefaultCaps uses SDK V2 caps instead of treating an
	// explicitly supplied Manager as an isolation boundary.
	AdaptiveConcurrencyUseSDKDefaultCaps bool
	// AdaptiveDownloadV2TargetClassifier optionally overrides the SDK's download
	// V2 target classifier. Custom targets use default SDK transfer behavior but
	// keep separate adaptive manager cache entries and telemetry target labels.
	AdaptiveDownloadV2TargetClassifier DownloadV2TargetClassifier
	// AdaptiveDownloadV2TuningSet applies V2 tuning overrides below.
	// When false, V2 uses built-in defaults.
	AdaptiveDownloadV2TuningSet bool
	// AdaptiveDownloadV2Tuning holds opt-in V2 transfer tuning.
	AdaptiveDownloadV2Tuning UploadV2Tuning
	// ZipBatch controls batched small-file downloads through the ZIP download endpoint.
	ZipBatch ZipBatchParams
}

func (c *Client) Downloader(params DownloaderParams, opts ...files_sdk.RequestResponseOption) *Job {
	params.config = c.Config
	if params.AdaptiveConcurrency && (params.AdaptiveConcurrencyUseSDKDefaultCaps || params.Manager == nil) {
		params.config = params.config.SetCustomClient(manager.New(manager.AdaptiveDownloadV2ConcurrentFiles, manager.EffectiveAdaptiveDownloadV2ConcurrentFileParts(), manager.ConcurrentDirectoryList).CreateMatchingClient(params.config.HTTPClient))
	}
	job := downloader(files_sdk.ContextOption(opts), (&FS{}).Init(params.config, true), params, opts...)
	registerSyncAfterActions(job, params.SyncAfterActions, params.DryRun, params.config, opts...)
	return job
}

type Entity struct {
	fs.File
	fs.FS
	error
}
