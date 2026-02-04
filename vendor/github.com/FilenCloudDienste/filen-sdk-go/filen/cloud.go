package filen

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/search"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/util"
	"github.com/google/uuid"
	"github.com/rclone/rclone/fs"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

// FindItem finds a cloud item (file or directory) by its path and returns it.
// The path should be in the format "dir1/dir2/item", with "/" as the separator.
// If the path is empty or "/", it returns the root directory.
// Returns nil if the item is not found.
func (api *Filen) FindItem(ctx context.Context, path string) (types.FileSystemObject, error) {
	var currentDir types.DirectoryInterface = &api.BaseFolder
	segments := strings.Split(path, "/")
	if len(strings.Join(segments, "")) == 0 {
		return currentDir, nil
	}

SegmentsLoop:
	for segmentIdx, segment := range segments {
		if segment == "" {
			continue
		}

		files, directories, err := api.ReadDirectory(ctx, currentDir)
		if err != nil {
			return nil, fmt.Errorf("read directory: %w", err)
		}
		for _, file := range files {
			if file.Name == segment {
				return file, nil
			}
		}
		for _, directory := range directories {
			if directory.Name == segment {
				if segmentIdx == len(segments)-1 {
					return directory, nil
				} else {
					currentDir = directory
					continue SegmentsLoop
				}
			}
		}
		return nil, nil
	}
	return nil, nil
}

// FindFile finds a cloud item by its path and then tries to map it to a file.
// Returns fs.ErrorIsDir if the item is a directory.
// Returns nil, nil if the file is not found.
func (api *Filen) FindFile(ctx context.Context, path string) (*types.File, error) {
	item, err := api.FindItem(ctx, path)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	file, ok := item.(*types.File)
	if !ok {
		return nil, fs.ErrorIsDir
	}
	return file, nil
}

// FindDirectory finds a cloud item by its path and then tries to map it to a directory.
// Returns fs.ErrorIsFile if the item is a file.
// Returns nil, nil if the directory is not found.
func (api *Filen) FindDirectory(ctx context.Context, path string) (types.DirectoryInterface, error) {
	item, err := api.FindItem(ctx, path)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	directory, ok := item.(types.DirectoryInterface)
	if !ok {
		return nil, fs.ErrorIsFile
	}
	return directory, nil
}

// FindDirectoryOrCreate finds a cloud directory by its path and returns it.
// If the directory cannot be found, it (and all non-existent parent directories) will be created.
// This is useful for ensuring a directory path exists before uploading files.
func (api *Filen) FindDirectoryOrCreate(ctx context.Context, path string) (types.DirectoryInterface, error) {
	segments := strings.Split(path, "/")

	var currentDir types.DirectoryInterface = &api.BaseFolder
SegmentsLoop:
	for _, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}

		_, directories, err := api.ReadDirectory(ctx, currentDir)
		if err != nil {
			return nil, err
		}
		for _, directory := range directories {
			if directory.Name == segment {
				// directory found
				currentDir = directory
				continue SegmentsLoop
			}
		}
		// create directory
		directory, err := api.CreateDirectory(ctx, currentDir, segment)
		if err != nil {
			return nil, err
		}
		currentDir = directory
	}
	return currentDir, nil
}

