package uploader

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	tags api.EndPoint = "tags"
)

type tagCommand string

const (
	addTag        tagCommand = "add"
	removeTag     tagCommand = "remove"
	replaceTag    tagCommand = "replace"
	removeAllTags tagCommand = "remove_all"
)

// AddTagParams are the parameters for
type AddTagParams struct {
	Tag          string   `json:"tag,omitempty"` // The name of the tag to add.
	PublicIDs    []string `json:"public_ids"`    // The public IDs of the assets to add the tag to.
	Type         string   `json:"type,omitempty"`
	ResourceType string   `json:"-"`
}

// AddTag adds a tag to the assets specified.
//
// https://cloudinary.com/documentation/image_upload_api_reference#tags_method
func (u *API) AddTag(ctx context.Context, params AddTagParams) (*AddTagResult, error) {
	res := &AddTagResult{}
	err := u.callTagsAPI(ctx, addTag, params, res)

	return res, err
}

// AddTagResult is the result of AddTag.
type AddTagResult struct {
	TagResult
}

// RemoveTagParams are the parameters for RemoveTag.
type RemoveTagParams struct {
	Tag          string   `json:"tag,omitempty"` // The name of the tag to remove.
	PublicIDs    []string `json:"public_ids"`    // The public IDs of the assets to remove the tags from.
	Type         string   `json:"type,omitempty"`
	ResourceType string   `json:"-"`
}

// RemoveTag removes a tag from the assets specified.
//
// https://cloudinary.com/documentation/image_upload_api_reference#tags_method
func (u *API) RemoveTag(ctx context.Context, params RemoveTagParams) (*RemoveTagResult, error) {
	res := &RemoveTagResult{}
	err := u.callTagsAPI(ctx, removeTag, params, res)

	return res, err
}

// RemoveTagResult is the result of RemoveTag.
type RemoveTagResult struct {
	TagResult
}

// RemoveAllTagsParams are the parameters for RemoveAllTags.
type RemoveAllTagsParams struct {
	PublicIDs    []string `json:"public_ids"` // The public IDs of the assets to remove all tags from.
	Type         string   `json:"type,omitempty"`
	ResourceType string   `json:"-"`
}

// RemoveAllTags removes all tags from the assets specified.
//
// https://cloudinary.com/documentation/image_upload_api_reference#tags_method
func (u *API) RemoveAllTags(ctx context.Context, params RemoveAllTagsParams) (*RemoveAllTagsResult, error) {
	res := &RemoveAllTagsResult{}
	err := u.callTagsAPI(ctx, removeAllTags, params, res)

	return res, err
}

// RemoveAllTagsResult is the result of RemoveAllTags.
type RemoveAllTagsResult struct {
	TagResult
}

// ReplaceTagParams are the parameters for ReplaceTag.
type ReplaceTagParams struct {
	Tag          string   `json:"tag"`        // The new tag with which to replace the existing tags.
	PublicIDs    []string `json:"public_ids"` // The public IDs of the assets to replace the tags of.
	Type         string   `json:"type,omitempty"`
	ResourceType string   `json:"-"`
}

// ReplaceTag replaces all existing tags on the assets specified with the tag specified.
//
// https://cloudinary.com/documentation/image_upload_api_reference#tags_method
func (u *API) ReplaceTag(ctx context.Context, params ReplaceTagParams) (*ReplaceTagResult, error) {
	res := &ReplaceTagResult{}
	err := u.callTagsAPI(ctx, replaceTag, params, res)

	return res, err
}

// ReplaceTagResult  is the result of ReplaceTag.
type ReplaceTagResult struct {
	TagResult
}

// TagResult represents the tag result.
type TagResult struct {
	PublicIDs []string      `json:"public_ids"` // The public IDs of the assets that were affected.
	Error     api.ErrorResp `json:"error,omitempty"`
	Response  interface{}
}

// callTagsAPI is an internal method that is used to call to the tags API.
func (u *API) callTagsAPI(ctx context.Context, command tagCommand, requestParams interface{}, result interface{}) error {
	formParams, err := api.StructToParams(requestParams)
	if err != nil {
		return err
	}

	formParams.Add("command", fmt.Sprintf("%v", command))

	return u.callUploadAPIWithParams(ctx, api.BuildPath(getAssetType(requestParams), tags), formParams, result)
}
