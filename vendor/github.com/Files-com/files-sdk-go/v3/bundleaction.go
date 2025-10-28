package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type BundleAction struct {
	Action                  string             `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	BundleRegistration      BundleRegistration `json:"bundle_registration,omitempty" path:"bundle_registration,omitempty" url:"bundle_registration,omitempty"`
	CreatedAt               *time.Time         `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Destination             string             `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	Path                    string             `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Source                  string             `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	BundleRegistrationId    int64              `json:"bundle_registration_id,omitempty" path:"bundle_registration_id,omitempty" url:"bundle_registration_id,omitempty"`
	BundleRegistrationName  string             `json:"bundle_registration_name,omitempty" path:"bundle_registration_name,omitempty" url:"bundle_registration_name,omitempty"`
	BundleRegistrationEmail string             `json:"bundle_registration_email,omitempty" path:"bundle_registration_email,omitempty" url:"bundle_registration_email,omitempty"`
	BundleRegistrationIp    string             `json:"bundle_registration_ip,omitempty" path:"bundle_registration_ip,omitempty" url:"bundle_registration_ip,omitempty"`
}

func (b BundleAction) Identifier() interface{} {
	return b.Path
}

type BundleActionCollection []BundleAction

type BundleActionListParams struct {
	UserId     int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy     map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     BundleAction           `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (b *BundleAction) UnmarshalJSON(data []byte) error {
	type bundleAction BundleAction
	var v bundleAction
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = BundleAction(v)
	return nil
}

func (b *BundleActionCollection) UnmarshalJSON(data []byte) error {
	type bundleActions BundleActionCollection
	var v bundleActions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BundleActionCollection(v)
	return nil
}

func (b *BundleActionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
