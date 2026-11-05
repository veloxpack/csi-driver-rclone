package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PartnerChannel struct {
	Id                             int64    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId                    int64    `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	PartnerId                      int64    `json:"partner_id,omitempty" path:"partner_id,omitempty" url:"partner_id,omitempty"`
	PartnerChannelTemplateId       int64    `json:"partner_channel_template_id,omitempty" path:"partner_channel_template_id,omitempty" url:"partner_channel_template_id,omitempty"`
	Path                           string   `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	ToPartnerFolderName            string   `json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name,omitempty" url:"to_partner_folder_name,omitempty"`
	FromPartnerFolderName          string   `json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name,omitempty" url:"from_partner_folder_name,omitempty"`
	FromPartnerRoutePath           string   `json:"from_partner_route_path,omitempty" path:"from_partner_route_path,omitempty" url:"from_partner_route_path,omitempty"`
	ToPartnerRoutePath             string   `json:"to_partner_route_path,omitempty" path:"to_partner_route_path,omitempty" url:"to_partner_route_path,omitempty"`
	ToPartnerManagedFolderPaths    []string `json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths,omitempty" url:"to_partner_managed_folder_paths,omitempty"`
	FromPartnerManagedFolderPaths  []string `json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths,omitempty" url:"from_partner_managed_folder_paths,omitempty"`
	EffectiveToPartnerFolderName   string   `json:"effective_to_partner_folder_name,omitempty" path:"effective_to_partner_folder_name,omitempty" url:"effective_to_partner_folder_name,omitempty"`
	EffectiveFromPartnerFolderName string   `json:"effective_from_partner_folder_name,omitempty" path:"effective_from_partner_folder_name,omitempty" url:"effective_from_partner_folder_name,omitempty"`
	ChannelPath                    string   `json:"channel_path,omitempty" path:"channel_path,omitempty" url:"channel_path,omitempty"`
	ToPartnerFolderPath            string   `json:"to_partner_folder_path,omitempty" path:"to_partner_folder_path,omitempty" url:"to_partner_folder_path,omitempty"`
	FromPartnerFolderPath          string   `json:"from_partner_folder_path,omitempty" path:"from_partner_folder_path,omitempty" url:"from_partner_folder_path,omitempty"`
}

func (p PartnerChannel) Identifier() interface{} {
	return p.Id
}

type PartnerChannelCollection []PartnerChannel

type PartnerChannelListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type PartnerChannelFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type PartnerChannelCreateParams struct {
	FromPartnerFolderName         string   `url:"from_partner_folder_name,omitempty" json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name"`
	FromPartnerManagedFolderPaths []string `url:"from_partner_managed_folder_paths,omitempty" json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths"`
	FromPartnerRoutePath          string   `url:"from_partner_route_path,omitempty" json:"from_partner_route_path,omitempty" path:"from_partner_route_path"`
	ToPartnerFolderName           string   `url:"to_partner_folder_name,omitempty" json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name"`
	ToPartnerManagedFolderPaths   []string `url:"to_partner_managed_folder_paths,omitempty" json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths"`
	ToPartnerRoutePath            string   `url:"to_partner_route_path,omitempty" json:"to_partner_route_path,omitempty" path:"to_partner_route_path"`
	PartnerId                     int64    `url:"partner_id" json:"partner_id" path:"partner_id"`
	Path                          string   `url:"path" json:"path" path:"path"`
	WorkspaceId                   int64    `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type PartnerChannelUpdateParams struct {
	Id                            int64    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	FromPartnerFolderName         string   `url:"from_partner_folder_name,omitempty" json:"from_partner_folder_name,omitempty" path:"from_partner_folder_name"`
	FromPartnerManagedFolderPaths []string `url:"from_partner_managed_folder_paths,omitempty" json:"from_partner_managed_folder_paths,omitempty" path:"from_partner_managed_folder_paths"`
	FromPartnerRoutePath          string   `url:"from_partner_route_path,omitempty" json:"from_partner_route_path,omitempty" path:"from_partner_route_path"`
	ToPartnerFolderName           string   `url:"to_partner_folder_name,omitempty" json:"to_partner_folder_name,omitempty" path:"to_partner_folder_name"`
	ToPartnerManagedFolderPaths   []string `url:"to_partner_managed_folder_paths,omitempty" json:"to_partner_managed_folder_paths,omitempty" path:"to_partner_managed_folder_paths"`
	ToPartnerRoutePath            string   `url:"to_partner_route_path,omitempty" json:"to_partner_route_path,omitempty" path:"to_partner_route_path"`
	Path                          string   `url:"path,omitempty" json:"path,omitempty" path:"path"`
}

type PartnerChannelDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *PartnerChannel) UnmarshalJSON(data []byte) error {
	type partnerChannel PartnerChannel
	var v partnerChannel
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PartnerChannel(v)
	return nil
}

func (p *PartnerChannelCollection) UnmarshalJSON(data []byte) error {
	type partnerChannels PartnerChannelCollection
	var v partnerChannels
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PartnerChannelCollection(v)
	return nil
}

func (p *PartnerChannelCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
