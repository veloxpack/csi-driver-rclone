package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SsoStrategy struct {
	Protocol                         string `json:"protocol,omitempty" path:"protocol,omitempty" url:"protocol,omitempty"`
	Provider                         string `json:"provider,omitempty" path:"provider,omitempty" url:"provider,omitempty"`
	Label                            string `json:"label,omitempty" path:"label,omitempty" url:"label,omitempty"`
	LogoUrl                          string `json:"logo_url,omitempty" path:"logo_url,omitempty" url:"logo_url,omitempty"`
	Id                               int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	UserCount                        int64  `json:"user_count,omitempty" path:"user_count,omitempty" url:"user_count,omitempty"`
	SamlProviderCertFingerprint      string `json:"saml_provider_cert_fingerprint,omitempty" path:"saml_provider_cert_fingerprint,omitempty" url:"saml_provider_cert_fingerprint,omitempty"`
	SamlProviderIssuerUrl            string `json:"saml_provider_issuer_url,omitempty" path:"saml_provider_issuer_url,omitempty" url:"saml_provider_issuer_url,omitempty"`
	SamlProviderMetadataContent      string `json:"saml_provider_metadata_content,omitempty" path:"saml_provider_metadata_content,omitempty" url:"saml_provider_metadata_content,omitempty"`
	SamlProviderMetadataUrl          string `json:"saml_provider_metadata_url,omitempty" path:"saml_provider_metadata_url,omitempty" url:"saml_provider_metadata_url,omitempty"`
	SamlProviderSloTargetUrl         string `json:"saml_provider_slo_target_url,omitempty" path:"saml_provider_slo_target_url,omitempty" url:"saml_provider_slo_target_url,omitempty"`
	SamlProviderSsoTargetUrl         string `json:"saml_provider_sso_target_url,omitempty" path:"saml_provider_sso_target_url,omitempty" url:"saml_provider_sso_target_url,omitempty"`
	ScimAuthenticationMethod         string `json:"scim_authentication_method,omitempty" path:"scim_authentication_method,omitempty" url:"scim_authentication_method,omitempty"`
	ScimUsername                     string `json:"scim_username,omitempty" path:"scim_username,omitempty" url:"scim_username,omitempty"`
	ScimOauthAccessToken             string `json:"scim_oauth_access_token,omitempty" path:"scim_oauth_access_token,omitempty" url:"scim_oauth_access_token,omitempty"`
	ScimOauthAccessTokenExpiresAt    string `json:"scim_oauth_access_token_expires_at,omitempty" path:"scim_oauth_access_token_expires_at,omitempty" url:"scim_oauth_access_token_expires_at,omitempty"`
	Subdomain                        string `json:"subdomain,omitempty" path:"subdomain,omitempty" url:"subdomain,omitempty"`
	ProvisionUsers                   *bool  `json:"provision_users,omitempty" path:"provision_users,omitempty" url:"provision_users,omitempty"`
	ProvisionGroups                  *bool  `json:"provision_groups,omitempty" path:"provision_groups,omitempty" url:"provision_groups,omitempty"`
	DeprovisionUsers                 *bool  `json:"deprovision_users,omitempty" path:"deprovision_users,omitempty" url:"deprovision_users,omitempty"`
	DeprovisionGroups                *bool  `json:"deprovision_groups,omitempty" path:"deprovision_groups,omitempty" url:"deprovision_groups,omitempty"`
	DeprovisionBehavior              string `json:"deprovision_behavior,omitempty" path:"deprovision_behavior,omitempty" url:"deprovision_behavior,omitempty"`
	ProvisionGroupDefault            string `json:"provision_group_default,omitempty" path:"provision_group_default,omitempty" url:"provision_group_default,omitempty"`
	ProvisionGroupExclusion          string `json:"provision_group_exclusion,omitempty" path:"provision_group_exclusion,omitempty" url:"provision_group_exclusion,omitempty"`
	ProvisionGroupInclusion          string `json:"provision_group_inclusion,omitempty" path:"provision_group_inclusion,omitempty" url:"provision_group_inclusion,omitempty"`
	ProvisionGroupRequired           string `json:"provision_group_required,omitempty" path:"provision_group_required,omitempty" url:"provision_group_required,omitempty"`
	ProvisionEmailSignupGroups       string `json:"provision_email_signup_groups,omitempty" path:"provision_email_signup_groups,omitempty" url:"provision_email_signup_groups,omitempty"`
	ProvisionReadonlySiteAdminGroups string `json:"provision_readonly_site_admin_groups,omitempty" path:"provision_readonly_site_admin_groups,omitempty" url:"provision_readonly_site_admin_groups,omitempty"`
	ProvisionSiteAdminGroups         string `json:"provision_site_admin_groups,omitempty" path:"provision_site_admin_groups,omitempty" url:"provision_site_admin_groups,omitempty"`
	ProvisionGroupAdminGroups        string `json:"provision_group_admin_groups,omitempty" path:"provision_group_admin_groups,omitempty" url:"provision_group_admin_groups,omitempty"`
	ProvisionAttachmentsPermission   *bool  `json:"provision_attachments_permission,omitempty" path:"provision_attachments_permission,omitempty" url:"provision_attachments_permission,omitempty"`
	ProvisionDavPermission           *bool  `json:"provision_dav_permission,omitempty" path:"provision_dav_permission,omitempty" url:"provision_dav_permission,omitempty"`
	ProvisionFtpPermission           *bool  `json:"provision_ftp_permission,omitempty" path:"provision_ftp_permission,omitempty" url:"provision_ftp_permission,omitempty"`
	ProvisionSftpPermission          *bool  `json:"provision_sftp_permission,omitempty" path:"provision_sftp_permission,omitempty" url:"provision_sftp_permission,omitempty"`
	ProvisionTimeZone                string `json:"provision_time_zone,omitempty" path:"provision_time_zone,omitempty" url:"provision_time_zone,omitempty"`
	ProvisionCompany                 string `json:"provision_company,omitempty" path:"provision_company,omitempty" url:"provision_company,omitempty"`
	ProvisionRequire2fa              string `json:"provision_require_2fa,omitempty" path:"provision_require_2fa,omitempty" url:"provision_require_2fa,omitempty"`
	ProvisionFilesystemLayout        string `json:"provision_filesystem_layout,omitempty" path:"provision_filesystem_layout,omitempty" url:"provision_filesystem_layout,omitempty"`
	ProviderIdentifier               string `json:"provider_identifier,omitempty" path:"provider_identifier,omitempty" url:"provider_identifier,omitempty"`
	LdapBaseDn                       string `json:"ldap_base_dn,omitempty" path:"ldap_base_dn,omitempty" url:"ldap_base_dn,omitempty"`
	LdapDomain                       string `json:"ldap_domain,omitempty" path:"ldap_domain,omitempty" url:"ldap_domain,omitempty"`
	Enabled                          *bool  `json:"enabled,omitempty" path:"enabled,omitempty" url:"enabled,omitempty"`
	LdapHost                         string `json:"ldap_host,omitempty" path:"ldap_host,omitempty" url:"ldap_host,omitempty"`
	LdapHost2                        string `json:"ldap_host_2,omitempty" path:"ldap_host_2,omitempty" url:"ldap_host_2,omitempty"`
	LdapHost3                        string `json:"ldap_host_3,omitempty" path:"ldap_host_3,omitempty" url:"ldap_host_3,omitempty"`
	LdapPort                         int64  `json:"ldap_port,omitempty" path:"ldap_port,omitempty" url:"ldap_port,omitempty"`
	LdapSecure                       *bool  `json:"ldap_secure,omitempty" path:"ldap_secure,omitempty" url:"ldap_secure,omitempty"`
	LdapUsername                     string `json:"ldap_username,omitempty" path:"ldap_username,omitempty" url:"ldap_username,omitempty"`
	LdapUsernameField                string `json:"ldap_username_field,omitempty" path:"ldap_username_field,omitempty" url:"ldap_username_field,omitempty"`
}

func (s SsoStrategy) Identifier() interface{} {
	return s.Id
}

type SsoStrategyCollection []SsoStrategy

type SsoStrategyListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type SsoStrategyFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Synchronize provisioning data with the SSO remote server
type SsoStrategySyncParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *SsoStrategy) UnmarshalJSON(data []byte) error {
	type ssoStrategy SsoStrategy
	var v ssoStrategy
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SsoStrategy(v)
	return nil
}

func (s *SsoStrategyCollection) UnmarshalJSON(data []byte) error {
	type ssoStrategys SsoStrategyCollection
	var v ssoStrategys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SsoStrategyCollection(v)
	return nil
}

func (s *SsoStrategyCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
