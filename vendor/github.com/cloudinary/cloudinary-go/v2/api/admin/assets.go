package admin

// Enables you to manage the assets in your account or cloud.
//
// https://cloudinary.com/documentation/admin_api#resources

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	assets        api.EndPoint = "resources"
	byAssetFolder api.EndPoint = "by_asset_folder"
	visualSearch  api.EndPoint = "visual_search"
	derivedAssets api.EndPoint = "derived_resources"
	relatedAssets api.EndPoint = "related_assets"
	tags          api.EndPoint = "tags"
	cldContext    api.EndPoint = "context"
	moderations   api.EndPoint = "moderations"
	restore       api.EndPoint = "restore"
)

// AssetTypes lists available asset types.
func (a *API) AssetTypes(ctx context.Context) (*AssetTypesResult, error) {
	res := &AssetTypesResult{}
	_, err := a.get(ctx, assets, nil, res)

	return res, err
}

// AssetTypesResult is the result of the AssetTypes.
type AssetTypesResult struct {
	AssetTypes []string      `json:"resource_types"`
	Error      api.ErrorResp `json:"error,omitempty"`
}

// AssetsParams are the parameters for Assets.
type AssetsParams struct {
	AssetType    api.AssetType   `json:"-"`
	DeliveryType string          `json:"-"`
	Prefix       string          `json:"prefix,omitempty"`
	StartAt      *time.Time      `json:"start_at,omitempty"`
	NextCursor   string          `json:"next_cursor,omitempty"`
	MaxResults   int             `json:"max_results,omitempty"`
	Tags         *bool           `json:"tags,omitempty"`
	Context      *bool           `json:"context,omitempty"`
	Moderations  *bool           `json:"moderations,omitempty"`
	Direction    string          `json:"direction,omitempty"`
	Fields       api.CldAPIArray `json:"fields,omitempty"`
}

// Assets lists all uploaded assets filtered by any specified AssetsParams.
//
// https://cloudinary.com/documentation/admin_api#get_resources
func (a *API) Assets(ctx context.Context, params AssetsParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// AssetsResult is the result of Assets.
type AssetsResult struct {
	Assets     []api.BriefAssetResult `json:"resources"`
	NextCursor string                 `json:"next_cursor"`
	Error      api.ErrorResp          `json:"error,omitempty"`
	Response   interface{}
}

// AssetsByTagParams are the parameters for AssetsByTag.
type AssetsByTagParams struct {
	AssetType   api.AssetType   `json:"-"`
	Tag         string          `json:"-"`
	NextCursor  string          `json:"next_cursor,omitempty"`
	MaxResults  int             `json:"max_results,omitempty"`
	Tags        *bool           `json:"tags,omitempty"`
	Context     *bool           `json:"context,omitempty"`
	Moderations *bool           `json:"moderations,omitempty"`
	Direction   string          `json:"direction,omitempty"`
	Fields      api.CldAPIArray `json:"fields,omitempty"`
}

// AssetsByTag lists assets with the specified tag.
//
// This method does not return matching deleted assets, even if they have been backed up.
//
// https://cloudinary.com/documentation/admin_api#get_resources_by_tag
func (a *API) AssetsByTag(ctx context.Context, params AssetsByTagParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, tags, params.Tag), params, res)

	return res, err
}

// AssetsByContextParams are the parameters for AssetsByContext.
type AssetsByContextParams struct {
	AssetType   api.AssetType   `json:"-"`
	Key         string          `json:"key"`
	Value       string          `json:"value,omitempty"`
	NextCursor  string          `json:"next_cursor,omitempty"`
	MaxResults  int             `json:"max_results,omitempty"`
	Tags        *bool           `json:"tags,omitempty"`
	Context     *bool           `json:"context,omitempty"`
	Moderations *bool           `json:"moderations,omitempty"`
	Direction   string          `json:"direction,omitempty"`
	Fields      api.CldAPIArray `json:"fields,omitempty"`
}

// AssetsByContext lists assets with the specified contextual metadata.
//
// This method does not return matching deleted assets, even if they have been backed up.
//
// https://cloudinary.com/documentation/admin_api#get_resources_by_context
func (a *API) AssetsByContext(ctx context.Context, params AssetsByContextParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, cldContext), params, res)

	return res, err
}

