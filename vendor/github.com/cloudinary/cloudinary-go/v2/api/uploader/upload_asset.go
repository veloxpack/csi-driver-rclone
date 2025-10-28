package uploader

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

// UploadParams struct allows to customize upload behaviour.
// For additional information about each one of the parameters please use the link below.
//
// https://cloudinary.com/documentation/image_upload_api_reference#upload_optional_parameters
type UploadParams struct {
	PublicID                       string                      `json:"public_id,omitempty"`
	PublicIDPrefix                 string                      `json:"public_id_prefix,omitempty"`
	PublicIDs                      api.CldAPIArray             `json:"public_ids,omitempty"`
	UseFilename                    *bool                       `json:"use_filename,omitempty"`
	UniqueFilename                 *bool                       `json:"unique_filename,omitempty"`
	UseFilenameAsDisplayName       *bool                       `json:"use_filename_as_display_name,omitempty"`
	FilenameOverride               string                      `json:"filename_override,omitempty"`
	DisplayName                    string                      `json:"display_name,omitempty"`
	UniqueDisplayName              *bool                       `json:"unique_display_name,omitempty"`
	Folder                         string                      `json:"folder,omitempty"`
	AssetFolder                    string                      `json:"asset_folder,omitempty"`
	UseAssetFolderAsPublicIDPrefix *bool                       `json:"use_asset_folder_as_public_id_prefix,omitempty"`
	Overwrite                      *bool                       `json:"overwrite,omitempty"`
	ResourceType                   string                      `json:"resource_type,omitempty"`
	Type                           api.DeliveryType            `json:"type,omitempty"`
	Tags                           api.CldAPIArray             `json:"tags,omitempty"`
	Context                        api.CldAPIMap               `json:"context,omitempty"`
	Metadata                       api.Metadata                `json:"metadata,omitempty"`
	Transformation                 string                      `json:"transformation,omitempty"`
	Format                         string                      `json:"format,omitempty"`
	AllowedFormats                 api.CldAPIArray             `json:"allowed_formats,omitempty"`
	Eager                          string                      `json:"eager,omitempty"`
	ResponsiveBreakpoints          ResponsiveBreakpointsParams `json:"responsive_breakpoints,omitempty"`
	AccessControl                  api.AccessControl           `json:"access_control,omitempty"`
	Eval                           string                      `json:"eval,omitempty"`
	OnSuccess                      string                      `json:"on_success,omitempty"`
	Async                          *bool                       `json:"async,omitempty"`
	EagerAsync                     *bool                       `json:"eager_async,omitempty"`
	Unsigned                       *bool                       `json:"unsigned,omitempty"`
	Proxy                          string                      `json:"proxy,omitempty"`
	Headers                        string                      `json:"headers,omitempty"`
	Callback                       string                      `json:"callback,omitempty"`
	NotificationURL                string                      `json:"notification_url,omitempty"`
	EagerNotificationURL           string                      `json:"eager_notification_url,omitempty"`
	Faces                          *bool                       `json:"faces,omitempty"`
	ImageMetadata                  *bool                       `json:"image_metadata,omitempty"`
	MediaMetadata                  *bool                       `json:"media_metadata,omitempty"`
	Exif                           *bool                       `json:"exif,omitempty"`
	Colors                         *bool                       `json:"colors,omitempty"`
	Phash                          *bool                       `json:"phash,omitempty"`
	FaceCoordinates                api.Coordinates             `json:"face_coordinates,omitempty"`
	CustomCoordinates              api.Coordinates             `json:"custom_coordinates,omitempty"`
	Backup                         *bool                       `json:"backup,omitempty"`
	ReturnDeleteToken              *bool                       `json:"return_delete_token,omitempty"`
	Invalidate                     *bool                       `json:"invalidate,omitempty"`
	DiscardOriginalFilename        *bool                       `json:"discard_original_filename,omitempty"`
	Moderation                     string                      `json:"moderation,omitempty"`
	UploadPreset                   string                      `json:"upload_preset,omitempty"`
	RawConvert                     string                      `json:"raw_convert,omitempty"`
	Categorization                 string                      `json:"categorization,omitempty"`
	VisualSearch                   *bool                       `json:"visual_search,omitempty"`
	AutoTagging                    float64                     `json:"auto_tagging,omitempty"`
	BackgroundRemoval              string                      `json:"background_removal,omitempty"`
	Detection                      string                      `json:"detection,omitempty"`
	OCR                            string                      `json:"ocr,omitempty"`
	Timestamp                      int64                       `json:"timestamp,omitempty"`
	QualityAnalysis                *bool                       `json:"quality_analysis,omitempty"`
	AccessibilityAnalysis          *bool                       `json:"accessibility_analysis,omitempty"`
	CinemagraphAnalysis            *bool                       `json:"cinemagraph_analysis,omitempty"`
	AutoChaptering                 *bool                       `json:"auto_chaptering,omitempty"`
	AutoTranscription              *api.AutoTranscription      `json:"auto_transcription,omitempty"`
	AutoVideoDetails               *api.AutoVideoDetails       `json:"auto_video_details,omitempty"`
}

// SingleResponsiveBreakpointsParams represents params for a single responsive breakpoints generation request.
type SingleResponsiveBreakpointsParams struct {
	CreateDerived  *bool  `json:"create_derived"`
	Transformation string `json:"transformation,omitempty"`
	MinWidth       int    `json:"min_width,omitempty"`
	MaxWidth       int    `json:"max_width,omitempty"`
	BytesStep      int    `json:"bytes_step,omitempty"`
	MaxImages      int    `json:"max_images,omitempty"`
	Format         string `json:"format,omitempty"`
}

