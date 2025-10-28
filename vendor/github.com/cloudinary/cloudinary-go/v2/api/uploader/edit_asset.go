package uploader

import (
	"context"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	upload   api.EndPoint = "upload"
	destroy  api.EndPoint = "destroy"
	rename   api.EndPoint = "rename"
	explicit api.EndPoint = "explicit"
	metadata api.EndPoint = "metadata"
)

// DestroyParams are the parameters for Destroy.
type DestroyParams struct {
	PublicID     string `json:"public_id,omitempty"`
	Type         string `json:"type,omitempty"`
	ResourceType string `json:"-"`
	Invalidate   *bool  `json:"invalidate,omitempty"`
}

// Destroy immediately and permanently deletes a single asset from your Cloudinary account.
//
// Backed up assets are not deleted, and any assets and transformed assets already downloaded by visitors to your
// website might still be accessible through cached copies on the CDN. You can invalidate any cached copies on the
// CDN with the `Invalidate` parameter.
//
// https://cloudinary.com/documentation/image_upload_api_reference#destroy_method
func (u *API) Destroy(ctx context.Context, params DestroyParams) (*DestroyResult, error) {
	res := &DestroyResult{}
	err := u.callUploadAPI(ctx, destroy, params, res)

	return res, err
}

// DestroyResult is the result of Destroy.
type DestroyResult struct {
	Result   string        `json:"result"`
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// RenameParams are the parameters for Rename.
type RenameParams struct {
	FromPublicID string `json:"from_public_id,omitempty"`
	ToPublicID   string `json:"to_public_id,omitempty"`
	Type         string `json:"type,omitempty"`
	ToType       string `json:"to_type,omitempty"`
	ResourceType string `json:"-"`
	Overwrite    *bool  `json:"overwrite,omitempty"`
	Invalidate   *bool  `json:"invalidate,omitempty"`
}

// Rename renames the specified asset in your Cloudinary account.
//
// The existing URLs of renamed assets and their associated derived resources are no longer valid, although any
// assets and transformed assets already downloaded by visitors to your website might still be accessible through
// cached copies on the CDN. You can invalidate any cached copies on the CDN with the `invalidate` parameter.
//
// https://cloudinary.com/documentation/image_upload_api_reference#rename_method
func (u *API) Rename(ctx context.Context, params RenameParams) (*RenameResult, error) {
	res := &RenameResult{}
	err := u.callUploadAPI(ctx, rename, params, res)

	return res, err
}

// RenameResult is the result of Rename.
type RenameResult struct {
	api.BriefAssetResult
	Error interface{} `json:"error,omitempty"`
}

// ExplicitParams are the parameters for Explicit.
type ExplicitParams = UploadParams

// Explicit applies actions to already uploaded assets.
//
// https://cloudinary.com/documentation/image_upload_api_reference#explicit_method
func (u *API) Explicit(ctx context.Context, params ExplicitParams) (*ExplicitResult, error) {
	res := &ExplicitResult{}
	err := u.callUploadAPI(ctx, explicit, params, res)

	return res, err
}

// ExplicitResult is the result of Explicit.
type ExplicitResult struct {
	UploadResult
}

// UpdateMetadataParams are the parameters for UpdateMetadata.
type UpdateMetadataParams struct {
	PublicIDs    []string      `json:"public_ids"`
	Metadata     api.CldAPIMap `json:"metadata"`
	Type         string        `json:"type,omitempty"`
	ResourceType string        `json:"-"`
}

// UpdateMetadata populates metadata fields with the given values. Existing values will be overwritten.
//
// Any metadata-value pairs given are merged with any existing metadata-value pairs
// (an empty value for an existing metadata field clears the value).
//
// https://cloudinary.com/documentation/image_upload_api_reference#metadata_method
func (u *API) UpdateMetadata(ctx context.Context, params UpdateMetadataParams) (*UpdateMetadataResult, error) {
	res := &UpdateMetadataResult{}
	err := u.callUploadAPI(ctx, metadata, params, res)

	return res, err
}

// UpdateMetadataResult is the result of UpdateMetadata.
type UpdateMetadataResult struct {
	PublicIDs []string    `json:"public_ids"`
	Error     interface{} `json:"error,omitempty"`
}
