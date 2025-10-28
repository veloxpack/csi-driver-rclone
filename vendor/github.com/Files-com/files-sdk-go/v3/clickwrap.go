package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Clickwrap struct {
	Id             int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name           string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Body           string `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	UseWithUsers   string `json:"use_with_users,omitempty" path:"use_with_users,omitempty" url:"use_with_users,omitempty"`
	UseWithBundles string `json:"use_with_bundles,omitempty" path:"use_with_bundles,omitempty" url:"use_with_bundles,omitempty"`
	UseWithInboxes string `json:"use_with_inboxes,omitempty" path:"use_with_inboxes,omitempty" url:"use_with_inboxes,omitempty"`
}

func (c Clickwrap) Identifier() interface{} {
	return c.Id
}

type ClickwrapCollection []Clickwrap

type ClickwrapUseWithBundlesEnum string

func (u ClickwrapUseWithBundlesEnum) String() string {
	return string(u)
}

func (u ClickwrapUseWithBundlesEnum) Enum() map[string]ClickwrapUseWithBundlesEnum {
	return map[string]ClickwrapUseWithBundlesEnum{
		"none":                   ClickwrapUseWithBundlesEnum("none"),
		"available":              ClickwrapUseWithBundlesEnum("available"),
		"require":                ClickwrapUseWithBundlesEnum("require"),
		"available_to_all_users": ClickwrapUseWithBundlesEnum("available_to_all_users"),
	}
}

type ClickwrapUseWithInboxesEnum string

func (u ClickwrapUseWithInboxesEnum) String() string {
	return string(u)
}

func (u ClickwrapUseWithInboxesEnum) Enum() map[string]ClickwrapUseWithInboxesEnum {
	return map[string]ClickwrapUseWithInboxesEnum{
		"none":                   ClickwrapUseWithInboxesEnum("none"),
		"available":              ClickwrapUseWithInboxesEnum("available"),
		"require":                ClickwrapUseWithInboxesEnum("require"),
		"available_to_all_users": ClickwrapUseWithInboxesEnum("available_to_all_users"),
	}
}

type ClickwrapUseWithUsersEnum string

func (u ClickwrapUseWithUsersEnum) String() string {
	return string(u)
}

func (u ClickwrapUseWithUsersEnum) Enum() map[string]ClickwrapUseWithUsersEnum {
	return map[string]ClickwrapUseWithUsersEnum{
		"none":    ClickwrapUseWithUsersEnum("none"),
		"require": ClickwrapUseWithUsersEnum("require"),
	}
}

type ClickwrapListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type ClickwrapFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ClickwrapCreateParams struct {
	Name           string                      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Body           string                      `url:"body,omitempty" json:"body,omitempty" path:"body"`
	UseWithBundles ClickwrapUseWithBundlesEnum `url:"use_with_bundles,omitempty" json:"use_with_bundles,omitempty" path:"use_with_bundles"`
	UseWithInboxes ClickwrapUseWithInboxesEnum `url:"use_with_inboxes,omitempty" json:"use_with_inboxes,omitempty" path:"use_with_inboxes"`
	UseWithUsers   ClickwrapUseWithUsersEnum   `url:"use_with_users,omitempty" json:"use_with_users,omitempty" path:"use_with_users"`
}

type ClickwrapUpdateParams struct {
	Id             int64                       `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name           string                      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Body           string                      `url:"body,omitempty" json:"body,omitempty" path:"body"`
	UseWithBundles ClickwrapUseWithBundlesEnum `url:"use_with_bundles,omitempty" json:"use_with_bundles,omitempty" path:"use_with_bundles"`
	UseWithInboxes ClickwrapUseWithInboxesEnum `url:"use_with_inboxes,omitempty" json:"use_with_inboxes,omitempty" path:"use_with_inboxes"`
	UseWithUsers   ClickwrapUseWithUsersEnum   `url:"use_with_users,omitempty" json:"use_with_users,omitempty" path:"use_with_users"`
}

type ClickwrapDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (c *Clickwrap) UnmarshalJSON(data []byte) error {
	type clickwrap Clickwrap
	var v clickwrap
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*c = Clickwrap(v)
	return nil
}

func (c *ClickwrapCollection) UnmarshalJSON(data []byte) error {
	type clickwraps ClickwrapCollection
	var v clickwraps
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*c = ClickwrapCollection(v)
	return nil
}

func (c *ClickwrapCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*c))
	for i, v := range *c {
		ret[i] = v
	}

	return &ret
}
