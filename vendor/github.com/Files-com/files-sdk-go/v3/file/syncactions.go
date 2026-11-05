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
	"github.com/Files-com/files-sdk-go/v3/folder"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
)

type SyncAfterActions struct {
	DeleteSourceFiles        bool
	DeleteSourceEmptyFolders bool
	// MoveSource is the destination path/root for moving synced source files after sync.
	MoveSource string
	// Log receives validation and cleanup action results. If nil, this helper cannot report those errors.
	Log func(status.Log, error)
}

func (a SyncAfterActions) Enabled() bool {
	return a.DeleteSourceFiles || a.DeleteSourceEmptyFolders || a.MoveSource != ""
}

func (a SyncAfterActions) Validate() error {
	if a.DeleteSourceFiles && a.MoveSource != "" {
		return errors.New("delete source files and move source cannot both be enabled")
	}
	return nil
}

func (a SyncAfterActions) log(log status.Log, err error) {
	if a.Log != nil {
		a.Log(log, err)
	}
}

func (a SyncAfterActions) runFile(f JobFile, config files_sdk.Config, opts ...files_sdk.RequestResponseOption) {
	if a.DeleteSourceFiles {
		a.log(DeleteSource{Direction: f.Direction, Config: config}.Call(f, opts...))
	}

	if a.MoveSource != "" {
		a.log(MoveSource{Direction: f.Direction, Config: config, Path: a.MoveSource}.Call(f, opts...))
	}
}

func (a SyncAfterActions) runEmptyFolders(job *Job, config files_sdk.Config, opts ...files_sdk.RequestResponseOption) {
	if !a.DeleteSourceEmptyFolders {
		return
	}

	a.log(DeleteEmptySourceFolders{Config: config, Direction: job.Direction}.call(job, opts...))
}

func runSyncAfterEmptyFolders(job *Job, actions SyncAfterActions, dryRun bool, config files_sdk.Config, opts ...files_sdk.RequestResponseOption) {
	if dryRun || !actions.DeleteSourceEmptyFolders {
		return
	}
	if err := actions.Validate(); err != nil {
		return
	}
	if job.Canceled.Called() || job.Count(status.Errored) != 0 {
		return
	}

	actions.runEmptyFolders(job, config, opts...)
}

func registerSyncAfterActions(job *Job, actions SyncAfterActions, dryRun bool, config files_sdk.Config, opts ...files_sdk.RequestResponseOption) {
	if dryRun || !actions.Enabled() {
		return
	}
	if err := actions.Validate(); err != nil {
		actions.log(status.Log{Action: "sync after actions"}, err)
		return
	}

	job.RegisterFileEvent(func(f JobFile) {
		if job.Canceled.Called() {
			return
		}
		actions.runFile(f, config, opts...)
	}, status.Complete, status.Skipped)
}

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

func (ad DeleteEmptySourceFolders) Call(job Job, opts ...files_sdk.RequestResponseOption) (status.Log, error) {
	return ad.call(&job, opts...)
}

func (ad DeleteEmptySourceFolders) call(job *Job, opts ...files_sdk.RequestResponseOption) (status.Log, error) {
	switch job.Direction {
	case direction.UploadType:
		localFolder := uploadEmptyFolderRoot(*job)
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
		return status.Log{Path: localFolder, Action: "delete source folder"}, nil
	case direction.DownloadType:
		err := ad.removeRemoteEmptyDirs(job.RemotePath, opts...)
		return status.Log{Path: job.RemotePath, Action: "delete source folder"}, err
	default:
		panic(fmt.Sprintf("unknown direction %v", job.Direction))
	}
}

func (ad DeleteEmptySourceFolders) removeRemoteEmptyDirs(path string, opts ...files_sdk.RequestResponseOption) error {
	childDirs, hasFile, err := ad.remoteFolderContents(path, opts...)
	if err != nil {
		return err
	}

	for _, childDir := range childDirs {
		if err := ad.removeRemoteEmptyDirs(childDir.Path, opts...); err != nil {
			return err
		}
	}

	childDirs, hasFile, err = ad.remoteFolderContents(path, opts...)
	if err != nil || hasFile || len(childDirs) > 0 {
		return err
	}

	client := Client{Config: ad.Config}
	return client.Delete(files_sdk.FileDeleteParams{Path: path}, opts...)
}

func (ad DeleteEmptySourceFolders) remoteFolderContents(path string, opts ...files_sdk.RequestResponseOption) ([]files_sdk.File, bool, error) {
	folderClient := folder.Client{Config: ad.Config}
	it, err := folderClient.ListFor(files_sdk.FolderListForParams{Path: path}, opts...)
	if err != nil {
		return nil, false, err
	}

	childDirs := make([]files_sdk.File, 0)
	hasFile := false
	for it.Next() {
		child := it.File()
		if child.Type != string(directory.Dir) {
			hasFile = true
			continue
		}
		childDirs = append(childDirs, child)
	}
	if it.Err() != nil {
		return nil, false, it.Err()
	}
	return childDirs, hasFile, nil
}

func uploadEmptyFolderRoot(f Job) string {
	if f.Type == directory.Dir {
		return f.LocalPath
	}
	return filepath.Dir(f.LocalPath)
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

	// Upstream this runs first; moving it to the bottom makes the walk depth first.
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

// MoveSource moves a successfully synced source file to Path.
// For uploads it moves the local source file. For downloads it moves the remote source file.
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
		if errors.Is(err, files_sdk.ErrDestinationParentDoesNotExist) {
			err := (&FS{}).Init(am.Config, true).WithContext(files_sdk.ContextOption(opts)).(*FS).MkdirAll(filepath.Dir(log.Path), 0755)
			if err != nil {
				return log, err
			}
			return am.Call(f, opts...)
		}
		if errors.Is(err, files_sdk.ErrDestinationExists) {
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
