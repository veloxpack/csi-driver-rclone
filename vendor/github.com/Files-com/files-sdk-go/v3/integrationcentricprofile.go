package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type IntegrationCentricProfile struct {
	Id                    int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                  string                   `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	WorkspaceId           int64                    `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	UseForAllUsers        *bool                    `json:"use_for_all_users,omitempty" path:"use_for_all_users,omitempty" url:"use_for_all_users,omitempty"`
	ExpectedRemoteServers []map[string]interface{} `json:"expected_remote_servers,omitempty" path:"expected_remote_servers,omitempty" url:"expected_remote_servers,omitempty"`
}

func (i IntegrationCentricProfile) Identifier() interface{} {
	return i.Id
}

type IntegrationCentricProfileCollection []IntegrationCentricProfile

type IntegrationCentricProfileListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type IntegrationCentricProfileFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type IntegrationCentricProfileCreateParams struct {
	Name                  string                   `url:"name" json:"name" path:"name"`
	ExpectedRemoteServers []map[string]interface{} `url:"expected_remote_servers" json:"expected_remote_servers" path:"expected_remote_servers"`
	WorkspaceId           int64                    `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	UseForAllUsers        *bool                    `url:"use_for_all_users,omitempty" json:"use_for_all_users,omitempty" path:"use_for_all_users"`
}

type IntegrationCentricProfileUpdateParams struct {
	Id                    int64                    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name                  string                   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	WorkspaceId           int64                    `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	ExpectedRemoteServers []map[string]interface{} `url:"expected_remote_servers,omitempty" json:"expected_remote_servers,omitempty" path:"expected_remote_servers"`
	UseForAllUsers        *bool                    `url:"use_for_all_users,omitempty" json:"use_for_all_users,omitempty" path:"use_for_all_users"`
}

type IntegrationCentricProfileDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (i *IntegrationCentricProfile) UnmarshalJSON(data []byte) error {
	type integrationCentricProfile IntegrationCentricProfile
	var v integrationCentricProfile
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = IntegrationCentricProfile(v)
	return nil
}

func (i *IntegrationCentricProfileCollection) UnmarshalJSON(data []byte) error {
	type integrationCentricProfiles IntegrationCentricProfileCollection
	var v integrationCentricProfiles
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = IntegrationCentricProfileCollection(v)
	return nil
}

func (i *IntegrationCentricProfileCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
