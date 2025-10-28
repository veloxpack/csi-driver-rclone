package admin

// Enables you to manage the assets in your account or cloud.
//
// https://cloudinary.com/documentation/admin_api#resources

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

// AssetParams are the parameters for Asset.
type AssetParams struct {
	AssetType             api.AssetType    `json:"-"`
	DeliveryType          api.DeliveryType `json:"-"`
	PublicID              string           `json:"-"`
	Exif                  *bool            `json:"exif,omitempty"`
	Colors                *bool            `json:"colors,omitempty"`
	Faces                 *bool            `json:"faces,omitempty"`
	QualityAnalysis       *bool            `json:"quality_analysis,omitempty"`
	ImageMetadata         *bool            `json:"image_metadata,omitempty"`
	MediaMetadata         *bool            `json:"media_metadata,omitempty"`
	Phash                 *bool            `json:"phash,omitempty"`
	Pages                 *bool            `json:"pages,omitempty"`
	AccessibilityAnalysis *bool            `json:"accessibility_analysis,omitempty"`
	CinemagraphAnalysis   *bool            `json:"cinemagraph_analysis,omitempty"`
	Coordinates           *bool            `json:"coordinates,omitempty"`
	MaxResults            int              `json:"max_results,omitempty"`
	DerivedNextCursor     string           `json:"derived_next_cursor,omitempty"`
	Related               *bool            `json:"related,omitempty"`
	RelatedNextCursor     string           `json:"related_next_cursor,omitempty"`
	Versions              *bool            `json:"versions,omitempty"`
}

// Asset returns the details of the specified asset and all its derived resources.
//
// Note that if you only need details about the original resource,
// you can also use the uploader.Upload or uploader.Explicit methods, which return the same information and
// are not rate limited.
//
// https://cloudinary.com/documentation/admin_api#get_the_details_of_a_single_resource
func (a *API) Asset(ctx context.Context, params AssetParams) (*AssetResult, error) {
	res := &AssetResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType,
		params.PublicID), params, res)

	return res, err
}

// AssetResult is the result of the Asset.
type AssetResult struct {
	AssetID               string                      `json:"asset_id"`
	PublicID              string                      `json:"public_id"`
	Format                string                      `json:"format"`
	AssetFolder           string                      `json:"asset_folder"`
	DisplayName           string                      `json:"display_name"`
	Version               int                         `json:"version"`
	ResourceType          string                      `json:"resource_type"`
	Type                  string                      `json:"type"`
	CreatedAt             time.Time                   `json:"created_at"`
	Bytes                 int                         `json:"bytes"`
	Width                 int                         `json:"width"`
	Height                int                         `json:"height"`
	Backup                bool                        `json:"backup"`
	AccessMode            string                      `json:"access_mode"`
	AccessControl         api.AccessControl           `json:"access_control,omitempty"`
	URL                   string                      `json:"url"`
	SecureURL             string                      `json:"secure_url"`
	Metadata              api.Metadata                `json:"metadata,omitempty"`
	Tags                  []string                    `json:"tags"`
	LastUpdated           api.LastUpdated             `json:"last_updated"`
	NextCursor            string                      `json:"next_cursor"`
	Derived               []interface{}               `json:"derived"`
	Etag                  string                      `json:"etag"`
	ImageMetadata         ImageMetadataResult         `json:"image_metadata"`
	VideoMetadata         MediaMetadataResult         `json:"video_metadata"`
	Coordinates           interface{}                 `json:"coordinates"`
	Info                  interface{}                 `json:"info"`
	Exif                  interface{}                 `json:"exif"`
	Faces                 [][]int                     `json:"faces"`
	IllustrationScore     float64                     `json:"illustration_score"`
	SemiTransparent       bool                        `json:"semi_transparent"`
	Grayscale             bool                        `json:"grayscale"`
	Colors                [][]interface{}             `json:"colors"`
	Predominant           PredominantResult           `json:"predominant"`
	Phash                 string                      `json:"phash"`
	QualityAnalysis       QualityAnalysisResult       `json:"quality_analysis"`
	QualityScore          float64                     `json:"quality_score"`
	AccessibilityAnalysis AccessibilityAnalysisResult `json:"accessibility_analysis"`
	Pages                 int                         `json:"pages"`
	CinemagraphAnalysis   CinemagraphAnalysisResult   `json:"cinemagraph_analysis"`
	Usage                 interface{}                 `json:"usage"`
	OriginalFilename      string                      `json:"original_filename"`
	Context               AssetContextResult          `json:"context"`
	Error                 api.ErrorResp               `json:"error,omitempty"`
	Response              interface{}
}

// QualityAnalysisResult contains the details about quality analysis.
type QualityAnalysisResult struct {
	JpegQuality       float64 `json:"jpeg_quality"`
	JpegChroma        float64 `json:"jpeg_chroma"`
	Focus             float64 `json:"focus"`
	Noise             float64 `json:"noise"`
	Contrast          float64 `json:"contrast"`
	Exposure          float64 `json:"exposure"`
	Saturation        float64 `json:"saturation"`
	Lighting          float64 `json:"lighting"`
	PixelScore        float64 `json:"pixel_score"`
	ColorScore        float64 `json:"color_score"`
	Dct               float64 `json:"dct"`
	Blockiness        float64 `json:"blockiness"`
	ChromaSubsampling float64 `json:"chroma_subsampling"`
	Resolution        float64 `json:"resolution"`
}

