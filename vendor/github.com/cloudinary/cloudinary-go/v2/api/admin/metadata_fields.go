package admin

import (
	"context"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin/metadata"
)

const (
	metadataFields    api.EndPoint = "metadata_fields"
	dataSource        api.EndPoint = "datasource"
	dataSourceRestore api.EndPoint = "datasource_restore"
	order             api.EndPoint = "order"
)

// ListMetadataFields lists all metadata field definitions.
//
// https://cloudinary.com/documentation/admin_api#get_metadata_fields
func (a *API) ListMetadataFields(ctx context.Context) (*ListMetadataFieldsResult, error) {
	res := &ListMetadataFieldsResult{}
	_, err := a.get(ctx, metadataFields, nil, res)

	return res, err
}

// ListMetadataFieldsResult is the result of ListMetadataFields.
type ListMetadataFieldsResult struct {
	MetadataFields []metadata.Field `json:"metadata_fields"`
	Error          api.ErrorResp    `json:"error,omitempty"`
	Response       interface{}
}

// MetadataFieldByFieldIDParams are the parameters for MetadataFieldByFieldID.
type MetadataFieldByFieldIDParams struct {
	FieldExternalID string `json:"-"`
}

// MetadataFieldByFieldID gets a single metadata field definition by external ID.
//
// https://cloudinary.com/documentation/admin_api#get_a_metadata_field_by_external_id
func (a *API) MetadataFieldByFieldID(ctx context.Context, params MetadataFieldByFieldIDParams) (*MetadataFieldByFieldIDResult, error) {
	res := &MetadataFieldByFieldIDResult{}
	_, err := a.get(ctx, api.BuildPath(metadataFields, params.FieldExternalID), params, res)

	return res, err
}

