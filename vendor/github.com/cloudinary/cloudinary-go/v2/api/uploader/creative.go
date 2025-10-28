package uploader

import (
	"context"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	sprite  api.EndPoint = "sprite"
	multi   api.EndPoint = "multi"
	explode api.EndPoint = "explode"
	text    api.EndPoint = "text"
)

// GenerateSpriteParams are the parameters for GenerateSprite.
type GenerateSpriteParams struct {
	Tag             string `json:"tag,omitempty"`
	NotificationURL string `json:"notification_url,omitempty"`
	Async           *bool  `json:"async,omitempty"`
	Transformation  string `json:"transformation,omitempty"`
	ResourceType    string `json:"-"`
}

// GenerateSprite creates a sprite from all images that have been assigned a specified tag.
//
// The process produces two files:
// * A single image file containing all the images with the specified tag (PNG by default).
// * A CSS file that includes the style class names and the location of the individual images in the sprite.
//
// https://cloudinary.com/documentation/image_upload_api_reference#sprite_method
func (u *API) GenerateSprite(ctx context.Context, params GenerateSpriteParams) (*GenerateSpriteResult, error) {
	res := &GenerateSpriteResult{}
	err := u.callUploadAPI(ctx, sprite, params, res)

	return res, err
}

// GenerateSpriteResult is the result of GenerateSprite.
type GenerateSpriteResult struct {
	CSSURL         string               `json:"css_url"`
	ImageURL       string               `json:"image_url"`
	SecureCSSURL   string               `json:"secure_css_url"`
	SecureImageURL string               `json:"secure_image_url"`
	JSONURL        string               `json:"json_url"`
	SecureJSONURL  string               `json:"secure_json_url"`
	Version        int                  `json:"version"`
	PublicID       string               `json:"public_id"`
	ImageInfos     map[string]ImageInfo `json:"image_infos"`
	Error          api.ErrorResp        `json:"error,omitempty"`
	Response       interface{}
}

// ImageInfo contains information about the image.
type ImageInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// MultiParams are the parameters for Multi.
type MultiParams struct {
	Tag             string `json:"tag,omitempty"`
	Format          string `json:"format,omitempty"`
	NotificationURL string `json:"notification_url,omitempty"`
	Async           *bool  `json:"async,omitempty"`
	Transformation  string `json:"transformation,omitempty"`
	ResourceType    string `json:"-"`
}

// Multi Creates a single animated image, video or PDF from all image assets that have been assigned a specified tag.
//
// https://cloudinary.com/documentation/image_upload_api_reference#multi_method
func (u *API) Multi(ctx context.Context, params MultiParams) (*MultiResult, error) {
	res := &MultiResult{}
	err := u.callUploadAPI(ctx, multi, params, res)

	return res, err
}

// MultiResult is the result of Multi.
type MultiResult struct {
	URL       string        `json:"url"`
	SecureURL string        `json:"secure_url"`
	AssetID   string        `json:"asset_id"`
	PublicID  string        `json:"public_id"`
	Version   int           `json:"version"`
	Error     api.ErrorResp `json:"error,omitempty"`
	Response  interface{}
}

// ExplodeParams are the parameters for Explode.
type ExplodeParams struct {
	PublicID        string `json:"public_id"`
	Format          string `json:"format,omitempty"`
	Type            string `json:"type,omitempty"`
	NotificationURL string `json:"notification_url,omitempty"`
	Transformation  string `json:"transformation,omitempty"`
	ResourceType    string `json:"-"`
}

// Explode creates derived images for all of the individual pages in a multi-page file (PDF or animated GIF).
//
// Each derived image is stored with the same public ID as the original file, and can be accessed using the page
// parameter, in order to deliver a specific image.
//
// https://cloudinary.com/documentation/image_upload_api_reference#explode_method
func (u *API) Explode(ctx context.Context, params ExplodeParams) (*ExplodeResult, error) {
	params.Transformation = "pg_all" // Transformation must contain exactly one "pg_all" transformation parameter
	res := &ExplodeResult{}
	err := u.callUploadAPI(ctx, explode, params, res)

	return res, err
}

// ExplodeResult is the result of Explode.
type ExplodeResult struct {
	Status   string        `json:"status"`
	BatchID  string        `json:"batch_id"`
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// TextParams are the parameters for Text.
type TextParams struct {
	Text           string `json:"text"`
	PublicID       string `json:"public_id,omitempty"`
	FontFamily     string `json:"font_family,omitempty"`
	FontSize       int    `json:"font_size,omitempty"`
	FontColor      string `json:"font_color,omitempty"`
	TextAlign      string `json:"text_align,omitempty"`
	FontWeight     string `json:"font_weight,omitempty"`
	FontStyle      string `json:"font_style,omitempty"`
	Background     string `json:"background,omitempty"`
	Opacity        string `json:"opacity,omitempty"`
	TextDecoration string `json:"text_decoration,omitempty"`
	ResourceType   string `json:"-"`
}

// Text dynamically generates an image from a given textual string.
//
// https://cloudinary.com/documentation/image_upload_api_reference#text_method
func (u *API) Text(ctx context.Context, params TextParams) (*UploadResult, error) {
	res := &UploadResult{}
	err := u.callUploadAPI(ctx, text, params, res)

	return res, err
}
