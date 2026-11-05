package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SsoEvent struct {
	Id            int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	EventType     string     `json:"event_type,omitempty" path:"event_type,omitempty" url:"event_type,omitempty"`
	Status        string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Body          string     `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	EventErrors   []string   `json:"event_errors,omitempty" path:"event_errors,omitempty" url:"event_errors,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	BodyUrl       string     `json:"body_url,omitempty" path:"body_url,omitempty" url:"body_url,omitempty"`
	UserId        int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username      string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	IdpUid        string     `json:"idp_uid,omitempty" path:"idp_uid,omitempty" url:"idp_uid,omitempty"`
	Provider      string     `json:"provider,omitempty" path:"provider,omitempty" url:"provider,omitempty"`
	ProviderLabel string     `json:"provider_label,omitempty" path:"provider_label,omitempty" url:"provider_label,omitempty"`
	Ip            string     `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	Region        string     `json:"region,omitempty" path:"region,omitempty" url:"region,omitempty"`
}

func (s SsoEvent) Identifier() interface{} {
	return s.Id
}

type SsoEventCollection []SsoEvent

type SsoEventListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type SsoEventFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *SsoEvent) UnmarshalJSON(data []byte) error {
	type ssoEvent SsoEvent
	var v ssoEvent
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SsoEvent(v)
	return nil
}

func (s *SsoEventCollection) UnmarshalJSON(data []byte) error {
	type ssoEvents SsoEventCollection
	var v ssoEvents
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SsoEventCollection(v)
	return nil
}

func (s *SsoEventCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
