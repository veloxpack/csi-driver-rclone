package folders

import (
	"encoding/json"
	"time"
)

// FolderStatus represents the status filter for file and folder operations
// Possible values: EXISTS, TRASHED, DELETED, ALL
type FolderStatus string

const (
	StatusExists  FolderStatus = "EXISTS"
	StatusTrashed FolderStatus = "TRASHED"
	StatusDeleted FolderStatus = "DELETED"
	StatusAll     FolderStatus = "ALL"
)

// CreateFolderRequest is the payload for POST /drive/folders
type CreateFolderRequest struct {
	PlainName        string `json:"plainName"`
	ParentFolderUUID string `json:"parentFolderUuid"`
	ModificationTime string `json:"modificationTime"`
	CreationTime     string `json:"creationTime"`
}

// UserResumeData represents user information as returned by the API
type UserResumeData struct {
	Avatar   *string `json:"avatar"`
	Email    string  `json:"email"`
	Lastname *string `json:"lastname"`
	Name     string  `json:"name"`
	UUID     string  `json:"uuid"`
}

// Folder represents the response from POST/GET /drive/folders
type Folder struct {
	Type             string          `json:"type"`
	ID               int64           `json:"id"`
	ParentID         int64           `json:"parentId"`
	ParentUUID       string          `json:"parentUuid"`
	Name             string          `json:"name"`
	Parent           *string         `json:"parent"`
	Bucket           *string         `json:"bucket"`
	UserID           int64           `json:"userId"`
	User             *UserResumeData `json:"user"`
	EncryptVersion   string          `json:"encryptVersion"`
	Deleted          bool            `json:"deleted"`
	DeletedAt        *time.Time      `json:"deletedAt"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	UUID             string          `json:"uuid"`
	PlainName        string          `json:"plainName"`
	Size             int64           `json:"size"`
	Removed          bool            `json:"removed"`
	RemovedAt        *time.Time      `json:"removedAt"`
	CreationTime     time.Time       `json:"creationTime"`
	ModificationTime time.Time       `json:"modificationTime"`
	Status           string          `json:"status"`
}

// Thumbnail represents a file thumbnail
type Thumbnail struct {
	ID             json.Number `json:"id"`
	FileID         json.Number `json:"file_id"`
	MaxWidth       json.Number `json:"max_width"`
	MaxHeight      json.Number `json:"max_height"`
	Type           string      `json:"type"`
	Size           json.Number `json:"size"`
	BucketID       string      `json:"bucket_id"`
	BucketFile     string      `json:"bucket_file"`
	EncryptVersion string      `json:"encrypt_version"`
	URLObject      string      `json:"urlObject,omitempty"`
}

// ShareLink represents a public share link for a file or folder
type ShareLink struct {
	ID             string      `json:"id"`
	Token          string      `json:"token"`
	Mnemonic       string      `json:"mnemonic"`
	User           any         `json:"user"` // Type not specified in SDK (any)
	Item           any         `json:"item"` // Type not specified in SDK (any)
	EncryptionKey  string      `json:"encryptionKey"`
	Bucket         string      `json:"bucket"`
	ItemToken      string      `json:"itemToken"`
	IsFolder       bool        `json:"isFolder"`
	Views          json.Number `json:"views"`
	TimesValid     json.Number `json:"timesValid"`
	Active         bool        `json:"active"`
	CreatedAt      string      `json:"createdAt"`
	UpdatedAt      string      `json:"updatedAt"`
	FileSize       json.Number `json:"fileSize"`
	HashedPassword *string     `json:"hashed_password"`
	Code           string      `json:"code"`
}

// File represents the response object for files in a folder
// under GET /drive/folders/content/{uuid}/files
type File struct {
	ID               int64           `json:"id"`
	FileID           string          `json:"fileId"`
	UUID             string          `json:"uuid"`
	Name             string          `json:"name"`
	PlainName        string          `json:"plainName"`
	Type             string          `json:"type"`
	FolderID         json.Number     `json:"folderId"`
	FolderUUID       string          `json:"folderUuid"`
	Folder           *string         `json:"folder"`
	Bucket           string          `json:"bucket"`
	UserID           json.Number     `json:"userId"`
	User             *UserResumeData `json:"user"`
	EncryptVersion   string          `json:"encryptVersion"`
	Size             json.Number     `json:"size"`
	Deleted          bool            `json:"deleted"`
	DeletedAt        *time.Time      `json:"deletedAt"`
	Removed          bool            `json:"removed"`
	RemovedAt        *time.Time      `json:"removedAt"`
	Shares           []ShareLink     `json:"shares"`
	Sharings         []any           `json:"sharings"`
	Thumbnails       []Thumbnail     `json:"thumbnails"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	CreationTime     time.Time       `json:"creationTime"`
	ModificationTime time.Time       `json:"modificationTime"`
	Status           string          `json:"status"`
}

// ListOptions defines common pagination and sorting parameters
// for list endpoints.
type ListOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
}

// TreeNode is a recursive structure representing a folder, its files, and its child folders.
type TreeNode struct {
	Folder
	Files    []File     `json:"files"`
	Children []TreeNode `json:"children"`
}
