package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ActionLog struct {
	Action             string     `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Destination        string     `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	FailureType        string     `json:"failure_type,omitempty" path:"failure_type,omitempty" url:"failure_type,omitempty"`
	Folder             string     `json:"folder,omitempty" path:"folder,omitempty" url:"folder,omitempty"`
	Interface          string     `json:"interface,omitempty" path:"interface,omitempty" url:"interface,omitempty"`
	Ip                 string     `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	MetadataDmId       int64      `json:"metadata_dm_id,omitempty" path:"metadata_dm_id,omitempty" url:"metadata_dm_id,omitempty"`
	ParentMetadataDmId int64      `json:"parent_metadata_dm_id,omitempty" path:"parent_metadata_dm_id,omitempty" url:"parent_metadata_dm_id,omitempty"`
	Path               string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	SiteId             int64      `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
	Src                string     `json:"src,omitempty" path:"src,omitempty" url:"src,omitempty"`
	UserId             int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username           string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
}

func (a ActionLog) Identifier() interface{} {
	return a.Path
}

type ActionLogCollection []ActionLog

type ActionLogListParams struct {
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (a *ActionLog) UnmarshalJSON(data []byte) error {
	type actionLog ActionLog
	var v actionLog
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = ActionLog(v)
	return nil
}

func (a *ActionLogCollection) UnmarshalJSON(data []byte) error {
	type actionLogs ActionLogCollection
	var v actionLogs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = ActionLogCollection(v)
	return nil
}

func (a *ActionLogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