// AssetsByModerationParams are the parameters for AssetsByModeration.
type AssetsByModerationParams struct {
	AssetType   api.AssetType   `json:"-"`
	Kind        string          `json:"-"`
	Status      string          `json:"-"`
	NextCursor  string          `json:"next_cursor,omitempty"`
	MaxResults  int             `json:"max_results,omitempty"`
	Tags        *bool           `json:"tags,omitempty"`
	Context     *bool           `json:"context,omitempty"`
	Moderations *bool           `json:"moderations,omitempty"`
	Direction   string          `json:"direction,omitempty"`
	Fields      api.CldAPIArray `json:"fields,omitempty"`
}

// AssetsByModeration lists assets currently in the specified moderation queue and status.
//
// https://cloudinary.com/documentation/admin_api#get_resources_in_moderation_queues
func (a *API) AssetsByModeration(ctx context.Context, params AssetsByModerationParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, moderations, params.Kind, params.Status), params, res)

	return res, err
}

// AssetsByIDsParams are the parameters for AssetsByIDs.
type AssetsByIDsParams struct {
	AssetType    api.AssetType    `json:"-"`
	DeliveryType api.DeliveryType `json:"-"`
	PublicIDs    api.CldAPIArray  `json:"public_ids"`
	Tags         *bool            `json:"tags,omitempty"`
	Context      *bool            `json:"context,omitempty"`
	Moderations  *bool            `json:"moderations,omitempty"`
	Fields       api.CldAPIArray  `json:"fields,omitempty"`
}

// AssetsByIDs lists assets with the specified public IDs.
//
// https://cloudinary.com/documentation/admin_api#get_resources
func (a *API) AssetsByIDs(ctx context.Context, params AssetsByIDsParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// AssetsByAssetFolderParams are the parameters for AssetsByAssetFolder.
type AssetsByAssetFolderParams struct {
	AssetFolder string          `json:"asset_folder"`
	Tags        *bool           `json:"tags,omitempty"`
	Context     *bool           `json:"context,omitempty"`
	Moderations *bool           `json:"moderations,omitempty"`
	NextCursor  string          `json:"next_cursor,omitempty"`
	MaxResults  int             `json:"max_results,omitempty"`
	Fields      api.CldAPIArray `json:"fields,omitempty"`
}

// AssetsByAssetFolder lists assets in the specified asset folder.
//
// https://cloudinary.com/documentation/admin_api#get_resources
func (a *API) AssetsByAssetFolder(ctx context.Context, params AssetsByAssetFolderParams) (*AssetsResult, error) {
	res := &AssetsResult{}
	_, err := a.get(ctx, api.BuildPath(assets, byAssetFolder), params, res)

	return res, err
}

// VisualSearchParams are the parameters for VisualSearch.
type VisualSearchParams struct {
	ImageURL     string      `json:"image_url,omitempty"`
	ImageAssetID string      `json:"image_asset_id,omitempty"`
	Text         string      `json:"text,omitempty"`
	ImageFile    interface{} `json:"-"`
}

// VisualSearch finds images based on their visual content.
func (a *API) VisualSearch(ctx context.Context, params VisualSearchParams) (*VisualSearchResult, error) {
	res := &VisualSearchResult{}
	visualSearchPath := api.BuildPath(assets, visualSearch)
	var err error
	if params.ImageFile != nil {
		_, err = a.postFile(ctx, visualSearchPath, params.ImageFile, "image_file", params, res)
	} else {
		_, err = a.get(ctx, visualSearchPath, params, res)
	}

	return res, err
}

type VisualSearchResult struct {
	Assets     []api.BriefAssetResult `json:"resources"`
	TotalCount int                    `json:"total_count"`
	Error      api.ErrorResp          `json:"error,omitempty"`
}

// RestoreAssetsParams are the parameters for RestoreAssets.
type RestoreAssetsParams struct {
	AssetType    api.AssetType    `json:"-"`
	DeliveryType api.DeliveryType `json:"-"`
	PublicIDs    api.CldAPIArray  `json:"public_ids"`
	Versions     api.CldAPIArray  `json:"versions"`
}

// RestoreAssets reverts to the latest backed up version of the specified deleted assets.
//
// https://cloudinary.com/documentation/admin_api#restore_resources
func (a *API) RestoreAssets(ctx context.Context, params RestoreAssetsParams) (*RestoreAssetsResult, error) {
	res := &RestoreAssetsResult{}
	_, err := a.post(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType, restore), params, res)

	return res, err
}

// RestoreAssetsResult is the result of RestoreAssets.
type RestoreAssetsResult map[string]api.BriefAssetResult

