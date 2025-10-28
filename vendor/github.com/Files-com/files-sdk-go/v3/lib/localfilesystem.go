package lib

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type WritableFs interface {
	MkdirAll(string, fs.FileMode) error
	MkdirTemp(dir, pattern string) (string, error)
	TempDir() string
	Create(string) (io.WriteCloser, error)
	RemoveAll(string) error
	Remove(string) error
	PathSeparator() string
	PathJoin(...string) string
	RelPath(parent, child string) (string, error)
	SplitPath(path string) (string, string)
	Chtimes(name string, atime time.Time, mtime time.Time) error
}

type StatefulDirectory interface {
	Chdir(string) error
	Getwd() (string, error)
}

type ReadWriteFs interface {
	WritableFs
	fs.FS
}

type LocalFileSystem struct{}

var _ = ReadWriteFs(LocalFileSystem{})
var _ = StatefulDirectory(LocalFileSystem{})

func (w LocalFileSystem) MkdirAll(path string, mode fs.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (w LocalFileSystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (w LocalFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (w LocalFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (w LocalFileSystem) PathSeparator() string {
	return string(os.PathSeparator)
}

func (w LocalFileSystem) PathJoin(paths ...string) string {
	return filepath.Join(paths...)
}

func (w LocalFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

func (w LocalFileSystem) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (w LocalFileSystem) Open(path string) (fs.File, error) {
	return os.Open(path)
}

func (w LocalFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (w LocalFileSystem) RelPath(parent, child string) (string, error) {
	return filepath.Rel(parent, child)
}

func (w LocalFileSystem) SplitPath(path string) (string, string) {
	return filepath.Split(path)
}

func (w LocalFileSystem) TempDir() string {
	return os.TempDir()
}

func (w LocalFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}

type FSWithContext interface {
	WithContext(ctx context.Context) fs.FS
}

type FileWithContext interface {
	WithContext(ctx context.Context) fs.File
}
