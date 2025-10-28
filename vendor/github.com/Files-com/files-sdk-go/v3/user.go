package files_sdk

import (
	"encoding/json"
	"io"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type User struct {
	Id                               int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Username                         string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	AdminGroupIds                    []int64    `json:"admin_group_ids,omitempty" path:"admin_group_ids,omitempty" url:"admin_group_ids,omitempty"`
	AllowedIps                       string     `json:"allowed_ips,omitempty" path:"allowed_ips,omitempty" url:"allowed_ips,omitempty"`
	AttachmentsPermission            *bool      `json:"attachments_permission,omitempty" path:"attachments_permission,omitempty" url:"attachments_permission,omitempty"`
	ApiKeysCount                     int64      `json:"api_keys_count,omitempty" path:"api_keys_count,omitempty" url:"api_keys_count,omitempty"`
	AuthenticateUntil                *time.Time `json:"authenticate_until,omitempty" path:"authenticate_until,omitempty" url:"authenticate_until,omitempty"`
	AuthenticationMethod             string     `json:"authentication_method,omitempty" path:"authentication_method,omitempty" url:"authentication_method,omitempty"`
	AvatarUrl                        string     `json:"avatar_url,omitempty" path:"avatar_url,omitempty" url:"avatar_url,omitempty"`
	Billable                         *bool      `json:"billable,omitempty" path:"billable,omitempty" url:"billable,omitempty"`
	BillingPermission                *bool      `json:"billing_permission,omitempty" path:"billing_permission,omitempty" url:"billing_permission,omitempty"`
	BypassSiteAllowedIps             *bool      `json:"bypass_site_allowed_ips,omitempty" path:"bypass_site_allowed_ips,omitempty" url:"bypass_site_allowed_ips,omitempty"`
	BypassUserLifecycleRules         *bool      `json:"bypass_user_lifecycle_rules,omitempty" path:"bypass_user_lifecycle_rules,omitempty" url:"bypass_user_lifecycle_rules,omitempty"`
	CreatedAt                        *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	DavPermission                    *bool      `json:"dav_permission,omitempty" path:"dav_permission,omitempty" url:"dav_permission,omitempty"`
	Disabled                         *bool      `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	DisabledExpiredOrInactive        *bool      `json:"disabled_expired_or_inactive,omitempty" path:"disabled_expired_or_inactive,omitempty" url:"disabled_expired_or_inactive,omitempty"`
	Email                            string     `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
	FirstLoginAt                     *time.Time `json:"first_login_at,omitempty" path:"first_login_at,omitempty" url:"first_login_at,omitempty"`
	FtpPermission                    *bool      `json:"ftp_permission,omitempty" path:"ftp_permission,omitempty" url:"ftp_permission,omitempty"`
	GroupIds                         string     `json:"group_ids,omitempty" path:"group_ids,omitempty" url:"group_ids,omitempty"`
	HeaderText                       string     `json:"header_text,omitempty" path:"header_text,omitempty" url:"header_text,omitempty"`
	Language                         string     `json:"language,omitempty" path:"language,omitempty" url:"language,omitempty"`
	LastLoginAt                      *time.Time `json:"last_login_at,omitempty" path:"last_login_at,omitempty" url:"last_login_at,omitempty"`
	LastWebLoginAt                   *time.Time `json:"last_web_login_at,omitempty" path:"last_web_login_at,omitempty" url:"last_web_login_at,omitempty"`
	LastFtpLoginAt                   *time.Time `json:"last_ftp_login_at,omitempty" path:"last_ftp_login_at,omitempty" url:"last_ftp_login_at,omitempty"`
	LastSftpLoginAt                  *time.Time `json:"last_sftp_login_at,omitempty" path:"last_sftp_login_at,omitempty" url:"last_sftp_login_at,omitempty"`
	LastDavLoginAt                   *time.Time `json:"last_dav_login_at,omitempty" path:"last_dav_login_at,omitempty" url:"last_dav_login_at,omitempty"`
	LastDesktopLoginAt               *time.Time `json:"last_desktop_login_at,omitempty" path:"last_desktop_login_at,omitempty" url:"last_desktop_login_at,omitempty"`
	LastRestapiLoginAt               *time.Time `json:"last_restapi_login_at,omitempty" path:"last_restapi_login_at,omitempty" url:"last_restapi_login_at,omitempty"`
	LastApiUseAt                     *time.Time `json:"last_api_use_at,omitempty" path:"last_api_use_at,omitempty" url:"last_api_use_at,omitempty"`
	LastActiveAt                     *time.Time `json:"last_active_at,omitempty" path:"last_active_at,omitempty" url:"last_active_at,omitempty"`
	LastProtocolCipher               string     `json:"last_protocol_cipher,omitempty" path:"last_protocol_cipher,omitempty" url:"last_protocol_cipher,omitempty"`
	LockoutExpires                   *time.Time `json:"lockout_expires,omitempty" path:"lockout_expires,omitempty" url:"lockout_expires,omitempty"`
	Name                             string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Company                          string     `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
	Notes                            string     `json:"notes,omitempty" path:"notes,omitempty" url:"notes,omitempty"`
	NotificationDailySendTime        int64      `json:"notification_daily_send_time,omitempty" path:"notification_daily_send_time,omitempty" url:"notification_daily_send_time,omitempty"`
	OfficeIntegrationEnabled         *bool      `json:"office_integration_enabled,omitempty" path:"office_integration_enabled,omitempty" url:"office_integration_enabled,omitempty"`
	PasswordSetAt                    *time.Time `json:"password_set_at,omitempty" path:"password_set_at,omitempty" url:"password_set_at,omitempty"`
	PasswordValidityDays             int64      `json:"password_validity_days,omitempty" path:"password_validity_days,omitempty" url:"password_validity_days,omitempty"`
	PublicKeysCount                  int64      `json:"public_keys_count,omitempty" path:"public_keys_count,omitempty" url:"public_keys_count,omitempty"`
	ReceiveAdminAlerts               *bool      `json:"receive_admin_alerts,omitempty" path:"receive_admin_alerts,omitempty" url:"receive_admin_alerts,omitempty"`
	Require2fa                       string     `json:"require_2fa,omitempty" path:"require_2fa,omitempty" url:"require_2fa,omitempty"`
	RequireLoginBy                   *time.Time `json:"require_login_by,omitempty" path:"require_login_by,omitempty" url:"require_login_by,omitempty"`
	Active2fa                        *bool      `json:"active_2fa,omitempty" path:"active_2fa,omitempty" url:"active_2fa,omitempty"`
	RequirePasswordChange            *bool      `json:"require_password_change,omitempty" path:"require_password_change,omitempty" url:"require_password_change,omitempty"`
	PasswordExpired                  *bool      `json:"password_expired,omitempty" path:"password_expired,omitempty" url:"password_expired,omitempty"`
	ReadonlySiteAdmin                *bool      `json:"readonly_site_admin,omitempty" path:"readonly_site_admin,omitempty" url:"readonly_site_admin,omitempty"`
	RestapiPermission                *bool      `json:"restapi_permission,omitempty" path:"restapi_permission,omitempty" url:"restapi_permission,omitempty"`
	SelfManaged                      *bool      `json:"self_managed,omitempty" path:"self_managed,omitempty" url:"self_managed,omitempty"`
	SftpPermission                   *bool      `json:"sftp_permission,omitempty" path:"sftp_permission,omitempty" url:"sftp_permission,omitempty"`
	SiteAdmin                        *bool      `json:"site_admin,omitempty" path:"site_admin,omitempty" url:"site_admin,omitempty"`
	SiteId                           int64      `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
	SkipWelcomeScreen                *bool      `json:"skip_welcome_screen,omitempty" path:"skip_welcome_screen,omitempty" url:"skip_welcome_screen,omitempty"`
	SslRequired                      string     `json:"ssl_required,omitempty" path:"ssl_required,omitempty" url:"ssl_required,omitempty"`
	SsoStrategyId                    int64      `json:"sso_strategy_id,omitempty" path:"sso_strategy_id,omitempty" url:"sso_strategy_id,omitempty"`
	SubscribeToNewsletter            *bool      `json:"subscribe_to_newsletter,omitempty" path:"subscribe_to_newsletter,omitempty" url:"subscribe_to_newsletter,omitempty"`
	ExternallyManaged                *bool      `json:"externally_managed,omitempty" path:"externally_managed,omitempty" url:"externally_managed,omitempty"`
	TimeZone                         string     `json:"time_zone,omitempty" path:"time_zone,omitempty" url:"time_zone,omitempty"`
	TypeOf2fa                        string     `json:"type_of_2fa,omitempty" path:"type_of_2fa,omitempty" url:"type_of_2fa,omitempty"`
	TypeOf2faForDisplay              string     `json:"type_of_2fa_for_display,omitempty" path:"type_of_2fa_for_display,omitempty" url:"type_of_2fa_for_display,omitempty"`
	UserRoot                         string     `json:"user_root,omitempty" path:"user_root,omitempty" url:"user_root,omitempty"`
	UserHome                         string     `json:"user_home,omitempty" path:"user_home,omitempty" url:"user_home,omitempty"`
	DaysRemainingUntilPasswordExpire int64      `json:"days_remaining_until_password_expire,omitempty" path:"days_remaining_until_password_expire,omitempty" url:"days_remaining_until_password_expire,omitempty"`
	PasswordExpireAt                 *time.Time `json:"password_expire_at,omitempty" path:"password_expire_at,omitempty" url:"password_expire_at,omitempty"`
	AvatarFile                       io.Reader  `json:"avatar_file,omitempty" path:"avatar_file,omitempty" url:"avatar_file,omitempty"`
	AvatarDelete                     *bool      `json:"avatar_delete,omitempty" path:"avatar_delete,omitempty" url:"avatar_delete,omitempty"`
	ChangePassword                   string     `json:"change_password,omitempty" path:"change_password,omitempty" url:"change_password,omitempty"`
	ChangePasswordConfirmation       string     `json:"change_password_confirmation,omitempty" path:"change_password_confirmation,omitempty" url:"change_password_confirmation,omitempty"`
	GrantPermission                  string     `json:"grant_permission,omitempty" path:"grant_permission,omitempty" url:"grant_permission,omitempty"`
	GroupId                          int64      `json:"group_id,omitempty" path:"group_id,omitempty" url:"group_id,omitempty"`
	ImportedPasswordHash             string     `json:"imported_password_hash,omitempty" path:"imported_password_hash,omitempty" url:"imported_password_hash,omitempty"`
	Password                         string     `json:"password,omitempty" path:"password,omitempty" url:"password,omitempty"`
	PasswordConfirmation             string     `json:"password_confirmation,omitempty" path:"password_confirmation,omitempty" url:"password_confirmation,omitempty"`
	AnnouncementsRead                *bool      `json:"announcements_read,omitempty" path:"announcements_read,omitempty" url:"announcements_read,omitempty"`
}

func (u User) Identifier() interface{} {
	return u.Id
}

type UserCollection []User

type UserAuthenticationMethodEnum string

func (u UserAuthenticationMethodEnum) String() string {
	return string(u)
}

func (u UserAuthenticationMethodEnum) Enum() map[string]UserAuthenticationMethodEnum {
	return map[string]UserAuthenticationMethodEnum{
		"password":                    UserAuthenticationMethodEnum("password"),
		"sso":                         UserAuthenticationMethodEnum("sso"),
		"none":                        UserAuthenticationMethodEnum("none"),
		"email_signup":                UserAuthenticationMethodEnum("email_signup"),
		"password_with_imported_hash": UserAuthenticationMethodEnum("password_with_imported_hash"),
		"password_and_ssh_key":        UserAuthenticationMethodEnum("password_and_ssh_key"),
	}
}

type UserSslRequiredEnum string

func (u UserSslRequiredEnum) String() string {
	return string(u)
}

func (u UserSslRequiredEnum) Enum() map[string]UserSslRequiredEnum {
	return map[string]UserSslRequiredEnum{
		"use_system_setting": UserSslRequiredEnum("use_system_setting"),
		"always_require":     UserSslRequiredEnum("always_require"),
		"never_require":      UserSslRequiredEnum("never_require"),
	}
}

type UserRequire2faEnum string

func (u UserRequire2faEnum) String() string {
	return string(u)
}

func (u UserRequire2faEnum) Enum() map[string]UserRequire2faEnum {
	return map[string]UserRequire2faEnum{
		"use_system_setting": UserRequire2faEnum("use_system_setting"),
		"always_require":     UserRequire2faEnum("always_require"),
		"never_require":      UserRequire2faEnum("never_require"),
	}
}

type UserListParams struct {
	SortBy                 map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter                 User                   `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt               map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq             map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix           map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt               map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq             map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	Ids                    string                 `url:"ids,omitempty" json:"ids,omitempty" path:"ids"`
	IncludeParentSiteUsers *bool                  `url:"include_parent_site_users,omitempty" json:"include_parent_site_users,omitempty" path:"include_parent_site_users"`
	Search                 string                 `url:"search,omitempty" json:"search,omitempty" path:"search"`
	ListParams
}

type UserFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type UserCreateParams struct {
	AvatarFile                 io.Writer                    `url:"avatar_file,omitempty" json:"avatar_file,omitempty" path:"avatar_file"`
	AvatarDelete               *bool                        `url:"avatar_delete,omitempty" json:"avatar_delete,omitempty" path:"avatar_delete"`
	ChangePassword             string                       `url:"change_password,omitempty" json:"change_password,omitempty" path:"change_password"`
	ChangePasswordConfirmation string                       `url:"change_password_confirmation,omitempty" json:"change_password_confirmation,omitempty" path:"change_password_confirmation"`
	Email                      string                       `url:"email,omitempty" json:"email,omitempty" path:"email"`
	GrantPermission            string                       `url:"grant_permission,omitempty" json:"grant_permission,omitempty" path:"grant_permission"`
	GroupId                    int64                        `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	GroupIds                   string                       `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	ImportedPasswordHash       string                       `url:"imported_password_hash,omitempty" json:"imported_password_hash,omitempty" path:"imported_password_hash"`
	Password                   string                       `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PasswordConfirmation       string                       `url:"password_confirmation,omitempty" json:"password_confirmation,omitempty" path:"password_confirmation"`
	AnnouncementsRead          *bool                        `url:"announcements_read,omitempty" json:"announcements_read,omitempty" path:"announcements_read"`
	AllowedIps                 string                       `url:"allowed_ips,omitempty" json:"allowed_ips,omitempty" path:"allowed_ips"`
	AttachmentsPermission      *bool                        `url:"attachments_permission,omitempty" json:"attachments_permission,omitempty" path:"attachments_permission"`
	AuthenticateUntil          *time.Time                   `url:"authenticate_until,omitempty" json:"authenticate_until,omitempty" path:"authenticate_until"`
	AuthenticationMethod       UserAuthenticationMethodEnum `url:"authentication_method,omitempty" json:"authentication_method,omitempty" path:"authentication_method"`
	BillingPermission          *bool                        `url:"billing_permission,omitempty" json:"billing_permission,omitempty" path:"billing_permission"`
	BypassUserLifecycleRules   *bool                        `url:"bypass_user_lifecycle_rules,omitempty" json:"bypass_user_lifecycle_rules,omitempty" path:"bypass_user_lifecycle_rules"`
	BypassSiteAllowedIps       *bool                        `url:"bypass_site_allowed_ips,omitempty" json:"bypass_site_allowed_ips,omitempty" path:"bypass_site_allowed_ips"`
	DavPermission              *bool                        `url:"dav_permission,omitempty" json:"dav_permission,omitempty" path:"dav_permission"`
	Disabled                   *bool                        `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	FtpPermission              *bool                        `url:"ftp_permission,omitempty" json:"ftp_permission,omitempty" path:"ftp_permission"`
	HeaderText                 string                       `url:"header_text,omitempty" json:"header_text,omitempty" path:"header_text"`
	Language                   string                       `url:"language,omitempty" json:"language,omitempty" path:"language"`
	NotificationDailySendTime  int64                        `url:"notification_daily_send_time,omitempty" json:"notification_daily_send_time,omitempty" path:"notification_daily_send_time"`
	Name                       string                       `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Company                    string                       `url:"company,omitempty" json:"company,omitempty" path:"company"`
	Notes                      string                       `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	OfficeIntegrationEnabled   *bool                        `url:"office_integration_enabled,omitempty" json:"office_integration_enabled,omitempty" path:"office_integration_enabled"`
	PasswordValidityDays       int64                        `url:"password_validity_days,omitempty" json:"password_validity_days,omitempty" path:"password_validity_days"`
	ReadonlySiteAdmin          *bool                        `url:"readonly_site_admin,omitempty" json:"readonly_site_admin,omitempty" path:"readonly_site_admin"`
	ReceiveAdminAlerts         *bool                        `url:"receive_admin_alerts,omitempty" json:"receive_admin_alerts,omitempty" path:"receive_admin_alerts"`
	RequireLoginBy             *time.Time                   `url:"require_login_by,omitempty" json:"require_login_by,omitempty" path:"require_login_by"`
	RequirePasswordChange      *bool                        `url:"require_password_change,omitempty" json:"require_password_change,omitempty" path:"require_password_change"`
	RestapiPermission          *bool                        `url:"restapi_permission,omitempty" json:"restapi_permission,omitempty" path:"restapi_permission"`
	SelfManaged                *bool                        `url:"self_managed,omitempty" json:"self_managed,omitempty" path:"self_managed"`
	SftpPermission             *bool                        `url:"sftp_permission,omitempty" json:"sftp_permission,omitempty" path:"sftp_permission"`
	SiteAdmin                  *bool                        `url:"site_admin,omitempty" json:"site_admin,omitempty" path:"site_admin"`
	SkipWelcomeScreen          *bool                        `url:"skip_welcome_screen,omitempty" json:"skip_welcome_screen,omitempty" path:"skip_welcome_screen"`
	SslRequired                UserSslRequiredEnum          `url:"ssl_required,omitempty" json:"ssl_required,omitempty" path:"ssl_required"`
	SsoStrategyId              int64                        `url:"sso_strategy_id,omitempty" json:"sso_strategy_id,omitempty" path:"sso_strategy_id"`
	SubscribeToNewsletter      *bool                        `url:"subscribe_to_newsletter,omitempty" json:"subscribe_to_newsletter,omitempty" path:"subscribe_to_newsletter"`
	Require2fa                 UserRequire2faEnum           `url:"require_2fa,omitempty" json:"require_2fa,omitempty" path:"require_2fa"`
	TimeZone                   string                       `url:"time_zone,omitempty" json:"time_zone,omitempty" path:"time_zone"`
	UserRoot                   string                       `url:"user_root,omitempty" json:"user_root,omitempty" path:"user_root"`
	UserHome                   string                       `url:"user_home,omitempty" json:"user_home,omitempty" path:"user_home"`
	Username                   string                       `url:"username" json:"username" path:"username"`
}

// Unlock user who has been locked out due to failed logins
type UserUnlockParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Resend user welcome email
type UserResendWelcomeEmailParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Trigger 2FA Reset process for user who has lost access to their existing 2FA methods
type UserUser2faResetParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type UserUpdateParams struct {
	Id                         int64                        `url:"-,omitempty" json:"-,omitempty" path:"id"`
	AvatarFile                 io.Writer                    `url:"avatar_file,omitempty" json:"avatar_file,omitempty" path:"avatar_file"`
	AvatarDelete               *bool                        `url:"avatar_delete,omitempty" json:"avatar_delete,omitempty" path:"avatar_delete"`
	ChangePassword             string                       `url:"change_password,omitempty" json:"change_password,omitempty" path:"change_password"`
	ChangePasswordConfirmation string                       `url:"change_password_confirmation,omitempty" json:"change_password_confirmation,omitempty" path:"change_password_confirmation"`
	Email                      string                       `url:"email,omitempty" json:"email,omitempty" path:"email"`
	GrantPermission            string                       `url:"grant_permission,omitempty" json:"grant_permission,omitempty" path:"grant_permission"`
	GroupId                    int64                        `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	GroupIds                   string                       `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	ImportedPasswordHash       string                       `url:"imported_password_hash,omitempty" json:"imported_password_hash,omitempty" path:"imported_password_hash"`
	Password                   string                       `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PasswordConfirmation       string                       `url:"password_confirmation,omitempty" json:"password_confirmation,omitempty" path:"password_confirmation"`
	AnnouncementsRead          *bool                        `url:"announcements_read,omitempty" json:"announcements_read,omitempty" path:"announcements_read"`
	AllowedIps                 string                       `url:"allowed_ips,omitempty" json:"allowed_ips,omitempty" path:"allowed_ips"`
	AttachmentsPermission      *bool                        `url:"attachments_permission,omitempty" json:"attachments_permission,omitempty" path:"attachments_permission"`
	AuthenticateUntil          *time.Time                   `url:"authenticate_until,omitempty" json:"authenticate_until,omitempty" path:"authenticate_until"`
	AuthenticationMethod       UserAuthenticationMethodEnum `url:"authentication_method,omitempty" json:"authentication_method,omitempty" path:"authentication_method"`
	BillingPermission          *bool                        `url:"billing_permission,omitempty" json:"billing_permission,omitempty" path:"billing_permission"`
	BypassUserLifecycleRules   *bool                        `url:"bypass_user_lifecycle_rules,omitempty" json:"bypass_user_lifecycle_rules,omitempty" path:"bypass_user_lifecycle_rules"`
	BypassSiteAllowedIps       *bool                        `url:"bypass_site_allowed_ips,omitempty" json:"bypass_site_allowed_ips,omitempty" path:"bypass_site_allowed_ips"`
	DavPermission              *bool                        `url:"dav_permission,omitempty" json:"dav_permission,omitempty" path:"dav_permission"`
	Disabled                   *bool                        `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	FtpPermission              *bool                        `url:"ftp_permission,omitempty" json:"ftp_permission,omitempty" path:"ftp_permission"`
	HeaderText                 string                       `url:"header_text,omitempty" json:"header_text,omitempty" path:"header_text"`
	Language                   string                       `url:"language,omitempty" json:"language,omitempty" path:"language"`
	NotificationDailySendTime  int64                        `url:"notification_daily_send_time,omitempty" json:"notification_daily_send_time,omitempty" path:"notification_daily_send_time"`
	Name                       string                       `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Company                    string                       `url:"company,omitempty" json:"company,omitempty" path:"company"`
	Notes                      string                       `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	OfficeIntegrationEnabled   *bool                        `url:"office_integration_enabled,omitempty" json:"office_integration_enabled,omitempty" path:"office_integration_enabled"`
	PasswordValidityDays       int64                        `url:"password_validity_days,omitempty" json:"password_validity_days,omitempty" path:"password_validity_days"`
	ReadonlySiteAdmin          *bool                        `url:"readonly_site_admin,omitempty" json:"readonly_site_admin,omitempty" path:"readonly_site_admin"`
	ReceiveAdminAlerts         *bool                        `url:"receive_admin_alerts,omitempty" json:"receive_admin_alerts,omitempty" path:"receive_admin_alerts"`
	RequireLoginBy             *time.Time                   `url:"require_login_by,omitempty" json:"require_login_by,omitempty" path:"require_login_by"`
	RequirePasswordChange      *bool                        `url:"require_password_change,omitempty" json:"require_password_change,omitempty" path:"require_password_change"`
	RestapiPermission          *bool                        `url:"restapi_permission,omitempty" json:"restapi_permission,omitempty" path:"restapi_permission"`
	SelfManaged                *bool                        `url:"self_managed,omitempty" json:"self_managed,omitempty" path:"self_managed"`
	SftpPermission             *bool                        `url:"sftp_permission,omitempty" json:"sftp_permission,omitempty" path:"sftp_permission"`
	SiteAdmin                  *bool                        `url:"site_admin,omitempty" json:"site_admin,omitempty" path:"site_admin"`
	SkipWelcomeScreen          *bool                        `url:"skip_welcome_screen,omitempty" json:"skip_welcome_screen,omitempty" path:"skip_welcome_screen"`
	SslRequired                UserSslRequiredEnum          `url:"ssl_required,omitempty" json:"ssl_required,omitempty" path:"ssl_required"`
	SsoStrategyId              int64                        `url:"sso_strategy_id,omitempty" json:"sso_strategy_id,omitempty" path:"sso_strategy_id"`
	SubscribeToNewsletter      *bool                        `url:"subscribe_to_newsletter,omitempty" json:"subscribe_to_newsletter,omitempty" path:"subscribe_to_newsletter"`
	Require2fa                 UserRequire2faEnum           `url:"require_2fa,omitempty" json:"require_2fa,omitempty" path:"require_2fa"`
	TimeZone                   string                       `url:"time_zone,omitempty" json:"time_zone,omitempty" path:"time_zone"`
	UserRoot                   string                       `url:"user_root,omitempty" json:"user_root,omitempty" path:"user_root"`
	UserHome                   string                       `url:"user_home,omitempty" json:"user_home,omitempty" path:"user_home"`
	Username                   string                       `url:"username,omitempty" json:"username,omitempty" path:"username"`
}

type UserDeleteParams struct {
	Id         int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	NewOwnerId int64 `url:"new_owner_id,omitempty" json:"new_owner_id,omitempty" path:"new_owner_id"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type user User
	var v user
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = User(v)
	return nil
}

func (u *UserCollection) UnmarshalJSON(data []byte) error {
	type users UserCollection
	var v users
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UserCollection(v)
	return nil
}

func (u *UserCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
