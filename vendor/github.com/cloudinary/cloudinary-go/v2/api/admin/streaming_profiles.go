package admin

// Enables you to manage streaming profiles for use with adaptive bitrate streaming.
//
// https://cloudinary.com/documentation/admin_api#adaptive_streaming_profiles

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/transformation"
)

const (
	streamingProfiles api.EndPoint = "streaming_profiles"
)

// ListStreamingProfiles lists streaming profiles including built-in and custom profiles.
//
// https://cloudinary.com/documentation/admin_api#get_adaptive_streaming_profiles
func (a *API) ListStreamingProfiles(ctx context.Context) (*ListStreamingProfilesResult, error) {
	res := &ListStreamingProfilesResult{}
	_, err := a.get(ctx, streamingProfiles, nil, res)

	return res, err
}

// ListStreamingProfilesResult represents the result of listing of streaming profiles.
type ListStreamingProfilesResult struct {
	Data  []StreamingProfile `json:"data"`
	Error api.ErrorResp      `json:"error,omitempty"`
}

// StreamingProfile represents a single streaming profile.
type StreamingProfile struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Predefined  bool   `json:"predefined"`
}

// GetStreamingProfileParams are the parameters for GetStreamingProfile.
type GetStreamingProfileParams struct {
	Name string `json:"-"`
}

// GetStreamingProfile gets details of a single streaming profile by name.
//
// https://cloudinary.com/documentation/admin_api#get_details_of_a_single_streaming_profile
func (a *API) GetStreamingProfile(ctx context.Context, params GetStreamingProfileParams) (*GetStreamingProfileResult, error) {
	res := &GetStreamingProfileResult{}
	_, err := a.get(ctx, api.BuildPath(streamingProfiles, params.Name), params, res)

	return res, err
}

// GetStreamingProfileResult is the result of GetStreamingProfile.
type GetStreamingProfileResult struct {
	Data  StreamingProfileDetails `json:"data"`
	Error api.ErrorResp           `json:"error,omitempty"`
}

// StreamingProfileDetails represents the details of a streaming profile.
type StreamingProfileDetails struct {
	Name            string                           `json:"name"`
	DisplayName     string                           `json:"display_name,omitempty"`
	Predefined      bool                             `json:"predefined"`
	Representations []StreamingProfileRepresentation `json:"representations,omitempty"`
}

// StreamingProfileRepresentation is a representation of a single streaming profile.
type StreamingProfileRepresentation struct {
	Transformation transformation.Transformation `json:"transformation"`
}

// StreamingProfileRepresentations contains multiple streaming profile representations.
type StreamingProfileRepresentations []RawStreamingProfileRepresentation

// MarshalJSON serializes StreamingProfileRepresentations to a string of a json encoded array.
func (spRepresentations StreamingProfileRepresentations) MarshalJSON() ([]byte, error) {
	spRepresentationsArray := ([]RawStreamingProfileRepresentation)(spRepresentations)
	paramsJSONObj, _ := json.Marshal(spRepresentationsArray)

	return []byte(strconv.Quote(string(paramsJSONObj))), nil
}

// RawStreamingProfileRepresentation is a raw representation of a steaming profile.
type RawStreamingProfileRepresentation struct {
	Transformation transformation.RawTransformation `json:"transformation"`
}

// CreateStreamingProfileParams are the parameters for CreateStreamingProfile.
type CreateStreamingProfileParams struct {
	Name            string                          `json:"name"` // The name to assign to the new streaming profile.
	DisplayName     string                          `json:"display_name,omitempty"`
	Representations StreamingProfileRepresentations `json:"representations"`
}

// CreateStreamingProfile creates a new, custom streaming profile.
//
// https://cloudinary.com/documentation/admin_api#create_a_streaming_profile
func (a *API) CreateStreamingProfile(ctx context.Context, params CreateStreamingProfileParams) (*GetStreamingProfileResult, error) {
	res := &GetStreamingProfileResult{}
	_, err := a.post(ctx, api.BuildPath(streamingProfiles), params, res)

	return res, err
}

// UpdateStreamingProfileParams are the parameters for UpdateStreamingProfile
type UpdateStreamingProfileParams struct {
	Name            string                          `json:"-"` // The name of the streaming profile to update.
	DisplayName     string                          `json:"display_name,omitempty"`
	Representations StreamingProfileRepresentations `json:"representations"`
}

// UpdateStreamingProfile updates an existing streaming profile.
//
// You can update both custom and built-in profiles. The specified list of representations replaces the previous list.
//
// https://cloudinary.com/documentation/admin_api#update_an_existing_streaming_profile
func (a *API) UpdateStreamingProfile(ctx context.Context, params UpdateStreamingProfileParams) (*GetStreamingProfileResult, error) {
	res := &GetStreamingProfileResult{}
	_, err := a.put(ctx, api.BuildPath(streamingProfiles, params.Name), params, res)

	return res, err
}

// DeleteStreamingProfileParams are the parameters for DeleteStreamingProfile.
type DeleteStreamingProfileParams struct {
	Name string `json:"-"` // The name of the streaming profile to delete or revert.
}

// DeleteStreamingProfile deletes or reverts the specified streaming profile.
//
// For custom streaming profiles, deletes the specified profile.
// For built-in streaming profiles, if the built-in profile was modified, reverts the profile to the original settings.
// For built-in streaming profiles that have not been modified, the Delete method returns an error.
//
// https://cloudinary.com/documentation/admin_api#delete_or_revert_the_specified_streaming_profile
func (a *API) DeleteStreamingProfile(ctx context.Context, params DeleteStreamingProfileParams) (*DeleteStreamingProfileResult, error) {
	res := &DeleteStreamingProfileResult{}
	_, err := a.delete(ctx, api.BuildPath(streamingProfiles, params.Name), params, res)

	return res, err
}

// DeleteStreamingProfileResult is the result of DeleteStreamingProfile.
type DeleteStreamingProfileResult struct {
	Message string        `json:"message"`
	Error   api.ErrorResp `json:"error,omitempty"`
}
