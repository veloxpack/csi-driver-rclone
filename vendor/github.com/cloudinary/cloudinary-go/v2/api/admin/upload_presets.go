package admin

// Enables you to manage upload presets.
//
// https://cloudinary.com/documentation/admin_api#upload_presets
import (
	"context"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const (
	uploadPresets api.EndPoint = "upload_presets"
)

// ListUploadPresetsParams are the parameters for ListUploadPresets.
type ListUploadPresetsParams struct {
	MaxResults int    `json:"max_results,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// ListUploadPresets lists existing upload presets.
//
// https://cloudinary.com/documentation/admin_api#get_upload_presets
func (a *API) ListUploadPresets(ctx context.Context, params ListUploadPresetsParams) (*ListUploadPresetsResult, error) {
	res := &ListUploadPresetsResult{}
	_, err := a.get(ctx, uploadPresets, params, res)

	return res, err
}

// ListUploadPresetsResult is the result of ListUploadPresets.
type ListUploadPresetsResult struct {
	Presets []UploadPreset `json:"presets"`
	Error   api.ErrorResp  `json:"error,omitempty"`
}

// UploadPreset represents the details of the upload preset.
type UploadPreset struct {
	Name     string      `json:"name"`
	Unsigned bool        `json:"unsigned"`
	Settings interface{} `json:"settings"`
}

// GetUploadPresetParams are the parameters for GetUploadPreset.
type GetUploadPresetParams struct {
	Name       string `json:"-"`
	MaxResults int    `json:"max_results,omitempty"`
}

// GetUploadPreset retrieves the details of the specified upload preset.
//
// https://cloudinary.com/documentation/admin_api#get_the_details_of_a_single_upload_preset
func (a *API) GetUploadPreset(ctx context.Context, params GetUploadPresetParams) (*GetUploadPresetResult, error) {
	res := &GetUploadPresetResult{}
	_, err := a.get(ctx, api.BuildPath(uploadPresets, params.Name), params, res)

	return res, err
}

// GetUploadPresetResult is the result of GetUploadPreset.
type GetUploadPresetResult struct {
	Name     string        `json:"name"`
	Unsigned bool          `json:"unsigned"`
	Settings interface{}   `json:"settings"`
	Error    api.ErrorResp `json:"error,omitempty"`
}

// CreateUploadPresetParams are the parameters for CreateUploadPreset.
type CreateUploadPresetParams struct {
	Name             string `json:"name,omitempty"`
	Unsigned         *bool  `json:"unsigned,omitempty"`
	DisallowPublicID *bool  `json:"disallow_public_id,omitempty"`
	Live             *bool  `json:"live,omitempty"`
	uploader.UploadParams
}

// CreateUploadPreset creates a new upload preset.
//
// https://cloudinary.com/documentation/admin_api#create_an_upload_preset
func (a *API) CreateUploadPreset(ctx context.Context, params CreateUploadPresetParams) (*CreateUploadPresetResult, error) {
	res := &CreateUploadPresetResult{}
	_, err := a.post(ctx, api.BuildPath(uploadPresets), params, res)

	return res, err
}

// CreateUploadPresetResult is the result of CreateUploadPreset.
type CreateUploadPresetResult struct {
	Message string        `json:"message"`
	Name    string        `json:"name"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// UpdateUploadPresetParams are the parameters for UpdateUploadPreset.
type UpdateUploadPresetParams struct {
	Name             string `json:"name"`
	Unsigned         *bool  `json:"unsigned,omitempty"`
	DisallowPublicID *bool  `json:"disallow_public_id,omitempty"`
	Live             *bool  `json:"live,omitempty"`
	uploader.UploadParams
}

// UpdateUploadPreset updates the specified upload preset.
//
// https://cloudinary.com/documentation/admin_api#update_an_upload_preset
func (a *API) UpdateUploadPreset(ctx context.Context, params UpdateUploadPresetParams) (*UploadPresetResult, error) {
	res := &UploadPresetResult{}
	_, err := a.put(ctx, api.BuildPath(uploadPresets, params.Name), params, res)

	return res, err
}

// UploadPresetResult is the result of UpdateUploadPreset, DeleteUploadPreset.
type UploadPresetResult struct {
	Message string        `json:"message"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// DeleteUploadPresetParams are the parameters for DeleteUploadPreset.
type DeleteUploadPresetParams struct {
	Name string `json:"-"`
}

// DeleteUploadPreset deletes the specified upload preset.
//
// https://cloudinary.com/documentation/admin_api#delete_an_upload_preset
func (a *API) DeleteUploadPreset(ctx context.Context, params DeleteUploadPresetParams) (*UploadPresetResult, error) {
	res := &UploadPresetResult{}
	_, err := a.delete(ctx, api.BuildPath(uploadPresets, params.Name), params, res)

	return res, err
}
