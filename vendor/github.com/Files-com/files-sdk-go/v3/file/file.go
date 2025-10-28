package file

import (
	"encoding/json"
	"sync"
	"time"

	filesSDK "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/status"
)

type JobFile struct {
	StatusName    string        `json:"status"`
	TransferBytes int64         `json:"transferred_bytes"`
	Size          int64         `json:"size_bytes"`
	LocalPath     string        `json:"local_path"`
	RemotePath    string        `json:"remote_path"`
	EndedAt       time.Time     `json:"completed_at"`
	StartedAt     time.Time     `json:"started_at"`
	Err           error         `json:"error"`
	Id            string        `json:"-"`
	Attempts      int           `json:"attempts"`
	Mutex         *sync.RWMutex `json:"-"`
	status.Status `json:"-"`
	filesSDK.File `json:"-"`
	*Job          `json:"-"`
}

type MashableError struct {
	error
}

func (m MashableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.error)
}

func (m MashableError) Err() error {
	if m.error == nil {
		return nil
	}

	return m
}

func (m MashableError) Unwrap() error {
	return m.error
}

type Reporter func(JobFile)