// DeleteAssetsParams are the parameters for DeleteAssets.
type DeleteAssetsParams struct {
	AssetType       api.AssetType    `json:"-"`
	DeliveryType    api.DeliveryType `json:"-"`
	PublicIDs       api.CldAPIArray  `json:"public_ids"` // The public IDs of the assets to delete (up to 100).
	KeepOriginal    *bool            `json:"keep_original,omitempty"`
	Invalidate      *bool            `json:"invalidate,omitempty"`
	Transformations string           `json:"transformations,omitempty"`
	NextCursor      string           `json:"next_cursor,omitempty"`
}

// DeleteAssets deletes the specified assets.
//
// https://cloudinary.com/documentation/admin_api#delete_resources
func (a *API) DeleteAssets(ctx context.Context, params DeleteAssetsParams) (*DeleteAssetsResult, error) {
	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// DeleteAssetsResult  is the result of DeleteAssets.
type DeleteAssetsResult struct {
	Deleted       map[string]string      `json:"deleted"`
	DeletedCounts map[string]interface{} `json:"deleted_counts"`
	Partial       bool                   `json:"partial"`
	NextCursor    string                 `json:"next_cursor,omitempty"`
	Error         api.ErrorResp          `json:"error,omitempty"`
}

// DeleteAssetsByPrefixParams are the parameters for DeleteAssetsByPrefix.
type DeleteAssetsByPrefixParams struct {
	AssetType       api.AssetType    `json:"-"`
	DeliveryType    api.DeliveryType `json:"-"`
	Prefix          api.CldAPIArray  `json:"prefix"`
	KeepOriginal    *bool            `json:"keep_original,omitempty"`
	Invalidate      *bool            `json:"invalidate,omitempty"`
	Transformations string           `json:"transformations,omitempty"`
	NextCursor      string           `json:"next_cursor,omitempty"`
}

// DeleteAssetsByPrefix deletes assets by prefix.
//
// https://cloudinary.com/documentation/admin_api#delete_resources
func (a *API) DeleteAssetsByPrefix(ctx context.Context, params DeleteAssetsByPrefixParams) (*DeleteAssetsResult, error) {
	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// DeleteAssetsByTagParams are the parameters for DeleteAssetsByTag.
type DeleteAssetsByTagParams struct {
	AssetType       api.AssetType `json:"-"`
	Tag             string        `json:"-"`
	KeepOriginal    *bool         `json:"keep_original,omitempty"`
	Invalidate      *bool         `json:"invalidate,omitempty"`
	Transformations string        `json:"transformations,omitempty"`
	NextCursor      string        `json:"next_cursor,omitempty"`
}

// DeleteAssetsByTag deletes assets with the specified tag, including their derived resources.
//
// Supports deleting up to a maximum of 1000 original assets in a single call.
//
// https://cloudinary.com/documentation/admin_api#delete_resources_by_tags
func (a *API) DeleteAssetsByTag(ctx context.Context, params DeleteAssetsByTagParams) (*DeleteAssetsResult, error) {
	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, params.AssetType, tags, params.Tag), params, res)

	return res, err
}

// DeleteAllAssetsParams are the parameters for DeleteAllAssets.
type DeleteAllAssetsParams struct {
	AssetType       api.AssetType    `json:"-"`
	DeliveryType    api.DeliveryType `json:"-"`
	All             *bool            `json:"all"`
	KeepOriginal    *bool            `json:"keep_original,omitempty"`
	Invalidate      *bool            `json:"invalidate,omitempty"`
	Transformations string           `json:"transformations,omitempty"`
	NextCursor      string           `json:"next_cursor,omitempty"`
}

// DeleteAllAssets deletes all assets of the specified asset and delivery type, including their derived resources.
//
// Supports deleting up to a maximum of 1000 original assets in a single call.
//
// https://cloudinary.com/documentation/admin_api#delete_resources
func (a *API) DeleteAllAssets(ctx context.Context, params DeleteAllAssetsParams) (*DeleteAssetsResult, error) {
	params.All = api.Bool(true)

	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// DeleteDerivedAssetsParams are the parameters for DeleteDerivedAssets.
type DeleteDerivedAssetsParams struct {
	DerivedAssetIDs api.CldAPIArray `json:"derived_resource_ids"`
}

// DeleteDerivedAssets deletes the specified derived resources by derived resource ID.
//
// The derived resource IDs for a particular original asset are returned when calling the `resource` method to return
// the details of a single asset.
//
// https://cloudinary.com/documentation/admin_api##delete_resources
func (a *API) DeleteDerivedAssets(ctx context.Context, params DeleteDerivedAssetsParams) (*DeleteAssetsResult, error) {
	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(derivedAssets), params, res)

	return res, err
}

