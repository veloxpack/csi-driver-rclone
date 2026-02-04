// Package types provides the core data types used throughout the Filen SDK.
// It defines file system objects, metadata structures, and utility types that
// represent the Filen cloud storage system.
package types

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/io"
	"github.com/google/uuid"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// IncompleteFile represents a file that has not been fully uploaded to Filen.
// It contains the metadata and encryption information needed to continue or
// complete the upload process.
type IncompleteFile struct {
	UUID          string               // The UUID of the cloud item
	Name          string               // The file name
	MimeType      string               // The MIME type of the file
	EncryptionKey crypto.EncryptionKey // The key used to encrypt the file data
	Created       time.Time            // When the file was created
	LastModified  time.Time            // When the file was last modified
	ParentUUID    string               // The Directory.UUID of the file's parent directory
}

// NewIncompleteFile creates a new incomplete file using the passed values.
// If mimeType is empty, the function will attempt to determine it from the file extension.
// Returns an error if the filename contains invalid characters.
func NewIncompleteFile(v crypto.FileEncryptionVersion, name string, mimeType string, created time.Time, lastModified time.Time, parent DirectoryInterface) (*IncompleteFile, error) {
	if strings.ContainsRune(name, '/') {
		return nil, fmt.Errorf("invalid file name")
	}
	key, err := crypto.MakeNewFileKey(v)
	if err != nil {
		return nil, fmt.Errorf("make new file key: %w", err)
	}
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(name))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	} else {
		mimeType, _, _ = strings.Cut(mimeType, ";")
	}

	return &IncompleteFile{
		UUID:          uuid.NewString(),
		Name:          name,
		MimeType:      mimeType,
		EncryptionKey: *key,
		Created:       created.Round(time.Millisecond),
		LastModified:  lastModified.Round(time.Millisecond),
		ParentUUID:    parent.GetUUID(),
	}, nil
}

// NewIncompleteFileFromOSFile creates a new IncompleteFile from a local file system file.
// It extracts metadata like creation time and modification time from the file.
func NewIncompleteFileFromOSFile(v crypto.FileEncryptionVersion, osFile *os.File, parent DirectoryInterface) (*IncompleteFile, error) {
	fileStat, err := osFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	created := io.GetCreationTime(fileStat)
	return NewIncompleteFile(v, filepath.Base(osFile.Name()), "", created, fileStat.ModTime(), parent)
}

// NewFromBase creates a new incomplete file based on an existing one.
// The new incomplete file will have a new UUID and encryption key,
// but will retain all other properties from the original.
// This is useful for retrying uploads or creating copies.
func (file *IncompleteFile) NewFromBase(v crypto.FileEncryptionVersion) (*IncompleteFile, error) {
	key, err := crypto.MakeNewFileKey(v)
	if err != nil {
		return nil, fmt.Errorf("make new file key: %w", err)
	}

	return &IncompleteFile{
		UUID:          uuid.NewString(),
		Name:          file.Name,
		MimeType:      file.MimeType,
		EncryptionKey: *key,
		Created:       file.Created,
		LastModified:  file.LastModified,
		ParentUUID:    file.ParentUUID,
	}, nil
}

// GetRawMeta returns the file metadata without the size and hash fields,
// since those cannot be calculated until the file is fully uploaded.
// This provides a base FileMetadata structure that can be completed later.
func (file *IncompleteFile) GetRawMeta(v crypto.FileEncryptionVersion) FileMetadata {
	return FileMetadata{
		Name:         file.Name,
		Size:         0,
		MimeType:     file.MimeType,
		Key:          file.EncryptionKey.ToStringWithVersion(v),
		LastModified: IntFromMaybeString(file.LastModified.UnixMilli()),
		Created:      int(file.Created.UnixMilli()),
		Hash:         "",
	}
}

func (file *IncompleteFile) SetMimeType(mimeType string) {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	} else {
		mimeType, _, _ = strings.Cut(mimeType, ";")
	}
	file.MimeType = mimeType
}

type IntFromMaybeString int

