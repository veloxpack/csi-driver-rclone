package uploader

import (
	"context"
	"net/url"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	generateArchive api.EndPoint = "generate_archive"
	download        api.EndPoint = "download"
	downloadBackup  api.EndPoint = "download_backup"
)

// ArchiveFormat is the supported archive format.
type ArchiveFormat = string

const (
	// Zip archive format.
	Zip ArchiveFormat = "zip"
	// Tgz archive format.
	Tgz ArchiveFormat = "tgz"
)

// ArchiveMode is the supported mode for Archive.
type ArchiveMode string

const (
	// CreateArchive creates the archive as a new asset in your cloud.
	CreateArchive ArchiveMode = "create"
	// DownloadArchive returns the archive contents in the response.
	DownloadArchive ArchiveMode = "download"
)

// CreateArchiveParams are the parameters for CreateArchive.
type CreateArchiveParams struct {
	AllowMissing            *bool            `json:"allow_missing,omitempty"`
	Async                   *bool            `json:"async,omitempty"`
	ExpiresAt               *time.Time       `json:"expires_at,omitempty"`
	FlattenFolders          *bool            `json:"flatten_folders,omitempty"`
	FlattenTransformations  *bool            `json:"flatten_transformations,omitempty"`
	FullyQualifiedPublicIDs api.CldAPIArray  `json:"fully_qualified_public_ids,omitempty"`
	KeepDerived             *bool            `json:"keep_derived,omitempty"`
	Mode                    ArchiveMode      `json:"mode,omitempty"`
	NotificationURL         string           `json:"notification_url,omitempty"`
	Phash                   string           `json:"phash,omitempty"`
	Prefixes                api.CldAPIArray  `json:"prefixes,omitempty"`
	PublicIDs               []string         `json:"public_ids,omitempty"`
	ResourceType            api.AssetType    `json:"-"`
	SkipTransformationName  *bool            `json:"skip_transformation_name,omitempty"`
	TargetFormat            ArchiveFormat    `json:"target_format,omitempty"`
	TargetPublicID          string           `json:"target_public_id,omitempty"`
	TargetTags              api.CldAPIArray  `json:"target_tags,omitempty"`
	Tags                    api.CldAPIArray  `json:"tags,omitempty"`
	Transformations         string           `json:"transformations,omitempty"`
	Type                    api.DeliveryType `json:"type,omitempty"`
	UseOriginalFilename     *bool            `json:"use_original_filename,omitempty"`
}

// CreateArchive creates a new archive in the server and returns information in JSON format.
func (u *API) CreateArchive(ctx context.Context, params CreateArchiveParams) (*CreateArchiveResult, error) {
	res := &CreateArchiveResult{}
	err := u.callUploadAPI(ctx, generateArchive, params, res)

	return res, err
}

// CreateArchiveResult is the result of CreateArchive.
type CreateArchiveResult struct {
	AssetID       string        `json:"asset_id"`
	PublicID      string        `json:"public_id"`
	Version       int           `json:"version"`
	VersionID     string        `json:"version_id"`
	Signature     string        `json:"signature"`
	ResourceType  string        `json:"resource_type"`
	CreatedAt     time.Time     `json:"created_at"`
	Tags          []string      `json:"tags"`
	Bytes         int           `json:"bytes"`
	Type          string        `json:"type"`
	Etag          string        `json:"etag"`
	Placeholder   bool          `json:"placeholder"`
	URL           string        `json:"url"`
	SecureURL     string        `json:"secure_url"`
	AccessMode    string        `json:"access_mode"`
	ResourceCount int           `json:"resource_count"`
	FileCount     int           `json:"file_count"`
	Error         api.ErrorResp `json:"error,omitempty"`
	Response      interface{}
}

// CreateZip creates a new zip archive in the server and returns information in JSON format.
func (u *API) CreateZip(ctx context.Context, params CreateArchiveParams) (*CreateArchiveResult, error) {
	params.TargetFormat = Zip

	return u.CreateArchive(ctx, params)
}

// DownloadArchiveURL creates a URL that when invoked generates an archive and returns it.
func (u *API) DownloadArchiveURL(params CreateArchiveParams) (string, error) {
	params.Mode = DownloadArchive

	queryParams, err := api.StructToParams(params)
	if err != nil {
		return "", err
	}
	queryParams, err = u.signRequest(queryParams)
	if err != nil {
		return "", err
	}

	assetType := getAssetType(params)

	archiveEndpointURL := u.getUploadURL(api.BuildPath(assetType, generateArchive))

	urlStruct, err := url.Parse(archiveEndpointURL)
	if err != nil {
		return "", err
	}

	urlStruct.RawQuery = queryParams.Encode()

	return urlStruct.String(), nil
}

// DownloadZipURL creates a URL that when invokes generates a zip archive and returns it.
func (u *API) DownloadZipURL(params CreateArchiveParams) (string, error) {
	params.TargetFormat = Zip

	return u.DownloadArchiveURL(params)
}

// DownloadFolder creates a URL that when invoked generates an archive of a folder.
func (u *API) DownloadFolder(folderPath string, params CreateArchiveParams) (string, error) {
	params.Prefixes = api.CldAPIArray{folderPath}
	if len(params.ResourceType) == 0 {
		params.ResourceType = api.All
	}

	return u.DownloadArchiveURL(params)
}

// DownloadBackedUpAssetParams are the parameters for DownloadBackedUpAsset.
type DownloadBackedUpAssetParams struct {
	AssetID   string `json:"asset_id"`
	VersionID string `json:"version_id,omitempty"`
}

// DownloadBackedUpAsset creates a URL that  allows downloading the backed-up asset
// based on the asset ID and the version ID.
func (u *API) DownloadBackedUpAsset(params DownloadBackedUpAssetParams) (string, error) {
	queryParams, err := api.StructToParams(params)
	if err != nil {
		return "", err
	}
	queryParams, err = u.signRequest(queryParams)
	if err != nil {
		return "", err
	}

	urlStruct, err := url.Parse(u.getUploadURL(downloadBackup))
	if err != nil {
		return "", err
	}

	urlStruct.RawQuery = queryParams.Encode()

	return urlStruct.String(), nil
}

// PrivateDownloadURLParams are the parameters for PrivateDownloadURL.
type PrivateDownloadURLParams struct {
	PublicID     string        `json:"public_id"`
	Format       string        `json:"format"`
	DeliveryType string        `json:"type,omitempty"`
	Attachment   string        `json:"attachment,omitempty"`
	ExpiresAt    *time.Time    `json:"expires_at,omitempty"`
	ResourceType api.AssetType `json:"-"`
}

// PrivateDownloadURL returns a URL that when invoked downloads the asset.
func (u *API) PrivateDownloadURL(params PrivateDownloadURLParams) (string, error) {
	queryParams, err := api.StructToParams(params)
	if err != nil {
		return "", err
	}
	queryParams, err = u.signRequest(queryParams)
	if err != nil {
		return "", err
	}

	assetType := getAssetType(params)
	urlStruct, err := url.Parse(u.getUploadURL(api.BuildPath(assetType, download)))
	if err != nil {
		return "", err
	}

	urlStruct.RawQuery = queryParams.Encode()

	return urlStruct.String(), nil
}
