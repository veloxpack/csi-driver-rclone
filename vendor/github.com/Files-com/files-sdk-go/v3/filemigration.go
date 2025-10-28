package files_sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FileMigration struct {
	Id         int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path       string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	DestPath   string `json:"dest_path,omitempty" path:"dest_path,omitempty" url:"dest_path,omitempty"`
	FilesMoved int64  `json:"files_moved,omitempty" path:"files_moved,omitempty" url:"files_moved,omitempty"`
	FilesTotal int64  `json:"files_total,omitempty" path:"files_total,omitempty" url:"files_total,omitempty"`
	Operation  string `json:"operation,omitempty" path:"operation,omitempty" url:"operation,omitempty"`
	Region     string `json:"region,omitempty" path:"region,omitempty" url:"region,omitempty"`
	Status     string `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	LogUrl     string `json:"log_url,omitempty" path:"log_url,omitempty" url:"log_url,omitempty"`
}

func (f FileMigration) Identifier() interface{} {
	return f.Id
}

type FileMigrationCollection []FileMigration

// this code is merged into the built version of filemigration.go

type FilesMigrationLog struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	FileType  string    `json:"file_type"`
	Path      string    `json:"path"`
	DestPath  string    `json:"dest_path"`
}

// FilesMigrationLogIter Transforms migrations into a log iterator
type FilesMigrationLogIter struct {
	context.Context
	Config
	FileMigration
	next func() bool
	more func() bool
	error
	current interface{}
}

func (l *FilesMigrationLogIter) Err() error {
	return l.error
}

func (l *FilesMigrationLogIter) Current() interface{} {
	return l.current
}

func (l *FilesMigrationLogIter) Next() bool {
	return l.next()
}

func (l FilesMigrationLogIter) Init() *FilesMigrationLogIter {
	l.next = func() bool {
		return false
	}
	if l.FileMigration.LogUrl != "" {
		req, err := http.NewRequestWithContext(l.Context, "GET", l.FileMigration.LogUrl, nil)
		if err != nil {
			l.error = err
			return &l
		}
		resp, err := l.Config.Do(req)
		if err != nil {
			l.error = err
			return &l
		}
		decoder := json.NewDecoder(resp.Body)
		l.more = decoder.More
		l.next = func() bool {
			if l.error != nil {
				return false
			}
			if l.more() {
				var log FilesMigrationLog
				if err := decoder.Decode(&log); err != nil {
					l.error = err
					return false
				}
				l.current = log
				return true
			}
			return false
		}
	}

	return &l
}

type FileMigrationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (f *FileMigration) UnmarshalJSON(data []byte) error {
	type fileMigration FileMigration
	var v fileMigration
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FileMigration(v)
	return nil
}

func (f *FileMigrationCollection) UnmarshalJSON(data []byte) error {
	type fileMigrations FileMigrationCollection
	var v fileMigrations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileMigrationCollection(v)
	return nil
}

func (f *FileMigrationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