// ReadDirectory fetches the files and directories that are direct children of a directory.
// It retrieves the encrypted metadata for each item and decrypts it to provide
// fully populated File and Directory objects.
func (api *Filen) ReadDirectory(ctx context.Context, dir types.DirectoryInterface) ([]*types.File, []*types.Directory, error) {
	// fetch directory content
	directoryContent, err := api.Client.PostV3DirContent(ctx, dir.GetUUID())
	if err != nil {
		return nil, nil, fmt.Errorf("ReadDirectory fetching directory: %w", err)
	}

	// transform files
	files := make([]*types.File, 0)
	for _, file := range directoryContent.Uploads {
		metadataStr, err := api.DecryptMeta(file.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadDirectory decrypting metadata: %v", err)
		}
		var metadata types.FileMetadata
		err = json.Unmarshal([]byte(metadataStr), &metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadDirectory unmarshalling metadata: %v", err)
		}

		encryptionKey, err := crypto.MakeEncryptionKeyFromUnknownStr(metadata.Key)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadDirectory creating encryption key: %v", err)
		}

		files = append(files, &types.File{
			IncompleteFile: types.IncompleteFile{
				UUID:          file.UUID,
				Name:          metadata.Name,
				MimeType:      metadata.MimeType,
				EncryptionKey: *encryptionKey,
				Created:       util.TimestampToTime(int64(metadata.Created)),
				LastModified:  util.TimestampToTime(int64(metadata.LastModified)),
				ParentUUID:    file.Parent,
			},
			Size:      metadata.Size,
			Favorited: file.Favorited == 1,
			Region:    file.Region,
			Bucket:    file.Bucket,
			Chunks:    file.Chunks,
			Hash:      metadata.Hash,
			Version:   file.Version,
		})
	}

	// transform directories
	directories := make([]*types.Directory, 0)
	for _, directory := range directoryContent.Folders {
		metaStr, err := api.DecryptMeta(directory.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadDirectory decrypting metadata: %v", err)
		}
		metaData := types.DirectoryMetaData{}
		err = json.Unmarshal([]byte(metaStr), &metaData)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadDirectory unmarshalling metadata: %v", err)
		}

		creationTimestamp := metaData.Creation
		if creationTimestamp == 0 {
			creationTimestamp = directory.Timestamp
		}

		directories = append(directories, &types.Directory{
			UUID:       directory.UUID,
			Name:       metaData.Name,
			ParentUUID: directory.Parent,
			Color:      directory.Color,
			Created:    util.TimestampToTime(int64(creationTimestamp)),
			Favorited:  directory.Favorited == 1,
		})
	}

	return files, directories, nil
}

// ListRecursive fetches all the files and directories that are descendants of a directory
// in a single backend API call. This is more efficient than multiple ReadDirectory calls
// when you need to retrieve the entire directory tree.
func (api *Filen) ListRecursive(ctx context.Context, dir types.DirectoryInterface) ([]*types.File, []*types.Directory, error) {
	resp, err := api.Client.PostV3DirDownload(ctx, dir.GetUUID())
	if err != nil {
		return nil, nil, fmt.Errorf("ListRecursive fetching directory: %w", err)
	}
	files := make([]*types.File, 0, len(resp.Files))
	dirs := make([]*types.Directory, 0, len(resp.Folders))

	for _, file := range resp.Files {
		metaStr, err := api.DecryptMeta(file.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ListRecursive decrypting metadata: %v", err)
		}
		metadata := types.FileMetadata{}
		err = json.Unmarshal([]byte(metaStr), &metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ListRecursive unmarshalling metadata: %v", err)
		}

		encryptionKey, err := crypto.MakeEncryptionKeyFromUnknownStr(metadata.Key)
		if err != nil {
			return nil, nil, fmt.Errorf("ListRecursive creating encryption key: %v", err)
		}

		files = append(files, &types.File{
			IncompleteFile: types.IncompleteFile{
				UUID:          file.UUID,
				Name:          metadata.Name,
				MimeType:      metadata.MimeType,
				EncryptionKey: *encryptionKey,
				Created:       util.TimestampToTime(int64(metadata.Created)),
				LastModified:  util.TimestampToTime(int64(metadata.LastModified)),
				ParentUUID:    file.Parent,
			},
			Size:      metadata.Size,
			Favorited: file.Favorited,
			Region:    file.Region,
			Bucket:    file.Bucket,
			Chunks:    file.Chunks,
			Hash:      metadata.Hash,
			Version:   file.Version,
		})
	}

	for _, directory := range resp.Folders {
		if directory.Parent == "base" {
			// /v3/dir/download returns the dir it was called on as well with parent base
			continue
		}
		metaStr, err := api.DecryptMeta(directory.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("ListRecursive decrypting metadata: %v", err)
		}
		metaData := types.DirectoryMetaData{}
		err = json.Unmarshal([]byte(metaStr), &metaData)
		if err != nil {
			return nil, nil, fmt.Errorf("ListRecursive unmarshalling metadata: %v", err)
		}

		creationTimestamp := metaData.Creation
		if creationTimestamp == 0 {
			creationTimestamp = directory.Timestamp
		}

		dirs = append(dirs, &types.Directory{
			UUID:       directory.UUID,
			Name:       metaData.Name,
			ParentUUID: directory.Parent,
			Color:      types.DirColor(directory.Color),
			Created:    util.TimestampToTime(int64(creationTimestamp)),
			Favorited:  directory.Favorited,
		})
	}
	return files, dirs, nil
}