func (i *IntFromMaybeString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == "" {
		*i = 0 // Default value
		return nil
	}

	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		var stringValue string
		if err := json.Unmarshal(data, &stringValue); err != nil {
			return err
		} else {
			value, err = strconv.Atoi(stringValue)
			if err != nil {
				return fmt.Errorf("couldn't unmarshal IntFromMaybeString: %w", err)
			}
		}
	}
	*i = IntFromMaybeString(value)
	return nil
}

// FileMetadata contains the metadata of a file in the Filen cloud.
// This structure is encrypted before being uploaded to the server
// to ensure end-to-end encryption of file details.
type FileMetadata struct {
	Name         string             `json:"name"`         // The file name
	Size         int64              `json:"size"`         // The file size in bytes
	MimeType     string             `json:"mime"`         // The MIME type
	Key          string             `json:"key"`          // The encryption key as a string
	LastModified IntFromMaybeString `json:"lastModified"` // Last modification timestamp in milliseconds
	Created      int                `json:"creation"`     // Creation timestamp in milliseconds
	Hash         string             `json:"blake3"`       // The file's Blake3 hash
}

// File represents a complete file in the Filen cloud storage.
// It extends IncompleteFile with additional properties that are
// only available once a file is fully uploaded.
type File struct {
	IncompleteFile                              // Embedded IncompleteFile with base properties
	Size           int64                        // The file size in bytes
	Favorited      bool                         // Whether the file is marked as a favorite
	Region         string                       // The file's storage region
	Bucket         string                       // The file's storage bucket
	Chunks         int64                        // How many 1 MiB chunks the file is partitioned into
	Hash           string                       // The file's Blake3 hash
	Version        crypto.FileEncryptionVersion // The crypto.FileEncryptionVersion version used to encrypt the file
}

// DirColor represents the color assigned to a directory in the Filen UI.
// This is primarily a visual marker in the user interface.
type DirColor string

// Directory color constants
const (
	DirColorDefault DirColor = ""       // Default color (no specific color)
	DirColorBlue    DirColor = "blue"   // Blue color
	DirColorGreen   DirColor = "green"  // Green color
	DirColorPurple  DirColor = "purple" // Purple color
	DirColorRed     DirColor = "red"    // Red color
	DirColorGray    DirColor = "gray"   // Gray color
)

// DirectoryMetaData contains the metadata of a directory.
// This structure is encrypted before being uploaded to the server
// to ensure end-to-end encryption of directory details.
type DirectoryMetaData struct {
	Name     string `json:"name"`     // The directory name
	Creation int    `json:"creation"` // Creation timestamp in seconds
}

// Directory represents a directory in the Filen cloud storage.
type Directory struct {
	UUID       string    // The UUID of the cloud item
	Name       string    // The directory name
	ParentUUID string    // The Directory.UUID of the directory's parent (empty for root)
	Color      DirColor  // The color assigned to the directory
	Created    time.Time // When the directory was created
	Favorited  bool      // Whether the directory is marked as a favorite
}

// RootDirectory represents the root directory of a user's Filen cloud storage.
// It has special properties compared to regular directories.
type RootDirectory struct {
	UUID string // The UUID of the root directory
}

// NewRootDirectory creates a new root directory with the specified UUID.
func NewRootDirectory(uuid string) RootDirectory {
	return RootDirectory{UUID: uuid}
}

// FileSystemObject is an interface implemented by both files and directories.
// It provides common methods to access basic properties of any file system object.
type FileSystemObject interface {
	// GetUUID returns the UUID of the cloud item
	GetUUID() string

	// GetName returns the name of the cloud item
	GetName() string

	// GetParent returns the Directory.UUID of the cloud item's parent directory
	GetParent() string
}

// NonRootFileSystemObject extends FileSystemObject for items that aren't the root directory.
// This allows access to metadata, which the root directory doesn't have.
type NonRootFileSystemObject interface {
	FileSystemObject

	// GetMeta returns the metadata of the cloud item
	// already marshalled into a JSON string
	GetMeta(fileEncryptionVersion crypto.FileEncryptionVersion) (string, error)
}

// DirectoryInterface is an interface for directories.
// It generalizes across both regular directories and the root directory.
type DirectoryInterface interface {
	FileSystemObject

	// IsRoot returns whether the directory is the root directory
	IsRoot() bool
}

