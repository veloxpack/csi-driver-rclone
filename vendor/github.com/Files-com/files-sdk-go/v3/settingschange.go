package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SettingsChange struct {
	ApiKeyId             int64      `json:"api_key_id,omitempty" path:"api_key_id,omitempty" url:"api_key_id,omitempty"`
	Changes              []string   `json:"changes,omitempty" path:"changes,omitempty" url:"changes,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UserId               int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	UserIsFilesSupport   *bool      `json:"user_is_files_support,omitempty" path:"user_is_files_support,omitempty" url:"user_is_files_support,omitempty"`
	UserIsFromParentSite *bool      `json:"user_is_from_parent_site,omitempty" path:"user_is_from_parent_site,omitempty" url:"user_is_from_parent_site,omitempty"`
	Username             string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
}

// Identifier no path or id

type SettingsChangeCollection []SettingsChange

type SettingsChangeListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter SettingsChange         `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

func (s *SettingsChange) UnmarshalJSON(data []byte) error {
	type settingsChange SettingsChange
	var v settingsChange
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SettingsChange(v)
	return nil
}

func (s *SettingsChangeCollection) UnmarshalJSON(data []byte) error {
	type settingsChanges SettingsChangeCollection
	var v settingsChanges
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SettingsChangeCollection(v)
	return nil
}

func (s *SettingsChangeCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
