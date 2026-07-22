package file

import (
	"errors"

	"github.com/Files-com/files-sdk-go/v3/file/status"
)

// ErrJobPaused is the cancel cause set when a job is paused rather than fully canceled.
// Use errors.Is(context.Cause(ctx), ErrJobPaused) to distinguish pause from cancel.
var ErrJobPaused = errors.New("job paused")

type JobDownloadCheckpoint struct {
	CompletedPaths []string
}

// DownloadCheckpoint builds a JobDownloadCheckpoint from the job's settled file statuses.
// Call at terminal time (Canceled or Finished) instead of tracking state incrementally.
func (j *Job) DownloadCheckpoint() *JobDownloadCheckpoint {
	completed := make(map[string]struct{})
	for p := range j.CompletedPaths {
		completed[p] = struct{}{}
	}

	j.statusesMutex.RLock()
	for _, f := range j.Statuses {
		if f.Status().Is(status.Complete) {
			completed[f.LocalPath()] = struct{}{}
		}
	}
	j.statusesMutex.RUnlock()

	paths := make([]string, 0, len(completed))
	for p := range completed {
		paths = append(paths, p)
	}
	return &JobDownloadCheckpoint{CompletedPaths: paths}
}