// TrashFile moves a file to the trash.
// This operation requires a lock to prevent race conditions with other operations.
func (api *Filen) TrashFile(ctx context.Context, file types.File) error {
	err := api.Lock(ctx)
	if err != nil {
		return err
	}
	defer api.Unlock()
	return api.Client.PostV3FileTrash(ctx, file.GetUUID())
}

// CreateDirectoryWithParentUUID creates a new directory as a child of the specified parent UUID.
// It handles encryption of directory metadata and updating search indexes.
// Returns the newly created Directory object.
func (api *Filen) CreateDirectoryWithParentUUID(ctx context.Context, parentUUID string, name string) (*types.Directory, error) {
	if strings.ContainsRune(name, '/') {
		return nil, fmt.Errorf("invalid directory name")
	}
	directoryUUID := uuid.New().String()
	creationTime := time.Now().Round(time.Millisecond)
	// encrypt metadata
	metadata := types.DirectoryMetaData{
		Name:     name,
		Creation: int(creationTime.UnixMilli()),
	}
	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	metadataEncrypted := api.EncryptMeta(string(metadataStr))

	// hash name
	nameHashed := api.HashFileName(name)

	// send
	response, err := api.Client.PostV3DirCreate(ctx, directoryUUID, metadataEncrypted, nameHashed, parentUUID)
	if err != nil {
		return nil, err
	}

	dir := &types.Directory{
		UUID:       response.UUID,
		Name:       name,
		ParentUUID: parentUUID,
		Color:      types.DirColorDefault,
		Created:    creationTime,
		Favorited:  false,
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return api.updateItemWithMaybeSharedParent(gCtx, dir) })
	g.Go(func() error { return api.updateSearchHashes(gCtx, dir) })
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return dir, nil
}

// CreateDirectory creates a new directory as a child of the specified parent directory.
// It uses CreateDirectoryWithParentUUID internally after extracting the parent's UUID.
func (api *Filen) CreateDirectory(ctx context.Context, parent types.DirectoryInterface, name string) (*types.Directory, error) {
	return api.CreateDirectoryWithParentUUID(ctx, parent.GetUUID(), name)
}

// TrashDirectory moves a directory to the trash.
// This operation requires a lock to prevent race conditions with other operations.
func (api *Filen) TrashDirectory(ctx context.Context, dir types.DirectoryInterface) error {
	err := api.Lock(ctx)
	if err != nil {
		return err
	}
	defer api.Unlock()
	return api.Client.PostV3DirTrash(ctx, dir.GetUUID())
}

// FileExists checks if a file with the given name exists in the specified parent directory.
// It uses the hashed filename for lookup to preserve end-to-end encryption.
func (api *Filen) FileExists(ctx context.Context, parentUUID string, name string) (*client.V3FileExistsResponse, error) {
	nameHashed := api.HashFileName(name)
	return api.Client.PostV3FileExists(ctx, nameHashed, parentUUID)
}

// DirExists checks if a directory with the given name exists in the specified parent directory.
// It uses the hashed directory name for lookup to preserve end-to-end encryption.
func (api *Filen) DirExists(ctx context.Context, parentUUID string, name string) (*client.V3DirExistsResponse, error) {
	nameHashed := api.HashFileName(name)
	return api.Client.PostV3DirExists(ctx, nameHashed, parentUUID)
}

// moveFile moves a file to a new parent directory.
// If overwrite is true, it will replace any existing file with the same name.
// Internal helper for MoveItem.
func (api *Filen) moveFile(ctx context.Context, file *types.File, newParentUUID string, overwrite bool) error {
	resp, err := api.FileExists(ctx, newParentUUID, file.GetName())
	if err != nil {
		return fmt.Errorf("FileExists: %w", err)
	}
	if resp.Exists {
		if overwrite {
			err := api.Client.PostV3FileTrash(ctx, resp.UUID)
			if err != nil {
				return fmt.Errorf("TrashFile: %w", err)
			}
		} else {
			return fmt.Errorf("file already exists")
		}
	}

	err = api.Client.PostV3FileMove(ctx, file.GetUUID(), newParentUUID)
	if err != nil {
		return fmt.Errorf("PostV3FileMove: %w", err)
	}
	file.ParentUUID = newParentUUID
	return api.updateItemWithMaybeSharedParent(ctx, file)
}

