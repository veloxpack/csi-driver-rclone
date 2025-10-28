package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/directory"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
)

// DeleteSource files after a sync
//
//	job.RegisterFileEvent(func(file status.File) {
//			log, err := file.DeleteSource{Direction: f.Direction, Config: config}.Call(ctx, f)
//	}, status.Complete, status.Skipped)
type DeleteSource struct {
	direction.Direction
	Config files_sdk.Config
}

func (ad DeleteSource) Call(f JobFile, opts ...files_sdk.RequestResponseOption) (status.Log, error) {
	switch f.Direction {
	case direction.UploadType:
		return status.Log{Path: f.LocalPath, Action: "delete source"}, os.Remove(f.LocalPath)
	case direction.DownloadType:
		client := Client{Config: ad.Config}
		err := client.Delete(files_sdk.FileDeleteParams{Path: f.RemotePath}, opts...)
		return status.Log{Path: f.RemotePath, Action: "delete source"}, err
	default:
		panic(fmt.Sprintf("unknown direction %v", f.Direction))
	}
}

// DeleteEmptySourceFolders folder after a sync
//
//	job.RegisterFileEvent(func(file status.File) {
//			log, err := file.DeleteEmptySourceFolders{Direction: f.Direction, Config: config}.Call(ctx, f)
//	}, status.Complete, status.Skipped)
type DeleteEmptySourceFolders struct {
	direction.Direction
	Config files_sdk.Config
}

func (ad DeleteEmptySourceFolders) Call(f Job, opts ...files_sdk.RequestResponseOption) (status.Log, error) {
	switch f.Direction {
	case direction.UploadType:
		localFolder := filepath.Dir(f.LocalPath)
		err := DepthFirstWalkDir(localFolder, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return removeEmptyDir(path)
			}
			return nil
		})
		if err != nil {
			return status.Log{Path: localFolder, Action: "delete source folder"}, err
		}
		return status.Log{Path: localFolder, Action: "delete source folder"}, os.Remove(localFolder)
	case direction.DownloadType:
		client := Client{Config: ad.Config}
		err := client.Delete(files_sdk.FileDeleteParams{Path: f.RemotePath, Recursive: lib.Bool(true)}, opts...)
		return status.Log{Path: f.RemotePath, Action: "delete source folder"}, err
	default:
		panic(fmt.Sprintf("unknown direction %v", f.Direction))
	}
}

// Depth first version of WalkDir
func depthFirstWalkDir(path string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		// Report ReadDir error.
		err = walkDirFn(path, d, err)
		if err != nil {
			if err == fs.SkipDir && d.IsDir() {
				err = nil
			}
			return err
		}
	}

	for _, d1 := range dirs {
		path1 := filepath.Join(path, d1.Name())
		if err := depthFirstWalkDir(path1, d1, walkDirFn); err != nil {
			if err == fs.SkipDir {
				break
			}
			return err
		}
	}

	// Upstream this runs first, by moving it to the botrtom, we ensure we are doing a depth first walk
	if err := walkDirFn(path, d, nil); err != nil || !d.IsDir() {
		if err == fs.SkipDir && d.IsDir() {
			// Successfully skipped directory.
			err = nil
		}
		return err
	}

	return nil
}

func DepthFirstWalkDir(root string, fn fs.WalkDirFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {

		err = depthFirstWalkDir(root, fs.FileInfoToDirEntry(info), fn)
	}
	if err == fs.SkipDir || err == fs.SkipAll {
		return nil
	}
	return err
}

func removeEmptyDir(path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	if len(files) != 0 {
		return nil
	}

	return os.Remove(path)
}

// MoveSource files after a sync
//
//	job.RegisterFileEvent(func(file status.File) {
//			log, err := file.MoveSource{Direction: f.Direction, Config: config}.Call(ctx, f)
//	}, status.Complete, status.Skipped)
type MoveSource struct {
	direction.Direction
	Path   string
	Config files_sdk.Config
}

func (am MoveSource) Call(f JobFile, opts ...files_sdk.RequestResponseOption) (status.Log, error) {
	var err error
	log := status.Log{Action: "move source"}
	log.Path = am.movePath(f)

	switch f.Direction {
	case direction.UploadType:
		dir, _ := filepath.Split(log.Path)
		err = os.MkdirAll(dir, 0755)
		if err != nil && !errors.Is(err, syscall.EEXIST) {
			return log, err
		}
		err = os.Rename(f.LocalPath, log.Path)
		if err != nil && errors.Is(err, syscall.EEXIST) {
			err = os.Remove(log.Path)
			if err != nil {
				return log, err
			}
			return am.Call(f, opts...)
		}
		return log, err
	case direction.DownloadType:
		client := &Client{Config: am.Config}
		_, err := client.Move(
			files_sdk.FileMoveParams{Path: f.RemotePath, Destination: lib.Path{Path: log.Path}.NormalizePathSystemForAPI().String()},
			opts...,
		)
		rErr, ok := err.(files_sdk.ResponseError)
		if ok && rErr.Type == "processing-failure/destination-parent-does-not-exist" {
			err := (&FS{}).Init(am.Config, true).WithContext(files_sdk.ContextOption(opts)).(*FS).MkdirAll(filepath.Dir(log.Path), 0755)
			if err != nil {
				return log, err
			}
			return am.Call(f, opts...)
		}
		if ok && rErr.Type == "processing-failure/destination-exists" {
			err := client.Delete(files_sdk.FileDeleteParams{Path: lib.Path{Path: log.Path}.NormalizePathSystemForAPI().String()}, opts...)
			if err != nil {
				return log, err
			}
			return am.Call(f, opts...)
		}
		return log, err
	default:
		panic(fmt.Sprintf("unknown direction %v", f.Direction))
	}
}

func (am MoveSource) movePath(f JobFile) string {
	switch f.Job.Type {
	case directory.Dir:
		return filepath.Join(
			append([]string{am.Path}, strings.Split(strings.TrimPrefix(f.RemotePath, f.Job.RemotePath), "/")...)...,
		)
	case directory.File:
		return am.Path
	default:
		panic("")
	}
}
