package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PartnerChannelTemplate struct {
	Id                             int64    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId                    int64    `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                           string   `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Path                           string   `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	ToPartnerFolderName            string   `json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name,omitempty" url:"to_partner_folder_name,omitempty"`
	FromPartnerFolderName          string   `json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name,omitempty" url:"from_partner_folder_name,omitempty"`
	FromPartnerRoutePathPattern    string   `json:"from_partner_route_path_pattern,omitempty" path:"from_partner_route_path_pattern,omitempty" url:"from_partner_route_path_pattern,omitempty"`
	ToPartnerRoutePathPattern      string   `json:"to_partner_route_path_pattern,omitempty" path:"to_partner_route_path_pattern,omitempty" url:"to_partner_route_path_pattern,omitempty"`
	ToPartnerManagedFolderPaths    []string `json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths,omitempty" url:"to_partner_managed_folder_paths,omitempty"`
	FromPartnerManagedFolderPaths  []string `json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths,omitempty" url:"from_partner_managed_folder_paths,omitempty"`
	EffectiveToPartnerFolderName   string   `json:"effective_to_partner_folder_name,omitempty" path:"effective_to_partner_folder_name,omitempty" url:"effective_to_partner_folder_name,omitempty"`
	EffectiveFromPartnerFolderName string   `json:"effective_from_partner_folder_name,omitempty" path:"effective_from_partner_folder_name,omitempty" url:"effective_from_partner_folder_name,omitempty"`
}

func (p PartnerChannelTemplate) Identifier() interface{} {
	return p.Id
}

type PartnerChannelTemplateCollection []PartnerChannelTemplate

type PartnerChannelTemplateListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type PartnerChannelTemplateFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type PartnerChannelTemplateCreateParams struct {
	FromPartnerFolderName         string   `url:"from_partner_folder_name,omitempty" json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name"`
	FromPartnerManagedFolderPaths []string `url:"from_partner_managed_folder_paths,omitempty" json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths"`
	FromPartnerRoutePathPattern   string   `url:"from_partner_route_path_pattern,omitempty" json:"from_partner_route_path_pattern,omitempty" path:"from_partner_route_path_pattern"`
	ToPartnerFolderName           string   `url:"to_partner_folder_name,omitempty" json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name"`
	ToPartnerManagedFolderPaths   []string `url:"to_partner_managed_folder_paths,omitempty" json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths"`
	ToPartnerRoutePathPattern     string   `url:"to_partner_route_path_pattern,omitempty" json:"to_partner_route_path_pattern,omitempty" path:"to_partner_route_path_pattern"`
	Name                          string   `url:"name" json:"name" path:"name"`
	Path                          string   `url:"path" json:"path" path:"path"`
	WorkspaceId                   int64    `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type PartnerChannelTemplateUpdateParams struct {
	Id                            int64    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	FromPartnerFolderName         string   `url:"from_partner_folder_name,omitempty" json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name"`
	FromPartnerManagedFolderPaths []string `url:"from_partner_managed_folder_paths,omitempty" json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths"`
	FromPartnerRoutePathPattern   string   `url:"from_partner_route_path_pattern,omitempty" json:"from_partner_route_path_pattern,omitempty" path:"from_partner_route_path_pattern"`
	ToPartnerFolderName           string   `url:"to_partner_folder_name,omitempty" json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name"`
	ToPartnerManagedFolderPaths   []string `url:"to_partner_managed_folder_paths,omitempty" json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths"`
	ToPartnerRoutePathPattern     string   `url:"to_partner_route_path_pattern,omitempty" json:"to_partner_route_path_pattern,omitempty" path:"to_partner_route_path_pattern"`
	Name                          string   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Path                          string   `url:"path,omitempty" json:"path,omitempty" path:"path"`
}

type PartnerChannelTemplateDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *PartnerChannelTemplate) UnmarshalJSON(data []byte) error {
	type partnerChannelTemplate PartnerChannelTemplate
	var v partnerChannelTemplate
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PartnerChannelTemplate(v)
	return nil
}

func (p *PartnerChannelTemplateCollection) UnmarshalJSON(data []byte) error {
	type partnerChannelTemplates PartnerChannelTemplateCollection
	var v partnerChannelTemplates
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PartnerChannelTemplateCollection(v)
	return nil
}

func (p *PartnerChannelTemplateCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