// AccessibilityAnalysisResult contains the details about accessibility analysis.
type AccessibilityAnalysisResult struct {
	ColorblindAccessibilityAnalysis struct {
		DistinctEdges      float64  `json:"distinct_edges"`
		DistinctColors     float64  `json:"distinct_colors"`
		MostIndistinctPair []string `json:"most_indistinct_pair"`
	} `json:"colorblind_accessibility_analysis"`
	ColorblindAccessibilityScore float64 `json:"colorblind_accessibility_score"`
}

// CinemagraphAnalysisResult contains the details about cinemagraph analysis.
type CinemagraphAnalysisResult struct {
	CinemagraphScore float64 `json:"cinemagraph_score"`
}

// ImageMetadataResult contains the image metadata.
type ImageMetadataResult map[string]string

// MediaMetadataResult contains the media metadata.
type MediaMetadataResult map[string]interface{}

// PredominantResult contains the details about predominant colors.
type PredominantResult struct {
	Google     [][]interface{} `json:"google"`
	Cloudinary [][]interface{} `json:"cloudinary"`
}

// AssetContextResult contains the contextual metadata
type AssetContextResult struct {
	Custom map[string]string `json:"custom"`
}

func (m *AssetContextResult) UnmarshalJSON(data []byte) error {
	// map format
	var mapData struct {
		Custom map[string]string `json:"custom"`
	}
	if err := json.Unmarshal(data, &mapData); err == nil {
		*m = AssetContextResult{
			Custom: mapData.Custom,
		}
		return nil
	}

	// array format
	var arrayData []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &arrayData); err != nil {
		return err
	}

	result := AssetContextResult{
		Custom: make(map[string]string),
	}
	for _, item := range arrayData {
		result.Custom[item.Key] = item.Value
	}
	*m = result

	return nil
}

// UpdateAssetParams are the parameters for UpdateAsset.
type UpdateAssetParams struct {
	AssetType         api.AssetType        `json:"-"`
	DeliveryType      api.DeliveryType     `json:"-"`
	PublicID          string               `json:"-"`
	AssetFolder       string               `json:"asset_folder,omitempty"`
	DisplayName       string               `json:"display_name,omitempty"`
	UniqueDisplayName *bool                `json:"unique_display_name,omitempty"`
	ModerationStatus  api.ModerationStatus `json:"moderation_status,omitempty"`
	RawConvert        string               `json:"raw_convert,omitempty"`
	OCR               string               `json:"ocr,omitempty"`
	Categorization    string               `json:"categorization,omitempty"`
	Detection         string               `json:"detection,omitempty"`
	SimilaritySearch  string               `json:"similarity_search,omitempty"`
	VisualSearch      *bool                `json:"visual_search,omitempty"`
	AutoTagging       float64              `json:"auto_tagging,omitempty"`
	BackgroundRemoval string               `json:"background_removal,omitempty"`
	QualityOverride   int                  `json:"quality_override,omitempty"`
	NotificationURL   string               `json:"notification_url,omitempty"`
	Tags              api.CldAPIArray      `json:"tags,omitempty,omitempty"`
	Context           api.CldAPIMap        `json:"context,omitempty"`
	FaceCoordinates   api.Coordinates      `json:"face_coordinates,omitempty"`
	CustomCoordinates api.Coordinates      `json:"custom_coordinates,omitempty"`
	AccessControl     interface{}          `json:"access_control,omitempty"`
}

// UpdateAsset updates details of an existing asset.
//
// Updates one or more of the attributes associated with a specified asset. Note that you can also update
// most attributes of an existing asset using the uploader.Explicit method, which is not rate limited.
//
// https://cloudinary.com/documentation/admin_api#update_details_of_an_existing_resource
func (a *API) UpdateAsset(ctx context.Context, params UpdateAssetParams) (*AssetResult, error) {
	res := &AssetResult{}
	_, err := a.post(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType, params.PublicID), params, res)

	return res, err
}

// AssetByAssetIDParams are the parameters for AssetByAssetID.
type AssetByAssetIDParams struct {
	AssetID               string `json:"-"`
	Colors                *bool  `json:"colors,omitempty"`
	Exif                  *bool  `json:"exif,omitempty"`
	Faces                 *bool  `json:"faces,omitempty"`
	QualityAnalysis       *bool  `json:"quality_analysis,omitempty"`
	ImageMetadata         *bool  `json:"image_metadata,omitempty"`
	MediaMetadata         *bool  `json:"media_metadata,omitempty"`
	Phash                 *bool  `json:"phash,omitempty"`
	Pages                 *bool  `json:"pages,omitempty"`
	CinemagraphAnalysis   *bool  `json:"cinemagraph_analysis,omitempty"`
	Coordinates           *bool  `json:"coordinates,omitempty"`
	MaxResults            int    `json:"max_results,omitempty"`
	DerivedNextCursor     string `json:"derived_next_cursor,omitempty"`
	AccessibilityAnalysis *bool  `json:"accessibility_analysis,omitempty"`
	Versions              *bool  `json:"versions,omitempty"`
}

// AssetByAssetID returns the details of the specified asset and all its derived assets by asset id.
//
// Note that if you only need details about the original asset,
// you can also use the uploader.Upload or uploader.Explicit methods, which return the same information and
// are not rate limited.
//
// https://cloudinary.com/documentation/admin_api#get_the_details_of_a_single_resource
func (a *API) AssetByAssetID(ctx context.Context, params AssetByAssetIDParams) (*AssetResult, error) {
	res := &AssetResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetID), params, res)

	return res, err
}
