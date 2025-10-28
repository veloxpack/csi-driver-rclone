package file

import (
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/manager"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	lib "github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
)

type UploadOption func(uploadIO) (uploadIO, error)

func UploadWithContext(ctx context.Context) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.passedInContext = ctx
		return params, nil
	}
}

func UploadWithReader(reader io.Reader) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		if lenReader, ok := reader.(Len); ok && params.Size == nil {
			params.Size = lib.Int64(int64(lenReader.Len()))
		}
		params.reader = reader
		return params, nil
	}
}

func UploadWithReaderAt(readerAt io.ReaderAt) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		if lenReader, ok := readerAt.(Len); ok && params.Size == nil {
			params.Size = lib.Int64(int64(lenReader.Len()))
		}
		params.readerAt = readerAt
		return params, nil
	}
}

func UploadWithFile(sourcePath string) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		file, err := os.Open(sourcePath)
		if err != nil {
			return params, err
		}
		info, err := file.Stat()
		if err != nil {
			return params, err
		}

		params, err = UploadWithReaderAt(file)(params)
		if err != nil {
			return params, err
		}
		return UploadWithSize(info.Size())(params)
	}
}

func UploadWithDestinationPath(destinationPath string) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.Path = destinationPath
		return params, nil
	}
}

func UploadWithProvidedMtime(providedMtime time.Time) UploadOption {
	return UploadWithProvidedMtimePtr(&providedMtime)
}

func UploadWithProvidedMtimePtr(providedMtime *time.Time) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.ProvidedMtime = providedMtime
		return params, nil
	}
}

func UploadWithManager(manager lib.ConcurrencyManagerWithSubWorker) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.Manager = manager
		return params, nil
	}
}

func UploadWithSize(size int64) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.Size = &size
		return params, nil
	}
}

func UploadWithProgress(progress Progress) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.Progress = progress
		return params, nil
	}
}

// UploadRewindAllProgressOnFailure on upload failure rewind all successfully parts
func UploadRewindAllProgressOnFailure() UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.RewindAllProgressOnFailure = true
		return params, nil
	}
}

func UploadWithResume(resume UploadResumable) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		if params.Path == "" {
			params.Path = resume.FileUploadPart.Path
		}
		params.Parts = resume.Parts
		params.FileUploadPart = resume.FileUploadPart
		return params, nil
	}
}

func UploadWithActionAttributes(attributes map[string]any) UploadOption {
	return func(params uploadIO) (uploadIO, error) {
		params.actionAttributes = attributes
		return params, nil
	}
}

func (c *Client) Upload(opts ...UploadOption) error {
	_, err := c.UploadWithResume(opts...)

	return err
}

func (c *Client) UploadWithResume(opts ...UploadOption) (UploadResumable, error) {
	uploadIo := uploadIO{
		Client:          c,
		Manager:         manager.Default().FilePartsManager,
		passedInContext: context.Background(),
		ByteOffset:      ByteOffset{PartSizes: lib.PartSizes},
	}

	for _, opt := range opts {
		var err error
		uploadIo, err = opt(uploadIo)
		if err != nil {
			return UploadResumable{}, err
		}
	}
	return (&uploadIo).Run(uploadIo.passedInContext)
}

// UploadFile Deprecated use c.Upload(UploadWithFile(sourcePath), UploadWithDestinationPath(destinationPath))
func (c *Client) UploadFile(sourcePath string, destinationPath string, opts ...UploadOption) error {
	return c.Upload(append(opts, UploadWithFile(sourcePath), UploadWithDestinationPath(destinationPath))...)
}

func (c *Client) UploadRetry(job Job, opts ...files_sdk.RequestResponseOption) *Job {
	newJob := job.ClearStatuses()
	return c.Uploader(
		UploaderParams{
			Sync:           newJob.Sync,
			LocalPath:      newJob.LocalPath,
			RemotePath:     newJob.RemotePath,
			EventsReporter: newJob.EventsReporter,
			Manager:        newJob.Manager,
			RetryPolicy:    newJob.RetryPolicy.(RetryPolicy),
			Ignore:         newJob.Params.(UploaderParams).Ignore,
		},
		opts...,
	)
}

type UploaderParams struct {
	// Ignore gitignore formatted pattern
	Ignore []string
	// Include gitignore formatted pattern
	Include []string
	// Sync compare destination and only upload if different or non-existent.
	Sync bool
	// LocalPaths files or directories to upload.
	LocalPaths []string
	// LocalPath a file or directory to recursively upload.
	LocalPath string
	// RemotePath destination path for files.com, formatted `/` path separator.
	RemotePath string
	// DryRun see what files would be uploaded.
	DryRun bool
	// NoOverwrite do not overwrite existing files.
	NoOverwrite bool
	// RetryPolicy config for retrying errored uploads.
	RetryPolicy
	// EventsReporter log file events
	EventsReporter
	// Manager limit concurrency
	*manager.Manager
	*Job
	config        files_sdk.Config
	PreserveTimes bool
}

func expand(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func (c *Client) Uploader(params UploaderParams, opts ...files_sdk.RequestResponseOption) *Job {
	job := (&Job{}).Init()
	params.config = c.Config
	SetJobParams(job, direction.UploadType, params, c.Logger, (&FS{}).Init(c.Config, true))
	job.Config = params.config
	job.CodeStart = func() {
		params.Job = job
		job.Params = params
		params.RemotePath = lib.Path{Path: params.RemotePath}.PruneStartingSlash().String()
		file := &UploadStatus{file: files_sdk.File{}, remotePath: params.RemotePath, localPath: params.LocalPath, status: status.Queued, job: job}
		expandedPath, err := expand(params.LocalPath)
		if err != nil {
			job.Add(file)
			job.UpdateStatus(status.Errored, file, err)
			return
		}
		absolutePath, err := filepath.Abs(expandedPath)
		if err != nil {
			job.Add(file)
			job.UpdateStatus(status.Errored, file, err)
			return
		}
		if (lib.Path{Path: params.LocalPath}).EndingSlash() {
			params.LocalPath = absolutePath + string(os.PathSeparator)
		} else {
			params.LocalPath = absolutePath
		}

		uploader(files_sdk.ContextOption(opts), c, params).CodeStart()
	}

	return job
}
