package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Action struct {
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

func (a Action) Identifier() interface{} {
	return a.Id
}

type ActionCollection []Action

func (a *Action) UnmarshalJSON(data []byte) error {
	type action Action
	var v action
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = Action(v)
	return nil
}

func (a *ActionCollection) UnmarshalJSON(data []byte) error {
	type actions ActionCollection
	var v actions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = ActionCollection(v)
	return nil
}

func (a *ActionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
