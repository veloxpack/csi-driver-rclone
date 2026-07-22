package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Restore struct {
	EarliestDate                           *time.Time `json:"earliest_date,omitempty" path:"earliest_date,omitempty" url:"earliest_date,omitempty"`
	Id                                     int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	DirsRestored                           int64      `json:"dirs_restored,omitempty" path:"dirs_restored,omitempty" url:"dirs_restored,omitempty"`
	DirsErrored                            int64      `json:"dirs_errored,omitempty" path:"dirs_errored,omitempty" url:"dirs_errored,omitempty"`
	DirsTotal                              int64      `json:"dirs_total,omitempty" path:"dirs_total,omitempty" url:"dirs_total,omitempty"`
	FilesRestored                          int64      `json:"files_restored,omitempty" path:"files_restored,omitempty" url:"files_restored,omitempty"`
	FilesErrored                           int64      `json:"files_errored,omitempty" path:"files_errored,omitempty" url:"files_errored,omitempty"`
	FilesTotal                             int64      `json:"files_total,omitempty" path:"files_total,omitempty" url:"files_total,omitempty"`
	Prefix                                 string     `json:"prefix,omitempty" path:"prefix,omitempty" url:"prefix,omitempty"`
	RestorationType                        string     `json:"restoration_type,omitempty" path:"restoration_type,omitempty" url:"restoration_type,omitempty"`
	RestoreInPlace                         *bool      `json:"restore_in_place,omitempty" path:"restore_in_place,omitempty" url:"restore_in_place,omitempty"`
	RestoreDeletedPermissions              *bool      `json:"restore_deleted_permissions,omitempty" path:"restore_deleted_permissions,omitempty" url:"restore_deleted_permissions,omitempty"`
	UsersRestored                          int64      `json:"users_restored,omitempty" path:"users_restored,omitempty" url:"users_restored,omitempty"`
	UsersErrored                           int64      `json:"users_errored,omitempty" path:"users_errored,omitempty" url:"users_errored,omitempty"`
	UsersTotal                             int64      `json:"users_total,omitempty" path:"users_total,omitempty" url:"users_total,omitempty"`
	ApiKeysRestored                        int64      `json:"api_keys_restored,omitempty" path:"api_keys_restored,omitempty" url:"api_keys_restored,omitempty"`
	PublicKeysRestored                     int64      `json:"public_keys_restored,omitempty" path:"public_keys_restored,omitempty" url:"public_keys_restored,omitempty"`
	TwoFactorAuthenticationMethodsRestored int64      `json:"two_factor_authentication_methods_restored,omitempty" path:"two_factor_authentication_methods_restored,omitempty" url:"two_factor_authentication_methods_restored,omitempty"`
	Status                                 string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	UpdateTimestamps                       *bool      `json:"update_timestamps,omitempty" path:"update_timestamps,omitempty" url:"update_timestamps,omitempty"`
	WorkspaceId                            int64      `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	ErrorMessages                          []string   `json:"error_messages,omitempty" path:"error_messages,omitempty" url:"error_messages,omitempty"`
}

func (r Restore) Identifier() interface{} {
	return r.Id
}

type RestoreCollection []Restore

type RestoreRestorationTypeEnum string

func (u RestoreRestorationTypeEnum) String() string {
	return string(u)
}

func (u RestoreRestorationTypeEnum) Enum() map[string]RestoreRestorationTypeEnum {
	return map[string]RestoreRestorationTypeEnum{
		"files": RestoreRestorationTypeEnum("files"),
		"users": RestoreRestorationTypeEnum("users"),
	}
}

type RestoreListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type RestoreCreateParams struct {
	EarliestDate              *time.Time                 `url:"earliest_date" json:"earliest_date" path:"earliest_date"`
	Prefix                    string                     `url:"prefix,omitempty" json:"prefix,omitempty" path:"prefix"`
	RestorationType           RestoreRestorationTypeEnum `url:"restoration_type,omitempty" json:"restoration_type,omitempty" path:"restoration_type"`
	RestoreDeletedPermissions *bool                      `url:"restore_deleted_permissions,omitempty" json:"restore_deleted_permissions,omitempty" path:"restore_deleted_permissions"`
	RestoreInPlace            *bool                      `url:"restore_in_place,omitempty" json:"restore_in_place,omitempty" path:"restore_in_place"`
	UpdateTimestamps          *bool                      `url:"update_timestamps,omitempty" json:"update_timestamps,omitempty" path:"update_timestamps"`
	WorkspaceId               int64                      `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

func (r *Restore) UnmarshalJSON(data []byte) error {
	type restore Restore
	var v restore
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = Restore(v)
	return nil
}

func (r *RestoreCollection) UnmarshalJSON(data []byte) error {
	type restores RestoreCollection
	var v restores
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RestoreCollection(v)
	return nil
}

func (r *RestoreCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
