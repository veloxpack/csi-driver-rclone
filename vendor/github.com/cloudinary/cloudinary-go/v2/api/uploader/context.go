package uploader

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	cldContext api.EndPoint = "context"
)

type contextCommand string

const (
	addContext       contextCommand = "add"
	removeAllContext contextCommand = "remove_all"
)

// AddContextParams are the parameters for AddContext.
type AddContextParams struct {
	Context      api.CldAPIMap   `json:"context,omitempty"`
	PublicIDs    api.CldAPIArray `json:"public_ids"`
	Type         string          `json:"type,omitempty"`
	ResourceType string          `json:"-"`
}

// AddContext adds context metadata as key-value pairs to the the specified assets.
//
// https://cloudinary.com/documentation/image_upload_api_reference#context_method
func (u *API) AddContext(ctx context.Context, params AddContextParams) (*AddContextResult, error) {
	res := &AddContextResult{}
	err := u.callContextAPI(ctx, addContext, params, res)

	return res, err
}

// AddContextResult is the result of AddContext.
type AddContextResult = ContextResult

// ContextResult is the result of Context APIs.
type ContextResult struct {
	PublicIDs []string      `json:"public_ids"`
	Error     api.ErrorResp `json:"error,omitempty"`
	Response  interface{}
}

// RemoveAllContextParams struct
type RemoveAllContextParams struct {
	PublicIDs    api.CldAPIArray `json:"public_ids"`
	Type         string          `json:"type,omitempty"`
	ResourceType string          `json:"-"`
}

// RemoveAllContext removes all context metadata from the specified assets.
//
// https://cloudinary.com/documentation/image_upload_api_reference#context_method
func (u *API) RemoveAllContext(ctx context.Context, params RemoveAllContextParams) (*RemoveAllContextResult, error) {
	res := &RemoveAllContextResult{}
	err := u.callContextAPI(ctx, removeAllContext, params, res)

	return res, err
}

// RemoveAllContextResult is the result of RemoveAllContext.
type RemoveAllContextResult = ContextResult

// callContextAPI is an internal method that is used to call to the context API.
func (u *API) callContextAPI(ctx context.Context, command contextCommand, requestParams interface{}, result interface{}) error {
	formParams, err := api.StructToParams(requestParams)
	if err != nil {
		return err
	}

	formParams.Add("command", fmt.Sprintf("%v", command))

	return u.callUploadAPIWithParams(ctx, api.BuildPath(getAssetType(requestParams), cldContext), formParams, result)
}
