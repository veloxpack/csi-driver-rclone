package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Partner struct {
	AllowBypassing2faPolicies  *bool   `json:"allow_bypassing_2fa_policies,omitempty" path:"allow_bypassing_2fa_policies,omitempty" url:"allow_bypassing_2fa_policies,omitempty"`
	AllowedIps                 string  `json:"allowed_ips,omitempty" path:"allowed_ips,omitempty" url:"allowed_ips,omitempty"`
	AllowCredentialChanges     *bool   `json:"allow_credential_changes,omitempty" path:"allow_credential_changes,omitempty" url:"allow_credential_changes,omitempty"`
	AllowProvidingGpgKeys      *bool   `json:"allow_providing_gpg_keys,omitempty" path:"allow_providing_gpg_keys,omitempty" url:"allow_providing_gpg_keys,omitempty"`
	AllowUserCreation          *bool   `json:"allow_user_creation,omitempty" path:"allow_user_creation,omitempty" url:"allow_user_creation,omitempty"`
	CcEmailsToResponsibleParty *bool   `json:"cc_emails_to_responsible_party,omitempty" path:"cc_emails_to_responsible_party,omitempty" url:"cc_emails_to_responsible_party,omitempty"`
	Id                         int64   `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	AiAssistantPersonalityId   int64   `json:"ai_assistant_personality_id,omitempty" path:"ai_assistant_personality_id,omitempty" url:"ai_assistant_personality_id,omitempty"`
	WorkspaceId                int64   `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                       string  `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Notes                      string  `json:"notes,omitempty" path:"notes,omitempty" url:"notes,omitempty"`
	PartnerAdminIds            []int64 `json:"partner_admin_ids,omitempty" path:"partner_admin_ids,omitempty" url:"partner_admin_ids,omitempty"`
	PartnerChannelTemplateId   int64   `json:"partner_channel_template_id,omitempty" path:"partner_channel_template_id,omitempty" url:"partner_channel_template_id,omitempty"`
	PartnershipRole            string  `json:"partnership_role,omitempty" path:"partnership_role,omitempty" url:"partnership_role,omitempty"`
	ResponsibleGroupId         int64   `json:"responsible_group_id,omitempty" path:"responsible_group_id,omitempty" url:"responsible_group_id,omitempty"`
	ResponsibleUserId          int64   `json:"responsible_user_id,omitempty" path:"responsible_user_id,omitempty" url:"responsible_user_id,omitempty"`
	RootFolder                 string  `json:"root_folder,omitempty" path:"root_folder,omitempty" url:"root_folder,omitempty"`
	Tags                       string  `json:"tags,omitempty" path:"tags,omitempty" url:"tags,omitempty"`
	UserIds                    []int64 `json:"user_ids,omitempty" path:"user_ids,omitempty" url:"user_ids,omitempty"`
}

func (p Partner) Identifier() interface{} {
	return p.Id
}

type PartnerCollection []Partner

type PartnerListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type PartnerFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type PartnerCreateParams struct {
	AiAssistantPersonalityId   int64  `url:"ai_assistant_personality_id,omitempty" json:"ai_assistant_personality_id,omitempty" path:"ai_assistant_personality_id"`
	AllowedIps                 string `url:"allowed_ips,omitempty" json:"allowed_ips,omitempty" path:"allowed_ips"`
	AllowBypassing2faPolicies  *bool  `url:"allow_bypassing_2fa_policies,omitempty" json:"allow_bypassing_2fa_policies,omitempty" path:"allow_bypassing_2fa_policies"`
	AllowCredentialChanges     *bool  `url:"allow_credential_changes,omitempty" json:"allow_credential_changes,omitempty" path:"allow_credential_changes"`
	AllowProvidingGpgKeys      *bool  `url:"allow_providing_gpg_keys,omitempty" json:"allow_providing_gpg_keys,omitempty" path:"allow_providing_gpg_keys"`
	AllowUserCreation          *bool  `url:"allow_user_creation,omitempty" json:"allow_user_creation,omitempty" path:"allow_user_creation"`
	CcEmailsToResponsibleParty *bool  `url:"cc_emails_to_responsible_party,omitempty" json:"cc_emails_to_responsible_party,omitempty" path:"cc_emails_to_responsible_party"`
	Notes                      string `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	PartnerChannelTemplateId   int64  `url:"partner_channel_template_id,omitempty" json:"partner_channel_template_id,omitempty" path:"partner_channel_template_id"`
	ResponsibleGroupId         int64  `url:"responsible_group_id,omitempty" json:"responsible_group_id,omitempty" path:"responsible_group_id"`
	ResponsibleUserId          int64  `url:"responsible_user_id,omitempty" json:"responsible_user_id,omitempty" path:"responsible_user_id"`
	Tags                       string `url:"tags,omitempty" json:"tags,omitempty" path:"tags"`
	Name                       string `url:"name" json:"name" path:"name"`
	RootFolder                 string `url:"root_folder" json:"root_folder" path:"root_folder"`
	WorkspaceId                int64  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type PartnerUpdateParams struct {
	Id                         int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	AiAssistantPersonalityId   int64  `url:"ai_assistant_personality_id,omitempty" json:"ai_assistant_personality_id,omitempty" path:"ai_assistant_personality_id"`
	AllowedIps                 string `url:"allowed_ips,omitempty" json:"allowed_ips,omitempty" path:"allowed_ips"`
	AllowBypassing2faPolicies  *bool  `url:"allow_bypassing_2fa_policies,omitempty" json:"allow_bypassing_2fa_policies,omitempty" path:"allow_bypassing_2fa_policies"`
	AllowCredentialChanges     *bool  `url:"allow_credential_changes,omitempty" json:"allow_credential_changes,omitempty" path:"allow_credential_changes"`
	AllowProvidingGpgKeys      *bool  `url:"allow_providing_gpg_keys,omitempty" json:"allow_providing_gpg_keys,omitempty" path:"allow_providing_gpg_keys"`
	AllowUserCreation          *bool  `url:"allow_user_creation,omitempty" json:"allow_user_creation,omitempty" path:"allow_user_creation"`
	CcEmailsToResponsibleParty *bool  `url:"cc_emails_to_responsible_party,omitempty" json:"cc_emails_to_responsible_party,omitempty" path:"cc_emails_to_responsible_party"`
	Notes                      string `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	PartnerChannelTemplateId   int64  `url:"partner_channel_template_id,omitempty" json:"partner_channel_template_id,omitempty" path:"partner_channel_template_id"`
	ResponsibleGroupId         int64  `url:"responsible_group_id,omitempty" json:"responsible_group_id,omitempty" path:"responsible_group_id"`
	ResponsibleUserId          int64  `url:"responsible_user_id,omitempty" json:"responsible_user_id,omitempty" path:"responsible_user_id"`
	Tags                       string `url:"tags,omitempty" json:"tags,omitempty" path:"tags"`
	Name                       string `url:"name,omitempty" json:"name,omitempty" path:"name"`
	RootFolder                 string `url:"root_folder,omitempty" json:"root_folder,omitempty" path:"root_folder"`
}

type PartnerDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *Partner) UnmarshalJSON(data []byte) error {
	type partner Partner
	var v partner
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = Partner(v)
	return nil
}

func (p *PartnerCollection) UnmarshalJSON(data []byte) error {
	type partners PartnerCollection
	var v partners
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PartnerCollection(v)
	return nil
}

func (p *PartnerCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
