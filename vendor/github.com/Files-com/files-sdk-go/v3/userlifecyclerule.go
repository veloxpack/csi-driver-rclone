package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type UserLifecycleRule struct {
	Id                   int64   `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	AuthenticationMethod string  `json:"authentication_method,omitempty" path:"authentication_method,omitempty" url:"authentication_method,omitempty"`
	GroupIds             []int64 `json:"group_ids,omitempty" path:"group_ids,omitempty" url:"group_ids,omitempty"`
	Action               string  `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	InactivityDays       int64   `json:"inactivity_days,omitempty" path:"inactivity_days,omitempty" url:"inactivity_days,omitempty"`
	IncludeFolderAdmins  *bool   `json:"include_folder_admins,omitempty" path:"include_folder_admins,omitempty" url:"include_folder_admins,omitempty"`
	IncludeSiteAdmins    *bool   `json:"include_site_admins,omitempty" path:"include_site_admins,omitempty" url:"include_site_admins,omitempty"`
	Name                 string  `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	PartnerTag           string  `json:"partner_tag,omitempty" path:"partner_tag,omitempty" url:"partner_tag,omitempty"`
	SiteId               int64   `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
	UserState            string  `json:"user_state,omitempty" path:"user_state,omitempty" url:"user_state,omitempty"`
	UserTag              string  `json:"user_tag,omitempty" path:"user_tag,omitempty" url:"user_tag,omitempty"`
}

func (u UserLifecycleRule) Identifier() interface{} {
	return u.Id
}

type UserLifecycleRuleCollection []UserLifecycleRule

type UserLifecycleRuleActionEnum string

func (u UserLifecycleRuleActionEnum) String() string {
	return string(u)
}

func (u UserLifecycleRuleActionEnum) Enum() map[string]UserLifecycleRuleActionEnum {
	return map[string]UserLifecycleRuleActionEnum{
		"disable": UserLifecycleRuleActionEnum("disable"),
		"delete":  UserLifecycleRuleActionEnum("delete"),
	}
}

type UserLifecycleRuleAuthenticationMethodEnum string

func (u UserLifecycleRuleAuthenticationMethodEnum) String() string {
	return string(u)
}

func (u UserLifecycleRuleAuthenticationMethodEnum) Enum() map[string]UserLifecycleRuleAuthenticationMethodEnum {
	return map[string]UserLifecycleRuleAuthenticationMethodEnum{
		"all":                         UserLifecycleRuleAuthenticationMethodEnum("all"),
		"password":                    UserLifecycleRuleAuthenticationMethodEnum("password"),
		"sso":                         UserLifecycleRuleAuthenticationMethodEnum("sso"),
		"none":                        UserLifecycleRuleAuthenticationMethodEnum("none"),
		"email_signup":                UserLifecycleRuleAuthenticationMethodEnum("email_signup"),
		"password_with_imported_hash": UserLifecycleRuleAuthenticationMethodEnum("password_with_imported_hash"),
		"password_and_ssh_key":        UserLifecycleRuleAuthenticationMethodEnum("password_and_ssh_key"),
	}
}

type UserLifecycleRuleUserStateEnum string

func (u UserLifecycleRuleUserStateEnum) String() string {
	return string(u)
}

func (u UserLifecycleRuleUserStateEnum) Enum() map[string]UserLifecycleRuleUserStateEnum {
	return map[string]UserLifecycleRuleUserStateEnum{
		"inactive": UserLifecycleRuleUserStateEnum("inactive"),
		"disabled": UserLifecycleRuleUserStateEnum("disabled"),
	}
}

type UserLifecycleRuleListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type UserLifecycleRuleFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type UserLifecycleRuleCreateParams struct {
	Action               UserLifecycleRuleActionEnum               `url:"action,omitempty" json:"action,omitempty" path:"action"`
	AuthenticationMethod UserLifecycleRuleAuthenticationMethodEnum `url:"authentication_method,omitempty" json:"authentication_method,omitempty" path:"authentication_method"`
	GroupIds             []int64                                   `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	InactivityDays       int64                                     `url:"inactivity_days,omitempty" json:"inactivity_days,omitempty" path:"inactivity_days"`
	IncludeSiteAdmins    *bool                                     `url:"include_site_admins,omitempty" json:"include_site_admins,omitempty" path:"include_site_admins"`
	IncludeFolderAdmins  *bool                                     `url:"include_folder_admins,omitempty" json:"include_folder_admins,omitempty" path:"include_folder_admins"`
	Name                 string                                    `url:"name,omitempty" json:"name,omitempty" path:"name"`
	PartnerTag           string                                    `url:"partner_tag,omitempty" json:"partner_tag,omitempty" path:"partner_tag"`
	UserState            UserLifecycleRuleUserStateEnum            `url:"user_state,omitempty" json:"user_state,omitempty" path:"user_state"`
	UserTag              string                                    `url:"user_tag,omitempty" json:"user_tag,omitempty" path:"user_tag"`
}

type UserLifecycleRuleUpdateParams struct {
	Id                   int64                                     `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Action               UserLifecycleRuleActionEnum               `url:"action,omitempty" json:"action,omitempty" path:"action"`
	AuthenticationMethod UserLifecycleRuleAuthenticationMethodEnum `url:"authentication_method,omitempty" json:"authentication_method,omitempty" path:"authentication_method"`
	GroupIds             []int64                                   `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	InactivityDays       int64                                     `url:"inactivity_days,omitempty" json:"inactivity_days,omitempty" path:"inactivity_days"`
	IncludeSiteAdmins    *bool                                     `url:"include_site_admins,omitempty" json:"include_site_admins,omitempty" path:"include_site_admins"`
	IncludeFolderAdmins  *bool                                     `url:"include_folder_admins,omitempty" json:"include_folder_admins,omitempty" path:"include_folder_admins"`
	Name                 string                                    `url:"name,omitempty" json:"name,omitempty" path:"name"`
	PartnerTag           string                                    `url:"partner_tag,omitempty" json:"partner_tag,omitempty" path:"partner_tag"`
	UserState            UserLifecycleRuleUserStateEnum            `url:"user_state,omitempty" json:"user_state,omitempty" path:"user_state"`
	UserTag              string                                    `url:"user_tag,omitempty" json:"user_tag,omitempty" path:"user_tag"`
}

type UserLifecycleRuleDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (u *UserLifecycleRule) UnmarshalJSON(data []byte) error {
	type userLifecycleRule UserLifecycleRule
	var v userLifecycleRule
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = UserLifecycleRule(v)
	return nil
}

func (u *UserLifecycleRuleCollection) UnmarshalJSON(data []byte) error {
	type userLifecycleRules UserLifecycleRuleCollection
	var v userLifecycleRules
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UserLifecycleRuleCollection(v)
	return nil
}

func (u *UserLifecycleRuleCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
