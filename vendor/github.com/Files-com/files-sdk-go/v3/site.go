package files_sdk

import (
	"encoding/json"
	"io"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Site struct {
	Id                                       int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                                     string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	AdditionalTextFileTypes                  []string               `json:"additional_text_file_types,omitempty" path:"additional_text_file_types,omitempty" url:"additional_text_file_types,omitempty"`
	Allowed2faMethodSms                      *bool                  `json:"allowed_2fa_method_sms,omitempty" path:"allowed_2fa_method_sms,omitempty" url:"allowed_2fa_method_sms,omitempty"`
	Allowed2faMethodTotp                     *bool                  `json:"allowed_2fa_method_totp,omitempty" path:"allowed_2fa_method_totp,omitempty" url:"allowed_2fa_method_totp,omitempty"`
	Allowed2faMethodWebauthn                 *bool                  `json:"allowed_2fa_method_webauthn,omitempty" path:"allowed_2fa_method_webauthn,omitempty" url:"allowed_2fa_method_webauthn,omitempty"`
	Allowed2faMethodYubi                     *bool                  `json:"allowed_2fa_method_yubi,omitempty" path:"allowed_2fa_method_yubi,omitempty" url:"allowed_2fa_method_yubi,omitempty"`
	Allowed2faMethodEmail                    *bool                  `json:"allowed_2fa_method_email,omitempty" path:"allowed_2fa_method_email,omitempty" url:"allowed_2fa_method_email,omitempty"`
	Allowed2faMethodStatic                   *bool                  `json:"allowed_2fa_method_static,omitempty" path:"allowed_2fa_method_static,omitempty" url:"allowed_2fa_method_static,omitempty"`
	Allowed2faMethodBypassForFtpSftpDav      *bool                  `json:"allowed_2fa_method_bypass_for_ftp_sftp_dav,omitempty" path:"allowed_2fa_method_bypass_for_ftp_sftp_dav,omitempty" url:"allowed_2fa_method_bypass_for_ftp_sftp_dav,omitempty"`
	AdminUserId                              int64                  `json:"admin_user_id,omitempty" path:"admin_user_id,omitempty" url:"admin_user_id,omitempty"`
	AdminsBypassLockedSubfolders             *bool                  `json:"admins_bypass_locked_subfolders,omitempty" path:"admins_bypass_locked_subfolders,omitempty" url:"admins_bypass_locked_subfolders,omitempty"`
	AllowBundleNames                         *bool                  `json:"allow_bundle_names,omitempty" path:"allow_bundle_names,omitempty" url:"allow_bundle_names,omitempty"`
	AllowedCountries                         string                 `json:"allowed_countries,omitempty" path:"allowed_countries,omitempty" url:"allowed_countries,omitempty"`
	AllowedIps                               string                 `json:"allowed_ips,omitempty" path:"allowed_ips,omitempty" url:"allowed_ips,omitempty"`
	AlwaysMkdirParents                       *bool                  `json:"always_mkdir_parents,omitempty" path:"always_mkdir_parents,omitempty" url:"always_mkdir_parents,omitempty"`
	As2MessageRetentionDays                  int64                  `json:"as2_message_retention_days,omitempty" path:"as2_message_retention_days,omitempty" url:"as2_message_retention_days,omitempty"`
	AskAboutOverwrites                       *bool                  `json:"ask_about_overwrites,omitempty" path:"ask_about_overwrites,omitempty" url:"ask_about_overwrites,omitempty"`
	BundleActivityNotifications              string                 `json:"bundle_activity_notifications,omitempty" path:"bundle_activity_notifications,omitempty" url:"bundle_activity_notifications,omitempty"`
	BundleExpiration                         int64                  `json:"bundle_expiration,omitempty" path:"bundle_expiration,omitempty" url:"bundle_expiration,omitempty"`
	BundleNotFoundMessage                    string                 `json:"bundle_not_found_message,omitempty" path:"bundle_not_found_message,omitempty" url:"bundle_not_found_message,omitempty"`
	BundlePasswordRequired                   *bool                  `json:"bundle_password_required,omitempty" path:"bundle_password_required,omitempty" url:"bundle_password_required,omitempty"`
	BundleRecipientBlacklistDomains          []string               `json:"bundle_recipient_blacklist_domains,omitempty" path:"bundle_recipient_blacklist_domains,omitempty" url:"bundle_recipient_blacklist_domains,omitempty"`
	BundleRecipientBlacklistFreeEmailDomains *bool                  `json:"bundle_recipient_blacklist_free_email_domains,omitempty" path:"bundle_recipient_blacklist_free_email_domains,omitempty" url:"bundle_recipient_blacklist_free_email_domains,omitempty"`
	BundleRegistrationNotifications          string                 `json:"bundle_registration_notifications,omitempty" path:"bundle_registration_notifications,omitempty" url:"bundle_registration_notifications,omitempty"`
	BundleRequireRegistration                *bool                  `json:"bundle_require_registration,omitempty" path:"bundle_require_registration,omitempty" url:"bundle_require_registration,omitempty"`
	BundleRequireShareRecipient              *bool                  `json:"bundle_require_share_recipient,omitempty" path:"bundle_require_share_recipient,omitempty" url:"bundle_require_share_recipient,omitempty"`
	BundleRequireNote                        *bool                  `json:"bundle_require_note,omitempty" path:"bundle_require_note,omitempty" url:"bundle_require_note,omitempty"`
	BundleSendSharedReceipts                 *bool                  `json:"bundle_send_shared_receipts,omitempty" path:"bundle_send_shared_receipts,omitempty" url:"bundle_send_shared_receipts,omitempty"`
	BundleUploadReceiptNotifications         string                 `json:"bundle_upload_receipt_notifications,omitempty" path:"bundle_upload_receipt_notifications,omitempty" url:"bundle_upload_receipt_notifications,omitempty"`
	BundleWatermarkAttachment                Image                  `json:"bundle_watermark_attachment,omitempty" path:"bundle_watermark_attachment,omitempty" url:"bundle_watermark_attachment,omitempty"`
	BundleWatermarkValue                     map[string]interface{} `json:"bundle_watermark_value,omitempty" path:"bundle_watermark_value,omitempty" url:"bundle_watermark_value,omitempty"`
	CalculateFileChecksumsCrc32              *bool                  `json:"calculate_file_checksums_crc32,omitempty" path:"calculate_file_checksums_crc32,omitempty" url:"calculate_file_checksums_crc32,omitempty"`
	CalculateFileChecksumsMd5                *bool                  `json:"calculate_file_checksums_md5,omitempty" path:"calculate_file_checksums_md5,omitempty" url:"calculate_file_checksums_md5,omitempty"`
	CalculateFileChecksumsSha1               *bool                  `json:"calculate_file_checksums_sha1,omitempty" path:"calculate_file_checksums_sha1,omitempty" url:"calculate_file_checksums_sha1,omitempty"`
	CalculateFileChecksumsSha256             *bool                  `json:"calculate_file_checksums_sha256,omitempty" path:"calculate_file_checksums_sha256,omitempty" url:"calculate_file_checksums_sha256,omitempty"`
	UploadsViaEmailAuthentication            *bool                  `json:"uploads_via_email_authentication,omitempty" path:"uploads_via_email_authentication,omitempty" url:"uploads_via_email_authentication,omitempty"`
	Color2Left                               string                 `json:"color2_left,omitempty" path:"color2_left,omitempty" url:"color2_left,omitempty"`
	Color2Link                               string                 `json:"color2_link,omitempty" path:"color2_link,omitempty" url:"color2_link,omitempty"`
	Color2Text                               string                 `json:"color2_text,omitempty" path:"color2_text,omitempty" url:"color2_text,omitempty"`
	Color2Top                                string                 `json:"color2_top,omitempty" path:"color2_top,omitempty" url:"color2_top,omitempty"`
	Color2TopText                            string                 `json:"color2_top_text,omitempty" path:"color2_top_text,omitempty" url:"color2_top_text,omitempty"`
	ContactName                              string                 `json:"contact_name,omitempty" path:"contact_name,omitempty" url:"contact_name,omitempty"`
	CreatedAt                                *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Currency                                 string                 `json:"currency,omitempty" path:"currency,omitempty" url:"currency,omitempty"`
	CustomNamespace                          *bool                  `json:"custom_namespace,omitempty" path:"custom_namespace,omitempty" url:"custom_namespace,omitempty"`
	DavEnabled                               *bool                  `json:"dav_enabled,omitempty" path:"dav_enabled,omitempty" url:"dav_enabled,omitempty"`
	DavUserRootEnabled                       *bool                  `json:"dav_user_root_enabled,omitempty" path:"dav_user_root_enabled,omitempty" url:"dav_user_root_enabled,omitempty"`
	DaysToRetainBackups                      int64                  `json:"days_to_retain_backups,omitempty" path:"days_to_retain_backups,omitempty" url:"days_to_retain_backups,omitempty"`
	DocumentEditsInBundleAllowed             *bool                  `json:"document_edits_in_bundle_allowed,omitempty" path:"document_edits_in_bundle_allowed,omitempty" url:"document_edits_in_bundle_allowed,omitempty"`
	DefaultTimeZone                          string                 `json:"default_time_zone,omitempty" path:"default_time_zone,omitempty" url:"default_time_zone,omitempty"`
	DesktopApp                               *bool                  `json:"desktop_app,omitempty" path:"desktop_app,omitempty" url:"desktop_app,omitempty"`
	DesktopAppSessionIpPinning               *bool                  `json:"desktop_app_session_ip_pinning,omitempty" path:"desktop_app_session_ip_pinning,omitempty" url:"desktop_app_session_ip_pinning,omitempty"`
	DesktopAppSessionLifetime                int64                  `json:"desktop_app_session_lifetime,omitempty" path:"desktop_app_session_lifetime,omitempty" url:"desktop_app_session_lifetime,omitempty"`
	LegacyChecksumsMode                      *bool                  `json:"legacy_checksums_mode,omitempty" path:"legacy_checksums_mode,omitempty" url:"legacy_checksums_mode,omitempty"`
	MigrateRemoteServerSyncToSync            *bool                  `json:"migrate_remote_server_sync_to_sync,omitempty" path:"migrate_remote_server_sync_to_sync,omitempty" url:"migrate_remote_server_sync_to_sync,omitempty"`
	MobileApp                                *bool                  `json:"mobile_app,omitempty" path:"mobile_app,omitempty" url:"mobile_app,omitempty"`
	MobileAppSessionIpPinning                *bool                  `json:"mobile_app_session_ip_pinning,omitempty" path:"mobile_app_session_ip_pinning,omitempty" url:"mobile_app_session_ip_pinning,omitempty"`
	MobileAppSessionLifetime                 int64                  `json:"mobile_app_session_lifetime,omitempty" path:"mobile_app_session_lifetime,omitempty" url:"mobile_app_session_lifetime,omitempty"`
	DisallowedCountries                      string                 `json:"disallowed_countries,omitempty" path:"disallowed_countries,omitempty" url:"disallowed_countries,omitempty"`
	DisableFilesCertificateGeneration        *bool                  `json:"disable_files_certificate_generation,omitempty" path:"disable_files_certificate_generation,omitempty" url:"disable_files_certificate_generation,omitempty"`
	DisableNotifications                     *bool                  `json:"disable_notifications,omitempty" path:"disable_notifications,omitempty" url:"disable_notifications,omitempty"`
	DisablePasswordReset                     *bool                  `json:"disable_password_reset,omitempty" path:"disable_password_reset,omitempty" url:"disable_password_reset,omitempty"`
	Domain                                   string                 `json:"domain,omitempty" path:"domain,omitempty" url:"domain,omitempty"`
	DomainHstsHeader                         *bool                  `json:"domain_hsts_header,omitempty" path:"domain_hsts_header,omitempty" url:"domain_hsts_header,omitempty"`
	DomainLetsencryptChain                   string                 `json:"domain_letsencrypt_chain,omitempty" path:"domain_letsencrypt_chain,omitempty" url:"domain_letsencrypt_chain,omitempty"`
	Email                                    string                 `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
	FtpEnabled                               *bool                  `json:"ftp_enabled,omitempty" path:"ftp_enabled,omitempty" url:"ftp_enabled,omitempty"`
	ReplyToEmail                             string                 `json:"reply_to_email,omitempty" path:"reply_to_email,omitempty" url:"reply_to_email,omitempty"`
	NonSsoGroupsAllowed                      *bool                  `json:"non_sso_groups_allowed,omitempty" path:"non_sso_groups_allowed,omitempty" url:"non_sso_groups_allowed,omitempty"`
	NonSsoUsersAllowed                       *bool                  `json:"non_sso_users_allowed,omitempty" path:"non_sso_users_allowed,omitempty" url:"non_sso_users_allowed,omitempty"`
	FolderPermissionsGroupsOnly              *bool                  `json:"folder_permissions_groups_only,omitempty" path:"folder_permissions_groups_only,omitempty" url:"folder_permissions_groups_only,omitempty"`
	Hipaa                                    *bool                  `json:"hipaa,omitempty" path:"hipaa,omitempty" url:"hipaa,omitempty"`
	Icon128                                  Image                  `json:"icon128,omitempty" path:"icon128,omitempty" url:"icon128,omitempty"`
	Icon16                                   Image                  `json:"icon16,omitempty" path:"icon16,omitempty" url:"icon16,omitempty"`
	Icon32                                   Image                  `json:"icon32,omitempty" path:"icon32,omitempty" url:"icon32,omitempty"`
	Icon48                                   Image                  `json:"icon48,omitempty" path:"icon48,omitempty" url:"icon48,omitempty"`
	ImmutableFilesSetAt                      *time.Time             `json:"immutable_files_set_at,omitempty" path:"immutable_files_set_at,omitempty" url:"immutable_files_set_at,omitempty"`
	IncludePasswordInWelcomeEmail            *bool                  `json:"include_password_in_welcome_email,omitempty" path:"include_password_in_welcome_email,omitempty" url:"include_password_in_welcome_email,omitempty"`
	Language                                 string                 `json:"language,omitempty" path:"language,omitempty" url:"language,omitempty"`
	LdapBaseDn                               string                 `json:"ldap_base_dn,omitempty" path:"ldap_base_dn,omitempty" url:"ldap_base_dn,omitempty"`
	LdapDomain                               string                 `json:"ldap_domain,omitempty" path:"ldap_domain,omitempty" url:"ldap_domain,omitempty"`
	LdapEnabled                              *bool                  `json:"ldap_enabled,omitempty" path:"ldap_enabled,omitempty" url:"ldap_enabled,omitempty"`
	LdapGroupAction                          string                 `json:"ldap_group_action,omitempty" path:"ldap_group_action,omitempty" url:"ldap_group_action,omitempty"`
	LdapGroupExclusion                       string                 `json:"ldap_group_exclusion,omitempty" path:"ldap_group_exclusion,omitempty" url:"ldap_group_exclusion,omitempty"`
	LdapGroupInclusion                       string                 `json:"ldap_group_inclusion,omitempty" path:"ldap_group_inclusion,omitempty" url:"ldap_group_inclusion,omitempty"`
	LdapHost                                 string                 `json:"ldap_host,omitempty" path:"ldap_host,omitempty" url:"ldap_host,omitempty"`
	LdapHost2                                string                 `json:"ldap_host_2,omitempty" path:"ldap_host_2,omitempty" url:"ldap_host_2,omitempty"`
	LdapHost3                                string                 `json:"ldap_host_3,omitempty" path:"ldap_host_3,omitempty" url:"ldap_host_3,omitempty"`
	LdapPort                                 int64                  `json:"ldap_port,omitempty" path:"ldap_port,omitempty" url:"ldap_port,omitempty"`
	LdapSecure                               *bool                  `json:"ldap_secure,omitempty" path:"ldap_secure,omitempty" url:"ldap_secure,omitempty"`
	LdapType                                 string                 `json:"ldap_type,omitempty" path:"ldap_type,omitempty" url:"ldap_type,omitempty"`
	LdapUserAction                           string                 `json:"ldap_user_action,omitempty" path:"ldap_user_action,omitempty" url:"ldap_user_action,omitempty"`
	LdapUserIncludeGroups                    string                 `json:"ldap_user_include_groups,omitempty" path:"ldap_user_include_groups,omitempty" url:"ldap_user_include_groups,omitempty"`
	LdapUsername                             string                 `json:"ldap_username,omitempty" path:"ldap_username,omitempty" url:"ldap_username,omitempty"`
	LdapUsernameField                        string                 `json:"ldap_username_field,omitempty" path:"ldap_username_field,omitempty" url:"ldap_username_field,omitempty"`
	LoginHelpText                            string                 `json:"login_help_text,omitempty" path:"login_help_text,omitempty" url:"login_help_text,omitempty"`
	Logo                                     Image                  `json:"logo,omitempty" path:"logo,omitempty" url:"logo,omitempty"`
	LoginPageBackgroundImage                 Image                  `json:"login_page_background_image,omitempty" path:"login_page_background_image,omitempty" url:"login_page_background_image,omitempty"`
	MaxPriorPasswords                        int64                  `json:"max_prior_passwords,omitempty" path:"max_prior_passwords,omitempty" url:"max_prior_passwords,omitempty"`
	ManagedSiteSettings                      map[string]interface{} `json:"managed_site_settings,omitempty" path:"managed_site_settings,omitempty" url:"managed_site_settings,omitempty"`
	MotdText                                 string                 `json:"motd_text,omitempty" path:"motd_text,omitempty" url:"motd_text,omitempty"`
	MotdUseForFtp                            *bool                  `json:"motd_use_for_ftp,omitempty" path:"motd_use_for_ftp,omitempty" url:"motd_use_for_ftp,omitempty"`
	MotdUseForSftp                           *bool                  `json:"motd_use_for_sftp,omitempty" path:"motd_use_for_sftp,omitempty" url:"motd_use_for_sftp,omitempty"`
	NextBillingAmount                        string                 `json:"next_billing_amount,omitempty" path:"next_billing_amount,omitempty" url:"next_billing_amount,omitempty"`
	NextBillingDate                          string                 `json:"next_billing_date,omitempty" path:"next_billing_date,omitempty" url:"next_billing_date,omitempty"`
	OfficeIntegrationAvailable               *bool                  `json:"office_integration_available,omitempty" path:"office_integration_available,omitempty" url:"office_integration_available,omitempty"`
	OfficeIntegrationType                    string                 `json:"office_integration_type,omitempty" path:"office_integration_type,omitempty" url:"office_integration_type,omitempty"`
	OncehubLink                              string                 `json:"oncehub_link,omitempty" path:"oncehub_link,omitempty" url:"oncehub_link,omitempty"`
	OptOutGlobal                             *bool                  `json:"opt_out_global,omitempty" path:"opt_out_global,omitempty" url:"opt_out_global,omitempty"`
	Overdue                                  *bool                  `json:"overdue,omitempty" path:"overdue,omitempty" url:"overdue,omitempty"`
	PasswordMinLength                        int64                  `json:"password_min_length,omitempty" path:"password_min_length,omitempty" url:"password_min_length,omitempty"`
	PasswordRequireLetter                    *bool                  `json:"password_require_letter,omitempty" path:"password_require_letter,omitempty" url:"password_require_letter,omitempty"`
	PasswordRequireMixed                     *bool                  `json:"password_require_mixed,omitempty" path:"password_require_mixed,omitempty" url:"password_require_mixed,omitempty"`
	PasswordRequireNumber                    *bool                  `json:"password_require_number,omitempty" path:"password_require_number,omitempty" url:"password_require_number,omitempty"`
	PasswordRequireSpecial                   *bool                  `json:"password_require_special,omitempty" path:"password_require_special,omitempty" url:"password_require_special,omitempty"`
	PasswordRequireUnbreached                *bool                  `json:"password_require_unbreached,omitempty" path:"password_require_unbreached,omitempty" url:"password_require_unbreached,omitempty"`
	PasswordRequirementsApplyToBundles       *bool                  `json:"password_requirements_apply_to_bundles,omitempty" path:"password_requirements_apply_to_bundles,omitempty" url:"password_requirements_apply_to_bundles,omitempty"`
	PasswordValidityDays                     int64                  `json:"password_validity_days,omitempty" path:"password_validity_days,omitempty" url:"password_validity_days,omitempty"`
	Phone                                    string                 `json:"phone,omitempty" path:"phone,omitempty" url:"phone,omitempty"`
	PinAllRemoteServersToSiteRegion          *bool                  `json:"pin_all_remote_servers_to_site_region,omitempty" path:"pin_all_remote_servers_to_site_region,omitempty" url:"pin_all_remote_servers_to_site_region,omitempty"`
	PreventRootPermissionsForNonSiteAdmins   *bool                  `json:"prevent_root_permissions_for_non_site_admins,omitempty" path:"prevent_root_permissions_for_non_site_admins,omitempty" url:"prevent_root_permissions_for_non_site_admins,omitempty"`
	ProtocolAccessGroupsOnly                 *bool                  `json:"protocol_access_groups_only,omitempty" path:"protocol_access_groups_only,omitempty" url:"protocol_access_groups_only,omitempty"`
	Require2fa                               *bool                  `json:"require_2fa,omitempty" path:"require_2fa,omitempty" url:"require_2fa,omitempty"`
	Require2faStopTime                       *time.Time             `json:"require_2fa_stop_time,omitempty" path:"require_2fa_stop_time,omitempty" url:"require_2fa_stop_time,omitempty"`
	RevokeBundleAccessOnDisableOrDelete      *bool                  `json:"revoke_bundle_access_on_disable_or_delete,omitempty" path:"revoke_bundle_access_on_disable_or_delete,omitempty" url:"revoke_bundle_access_on_disable_or_delete,omitempty"`
	Require2faUserType                       string                 `json:"require_2fa_user_type,omitempty" path:"require_2fa_user_type,omitempty" url:"require_2fa_user_type,omitempty"`
	RequireLogoutFromBundlesAndInboxes       *bool                  `json:"require_logout_from_bundles_and_inboxes,omitempty" path:"require_logout_from_bundles_and_inboxes,omitempty" url:"require_logout_from_bundles_and_inboxes,omitempty"`
	Session                                  Session                `json:"session,omitempty" path:"session,omitempty" url:"session,omitempty"`
	SftpEnabled                              *bool                  `json:"sftp_enabled,omitempty" path:"sftp_enabled,omitempty" url:"sftp_enabled,omitempty"`
	SftpHostKeyType                          string                 `json:"sftp_host_key_type,omitempty" path:"sftp_host_key_type,omitempty" url:"sftp_host_key_type,omitempty"`
	ActiveSftpHostKeyId                      int64                  `json:"active_sftp_host_key_id,omitempty" path:"active_sftp_host_key_id,omitempty" url:"active_sftp_host_key_id,omitempty"`
	SftpInsecureCiphers                      *bool                  `json:"sftp_insecure_ciphers,omitempty" path:"sftp_insecure_ciphers,omitempty" url:"sftp_insecure_ciphers,omitempty"`
	SftpInsecureDiffieHellman                *bool                  `json:"sftp_insecure_diffie_hellman,omitempty" path:"sftp_insecure_diffie_hellman,omitempty" url:"sftp_insecure_diffie_hellman,omitempty"`
	SftpUserRootEnabled                      *bool                  `json:"sftp_user_root_enabled,omitempty" path:"sftp_user_root_enabled,omitempty" url:"sftp_user_root_enabled,omitempty"`
	SharingEnabled                           *bool                  `json:"sharing_enabled,omitempty" path:"sharing_enabled,omitempty" url:"sharing_enabled,omitempty"`
	ShowUserNotificationsLogInLink           *bool                  `json:"show_user_notifications_log_in_link,omitempty" path:"show_user_notifications_log_in_link,omitempty" url:"show_user_notifications_log_in_link,omitempty"`
	ShowRequestAccessLink                    *bool                  `json:"show_request_access_link,omitempty" path:"show_request_access_link,omitempty" url:"show_request_access_link,omitempty"`
	SiteFooter                               string                 `json:"site_footer,omitempty" path:"site_footer,omitempty" url:"site_footer,omitempty"`
	SiteHeader                               string                 `json:"site_header,omitempty" path:"site_header,omitempty" url:"site_header,omitempty"`
	SitePublicFooter                         string                 `json:"site_public_footer,omitempty" path:"site_public_footer,omitempty" url:"site_public_footer,omitempty"`
	SitePublicHeader                         string                 `json:"site_public_header,omitempty" path:"site_public_header,omitempty" url:"site_public_header,omitempty"`
	SmtpAddress                              string                 `json:"smtp_address,omitempty" path:"smtp_address,omitempty" url:"smtp_address,omitempty"`
	SmtpAuthentication                       string                 `json:"smtp_authentication,omitempty" path:"smtp_authentication,omitempty" url:"smtp_authentication,omitempty"`
	SmtpFrom                                 string                 `json:"smtp_from,omitempty" path:"smtp_from,omitempty" url:"smtp_from,omitempty"`
	SmtpPort                                 int64                  `json:"smtp_port,omitempty" path:"smtp_port,omitempty" url:"smtp_port,omitempty"`
	SmtpUsername                             string                 `json:"smtp_username,omitempty" path:"smtp_username,omitempty" url:"smtp_username,omitempty"`
	SessionExpiry                            string                 `json:"session_expiry,omitempty" path:"session_expiry,omitempty" url:"session_expiry,omitempty"`
	SessionExpiryMinutes                     int64                  `json:"session_expiry_minutes,omitempty" path:"session_expiry_minutes,omitempty" url:"session_expiry_minutes,omitempty"`
	SnapshotSharingEnabled                   *bool                  `json:"snapshot_sharing_enabled,omitempty" path:"snapshot_sharing_enabled,omitempty" url:"snapshot_sharing_enabled,omitempty"`
	SslRequired                              *bool                  `json:"ssl_required,omitempty" path:"ssl_required,omitempty" url:"ssl_required,omitempty"`
	Subdomain                                string                 `json:"subdomain,omitempty" path:"subdomain,omitempty" url:"subdomain,omitempty"`
	SwitchToPlanDate                         *time.Time             `json:"switch_to_plan_date,omitempty" path:"switch_to_plan_date,omitempty" url:"switch_to_plan_date,omitempty"`
	TrialDaysLeft                            int64                  `json:"trial_days_left,omitempty" path:"trial_days_left,omitempty" url:"trial_days_left,omitempty"`
	TrialUntil                               *time.Time             `json:"trial_until,omitempty" path:"trial_until,omitempty" url:"trial_until,omitempty"`
	UseDedicatedIpsForSmtp                   *bool                  `json:"use_dedicated_ips_for_smtp,omitempty" path:"use_dedicated_ips_for_smtp,omitempty" url:"use_dedicated_ips_for_smtp,omitempty"`
	UseProvidedModifiedAt                    *bool                  `json:"use_provided_modified_at,omitempty" path:"use_provided_modified_at,omitempty" url:"use_provided_modified_at,omitempty"`
	User                                     User                   `json:"user,omitempty" path:"user,omitempty" url:"user,omitempty"`
	UserLockout                              *bool                  `json:"user_lockout,omitempty" path:"user_lockout,omitempty" url:"user_lockout,omitempty"`
	UserLockoutLockPeriod                    int64                  `json:"user_lockout_lock_period,omitempty" path:"user_lockout_lock_period,omitempty" url:"user_lockout_lock_period,omitempty"`
	UserLockoutTries                         int64                  `json:"user_lockout_tries,omitempty" path:"user_lockout_tries,omitempty" url:"user_lockout_tries,omitempty"`
	UserLockoutWithin                        int64                  `json:"user_lockout_within,omitempty" path:"user_lockout_within,omitempty" url:"user_lockout_within,omitempty"`
	UserRequestsEnabled                      *bool                  `json:"user_requests_enabled,omitempty" path:"user_requests_enabled,omitempty" url:"user_requests_enabled,omitempty"`
	UserRequestsNotifyAdmins                 *bool                  `json:"user_requests_notify_admins,omitempty" path:"user_requests_notify_admins,omitempty" url:"user_requests_notify_admins,omitempty"`
	UsersCanCreateApiKeys                    *bool                  `json:"users_can_create_api_keys,omitempty" path:"users_can_create_api_keys,omitempty" url:"users_can_create_api_keys,omitempty"`
	UsersCanCreateSshKeys                    *bool                  `json:"users_can_create_ssh_keys,omitempty" path:"users_can_create_ssh_keys,omitempty" url:"users_can_create_ssh_keys,omitempty"`
	WelcomeCustomText                        string                 `json:"welcome_custom_text,omitempty" path:"welcome_custom_text,omitempty" url:"welcome_custom_text,omitempty"`
	WelcomeEmailCc                           string                 `json:"welcome_email_cc,omitempty" path:"welcome_email_cc,omitempty" url:"welcome_email_cc,omitempty"`
	WelcomeEmailSubject                      string                 `json:"welcome_email_subject,omitempty" path:"welcome_email_subject,omitempty" url:"welcome_email_subject,omitempty"`
	WelcomeEmailEnabled                      *bool                  `json:"welcome_email_enabled,omitempty" path:"welcome_email_enabled,omitempty" url:"welcome_email_enabled,omitempty"`
	WelcomeScreen                            string                 `json:"welcome_screen,omitempty" path:"welcome_screen,omitempty" url:"welcome_screen,omitempty"`
	WindowsModeFtp                           *bool                  `json:"windows_mode_ftp,omitempty" path:"windows_mode_ftp,omitempty" url:"windows_mode_ftp,omitempty"`
	GroupAdminsCanSetUserPassword            *bool                  `json:"group_admins_can_set_user_password,omitempty" path:"group_admins_can_set_user_password,omitempty" url:"group_admins_can_set_user_password,omitempty"`
}

func (s Site) Identifier() interface{} {
	return s.Id
}

type SiteCollection []Site

type SiteUpdateParams struct {
	Name                                     string                 `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Subdomain                                string                 `url:"subdomain,omitempty" json:"subdomain,omitempty" path:"subdomain"`
	Domain                                   string                 `url:"domain,omitempty" json:"domain,omitempty" path:"domain"`
	DomainHstsHeader                         *bool                  `url:"domain_hsts_header,omitempty" json:"domain_hsts_header,omitempty" path:"domain_hsts_header"`
	DomainLetsencryptChain                   string                 `url:"domain_letsencrypt_chain,omitempty" json:"domain_letsencrypt_chain,omitempty" path:"domain_letsencrypt_chain"`
	Email                                    string                 `url:"email,omitempty" json:"email,omitempty" path:"email"`
	ReplyToEmail                             string                 `url:"reply_to_email,omitempty" json:"reply_to_email,omitempty" path:"reply_to_email"`
	AllowBundleNames                         *bool                  `url:"allow_bundle_names,omitempty" json:"allow_bundle_names,omitempty" path:"allow_bundle_names"`
	BundleExpiration                         int64                  `url:"bundle_expiration,omitempty" json:"bundle_expiration,omitempty" path:"bundle_expiration"`
	WelcomeEmailEnabled                      *bool                  `url:"welcome_email_enabled,omitempty" json:"welcome_email_enabled,omitempty" path:"welcome_email_enabled"`
	AskAboutOverwrites                       *bool                  `url:"ask_about_overwrites,omitempty" json:"ask_about_overwrites,omitempty" path:"ask_about_overwrites"`
	ShowRequestAccessLink                    *bool                  `url:"show_request_access_link,omitempty" json:"show_request_access_link,omitempty" path:"show_request_access_link"`
	AlwaysMkdirParents                       *bool                  `url:"always_mkdir_parents,omitempty" json:"always_mkdir_parents,omitempty" path:"always_mkdir_parents"`
	WelcomeEmailCc                           string                 `url:"welcome_email_cc,omitempty" json:"welcome_email_cc,omitempty" path:"welcome_email_cc"`
	WelcomeEmailSubject                      string                 `url:"welcome_email_subject,omitempty" json:"welcome_email_subject,omitempty" path:"welcome_email_subject"`
	WelcomeCustomText                        string                 `url:"welcome_custom_text,omitempty" json:"welcome_custom_text,omitempty" path:"welcome_custom_text"`
	Language                                 string                 `url:"language,omitempty" json:"language,omitempty" path:"language"`
	WindowsModeFtp                           *bool                  `url:"windows_mode_ftp,omitempty" json:"windows_mode_ftp,omitempty" path:"windows_mode_ftp"`
	DefaultTimeZone                          string                 `url:"default_time_zone,omitempty" json:"default_time_zone,omitempty" path:"default_time_zone"`
	DesktopApp                               *bool                  `url:"desktop_app,omitempty" json:"desktop_app,omitempty" path:"desktop_app"`
	DesktopAppSessionIpPinning               *bool                  `url:"desktop_app_session_ip_pinning,omitempty" json:"desktop_app_session_ip_pinning,omitempty" path:"desktop_app_session_ip_pinning"`
	DesktopAppSessionLifetime                int64                  `url:"desktop_app_session_lifetime,omitempty" json:"desktop_app_session_lifetime,omitempty" path:"desktop_app_session_lifetime"`
	MobileApp                                *bool                  `url:"mobile_app,omitempty" json:"mobile_app,omitempty" path:"mobile_app"`
	MobileAppSessionIpPinning                *bool                  `url:"mobile_app_session_ip_pinning,omitempty" json:"mobile_app_session_ip_pinning,omitempty" path:"mobile_app_session_ip_pinning"`
	MobileAppSessionLifetime                 int64                  `url:"mobile_app_session_lifetime,omitempty" json:"mobile_app_session_lifetime,omitempty" path:"mobile_app_session_lifetime"`
	FolderPermissionsGroupsOnly              *bool                  `url:"folder_permissions_groups_only,omitempty" json:"folder_permissions_groups_only,omitempty" path:"folder_permissions_groups_only"`
	WelcomeScreen                            string                 `url:"welcome_screen,omitempty" json:"welcome_screen,omitempty" path:"welcome_screen"`
	OfficeIntegrationAvailable               *bool                  `url:"office_integration_available,omitempty" json:"office_integration_available,omitempty" path:"office_integration_available"`
	OfficeIntegrationType                    string                 `url:"office_integration_type,omitempty" json:"office_integration_type,omitempty" path:"office_integration_type"`
	PinAllRemoteServersToSiteRegion          *bool                  `url:"pin_all_remote_servers_to_site_region,omitempty" json:"pin_all_remote_servers_to_site_region,omitempty" path:"pin_all_remote_servers_to_site_region"`
	MotdText                                 string                 `url:"motd_text,omitempty" json:"motd_text,omitempty" path:"motd_text"`
	MotdUseForFtp                            *bool                  `url:"motd_use_for_ftp,omitempty" json:"motd_use_for_ftp,omitempty" path:"motd_use_for_ftp"`
	MotdUseForSftp                           *bool                  `url:"motd_use_for_sftp,omitempty" json:"motd_use_for_sftp,omitempty" path:"motd_use_for_sftp"`
	LeftNavigationVisibility                 map[string]interface{} `url:"left_navigation_visibility,omitempty" json:"left_navigation_visibility,omitempty" path:"left_navigation_visibility"`
	AdditionalTextFileTypes                  []string               `url:"additional_text_file_types,omitempty" json:"additional_text_file_types,omitempty" path:"additional_text_file_types"`
	BundleRequireNote                        *bool                  `url:"bundle_require_note,omitempty" json:"bundle_require_note,omitempty" path:"bundle_require_note"`
	BundleSendSharedReceipts                 *bool                  `url:"bundle_send_shared_receipts,omitempty" json:"bundle_send_shared_receipts,omitempty" path:"bundle_send_shared_receipts"`
	CalculateFileChecksumsCrc32              *bool                  `url:"calculate_file_checksums_crc32,omitempty" json:"calculate_file_checksums_crc32,omitempty" path:"calculate_file_checksums_crc32"`
	CalculateFileChecksumsMd5                *bool                  `url:"calculate_file_checksums_md5,omitempty" json:"calculate_file_checksums_md5,omitempty" path:"calculate_file_checksums_md5"`
	CalculateFileChecksumsSha1               *bool                  `url:"calculate_file_checksums_sha1,omitempty" json:"calculate_file_checksums_sha1,omitempty" path:"calculate_file_checksums_sha1"`
	CalculateFileChecksumsSha256             *bool                  `url:"calculate_file_checksums_sha256,omitempty" json:"calculate_file_checksums_sha256,omitempty" path:"calculate_file_checksums_sha256"`
	LegacyChecksumsMode                      *bool                  `url:"legacy_checksums_mode,omitempty" json:"legacy_checksums_mode,omitempty" path:"legacy_checksums_mode"`
	MigrateRemoteServerSyncToSync            *bool                  `url:"migrate_remote_server_sync_to_sync,omitempty" json:"migrate_remote_server_sync_to_sync,omitempty" path:"migrate_remote_server_sync_to_sync"`
	As2MessageRetentionDays                  int64                  `url:"as2_message_retention_days,omitempty" json:"as2_message_retention_days,omitempty" path:"as2_message_retention_days"`
	SessionExpiry                            string                 `url:"session_expiry,omitempty" json:"session_expiry,omitempty" path:"session_expiry"`
	SslRequired                              *bool                  `url:"ssl_required,omitempty" json:"ssl_required,omitempty" path:"ssl_required"`
	SftpInsecureCiphers                      *bool                  `url:"sftp_insecure_ciphers,omitempty" json:"sftp_insecure_ciphers,omitempty" path:"sftp_insecure_ciphers"`
	SftpInsecureDiffieHellman                *bool                  `url:"sftp_insecure_diffie_hellman,omitempty" json:"sftp_insecure_diffie_hellman,omitempty" path:"sftp_insecure_diffie_hellman"`
	DisableFilesCertificateGeneration        *bool                  `url:"disable_files_certificate_generation,omitempty" json:"disable_files_certificate_generation,omitempty" path:"disable_files_certificate_generation"`
	UserLockout                              *bool                  `url:"user_lockout,omitempty" json:"user_lockout,omitempty" path:"user_lockout"`
	UserLockoutTries                         int64                  `url:"user_lockout_tries,omitempty" json:"user_lockout_tries,omitempty" path:"user_lockout_tries"`
	UserLockoutWithin                        int64                  `url:"user_lockout_within,omitempty" json:"user_lockout_within,omitempty" path:"user_lockout_within"`
	UserLockoutLockPeriod                    int64                  `url:"user_lockout_lock_period,omitempty" json:"user_lockout_lock_period,omitempty" path:"user_lockout_lock_period"`
	IncludePasswordInWelcomeEmail            *bool                  `url:"include_password_in_welcome_email,omitempty" json:"include_password_in_welcome_email,omitempty" path:"include_password_in_welcome_email"`
	AllowedCountries                         string                 `url:"allowed_countries,omitempty" json:"allowed_countries,omitempty" path:"allowed_countries"`
	AllowedIps                               string                 `url:"allowed_ips,omitempty" json:"allowed_ips,omitempty" path:"allowed_ips"`
	DisallowedCountries                      string                 `url:"disallowed_countries,omitempty" json:"disallowed_countries,omitempty" path:"disallowed_countries"`
	DaysToRetainBackups                      int64                  `url:"days_to_retain_backups,omitempty" json:"days_to_retain_backups,omitempty" path:"days_to_retain_backups"`
	MaxPriorPasswords                        int64                  `url:"max_prior_passwords,omitempty" json:"max_prior_passwords,omitempty" path:"max_prior_passwords"`
	PasswordValidityDays                     int64                  `url:"password_validity_days,omitempty" json:"password_validity_days,omitempty" path:"password_validity_days"`
	PasswordMinLength                        int64                  `url:"password_min_length,omitempty" json:"password_min_length,omitempty" path:"password_min_length"`
	PasswordRequireLetter                    *bool                  `url:"password_require_letter,omitempty" json:"password_require_letter,omitempty" path:"password_require_letter"`
	PasswordRequireMixed                     *bool                  `url:"password_require_mixed,omitempty" json:"password_require_mixed,omitempty" path:"password_require_mixed"`
	PasswordRequireSpecial                   *bool                  `url:"password_require_special,omitempty" json:"password_require_special,omitempty" path:"password_require_special"`
	PasswordRequireNumber                    *bool                  `url:"password_require_number,omitempty" json:"password_require_number,omitempty" path:"password_require_number"`
	PasswordRequireUnbreached                *bool                  `url:"password_require_unbreached,omitempty" json:"password_require_unbreached,omitempty" path:"password_require_unbreached"`
	RequireLogoutFromBundlesAndInboxes       *bool                  `url:"require_logout_from_bundles_and_inboxes,omitempty" json:"require_logout_from_bundles_and_inboxes,omitempty" path:"require_logout_from_bundles_and_inboxes"`
	DavUserRootEnabled                       *bool                  `url:"dav_user_root_enabled,omitempty" json:"dav_user_root_enabled,omitempty" path:"dav_user_root_enabled"`
	SftpUserRootEnabled                      *bool                  `url:"sftp_user_root_enabled,omitempty" json:"sftp_user_root_enabled,omitempty" path:"sftp_user_root_enabled"`
	DisablePasswordReset                     *bool                  `url:"disable_password_reset,omitempty" json:"disable_password_reset,omitempty" path:"disable_password_reset"`
	ImmutableFiles                           *bool                  `url:"immutable_files,omitempty" json:"immutable_files,omitempty" path:"immutable_files"`
	BundleNotFoundMessage                    string                 `url:"bundle_not_found_message,omitempty" json:"bundle_not_found_message,omitempty" path:"bundle_not_found_message"`
	BundlePasswordRequired                   *bool                  `url:"bundle_password_required,omitempty" json:"bundle_password_required,omitempty" path:"bundle_password_required"`
	BundleRequireRegistration                *bool                  `url:"bundle_require_registration,omitempty" json:"bundle_require_registration,omitempty" path:"bundle_require_registration"`
	BundleRequireShareRecipient              *bool                  `url:"bundle_require_share_recipient,omitempty" json:"bundle_require_share_recipient,omitempty" path:"bundle_require_share_recipient"`
	BundleRegistrationNotifications          string                 `url:"bundle_registration_notifications,omitempty" json:"bundle_registration_notifications,omitempty" path:"bundle_registration_notifications"`
	BundleActivityNotifications              string                 `url:"bundle_activity_notifications,omitempty" json:"bundle_activity_notifications,omitempty" path:"bundle_activity_notifications"`
	BundleUploadReceiptNotifications         string                 `url:"bundle_upload_receipt_notifications,omitempty" json:"bundle_upload_receipt_notifications,omitempty" path:"bundle_upload_receipt_notifications"`
	DocumentEditsInBundleAllowed             *bool                  `url:"document_edits_in_bundle_allowed,omitempty" json:"document_edits_in_bundle_allowed,omitempty" path:"document_edits_in_bundle_allowed"`
	PasswordRequirementsApplyToBundles       *bool                  `url:"password_requirements_apply_to_bundles,omitempty" json:"password_requirements_apply_to_bundles,omitempty" path:"password_requirements_apply_to_bundles"`
	PreventRootPermissionsForNonSiteAdmins   *bool                  `url:"prevent_root_permissions_for_non_site_admins,omitempty" json:"prevent_root_permissions_for_non_site_admins,omitempty" path:"prevent_root_permissions_for_non_site_admins"`
	OptOutGlobal                             *bool                  `url:"opt_out_global,omitempty" json:"opt_out_global,omitempty" path:"opt_out_global"`
	UseProvidedModifiedAt                    *bool                  `url:"use_provided_modified_at,omitempty" json:"use_provided_modified_at,omitempty" path:"use_provided_modified_at"`
	CustomNamespace                          *bool                  `url:"custom_namespace,omitempty" json:"custom_namespace,omitempty" path:"custom_namespace"`
	NonSsoGroupsAllowed                      *bool                  `url:"non_sso_groups_allowed,omitempty" json:"non_sso_groups_allowed,omitempty" path:"non_sso_groups_allowed"`
	NonSsoUsersAllowed                       *bool                  `url:"non_sso_users_allowed,omitempty" json:"non_sso_users_allowed,omitempty" path:"non_sso_users_allowed"`
	SharingEnabled                           *bool                  `url:"sharing_enabled,omitempty" json:"sharing_enabled,omitempty" path:"sharing_enabled"`
	SnapshotSharingEnabled                   *bool                  `url:"snapshot_sharing_enabled,omitempty" json:"snapshot_sharing_enabled,omitempty" path:"snapshot_sharing_enabled"`
	UserRequestsEnabled                      *bool                  `url:"user_requests_enabled,omitempty" json:"user_requests_enabled,omitempty" path:"user_requests_enabled"`
	UserRequestsNotifyAdmins                 *bool                  `url:"user_requests_notify_admins,omitempty" json:"user_requests_notify_admins,omitempty" path:"user_requests_notify_admins"`
	DavEnabled                               *bool                  `url:"dav_enabled,omitempty" json:"dav_enabled,omitempty" path:"dav_enabled"`
	FtpEnabled                               *bool                  `url:"ftp_enabled,omitempty" json:"ftp_enabled,omitempty" path:"ftp_enabled"`
	SftpEnabled                              *bool                  `url:"sftp_enabled,omitempty" json:"sftp_enabled,omitempty" path:"sftp_enabled"`
	UsersCanCreateApiKeys                    *bool                  `url:"users_can_create_api_keys,omitempty" json:"users_can_create_api_keys,omitempty" path:"users_can_create_api_keys"`
	UsersCanCreateSshKeys                    *bool                  `url:"users_can_create_ssh_keys,omitempty" json:"users_can_create_ssh_keys,omitempty" path:"users_can_create_ssh_keys"`
	ShowUserNotificationsLogInLink           *bool                  `url:"show_user_notifications_log_in_link,omitempty" json:"show_user_notifications_log_in_link,omitempty" path:"show_user_notifications_log_in_link"`
	SftpHostKeyType                          string                 `url:"sftp_host_key_type,omitempty" json:"sftp_host_key_type,omitempty" path:"sftp_host_key_type"`
	ActiveSftpHostKeyId                      int64                  `url:"active_sftp_host_key_id,omitempty" json:"active_sftp_host_key_id,omitempty" path:"active_sftp_host_key_id"`
	ProtocolAccessGroupsOnly                 *bool                  `url:"protocol_access_groups_only,omitempty" json:"protocol_access_groups_only,omitempty" path:"protocol_access_groups_only"`
	RevokeBundleAccessOnDisableOrDelete      *bool                  `url:"revoke_bundle_access_on_disable_or_delete,omitempty" json:"revoke_bundle_access_on_disable_or_delete,omitempty" path:"revoke_bundle_access_on_disable_or_delete"`
	BundleWatermarkValue                     map[string]interface{} `url:"bundle_watermark_value,omitempty" json:"bundle_watermark_value,omitempty" path:"bundle_watermark_value"`
	GroupAdminsCanSetUserPassword            *bool                  `url:"group_admins_can_set_user_password,omitempty" json:"group_admins_can_set_user_password,omitempty" path:"group_admins_can_set_user_password"`
	BundleRecipientBlacklistFreeEmailDomains *bool                  `url:"bundle_recipient_blacklist_free_email_domains,omitempty" json:"bundle_recipient_blacklist_free_email_domains,omitempty" path:"bundle_recipient_blacklist_free_email_domains"`
	BundleRecipientBlacklistDomains          []string               `url:"bundle_recipient_blacklist_domains,omitempty" json:"bundle_recipient_blacklist_domains,omitempty" path:"bundle_recipient_blacklist_domains"`
	AdminsBypassLockedSubfolders             *bool                  `url:"admins_bypass_locked_subfolders,omitempty" json:"admins_bypass_locked_subfolders,omitempty" path:"admins_bypass_locked_subfolders"`
	Allowed2faMethodSms                      *bool                  `url:"allowed_2fa_method_sms,omitempty" json:"allowed_2fa_method_sms,omitempty" path:"allowed_2fa_method_sms"`
	Allowed2faMethodTotp                     *bool                  `url:"allowed_2fa_method_totp,omitempty" json:"allowed_2fa_method_totp,omitempty" path:"allowed_2fa_method_totp"`
	Allowed2faMethodWebauthn                 *bool                  `url:"allowed_2fa_method_webauthn,omitempty" json:"allowed_2fa_method_webauthn,omitempty" path:"allowed_2fa_method_webauthn"`
	Allowed2faMethodYubi                     *bool                  `url:"allowed_2fa_method_yubi,omitempty" json:"allowed_2fa_method_yubi,omitempty" path:"allowed_2fa_method_yubi"`
	Allowed2faMethodEmail                    *bool                  `url:"allowed_2fa_method_email,omitempty" json:"allowed_2fa_method_email,omitempty" path:"allowed_2fa_method_email"`
	Allowed2faMethodStatic                   *bool                  `url:"allowed_2fa_method_static,omitempty" json:"allowed_2fa_method_static,omitempty" path:"allowed_2fa_method_static"`
	Allowed2faMethodBypassForFtpSftpDav      *bool                  `url:"allowed_2fa_method_bypass_for_ftp_sftp_dav,omitempty" json:"allowed_2fa_method_bypass_for_ftp_sftp_dav,omitempty" path:"allowed_2fa_method_bypass_for_ftp_sftp_dav"`
	Require2fa                               *bool                  `url:"require_2fa,omitempty" json:"require_2fa,omitempty" path:"require_2fa"`
	Require2faUserType                       string                 `url:"require_2fa_user_type,omitempty" json:"require_2fa_user_type,omitempty" path:"require_2fa_user_type"`
	Color2Top                                string                 `url:"color2_top,omitempty" json:"color2_top,omitempty" path:"color2_top"`
	Color2Left                               string                 `url:"color2_left,omitempty" json:"color2_left,omitempty" path:"color2_left"`
	Color2Link                               string                 `url:"color2_link,omitempty" json:"color2_link,omitempty" path:"color2_link"`
	Color2Text                               string                 `url:"color2_text,omitempty" json:"color2_text,omitempty" path:"color2_text"`
	Color2TopText                            string                 `url:"color2_top_text,omitempty" json:"color2_top_text,omitempty" path:"color2_top_text"`
	SiteHeader                               string                 `url:"site_header,omitempty" json:"site_header,omitempty" path:"site_header"`
	SiteFooter                               string                 `url:"site_footer,omitempty" json:"site_footer,omitempty" path:"site_footer"`
	SitePublicHeader                         string                 `url:"site_public_header,omitempty" json:"site_public_header,omitempty" path:"site_public_header"`
	SitePublicFooter                         string                 `url:"site_public_footer,omitempty" json:"site_public_footer,omitempty" path:"site_public_footer"`
	LoginHelpText                            string                 `url:"login_help_text,omitempty" json:"login_help_text,omitempty" path:"login_help_text"`
	UseDedicatedIpsForSmtp                   *bool                  `url:"use_dedicated_ips_for_smtp,omitempty" json:"use_dedicated_ips_for_smtp,omitempty" path:"use_dedicated_ips_for_smtp"`
	SmtpAddress                              string                 `url:"smtp_address,omitempty" json:"smtp_address,omitempty" path:"smtp_address"`
	SmtpAuthentication                       string                 `url:"smtp_authentication,omitempty" json:"smtp_authentication,omitempty" path:"smtp_authentication"`
	SmtpFrom                                 string                 `url:"smtp_from,omitempty" json:"smtp_from,omitempty" path:"smtp_from"`
	SmtpUsername                             string                 `url:"smtp_username,omitempty" json:"smtp_username,omitempty" path:"smtp_username"`
	SmtpPort                                 int64                  `url:"smtp_port,omitempty" json:"smtp_port,omitempty" path:"smtp_port"`
	LdapEnabled                              *bool                  `url:"ldap_enabled,omitempty" json:"ldap_enabled,omitempty" path:"ldap_enabled"`
	LdapType                                 string                 `url:"ldap_type,omitempty" json:"ldap_type,omitempty" path:"ldap_type"`
	LdapHost                                 string                 `url:"ldap_host,omitempty" json:"ldap_host,omitempty" path:"ldap_host"`
	LdapHost2                                string                 `url:"ldap_host_2,omitempty" json:"ldap_host_2,omitempty" path:"ldap_host_2"`
	LdapHost3                                string                 `url:"ldap_host_3,omitempty" json:"ldap_host_3,omitempty" path:"ldap_host_3"`
	LdapPort                                 int64                  `url:"ldap_port,omitempty" json:"ldap_port,omitempty" path:"ldap_port"`
	LdapSecure                               *bool                  `url:"ldap_secure,omitempty" json:"ldap_secure,omitempty" path:"ldap_secure"`
	LdapUsername                             string                 `url:"ldap_username,omitempty" json:"ldap_username,omitempty" path:"ldap_username"`
	LdapUsernameField                        string                 `url:"ldap_username_field,omitempty" json:"ldap_username_field,omitempty" path:"ldap_username_field"`
	LdapDomain                               string                 `url:"ldap_domain,omitempty" json:"ldap_domain,omitempty" path:"ldap_domain"`
	LdapUserAction                           string                 `url:"ldap_user_action,omitempty" json:"ldap_user_action,omitempty" path:"ldap_user_action"`
	LdapGroupAction                          string                 `url:"ldap_group_action,omitempty" json:"ldap_group_action,omitempty" path:"ldap_group_action"`
	LdapUserIncludeGroups                    string                 `url:"ldap_user_include_groups,omitempty" json:"ldap_user_include_groups,omitempty" path:"ldap_user_include_groups"`
	LdapGroupExclusion                       string                 `url:"ldap_group_exclusion,omitempty" json:"ldap_group_exclusion,omitempty" path:"ldap_group_exclusion"`
	LdapGroupInclusion                       string                 `url:"ldap_group_inclusion,omitempty" json:"ldap_group_inclusion,omitempty" path:"ldap_group_inclusion"`
	LdapBaseDn                               string                 `url:"ldap_base_dn,omitempty" json:"ldap_base_dn,omitempty" path:"ldap_base_dn"`
	UploadsViaEmailAuthentication            *bool                  `url:"uploads_via_email_authentication,omitempty" json:"uploads_via_email_authentication,omitempty" path:"uploads_via_email_authentication"`
	Icon16File                               io.Writer              `url:"icon16_file,omitempty" json:"icon16_file,omitempty" path:"icon16_file"`
	Icon16Delete                             *bool                  `url:"icon16_delete,omitempty" json:"icon16_delete,omitempty" path:"icon16_delete"`
	Icon32File                               io.Writer              `url:"icon32_file,omitempty" json:"icon32_file,omitempty" path:"icon32_file"`
	Icon32Delete                             *bool                  `url:"icon32_delete,omitempty" json:"icon32_delete,omitempty" path:"icon32_delete"`
	Icon48File                               io.Writer              `url:"icon48_file,omitempty" json:"icon48_file,omitempty" path:"icon48_file"`
	Icon48Delete                             *bool                  `url:"icon48_delete,omitempty" json:"icon48_delete,omitempty" path:"icon48_delete"`
	Icon128File                              io.Writer              `url:"icon128_file,omitempty" json:"icon128_file,omitempty" path:"icon128_file"`
	Icon128Delete                            *bool                  `url:"icon128_delete,omitempty" json:"icon128_delete,omitempty" path:"icon128_delete"`
	LogoFile                                 io.Writer              `url:"logo_file,omitempty" json:"logo_file,omitempty" path:"logo_file"`
	LogoDelete                               *bool                  `url:"logo_delete,omitempty" json:"logo_delete,omitempty" path:"logo_delete"`
	BundleWatermarkAttachmentFile            io.Writer              `url:"bundle_watermark_attachment_file,omitempty" json:"bundle_watermark_attachment_file,omitempty" path:"bundle_watermark_attachment_file"`
	BundleWatermarkAttachmentDelete          *bool                  `url:"bundle_watermark_attachment_delete,omitempty" json:"bundle_watermark_attachment_delete,omitempty" path:"bundle_watermark_attachment_delete"`
	LoginPageBackgroundImageFile             io.Writer              `url:"login_page_background_image_file,omitempty" json:"login_page_background_image_file,omitempty" path:"login_page_background_image_file"`
	LoginPageBackgroundImageDelete           *bool                  `url:"login_page_background_image_delete,omitempty" json:"login_page_background_image_delete,omitempty" path:"login_page_background_image_delete"`
	Disable2faWithDelay                      *bool                  `url:"disable_2fa_with_delay,omitempty" json:"disable_2fa_with_delay,omitempty" path:"disable_2fa_with_delay"`
	LdapPasswordChange                       string                 `url:"ldap_password_change,omitempty" json:"ldap_password_change,omitempty" path:"ldap_password_change"`
	LdapPasswordChangeConfirmation           string                 `url:"ldap_password_change_confirmation,omitempty" json:"ldap_password_change_confirmation,omitempty" path:"ldap_password_change_confirmation"`
	SmtpPassword                             string                 `url:"smtp_password,omitempty" json:"smtp_password,omitempty" path:"smtp_password"`
	SessionExpiryMinutes                     int64                  `url:"session_expiry_minutes,omitempty" json:"session_expiry_minutes,omitempty" path:"session_expiry_minutes"`
}

func (s *Site) UnmarshalJSON(data []byte) error {
	type site Site
	var v site
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Site(v)
	return nil
}

func (s *SiteCollection) UnmarshalJSON(data []byte) error {
	type sites SiteCollection
	var v sites
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SiteCollection(v)
	return nil
}

func (s *SiteCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
