package admin

// The Cloudinary API search method allows you fine control on filtering and retrieving information on all the assets
// in your account with the help of query expressions in a Lucene-like query language. A few examples of what you can
// accomplish using the search method include:
//
//  * Searching by descriptive attributes such as public ID, filename, folders, tags, context, etc.
//  * Searching by file details such as type, format, file size, dimensions, etc.
//  * Searching by embedded data such as Exif, XMP, etc.
//  * Searching by analyzed data such as the number of faces, predominant colors, auto-tags, etc.
//  * Requesting aggregation counts on specified parameters, for example the number of assets found broken down by file
// format.
//
// https://cloudinary.com/documentation/search_api
import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin/search"
)

const searchEndPoint api.EndPoint = "resources/search"

// Search executes the search API request.
func (a *API) Search(ctx context.Context, searchQuery search.Query) (*SearchResult, error) {
	res := &SearchResult{}
	_, err := a.post(ctx, api.BuildPath(searchEndPoint), searchQuery, res)

	return res, err
}

// SearchResult is the result of Search.
type SearchResult struct {
	TotalCount int           `json:"total_count"`
	Time       int           `json:"time"`
	Assets     []SearchAsset `json:"resources"`
	Error      api.ErrorResp `json:"error,omitempty"`
	NextCursor string        `json:"next_cursor,omitempty"`
	Response   interface{}
}

// SearchAsset represents the details of a single asset that was found.
type SearchAsset struct {
	PublicID      string              `json:"public_id"`
	AssetID       string              `json:"asset_id"`
	Folder        string              `json:"folder"`
	AssetFolder   string              `json:"asset_folder"`
	Filename      string              `json:"filename"`
	DisplayName   string              `json:"display_name"`
	Format        string              `json:"format"`
	Version       int                 `json:"version"`
	ResourceType  string              `json:"resource_type"`
	Type          string              `json:"type"`
	CreatedAt     time.Time           `json:"created_at"`
	UploadedAt    time.Time           `json:"uploaded_at"`
	Bytes         int                 `json:"bytes"`
	BackupBytes   int                 `json:"backup_bytes"`
	Width         int                 `json:"width"`
	Height        int                 `json:"height"`
	AspectRatio   float64             `json:"aspect_ratio"`
	Pixels        int                 `json:"pixels"`
	Tags          []string            `json:"tags"`
	Context       ImageMetadataResult `json:"context"`
	ImageMetadata ImageMetadataResult `json:"image_metadata"`
	VideoMetadata MediaMetadataResult `json:"video_metadata"`
	ImageAnalysis ImageAnalysis       `json:"image_analysis"`
	URL           string              `json:"url"`
	SecureURL     string              `json:"secure_url"`
	Status        string              `json:"status"`
	AccessMode    string              `json:"access_mode"`
	AccessControl interface{}         `json:"access_control"`
	Etag          string              `json:"etag"`
	CreatedBy     SearchUser          `json:"created_by"`
	UploadedBy    SearchUser          `json:"uploaded_by"`
	LastUpdated   api.LastUpdated     `json:"last_updated"`
}

// ImageAnalysis contains details about image analysis.
type ImageAnalysis struct {
	FaceCount         int                `json:"face_count"`
	Faces             [][]int            `json:"faces"`
	Grayscale         bool               `json:"grayscale"`
	IllustrationScore int                `json:"illustration_score"`
	Transparent       bool               `json:"transparent"`
	Etag              string             `json:"etag"`
	Colors            map[string]float64 `json:"colors"`
}

// SearchUser contains details about the user.
type SearchUser struct {
	AccessKey string `json:"access_key"`
}
