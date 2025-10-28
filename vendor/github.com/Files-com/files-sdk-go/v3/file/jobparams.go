package file

import (
	"io/fs"

	"github.com/Files-com/files-sdk-go/v3/lib/direction"
	"github.com/hashicorp/go-retryablehttp"
)

func SetJobParams(r *Job, d direction.Direction, params interface{}, logger retryablehttp.Logger, remoteFs fs.FS) {
	r.Params = params
	r.Direction = d
	r.Logger = logger
	r.RemoteFs = remoteFs
	switch d {
	case direction.DownloadType:
		p := params.(DownloaderParams)
		r.SetManager(p.Manager)
		r.SetEventsReporter(p.EventsReporter)
		r.RetryPolicy = p.RetryPolicy
		r.LocalPath = p.LocalPath
		r.RemotePath = p.RemotePath
		r.Sync = p.Sync
	case direction.UploadType:
		p := params.(UploaderParams)
		r.SetManager(p.Manager)
		r.SetEventsReporter(p.EventsReporter)
		r.RetryPolicy = p.RetryPolicy
		r.LocalPath = p.LocalPath
		r.RemotePath = p.RemotePath
		r.Sync = p.Sync
	}
}
