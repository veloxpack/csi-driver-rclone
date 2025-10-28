package admin

// Enables you to manage the folders in your account or cloud.
//
// https://cloudinary.com/documentation/admin_api#folders
import (
	"context"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	folders api.EndPoint = "folders"
)

// RootFoldersParams are the parameters for RootFolders.
type RootFoldersParams struct {
	MaxResults int    `json:"max_results,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// RootFolders lists all root folders.
//
// https://cloudinary.com/documentation/admin_api#get_root_folders
func (a *API) RootFolders(ctx context.Context, params RootFoldersParams) (*FoldersResult, error) {
	res := &FoldersResult{}
	_, err := a.get(ctx, folders, params, res)

	return res, err
}

// FoldersResult is the result of RootFolders, SubFolders.
type FoldersResult struct {
	Folders    []FolderResult `json:"folders"`
	TotalCount int            `json:"total_count"`
	NextCursor string         `json:"next_cursor"`
	Error      api.ErrorResp  `json:"error,omitempty"`
}

// FolderResult contains details of a single folder.
type FolderResult struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SubFoldersParams are the parameters for SubFolders.
type SubFoldersParams struct {
	Folder     string `json:"-"`
	MaxResults int    `json:"max_results,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// SubFolders lists sub-folders.
//
// Returns the name and path of all the sub-folders of a specified parent folder. Limited to 2000 results.
//
// https://cloudinary.com/documentation/admin_api#get_subfolders
func (a *API) SubFolders(ctx context.Context, params SubFoldersParams) (*FoldersResult, error) {
	res := &FoldersResult{}
	_, err := a.get(ctx, api.BuildPath(folders, params.Folder), params, res)

	return res, err
}

// CreateFolderParams are the parameters for CreateFolder.
type CreateFolderParams struct {
	Folder string `json:"-"` // The full path of the new folder to create.
}

// CreateFolder creates a new empty folder.
//
// https://cloudinary.com/documentation/admin_api#create_folder
func (a *API) CreateFolder(ctx context.Context, params CreateFolderParams) (*CreateFolderResult, error) {
	res := &CreateFolderResult{}
	_, err := a.post(ctx, api.BuildPath(folders, params.Folder), params, res)

	return res, err
}

// CreateFolderResult is the result of CreateFolder.
type CreateFolderResult struct {
	Success bool          `json:"success"`
	Path    string        `json:"path"`
	Name    string        `json:"name"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// RenameFolderParams are the parameters for RenameFolder.
type RenameFolderParams struct {
	FromPath string `json:"-"`         // The full path of the existing folder.
	ToPath   string `json:"to_folder"` // The full path of the new folder.
}

// RenameFolder renames an existing asset folder.
//
// https://cloudinary.com/documentation/admin_api#update_folder
func (a *API) RenameFolder(ctx context.Context, params RenameFolderParams) (*RenameFolderResult, error) {
	res := &RenameFolderResult{}
	_, err := a.put(ctx, api.BuildPath(folders, params.FromPath), params, res)

	return res, err
}

type RenameFolderResult struct {
	From  FolderResult  `json:"from"`
	To    FolderResult  `json:"to"`
	Error api.ErrorResp `json:"error,omitempty"`
}

// DeleteFolderParams are the parameters for DeleteFolder.
type DeleteFolderParams struct {
	Folder string `json:"-"` // The full path of the empty folder to delete.
}

// DeleteFolder deletes an empty folder.
//
// The specified folder cannot contain any assets, but can have empty descendant sub-folders.
//
// https://cloudinary.com/documentation/admin_api#delete_folder
func (a *API) DeleteFolder(ctx context.Context, params DeleteFolderParams) (*DeleteFolderResult, error) {
	res := &DeleteFolderResult{}
	_, err := a.delete(ctx, api.BuildPath(folders, params.Folder), params, res)

	return res, err
}

// DeleteFolderResult is the result of DeleteFolder.
type DeleteFolderResult struct {
	Deleted []string      `json:"deleted"`
	Error   api.ErrorResp `json:"error,omitempty"`
}