// MetadataFieldByFieldIDResult is the result of MetadataFieldByFieldID.
type MetadataFieldByFieldIDResult struct {
	metadata.Field
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// AddMetadataField creates a new metadata field definition.
//
// https://cloudinary.com/documentation/admin_api#create_a_metadata_field
func (a *API) AddMetadataField(ctx context.Context, params metadata.Field) (*AddMetadataFieldResult, error) {
	res := &AddMetadataFieldResult{}
	_, err := a.post(ctx, metadataFields, params, res)

	return res, err
}

// AddMetadataFieldResult is the result of AddMetadataField.
type AddMetadataFieldResult struct {
	metadata.Field
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// UpdateMetadataFieldParams are the parameters for UpdateMetadataField.
type UpdateMetadataFieldParams struct {
	metadata.Field
	FieldExternalID string `json:"-"`
}

// UpdateMetadataField updates a metadata field by external ID.
//
// Updates a metadata field definition (partially, no need to pass the entire object) passed as JSON data.
//
// https://cloudinary.com/documentation/admin_api#update_a_metadata_field_by_external_id
func (a *API) UpdateMetadataField(ctx context.Context, params UpdateMetadataFieldParams) (*UpdateMetadataFieldResult, error) {
	res := &UpdateMetadataFieldResult{}
	_, err := a.put(ctx, api.BuildPath(metadataFields, params.FieldExternalID), params, res)

	return res, err
}

// UpdateMetadataFieldResult is the result of UpdateMetadataField.
type UpdateMetadataFieldResult struct {
	metadata.Field
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// DeleteMetadataFieldParams are the parameters for DeleteMetadataField.
type DeleteMetadataFieldParams struct {
	FieldExternalID string `json:"-"`
}

// DeleteMetadataField deletes a metadata field definition by external ID.
//
// The external ID is immutable. Therefore, once deleted, the field's external ID can no longer be used for future purposes.
//
// https://cloudinary.com/documentation/admin_api#delete_a_metadata_field_by_external_id
func (a *API) DeleteMetadataField(ctx context.Context, params DeleteMetadataFieldParams) (*DeleteMetadataFieldResult, error) {
	res := &DeleteMetadataFieldResult{}
	_, err := a.delete(ctx, api.BuildPath(metadataFields, params.FieldExternalID), params, res)

	return res, err
}

// DeleteMetadataFieldResult is the result of DeleteMetadataField.
type DeleteMetadataFieldResult struct {
	Message  string        `json:"message"`
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// DeleteDataSourceEntriesParams are the parameters for DeleteDataSourceEntries.
type DeleteDataSourceEntriesParams struct {
	FieldExternalID    string   `json:"-"`
	EntriesExternalIDs []string `json:"external_ids"`
}

// DeleteDataSourceEntries deletes entries in a metadata single or multi-select field's datasource.
//
// Deletes (blocks) the datasource (list) entries from the specified metadata field definition. Sets the state of
// the entries to inactive. This is a soft delete. The entries still exist in the database and can be reactivated
// using the RestoreMetadataFieldDataSource method.
//
// https://cloudinary.com/documentation/admin_api#delete_entries_in_a_metadata_field_datasource
func (a *API) DeleteDataSourceEntries(ctx context.Context, params DeleteDataSourceEntriesParams) (*DeleteDataSourceEntriesResult, error) {
	res := &DeleteDataSourceEntriesResult{}
	_, err := a.delete(ctx, api.BuildPath(metadataFields, params.FieldExternalID, dataSource), params, res)

	return res, err
}

// DeleteDataSourceEntriesResult is the result of DeleteDataSourceEntries.
type DeleteDataSourceEntriesResult struct {
	metadata.DataSource
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// UpdateMetadataFieldDataSourceParams are the parameters for UpdateMetadataFieldDataSource.
type UpdateMetadataFieldDataSourceParams struct {
	metadata.DataSource
	FieldExternalID string `json:"-"`
}

// UpdateMetadataFieldDataSource updates a metadata field datasource.
//
// Updates the datasource of a supported field type (currently enum or set), passed as JSON data. The
// update is partial: datasource entries with an existing external_id will be updated and entries with new
// external_id’s (or without external_id’s) will be appended.
//
// https://cloudinary.com/documentation/admin_api#update_a_metadata_field_datasource
func (a *API) UpdateMetadataFieldDataSource(ctx context.Context, params UpdateMetadataFieldDataSourceParams) (*UpdateMetadataFieldDataSourceResult, error) {
	res := &UpdateMetadataFieldDataSourceResult{}
	_, err := a.put(ctx, api.BuildPath(metadataFields, params.FieldExternalID, dataSource), params, res)

	return res, err
}

// UpdateMetadataFieldDataSourceResult is the result of UpdateMetadataFieldDataSource.
type UpdateMetadataFieldDataSourceResult struct {
	metadata.DataSource
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// RestoreDatasourceEntriesParams are the parameters for RestoreDatasourceEntries.
type RestoreDatasourceEntriesParams struct {
	FieldExternalID    string   `json:"-"`
	EntriesExternalIDs []string `json:"external_ids"`
}

// RestoreDatasourceEntries restores entries in a metadata field datasource.
//
// Restores (unblocks) any previously deleted datasource entries for a specified metadata field definition.
// Sets the state of the entries to active.
//
// https://cloudinary.com/documentation/admin_api#restore_entries_in_a_metadata_field_datasource
func (a *API) RestoreDatasourceEntries(ctx context.Context, params RestoreDatasourceEntriesParams) (*RestoreDatasourceEntriesResult, error) {
	res := &RestoreDatasourceEntriesResult{}
	_, err := a.post(ctx, api.BuildPath(metadataFields, params.FieldExternalID, dataSourceRestore), params, res)

	return res, err
}

// RestoreDatasourceEntriesResult is the result of RestoreDatasourceEntries.
type RestoreDatasourceEntriesResult struct {
	metadata.DataSource
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// ReorderMetadataFieldDatasourceParams are the parameters for ReorderMetadataFieldDatasource.
type ReorderMetadataFieldDatasourceParams struct {
	FieldExternalID string       `json:"-"`
	FieldOrderBy    OrderByField `json:"order_by"`
	FieldDirection  Direction    `json:"direction,omitempty"`
}

// ReorderMetadataFieldDatasource reorders metadata fields datasource. Currently, supports only value.
func (a *API) ReorderMetadataFieldDatasource(ctx context.Context, params ReorderMetadataFieldDatasourceParams) (*ReorderMetadataFieldDatasourceResult, error) {
	res := &ReorderMetadataFieldDatasourceResult{}
	_, err := a.post(ctx, api.BuildPath(metadataFields, params.FieldExternalID, dataSource, order), params, res)

	return res, err
}

// ReorderMetadataFieldDatasourceResult is the result of ReorderMetadataFieldDatasource.
type ReorderMetadataFieldDatasourceResult struct {
	metadata.DataSource
	Error    api.ErrorResp `json:"error,omitempty"`
	Response interface{}
}

// ReorderMetadataFieldsParams are the parameters for ReorderMetadataFields.
type ReorderMetadataFieldsParams struct {
	FieldOrderBy   OrderByField `json:"order_by"`
	FieldDirection Direction    `json:"direction,omitempty"`
}

// ReorderMetadataFields reorders metadata fields.
func (a *API) ReorderMetadataFields(ctx context.Context, params ReorderMetadataFieldsParams) (*ReorderMetadataFieldsResult, error) {
	res := &ReorderMetadataFieldsResult{}
	_, err := a.put(ctx, api.BuildPath(metadataFields, order), params, res)

	return res, err
}

// ReorderMetadataFieldsResult is the result of ReorderMetadataFields.
type ReorderMetadataFieldsResult struct {
	MetadataFields []metadata.Field `json:"metadata_fields"`
	Error          api.ErrorResp    `json:"error,omitempty"`
	Response       interface{}
}
