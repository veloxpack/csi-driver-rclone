package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AiAssistantPersonality struct {
	Id                   int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId          int64      `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                 string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	SystemPrompt         string     `json:"system_prompt,omitempty" path:"system_prompt,omitempty" url:"system_prompt,omitempty"`
	UseByDefault         *bool      `json:"use_by_default,omitempty" path:"use_by_default,omitempty" url:"use_by_default,omitempty"`
	ApplyToAllWorkspaces *bool      `json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces,omitempty" url:"apply_to_all_workspaces,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (a AiAssistantPersonality) Identifier() interface{} {
	return a.Id
}

type AiAssistantPersonalityCollection []AiAssistantPersonality

type AiAssistantPersonalityListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type AiAssistantPersonalityFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type AiAssistantPersonalityCreateParams struct {
	ApplyToAllWorkspaces *bool  `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	Name                 string `url:"name" json:"name" path:"name"`
	SystemPrompt         string `url:"system_prompt" json:"system_prompt" path:"system_prompt"`
	UseByDefault         *bool  `url:"use_by_default,omitempty" json:"use_by_default,omitempty" path:"use_by_default"`
	WorkspaceId          int64  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type AiAssistantPersonalityUpdateParams struct {
	Id                   int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	ApplyToAllWorkspaces *bool  `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	Name                 string `url:"name,omitempty" json:"name,omitempty" path:"name"`
	SystemPrompt         string `url:"system_prompt,omitempty" json:"system_prompt,omitempty" path:"system_prompt"`
	UseByDefault         *bool  `url:"use_by_default,omitempty" json:"use_by_default,omitempty" path:"use_by_default"`
	WorkspaceId          int64  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type AiAssistantPersonalityDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *AiAssistantPersonality) UnmarshalJSON(data []byte) error {
	type aiAssistantPersonality AiAssistantPersonality
	var v aiAssistantPersonality
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AiAssistantPersonality(v)
	return nil
}

func (a *AiAssistantPersonalityCollection) UnmarshalJSON(data []byte) error {
	type aiAssistantPersonalitys AiAssistantPersonalityCollection
	var v aiAssistantPersonalitys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AiAssistantPersonalityCollection(v)
	return nil
}

func (a *AiAssistantPersonalityCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
