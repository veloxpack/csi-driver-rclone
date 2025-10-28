package asset

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin/search"
	"github.com/cloudinary/cloudinary-go/v2/config"
	"github.com/cloudinary/cloudinary-go/v2/internal/signature"
	"github.com/cloudinary/cloudinary-go/v2/logger"
	"strconv"
)

const searchAssetType = "search"
const searchAssetURLTTL = 300

// SearchURLAsset is the SearchURLAsset struct.
type SearchURLAsset struct {
	SearchQuery search.Query
	TTL         int
	NextCursor  string
	Config      config.Configuration
	logger      logger.Logger
}

// SearchURL returns a new SearchURLAsset instance from the provided query and configuration.
func SearchURL(query search.Query, conf *config.Configuration) (*SearchURLAsset, error) {
	if conf == nil {
		var err error
		conf, err = config.New()
		if err != nil {
			return nil, err
		}
	}

	asset := SearchURLAsset{SearchQuery: query, Config: *conf}
	asset.setDefaults()

	return &asset, nil
}

// setDefaults sets the default values.
func (sa *SearchURLAsset) setDefaults() {
	sa.TTL = searchAssetURLTTL
}

// String serializes SearchURLAsset to string.
func (sa SearchURLAsset) String() (result string, err error) {
	return sa.ToURL(0, "")
}

// ToURLWithNextCursor serializes SearchURLAsset to string.
func (sa SearchURLAsset) ToURLWithNextCursor(nextCursor string) (result string, err error) {
	return sa.ToURL(0, nextCursor)
}

// ToURL serializes SearchURLAsset to string.
func (sa SearchURLAsset) ToURL(ttl int, nextCursor string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("failed to build URL: %v", r)
			sa.logger.Error(msg)
			result = ""
			err = errors.New(msg)
		}
	}()

	path, err := sa.path(ttl, nextCursor)
	if err != nil {
		return "", err
	}

	assetURL := joinURL([]interface{}{distribution("", sa.Config), path})

	query := sa.query()

	return joinNonEmpty([]interface{}{assetURL, query}, "?"), nil
}

// path builds the URL path.
func (sa SearchURLAsset) path(ttl int, nextCursor string) (result string, err error) {
	if ttl == 0 {
		ttl = sa.TTL
	}

	if nextCursor == "" {
		nextCursor = sa.SearchQuery.NextCursor
	}

	b64Query, err := sa.b64SearchQuery()
	if err != nil {
		return "", err
	}

	toSign := strconv.Itoa(ttl) + b64Query
	sig, err := sa.signature(toSign)
	if err != nil {
		return "", err
	}

	return joinURL([]interface{}{searchAssetType, sig, ttl, b64Query, nextCursor}), nil
}

// query builds the URL query string.
func (sa SearchURLAsset) query() string {
	if !sa.Config.URL.Analytics {
		return ""
	}

	return fmt.Sprintf("%s=%s", queryString, sdkAnalyticsSignature())
}

// signature builds the signature.
func (sa SearchURLAsset) signature(toSign string) (result string, err error) {
	rawSignature, err := signature.Sign(toSign, sa.Config.Cloud.APISecret, signature.SHA256)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(rawSignature), nil
}

// b64SearchQuery encodes the search query using base64 encoding.
func (sa SearchURLAsset) b64SearchQuery() (result string, err error) {
	query := sa.SearchQuery
	query.NextCursor = "" // drop next_cursor
	jsonBytes, err := api.MarshalJSONRaw(query)
	if err != nil {
		return "", err
	}

	jsonBytes, err = api.ReMarshalJSON(jsonBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}