// GetUUID implements FileSystemObject.GetUUID for File.
func (file File) GetUUID() string {
	return file.IncompleteFile.UUID
}

// GetName implements FileSystemObject.GetName for File.
func (file File) GetName() string {
	return file.Name
}

// GetParent implements FileSystemObject.GetParent for File.
func (file File) GetParent() string {
	return file.ParentUUID
}

// GetMeta implements NonRootFileSystemObject.GetMeta for File.
// It returns the file's metadata as a JSON string, ready for encryption.
func (file File) GetMeta(v crypto.FileEncryptionVersion) (string, error) {
	meta := file.GetRawMeta(v)
	meta.Size = file.Size
	meta.Hash = file.Hash
	metaStr, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal file metadata: %w", err)
	}
	return string(metaStr), nil
}

// GetUUID implements FileSystemObject.GetUUID for Directory.
func (dir Directory) GetUUID() string {
	return dir.UUID
}

// GetName implements FileSystemObject.GetName for Directory.
func (dir Directory) GetName() string {
	return dir.Name
}

// GetParent implements FileSystemObject.GetParent for Directory.
func (dir Directory) GetParent() string {
	return dir.ParentUUID
}

// IsRoot implements DirectoryInterface.IsRoot for Directory.
// Regular directories are never the root, so this always returns false.
func (dir Directory) IsRoot() bool {
	return false
}

// GetMeta implements NonRootFileSystemObject.GetMeta for Directory.
// It returns the directory's metadata as a JSON string, ready for encryption.
func (dir Directory) GetMeta(v crypto.FileEncryptionVersion) (string, error) {
	meta := DirectoryMetaData{
		Name:     dir.Name,
		Creation: int(dir.Created.Unix()),
	}
	metaStr, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal directory metadata: %w", err)
	}
	return string(metaStr), nil
}

// GetUUID implements FileSystemObject.GetUUID for RootDirectory.
func (root RootDirectory) GetUUID() string {
	return root.UUID
}

// GetName implements FileSystemObject.GetName for RootDirectory.
// The root directory has no name, so this returns an empty string.
func (root RootDirectory) GetName() string {
	return ""
}

// GetParent implements FileSystemObject.GetParent for RootDirectory.
// The root directory has no parent, so this returns an empty string.
func (root RootDirectory) GetParent() string {
	return ""
}

// IsRoot implements DirectoryInterface.IsRoot for RootDirectory.
// This always returns true since this is the root directory.
func (root RootDirectory) IsRoot() bool {
	return true
}

// CtxMutex is a mutex implementation that can be canceled through a context.Context.
// Unlike sync.Mutex, CtxMutex allows for cancellation and timeout through context.
type CtxMutex struct {
	channel chan struct{} // Channel for mutex operations
}

// NewCtxMutex returns a new CtxMutex.
// The returned mutex is initially unlocked.
func NewCtxMutex() CtxMutex {
	return CtxMutex{
		channel: make(chan struct{}, 1),
	}
}

// Lock attempts to lock the mutex. It blocks until the mutex is available or the
// provided context is canceled.
//
// If the context is canceled before acquiring the lock, Lock returns the error from
// context.Cause(ctx). Otherwise, it returns nil when the lock is acquired.
func (m *CtxMutex) Lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case m.channel <- struct{}{}:
		return nil
	}
}

// BlockUntilLock blocks until the mutex can be locked.
// This method will wait indefinitely until the mutex becomes available,
// with no option for cancellation.
func (m *CtxMutex) BlockUntilLock() {
	select {
	case m.channel <- struct{}{}:
		return
	}
}

// MustLock attempts to lock the mutex without blocking.
// If the mutex is already locked, it will panic with "locking locked mutex".
func (m *CtxMutex) MustLock() {
	select {
	case m.channel <- struct{}{}:
		return
	default:
		panic("locking locked mutex")
	}
}

// Unlock releases the mutex.
// If the mutex is not currently locked, it will panic with "unlocking unlocked mutex".
func (m *CtxMutex) Unlock() {
	select {
	case <-m.channel:
		return
	default:
		panic("unlocking unlocked mutex")
	}
}