// ResponsiveBreakpointsParams represents params for responsive breakpoints generation request.
type ResponsiveBreakpointsParams []SingleResponsiveBreakpointsParams

// MarshalJSON writes a quoted string in the custom format.
func (rbpParams ResponsiveBreakpointsParams) MarshalJSON() ([]byte, error) {
	rbpParamsArray := ([]SingleResponsiveBreakpointsParams)(rbpParams)
	paramsJSONObj, _ := json.Marshal(rbpParamsArray)

	return []byte(strconv.Quote(string(paramsJSONObj))), nil
}

// Upload uploads an asset to a Cloudinary account.
//
// The asset can be:
//   - a local file path
//   - the actual data (io.Reader)
//   - the Data URI (Base64 encoded), max ~60 MB (62,910,000 chars)
//   - the remote FTP, HTTP or HTTPS URL address of an existing file
//   - a private storage bucket (S3 or Google Storage) URL of a whitelisted bucket
//
// https://cloudinary.com/documentation/image_upload_api_reference#upload_method
func (u *API) Upload(ctx context.Context, file interface{}, uploadParams UploadParams) (*UploadResult, error) {
	formParams, err := api.StructToParams(uploadParams)
	if err != nil {
		return nil, err
	}

	body, err := u.postFile(ctx, file, formParams)
	if err != nil {
		return nil, err
	}

	result := &UploadResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}

	err = api.HandleRawResponse(body, result)
	return result, nil
}

// Eager contains information about eagerly transformed derived assets.
type Eager struct {
	Transformation string `json:"transformation"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Bytes          int    `json:"bytes"`
	Format         string `json:"format"`
	URL            string `json:"url"`
	SecureURL      string `json:"secure_url"`
}

// ModerationLabel represents moderation label.
type ModerationLabel struct {
	Confidence float64 `json:"confidence"`
	Name       string  `json:"name"`
	ParentName string  `json:"parent_name"`
}

// ModerationResponse represents moderation response.
type ModerationResponse struct {
	ModerationLabels []ModerationLabel `json:"moderation_labels"`
}

// Moderation represents moderation result.
type Moderation struct {
	Status    api.ModerationStatus `json:"status"`
	Kind      string               `json:"kind"`
	Response  ModerationResponse   `json:"response"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// BreakpointResult represents a result of a single responsive breakpoints image.
type BreakpointResult struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Bytes     int    `json:"bytes"`
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
}

// ResponsiveBreakpointsResult represents a result of responsive breakpoints of an image.
type ResponsiveBreakpointsResult struct {
	Breakpoints    []BreakpointResult `json:"breakpoints"`
	Transformation string             `json:"transformation"`
}

// UploadResult image success response struct.
type UploadResult struct {
	AssetID               string                        `json:"asset_id"`
	PublicID              string                        `json:"public_id"`
	AssetFolder           string                        `json:"asset_folder"`
	DisplayName           string                        `json:"display_name"`
	Version               int                           `json:"version"`
	VersionID             string                        `json:"version_id"`
	Signature             string                        `json:"signature"`
	Width                 int                           `json:"width,omitempty"`
	Height                int                           `json:"height,omitempty"`
	Format                string                        `json:"format"`
	ResourceType          string                        `json:"resource_type"`
	CreatedAt             time.Time                     `json:"created_at"`
	Tags                  api.CldAPIArray               `json:"tags,omitempty"`
	Pages                 int                           `json:"pages,omitempty"`
	Bytes                 int                           `json:"bytes"`
	Type                  string                        `json:"type"`
	Etag                  string                        `json:"etag"`
	Phash                 string                        `json:"phash,omitempty"`
	Placeholder           bool                          `json:"placeholder,omitempty"`
	URL                   string                        `json:"url"`
	SecureURL             string                        `json:"secure_url"`
	PlaybackURL           string                        `json:"playback_url,omitempty"`
	AccessMode            string                        `json:"access_mode"`
	AccessControl         api.AccessControl             `json:"access_control,omitempty"`
	Context               api.Metadata                  `json:"context,omitempty"`
	Metadata              api.Metadata                  `json:"metadata,omitempty"`
	Moderation            []Moderation                  `json:"moderation,omitempty"`
	Overwritten           bool                          `json:"overwritten"`
	OriginalFilename      string                        `json:"original_filename"`
	Eager                 []Eager                       `json:"eager"`
	ResponsiveBreakpoints []ResponsiveBreakpointsResult `json:"responsive_breakpoints"`
	HookExecution         api.HookExecution             `json:"hook_execution"`
	Error                 api.ErrorResp                 `json:"error,omitempty"`
	Response              interface{}
}

// UnsignedUpload uploads an asset to a Cloudinary account.
//
// The upload is not signed so an upload preset is required.
//
// https://cloudinary.com/documentation/image_upload_api_reference#unsigned_upload_syntax
func (u *API) UnsignedUpload(ctx context.Context, file interface{}, uploadPreset string, uploadParams UploadParams) (*UploadResult, error) {
	uploadParams.Unsigned = api.Bool(true)
	uploadParams.UploadPreset = uploadPreset

	return u.Upload(ctx, file, uploadParams)
}
