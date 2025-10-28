package file

import (
	"sync"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/status"
)

type UploadStatus struct {
	file          files_sdk.File
	status        status.Status
	job           *Job
	localPath     string
	remotePath    string
	uploadedBytes int64
	Sync          bool
	NoOverwrite   bool
	Uploader
	UploadResumable
	Mutex *sync.RWMutex
	error
	lastError   error
	missingStat bool
	dryRun      bool
	status.Changes
	endedAt   time.Time
	startedAt time.Time
}

var _ IFile = &UploadStatus{}

func (u *UploadStatus) EndedAt() time.Time {
	return u.endedAt
}

func (u *UploadStatus) StartedAt() time.Time {
	return u.startedAt
}

func (u *UploadStatus) Size() int64 {
	return u.File().Size
}

func (u *UploadStatus) RecentError() error {
	if u.error != nil {
		return u.error
	}

	return u.lastError
}

func (u *UploadStatus) Job() *Job {
	return u.job
}

func (u *UploadStatus) TransferBytes() int64 {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()
	return u.uploadedBytes
}

func (u *UploadStatus) File() files_sdk.File {
	return u.file
}

func (u *UploadStatus) LocalPath() string {
	return u.localPath
}

func (u *UploadStatus) RemotePath() string {
	return u.remotePath
}

func (u *UploadStatus) Status() status.Status {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()
	return u.status
}

func (u *UploadStatus) Err() error {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()
	return u.error
}

func (u *UploadStatus) SetStatus(s status.Status, err error) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	var setError bool
	u.status, setError = status.SetStatus(u.status, s, err)
	if setError {
		if u.error != nil {
			u.lastError = u.error
		}
		u.error = err
	}

	if s.Is(status.Uploading) && u.startedAt.IsZero() {
		u.startedAt = time.Now()
	}

	if s.Is(status.Retrying) {
		u.uploadedBytes = 0
	}

	if s.Is(status.Ended...) {
		u.endedAt = time.Now()
	}

	u.Changes = append(u.Changes, status.Change{Status: u.status, Err: u.error, Time: time.Now()})
}

func (u *UploadStatus) StatusChanges() status.Changes {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()

	return u.Changes
}

func (u *UploadStatus) Id() string {
	return u.job.Id + ":" + u.file.Path
}

func (u *UploadStatus) incrementUploadedBytes(b int64) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	u.uploadedBytes += b
}

func (u *UploadStatus) IncrementTransferBytes(b int64) {
	u.incrementUploadedBytes(b)
}

func (u *UploadStatus) SetUploadedBytes(b int64) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	u.uploadedBytes = b
}
