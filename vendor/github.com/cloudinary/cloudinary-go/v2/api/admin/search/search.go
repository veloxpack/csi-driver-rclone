package search

// Query struct includes the search query.
type Query struct {
	// Expression is the (Lucene-like) string expression specifying the search query.
	// If not provided then all resources are listed (up to MaxResults).
	Expression string `json:"expression,omitempty"`
	// SortBy is the field to sort by. You can specify more than one SortBy parameter; results will be sorted
	// according to the order of the fields provided.
	SortBy []SortByField `json:"sort_by,omitempty"`
	// Aggregate is the name of a field (attribute) for which an aggregation count should be calculated and returned in the response.
	// (Tier 2 only)
	// You can specify more than one aggregate parameter.
	// For aggregation fields without discrete values, the results are divided into categories.
	Aggregate []Aggregation `json:"aggregate,omitempty"`
	// WithField contains names of additional asset attributes to include for each asset in the response.
	WithField []WithField `json:"with_field,omitempty"`
	// Fields contains names of the asset attributes to keep for each asset in the response.
	Fields []string `json:"fields,omitempty"`
	// MaxResults is the maximum number of results to return. Default 50. Maximum 500.
	MaxResults int `json:"max_results,omitempty"`
	// NextCursor value is returned as part of the response when a search request has more results to return than MaxResults.
	// You can then specify this value as the NextCursor parameter of the following request.
	NextCursor string `json:"next_cursor,omitempty"`
}

// Aggregation is the aggregation field.
type Aggregation = string

const (
	// AssetType aggregation field.
	AssetType Aggregation = "resource_type"
	// DeliveryType aggregation field.
	DeliveryType = "type"
	// Pixels aggregation field. Only the image assets in the response are aggregated.
	Pixels = "pixels"
	// Duration aggregation field. Only the video assets in the response are aggregated.
	Duration = "duration"
	// Format aggregation field.
	Format = "format"
	// Bytes aggregation field.
	Bytes = "bytes"
)

// WithField is the name of the addition filed to include in result.
type WithField = string

const (
	// ContextField is the context field.
	ContextField WithField = "context"
	// TagsField is the tags field.
	TagsField = "tags"
	// ImageMetadataField is the image metadata field.
	ImageMetadataField = "image_metadata"
	// ImageAnalysisField is the image analysis field.
	ImageAnalysisField = "image_analysis"
)

// Direction is the sorting direction.
type Direction string

const (
	// Ascending direction.
	Ascending Direction = "asc"
	// Descending direction.
	Descending = "desc"
)

// SortByField is the field to sort by and direction.
type SortByField map[string]Direction
