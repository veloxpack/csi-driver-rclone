package files_sdk

import (
	"io/fs"
	"time"
)

type DirEntry struct {
	File
}

func (f DirEntry) Mode() fs.FileMode {
	return fs.ModePerm
}

func (f DirEntry) ModTime() time.Time {
	if f.File.Mtime == nil {
		return time.Time{}
	}
	return *f.File.Mtime
}

func (f DirEntry) Sys() any {
	return f.File
}

func (f DirEntry) Type() fs.FileMode {
	return fs.ModePerm
}

func (f DirEntry) Name() string {
	return f.DisplayName
}

func (f DirEntry) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f DirEntry) Size() int64 {
	return f.File.Size
}
