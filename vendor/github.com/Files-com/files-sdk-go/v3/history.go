package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type History struct {
	Id                   int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path                 string                 `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	When                 *time.Time             `json:"when,omitempty" path:"when,omitempty" url:"when,omitempty"`
	Destination          string                 `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	Display              string                 `json:"display,omitempty" path:"display,omitempty" url:"display,omitempty"`
	Ip                   string                 `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	Source               string                 `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	Targets              map[string]interface{} `json:"targets,omitempty" path:"targets,omitempty" url:"targets,omitempty"`
	UserId               int64                  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username             string                 `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	UserIsFromParentSite *bool                  `json:"user_is_from_parent_site,omitempty" path:"user_is_from_parent_site,omitempty" url:"user_is_from_parent_site,omitempty"`
	Action               string                 `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	FailureType          string                 `json:"failure_type,omitempty" path:"failure_type,omitempty" url:"failure_type,omitempty"`
	Interface            string                 `json:"interface,omitempty" path:"interface,omitempty" url:"interface,omitempty"`
}

func (h History) Identifier() interface{} {
	return h.Id
}

type HistoryCollection []History

type HistoryListForFileParams struct {
	StartAt *time.Time             `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt   *time.Time             `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	Display string                 `url:"display,omitempty" json:"display,omitempty" path:"display"`
	SortBy  map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Path    string                 `url:"-,omitempty" json:"-,omitempty" path:"path"`
	ListParams
}

type HistoryListForFolderParams struct {
	StartAt *time.Time             `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt   *time.Time             `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	Display string                 `url:"display,omitempty" json:"display,omitempty" path:"display"`
	SortBy  map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Path    string                 `url:"-,omitempty" json:"-,omitempty" path:"path"`
	ListParams
}

type HistoryListForUserParams struct {
	StartAt *time.Time             `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt   *time.Time             `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	Display string                 `url:"display,omitempty" json:"display,omitempty" path:"display"`
	SortBy  map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	UserId  int64                  `url:"-,omitempty" json:"-,omitempty" path:"user_id"`
	ListParams
}

type HistoryListLoginsParams struct {
	StartAt *time.Time             `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt   *time.Time             `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	Display string                 `url:"display,omitempty" json:"display,omitempty" path:"display"`
	SortBy  map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type HistoryListParams struct {
	StartAt      *time.Time             `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt        *time.Time             `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	Display      string                 `url:"display,omitempty" json:"display,omitempty" path:"display"`
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       History                `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

func (h *History) UnmarshalJSON(data []byte) error {
	type history History
	var v history
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*h = History(v)
	return nil
}

func (h *HistoryCollection) UnmarshalJSON(data []byte) error {
	type historys HistoryCollection
	var v historys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*h = HistoryCollection(v)
	return nil
}

func (h *HistoryCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*h))
	for i, v := range *h {
		ret[i] = v
	}

	return &ret
}