// moveDir moves a directory to a new parent directory.
// If overwrite is true, it will replace any existing directory with the same name.
// Internal helper for MoveItem.
func (api *Filen) moveDir(ctx context.Context, dir *types.Directory, newParentUUID string, overwrite bool) error {
	resp, err := api.DirExists(ctx, newParentUUID, dir.GetName())
	if err != nil {
		return fmt.Errorf("DirExists: %w", err)
	}
	if resp.Exists {
		if overwrite {
			err := api.Client.PostV3FileTrash(ctx, resp.UUID)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("directory already exists")
		}
	}

	err = api.Client.PostV3DirMove(ctx, dir.GetUUID(), newParentUUID)
	if err != nil {
		return fmt.Errorf("PostV3DirMove: %w", err)
	}
	dir.ParentUUID = newParentUUID
	return api.updateItemWithMaybeSharedParent(ctx, dir)
}

// MoveItem moves a file or directory to a new parent directory.
// If overwrite is true, it will replace any existing item with the same name.
// This operation requires a lock to prevent race conditions with other operations.
func (api *Filen) MoveItem(ctx context.Context, item types.NonRootFileSystemObject, newParentUUID string, overwrite bool) error {
	err := api.Lock(ctx)
	if err != nil {
		return err
	}
	defer api.Unlock()
	if dir, ok := item.(*types.Directory); ok {
		return api.moveDir(ctx, dir, newParentUUID, overwrite)
	} else if file, ok := item.(*types.File); ok {
		return api.moveFile(ctx, file, newParentUUID, overwrite)
	} else {
		return fmt.Errorf("unknown item type")
	}
}

// EmptyTrash permanently deletes all items in the trash.
// This operation cannot be undone.
func (api *Filen) EmptyTrash(ctx context.Context) error {
	return api.Client.PostV3TrashEmpty(ctx)
}

// GetUserInfo retrieves information about the current user,
// including account details, storage usage, and quotas.
func (api *Filen) GetUserInfo(ctx context.Context) (*client.V3UserInfoResponse, error) {
	return api.Client.GetV3UserInfo(ctx)
}

// GetDirSize returns the total size, file count, and folder count of a directory,
// including all its subdirectories and files.
func (api *Filen) GetDirSize(ctx context.Context, dir *types.Directory) (*client.V3DirSizeResponse, error) {
	return api.Client.PostV3DirSize(ctx, dir.GetUUID())
}

// DownloadToPath downloads a file from the cloud to the given local path.
// The file is first downloaded to a temporary file in the same directory,
// then renamed to the final path. If an error occurs during download or rename,
// the temporary file is removed.
func (api *Filen) DownloadToPath(ctx context.Context, file *types.File, downloadPath string) error {
	downloadDir := path.Dir(downloadPath)
	// needs to be removed or renamed
	f, err := os.CreateTemp(downloadDir, fmt.Sprintf("%s-download-*.tmp", file.Name))
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	fName := f.Name()
	downloader := api.GetDownloadReader(ctx, file)
	_, err = f.ReadFrom(downloader)
	errClose := f.Close()
	if err != nil {
		_ = os.Remove(fName)
		maybeErr := context.Cause(ctx)
		if maybeErr != nil {
			return fmt.Errorf("download file: %w", maybeErr)
		}
		return fmt.Errorf("download file: %w", err)
	}

	err = downloader.Close()
	if err != nil {
		_ = os.Remove(fName)
		return fmt.Errorf("close downloader: %w", err)
	}

	if errClose != nil {
		_ = os.Remove(fName)
		return fmt.Errorf("close file: %w", errClose)
	}
	// should be okay because the temp file is in the same directory
	err = os.Rename(f.Name(), downloadPath)
	if err != nil {
		_ = os.Remove(fName)
		return fmt.Errorf("rename file: %w", err)
	}
	return nil
}

// GetDownloadReader returns a reader which can be used to stream a file from the cloud.
// The returned io.ReadCloser should be closed after use to release resources.
// The reader handles decryption and integrity verification automatically.
func (api *Filen) GetDownloadReader(ctx context.Context, file *types.File) io.ReadCloser {
	return newChunkedReader(ctx, api, file)
}

// GetDownloadReaderWithOffset returns a reader which can be used to stream a file
// starting at the given offset, reading up to the specified limit.
// This is useful for range requests or partial downloads.
// The returned io.ReadCloser should be closed after use to release resources.
func (api *Filen) GetDownloadReaderWithOffset(ctx context.Context, file *types.File, offset int64, limit int64) io.ReadCloser {
	return newChunkedReaderWithOffset(ctx, api, file, offset, limit)
}

