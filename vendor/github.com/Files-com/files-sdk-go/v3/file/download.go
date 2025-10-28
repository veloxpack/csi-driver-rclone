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
			RemotePath:     newJob.RemotePath,
			Sync:           newJob.Sync,
			Manager:        newJob.Manager,
			LocalPath:      newJob.LocalPath,
			RetryPolicy:    newJob.RetryPolicy.(RetryPolicy),
			EventsReporter: newJob.EventsReporter,
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
	RetryPolicy
	*manager.Manager
	EventsReporter
	config files_sdk.Config
	DryRun bool
}

func (c *Client) Downloader(params DownloaderParams, opts ...files_sdk.RequestResponseOption) *Job {
	params.config = c.Config
	return downloader(files_sdk.ContextOption(opts), (&FS{}).Init(c.Config, true), params)
}

type Entity struct {
	fs.File
	fs.FS
	error
}
