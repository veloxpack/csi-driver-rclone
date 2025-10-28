package admin

// Enables you to manage the upload mappings.
//
// https://cloudinary.com/documentation/admin_api#upload_mappings
import (
	"context"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	uploadMappings api.EndPoint = "upload_mappings"
)

// ListUploadMappingsParams are the parameters for ListUploadMappings.
type ListUploadMappingsParams struct {
	MaxResults int    `json:"max_results,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// ListUploadMappings lists upload mappings by folder and its mapped template (URL).
//
// https://cloudinary.com/documentation/admin_api#get_upload_mappings
func (a *API) ListUploadMappings(ctx context.Context, params ListUploadMappingsParams) (*ListUploadMappingsResult, error) {
	res := &ListUploadMappingsResult{}
	_, err := a.get(ctx, uploadMappings, params, res)

	return res, err
}

// ListUploadMappingsResult is the result of ListUploadMappings.
type ListUploadMappingsResult struct {
	Mappings []UploadMapping `json:"mappings"`
	Error    api.ErrorResp   `json:"error,omitempty"`
}

// UploadMapping represents a single upload mapping.
type UploadMapping struct {
	Folder   string `json:"folder"`
	Template string `json:"template"`
}

// GetUploadMappingParams are the parameters for GetUploadMapping.
type GetUploadMappingParams struct {
	Folder string `json:"folder"` // The name of the upload mapping folder.
}

// GetUploadMapping returns the details of the specified upload mapping.
//
// Retrieve the mapped template (URL) of a specified upload mapping folder.
//
// https://cloudinary.com/documentation/admin_api#get_the_details_of_a_single_upload_mapping
func (a *API) GetUploadMapping(ctx context.Context, params GetUploadMappingParams) (*GetUploadMappingResult, error) {
	res := &GetUploadMappingResult{}
	_, err := a.get(ctx, api.BuildPath(uploadMappings), params, res)

	return res, err
}

// GetUploadMappingResult is the result of GetUploadMapping.
type GetUploadMappingResult struct {
	Folder   string        `json:"folder"`
	Template string        `json:"template"`
	Error    api.ErrorResp `json:"error,omitempty"`
}

// CreateUploadMappingParams are the parameters for CreateUploadMapping.
type CreateUploadMappingParams struct {
	Folder   string `json:"folder"`   // The name of the folder to map.
	Template string `json:"template"` // The URL to be mapped to the folder.
}

// CreateUploadMapping creates a new upload mapping.
//
// https://cloudinary.com/documentation/admin_api#create_an_upload_mapping
func (a *API) CreateUploadMapping(ctx context.Context, params CreateUploadMappingParams) (*CreateUploadMappingResult, error) {
	res := &CreateUploadMappingResult{}
	_, err := a.post(ctx, api.BuildPath(uploadMappings), params, res)

	return res, err
}

// CreateUploadMappingResult is the result of CreateUploadMapping.
type CreateUploadMappingResult struct {
	Message string        `json:"message"`
	Folder  string        `json:"folder"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// UploadMappingResult is the result of UpdateUploadMapping, DeleteUploadMapping.
type UploadMappingResult struct {
	Message string        `json:"message"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// UpdateUploadMappingParams are the parameters for UpdateUploadMapping.
type UpdateUploadMappingParams struct {
	Folder   string `json:"folder"` // The name of the upload mapping folder to remap.
	Template string `json:"template"`
}

// UpdateUploadMapping updates an existing upload mapping with a new template (URL).
//
// https://cloudinary.com/documentation/admin_api#update_an_upload_mapping
func (a *API) UpdateUploadMapping(ctx context.Context, params UpdateUploadMappingParams) (*UploadMappingResult, error) {
	res := &UploadMappingResult{}
	_, err := a.put(ctx, api.BuildPath(uploadMappings), params, res)

	return res, err
}

// DeleteUploadMappingParams are the parameters for DeleteUploadMapping.
type DeleteUploadMappingParams struct {
	Folder string `json:"folder"` // The name of the upload mapping folder to delete.
}

// DeleteUploadMapping deletes an upload mapping.
//
// https://cloudinary.com/documentation/admin_api#delete_an_upload_mapping
func (a *API) DeleteUploadMapping(ctx context.Context, params DeleteUploadMappingParams) (*UploadMappingResult, error) {
	res := &UploadMappingResult{}
	_, err := a.delete(ctx, api.BuildPath(uploadMappings), params, res)

	return res, err
}