// UploadFromReader uploads a file to the cloud using the provided reader as the data source.
// The file metadata is taken from the IncompleteFile parameter.
// The function handles chunking, encryption, and verification automatically.
func (api *Filen) UploadFromReader(ctx context.Context, file *types.IncompleteFile, r io.Reader) (*types.File, error) {
	return api.UploadFile(ctx, file, r)
}

// updateFileMeta updates the metadata of a file on the server.
// This is an internal helper used by UpdateMeta.
func (api *Filen) updateFileMeta(ctx context.Context, file *types.File, metaEncrypted crypto.EncryptedString, nameHashed string) error {
	metaKey, err := file.EncryptionKey.ToMasterKey()
	if err != nil {
		return fmt.Errorf("encrypt name: %w", err)
	}
	nameEncrypted := metaKey.EncryptMeta(file.Name)
	return api.Client.PostV3FileMetadata(ctx, file.UUID, nameEncrypted, nameHashed, metaEncrypted)
}

// updateDirMeta updates the metadata of a directory on the server.
// This is an internal helper used by UpdateMeta.
func (api *Filen) updateDirMeta(ctx context.Context, dir *types.Directory, metaEncrypted crypto.EncryptedString, nameHashed string) error {
	return api.Client.PostV3DirMetadata(ctx, dir.UUID, nameHashed, metaEncrypted)
}

// UpdateMeta updates the metadata of a file or directory on the server.
// This operation requires a lock to prevent race conditions with other operations.
// It also updates search indexes and shared parent metadata.
func (api *Filen) UpdateMeta(ctx context.Context, item types.NonRootFileSystemObject) error {
	err := api.Lock(ctx)
	if err != nil {
		return err
	}
	defer api.Unlock()
	metaStr, err := item.GetMeta(api.FileEncryptionVersion)
	if err != nil {
		return fmt.Errorf("get meta: %w", err)
	}
	metaEncrypted := api.EncryptMeta(metaStr)

	nameHashed := api.HashFileName(item.GetName())

	if dir, ok := item.(*types.Directory); ok {
		err = api.updateDirMeta(ctx, dir, metaEncrypted, nameHashed)
	} else if file, ok := item.(*types.File); ok {
		err = api.updateFileMeta(ctx, file, metaEncrypted, nameHashed)
	} else {
		return fmt.Errorf("unknown item type")
	}
	if err != nil {
		return fmt.Errorf("update meta: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return api.updateMaybeSharedItem(gCtx, item) })
	g.Go(func() error { return api.updateSearchHashes(gCtx, item) })
	return g.Wait()
}

// Rename renames a file or directory.
// This uses UpdateMeta under the hood, but cleanly handles errors and
// only updates the name in memory if the server update is successful.
// If the operation fails, the original name is preserved.
func (api *Filen) Rename(ctx context.Context, item types.NonRootFileSystemObject, newName string) error {
	oldName := item.GetName()
	if dir, ok := item.(*types.Directory); ok {
		dir.Name = newName
		err := api.UpdateMeta(ctx, item)
		if err != nil {
			dir.Name = oldName
			return fmt.Errorf("update meta: %w", err)
		}
	} else if file, ok := item.(*types.File); ok {
		file.Name = newName
		err := api.UpdateMeta(ctx, item)
		if err != nil {
			file.Name = oldName
			return fmt.Errorf("update meta: %w", err)
		}
	} else {
		return fmt.Errorf("unknown item type")
	}
	return nil
}

// updateSearchHashes updates the search index for a file or directory.
// This is called automatically when items are created, renamed, or moved.
// It generates search hashes that enable encrypted search functionality.
func (api *Filen) updateSearchHashes(ctx context.Context, item types.NonRootFileSystemObject) error {
	var typ string
	if _, ok := item.(*types.Directory); ok {
		typ = "directory"
	} else if _, ok := item.(*types.File); ok {
		typ = "file"
	} else {
		return fmt.Errorf("unknown item type")
	}
	nameHashes := search.GenerateSearchIndexHashes(item.GetName(), api.HMACKey, item.GetUUID(), typ)
	return api.Client.PostV3SearchAdd(ctx, nameHashes)
}