// DeleteDerivedAssetsByTransformationParams are the parameters for DeleteDerivedAssetsByTransformation.
type DeleteDerivedAssetsByTransformationParams struct {
	AssetType       api.AssetType    `json:"-"`
	DeliveryType    api.DeliveryType `json:"-"`
	PublicIDs       api.CldAPIArray  `json:"public_ids"`      // The public IDs for which you want to delete derived resources.
	Transformations string           `json:"transformations"` // The transformation(s) associated with the derived resources to delete.
	KeepOriginal    *bool            `json:"keep_original"`
	Invalidate      *bool            `json:"invalidate,omitempty"`
}

// DeleteDerivedAssetsByTransformation deletes derived resources identified by transformation and public_ids.
func (a *API) DeleteDerivedAssetsByTransformation(ctx context.Context, params DeleteDerivedAssetsByTransformationParams) (*DeleteAssetsResult, error) {
	params.KeepOriginal = api.Bool(true)

	res := &DeleteAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, params.AssetType, params.DeliveryType), params, res)

	return res, err
}

// AddRelatedAssetsParams are the parameters for AddRelatedAssets.
type AddRelatedAssetsParams struct {
	AssetType      api.AssetType    `json:"-"`
	DeliveryType   api.DeliveryType `json:"-"`
	PublicID       string           `json:"-"`
	AssetsToRelate []string         `json:"assets_to_relate"`
}

// AddRelatedAssets relates an asset to other assets by public IDs.
func (a *API) AddRelatedAssets(ctx context.Context, params AddRelatedAssetsParams) (*AddRelatedAssetsResult, error) {
	res := &AddRelatedAssetsResult{}
	_, err := a.post(ctx, api.BuildPath(assets, relatedAssets, params.AssetType, params.DeliveryType, params.PublicID), params, res)

	return res, err
}

// AddRelatedAssetsResult is the result of AddRelatedAssets.
type AddRelatedAssetsResult struct {
	Success []RelatedAssetResult `json:"success"`
	Failed  []RelatedAssetResult `json:"failed"`
}

// AddRelatedAssetsByAssetIDsParams are the parameters for AddRelatedAssetsByAssetIDs.
type AddRelatedAssetsByAssetIDsParams struct {
	AssetID        string   `json:"-"`
	AssetsToRelate []string `json:"assets_to_relate"`
}

// AddRelatedAssetsByAssetIDs relates an asset to other assets by asset IDs.
func (a *API) AddRelatedAssetsByAssetIDs(ctx context.Context, params AddRelatedAssetsByAssetIDsParams) (*AddRelatedAssetsResult, error) {
	res := &AddRelatedAssetsResult{}
	_, err := a.post(ctx, api.BuildPath(assets, relatedAssets, params.AssetID), params, res)

	return res, err
}

type RelatedAssetResult struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Asset   string `json:"asset"`
	Status  int    `json:"status"`
}

// DeleteRelatedAssetsParams are the parameters for DeleteRelatedAssets.
type DeleteRelatedAssetsParams struct {
	AssetType        api.AssetType    `json:"-"`
	DeliveryType     api.DeliveryType `json:"-"`
	PublicID         string           `json:"-"`
	AssetsToUnrelate []string         `json:"assets_to_unrelate"`
}

// DeleteRelatedAssets unrelates an asset from other assets by public IDs.
func (a *API) DeleteRelatedAssets(ctx context.Context, params DeleteRelatedAssetsParams) (*DeleteRelatedAssetsResult, error) {
	res := &DeleteRelatedAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, relatedAssets, params.AssetType, params.DeliveryType, params.PublicID), params, res)

	return res, err
}

// DeleteRelatedAssetsResult is the result of DeleteRelatedAssets.
type DeleteRelatedAssetsResult struct {
	Success []RelatedAssetResult `json:"success"`
	Failed  []RelatedAssetResult `json:"failed"`
}

// DeleteRelatedAssetsByAssetIDsParams are the parameters for DeleteRelatedAssetsByAssetIDs.
type DeleteRelatedAssetsByAssetIDsParams struct {
	AssetID          string   `json:"-"`
	AssetsToUnrelate []string `json:"assets_to_unrelate"`
}

// DeleteRelatedAssetsByAssetIDs unrelates an asset from other assets by asset IDs.
func (a *API) DeleteRelatedAssetsByAssetIDs(ctx context.Context, params DeleteRelatedAssetsByAssetIDsParams) (*DeleteRelatedAssetsResult, error) {
	res := &DeleteRelatedAssetsResult{}
	_, err := a.delete(ctx, api.BuildPath(assets, relatedAssets, params.AssetID), params, res)

	return res, err
}
