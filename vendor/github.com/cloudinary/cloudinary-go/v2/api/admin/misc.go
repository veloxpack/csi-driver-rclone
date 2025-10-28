package admin

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	ping        api.EndPoint = "ping"
	cloudConfig api.EndPoint = "config"
	usage       api.EndPoint = "usage"
)

// Ping tests the reachability of the Cloudinary API.
//
// https://cloudinary.com/documentation/admin_api#ping
func (a *API) Ping(ctx context.Context) (*PingResult, error) {
	res := &PingResult{}
	_, err := a.get(ctx, ping, nil, res)

	return res, err
}

// PingResult represents the result of the Ping request.
type PingResult struct {
	Status   string        `json:"status"`
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// GetConfigParams are the parameters for GetConfig.
type GetConfigParams struct {
	Settings *bool `json:"settings,omitempty"`
}

// GetConfig gets cloud config details.
//
// https://cloudinary.com/documentation/admin_api#config
func (a *API) GetConfig(ctx context.Context, params GetConfigParams) (*GetConfigResult, error) {
	res := &GetConfigResult{}
	_, err := a.get(ctx, cloudConfig, params, res)

	return res, err
}

// GetConfigResult represents the result of the GetConfig request.
type GetConfigResult struct {
	CloudName string        `json:"cloud_name"`
	CreatedAt time.Time     `json:"created_at"`
	Settings  CloudSettings `json:"settings"`
	Error     api.ErrorResp `json:"error,omitempty"`
	Response  interface{}
}

// CloudSettings represents the settings of the cloud.
type CloudSettings struct {
	FolderMode string `json:"folder_mode"`
}

// UsageParams are the parameters for Usage.
type UsageParams struct {
	Date time.Time `json:"-"`
}

// Usage gets account usage details.
//
// Returns a report detailing your current Cloudinary account usage details, including
// storage, bandwidth, requests, number of resources, and add-on usage.
// Note that numbers are updated periodically.
//
// https://cloudinary.com/documentation/admin_api#usage
func (a *API) Usage(ctx context.Context, params UsageParams) (*UsageResult, error) {
	date := ""
	if !params.Date.IsZero() {
		date = params.Date.Format("02-01-2006")
	}
	res := &UsageResult{}
	_, err := a.get(ctx, api.BuildPath(usage, date), params, res)

	return res, err
}

// UsageResult is the result of Usage.
type UsageResult struct {
	Plan            string `json:"plan"`
	LastUpdated     string `json:"last_updated"`
	Transformations struct {
		Usage        int     `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
		Limit        int     `json:"limit"`
		UsedPercent  float64 `json:"used_percent"`
	} `json:"transformations"`
	Objects struct {
		Usage       int     `json:"usage"`
		Limit       int     `json:"limit"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"objects"`
	Bandwidth struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
		Limit        int64   `json:"limit"`
		UsedPercent  float64 `json:"used_percent"`
	} `json:"bandwidth"`
	Storage struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
		Limit        int64   `json:"limit"`
		UsedPercent  float64 `json:"used_percent"`
	} `json:"storage"`
	Credits struct {
		Usage float64 `json:"usage"`
	} `json:"credits"`
	Requests         int64 `json:"requests"`
	Resources        int   `json:"resources"`
	DerivedResources int   `json:"derived_resources"`
	MediaLimits      struct {
		ImageMaxSizeBytes int `json:"image_max_size_bytes"`
		VideoMaxSizeBytes int `json:"video_max_size_bytes"`
		RawMaxSizeBytes   int `json:"raw_max_size_bytes"`
		ImageMaxPx        int `json:"image_max_px"`
		AssetMaxTotalPx   int `json:"asset_max_total_px"`
	} `json:"media_limits"`
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// TagsParams are the parameters for Tags.
type TagsParams struct {
	AssetType  api.AssetType `json:"-"`                     // The type of asset.
	NextCursor string        `json:"next_cursor,omitempty"` // The cursor used for pagination.
	MaxResults int           `json:"max_results,omitempty"` // Maximum number of tags to return (up to 500). Default: 10.
	Prefix     string        `json:"prefix,omitempty"`      // Find all tags that start with the given prefix.
}

// Tags lists all the tags currently used for a specified asset type.
//
// https://cloudinary.com/documentation/admin_api#get_tags
func (a *API) Tags(ctx context.Context, params TagsParams) (*TagsResult, error) {
	res := &TagsResult{}
	_, err := a.get(ctx, api.BuildPath(tags, params.AssetType), params, res)

	return res, err
}

// TagsResult is the result of Tags.
type TagsResult struct {
	Tags       []string      `json:"tags"`
	NextCursor string        `json:"next_cursor"`
	Error      api.ErrorResp `json:"error,omitempty"`
	Response   interface{}
}
