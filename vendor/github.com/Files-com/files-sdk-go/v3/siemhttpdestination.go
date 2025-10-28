package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SiemHttpDestination struct {
	Id                                            int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                                          string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	DestinationType                               string                 `json:"destination_type,omitempty" path:"destination_type,omitempty" url:"destination_type,omitempty"`
	DestinationUrl                                string                 `json:"destination_url,omitempty" path:"destination_url,omitempty" url:"destination_url,omitempty"`
	AdditionalHeaders                             map[string]interface{} `json:"additional_headers,omitempty" path:"additional_headers,omitempty" url:"additional_headers,omitempty"`
	SendingActive                                 *bool                  `json:"sending_active,omitempty" path:"sending_active,omitempty" url:"sending_active,omitempty"`
	GenericPayloadType                            string                 `json:"generic_payload_type,omitempty" path:"generic_payload_type,omitempty" url:"generic_payload_type,omitempty"`
	SplunkTokenMasked                             string                 `json:"splunk_token_masked,omitempty" path:"splunk_token_masked,omitempty" url:"splunk_token_masked,omitempty"`
	AzureDcrImmutableId                           string                 `json:"azure_dcr_immutable_id,omitempty" path:"azure_dcr_immutable_id,omitempty" url:"azure_dcr_immutable_id,omitempty"`
	AzureStreamName                               string                 `json:"azure_stream_name,omitempty" path:"azure_stream_name,omitempty" url:"azure_stream_name,omitempty"`
	AzureOauthClientCredentialsTenantId           string                 `json:"azure_oauth_client_credentials_tenant_id,omitempty" path:"azure_oauth_client_credentials_tenant_id,omitempty" url:"azure_oauth_client_credentials_tenant_id,omitempty"`
	AzureOauthClientCredentialsClientId           string                 `json:"azure_oauth_client_credentials_client_id,omitempty" path:"azure_oauth_client_credentials_client_id,omitempty" url:"azure_oauth_client_credentials_client_id,omitempty"`
	AzureOauthClientCredentialsClientSecretMasked string                 `json:"azure_oauth_client_credentials_client_secret_masked,omitempty" path:"azure_oauth_client_credentials_client_secret_masked,omitempty" url:"azure_oauth_client_credentials_client_secret_masked,omitempty"`
	QradarUsername                                string                 `json:"qradar_username,omitempty" path:"qradar_username,omitempty" url:"qradar_username,omitempty"`
	QradarPasswordMasked                          string                 `json:"qradar_password_masked,omitempty" path:"qradar_password_masked,omitempty" url:"qradar_password_masked,omitempty"`
	SolarWindsTokenMasked                         string                 `json:"solar_winds_token_masked,omitempty" path:"solar_winds_token_masked,omitempty" url:"solar_winds_token_masked,omitempty"`
	NewRelicApiKeyMasked                          string                 `json:"new_relic_api_key_masked,omitempty" path:"new_relic_api_key_masked,omitempty" url:"new_relic_api_key_masked,omitempty"`
	DatadogApiKeyMasked                           string                 `json:"datadog_api_key_masked,omitempty" path:"datadog_api_key_masked,omitempty" url:"datadog_api_key_masked,omitempty"`
	SftpActionSendEnabled                         *bool                  `json:"sftp_action_send_enabled,omitempty" path:"sftp_action_send_enabled,omitempty" url:"sftp_action_send_enabled,omitempty"`
	SftpActionEntriesSent                         int64                  `json:"sftp_action_entries_sent,omitempty" path:"sftp_action_entries_sent,omitempty" url:"sftp_action_entries_sent,omitempty"`
	FtpActionSendEnabled                          *bool                  `json:"ftp_action_send_enabled,omitempty" path:"ftp_action_send_enabled,omitempty" url:"ftp_action_send_enabled,omitempty"`
	FtpActionEntriesSent                          int64                  `json:"ftp_action_entries_sent,omitempty" path:"ftp_action_entries_sent,omitempty" url:"ftp_action_entries_sent,omitempty"`
	WebDavActionSendEnabled                       *bool                  `json:"web_dav_action_send_enabled,omitempty" path:"web_dav_action_send_enabled,omitempty" url:"web_dav_action_send_enabled,omitempty"`
	WebDavActionEntriesSent                       int64                  `json:"web_dav_action_entries_sent,omitempty" path:"web_dav_action_entries_sent,omitempty" url:"web_dav_action_entries_sent,omitempty"`
	SyncSendEnabled                               *bool                  `json:"sync_send_enabled,omitempty" path:"sync_send_enabled,omitempty" url:"sync_send_enabled,omitempty"`
	SyncEntriesSent                               int64                  `json:"sync_entries_sent,omitempty" path:"sync_entries_sent,omitempty" url:"sync_entries_sent,omitempty"`
	OutboundConnectionSendEnabled                 *bool                  `json:"outbound_connection_send_enabled,omitempty" path:"outbound_connection_send_enabled,omitempty" url:"outbound_connection_send_enabled,omitempty"`
	OutboundConnectionEntriesSent                 int64                  `json:"outbound_connection_entries_sent,omitempty" path:"outbound_connection_entries_sent,omitempty" url:"outbound_connection_entries_sent,omitempty"`
	AutomationSendEnabled                         *bool                  `json:"automation_send_enabled,omitempty" path:"automation_send_enabled,omitempty" url:"automation_send_enabled,omitempty"`
	AutomationEntriesSent                         int64                  `json:"automation_entries_sent,omitempty" path:"automation_entries_sent,omitempty" url:"automation_entries_sent,omitempty"`
	ApiRequestSendEnabled                         *bool                  `json:"api_request_send_enabled,omitempty" path:"api_request_send_enabled,omitempty" url:"api_request_send_enabled,omitempty"`
	ApiRequestEntriesSent                         int64                  `json:"api_request_entries_sent,omitempty" path:"api_request_entries_sent,omitempty" url:"api_request_entries_sent,omitempty"`
	PublicHostingRequestSendEnabled               *bool                  `json:"public_hosting_request_send_enabled,omitempty" path:"public_hosting_request_send_enabled,omitempty" url:"public_hosting_request_send_enabled,omitempty"`
	PublicHostingRequestEntriesSent               int64                  `json:"public_hosting_request_entries_sent,omitempty" path:"public_hosting_request_entries_sent,omitempty" url:"public_hosting_request_entries_sent,omitempty"`
	EmailSendEnabled                              *bool                  `json:"email_send_enabled,omitempty" path:"email_send_enabled,omitempty" url:"email_send_enabled,omitempty"`
	EmailEntriesSent                              int64                  `json:"email_entries_sent,omitempty" path:"email_entries_sent,omitempty" url:"email_entries_sent,omitempty"`
	ExavaultApiRequestSendEnabled                 *bool                  `json:"exavault_api_request_send_enabled,omitempty" path:"exavault_api_request_send_enabled,omitempty" url:"exavault_api_request_send_enabled,omitempty"`
	ExavaultApiRequestEntriesSent                 int64                  `json:"exavault_api_request_entries_sent,omitempty" path:"exavault_api_request_entries_sent,omitempty" url:"exavault_api_request_entries_sent,omitempty"`
	SettingsChangeSendEnabled                     *bool                  `json:"settings_change_send_enabled,omitempty" path:"settings_change_send_enabled,omitempty" url:"settings_change_send_enabled,omitempty"`
	SettingsChangeEntriesSent                     int64                  `json:"settings_change_entries_sent,omitempty" path:"settings_change_entries_sent,omitempty" url:"settings_change_entries_sent,omitempty"`
	LastHttpCallTargetType                        string                 `json:"last_http_call_target_type,omitempty" path:"last_http_call_target_type,omitempty" url:"last_http_call_target_type,omitempty"`
	LastHttpCallSuccess                           *bool                  `json:"last_http_call_success,omitempty" path:"last_http_call_success,omitempty" url:"last_http_call_success,omitempty"`
	LastHttpCallResponseCode                      int64                  `json:"last_http_call_response_code,omitempty" path:"last_http_call_response_code,omitempty" url:"last_http_call_response_code,omitempty"`
	LastHttpCallResponseBody                      string                 `json:"last_http_call_response_body,omitempty" path:"last_http_call_response_body,omitempty" url:"last_http_call_response_body,omitempty"`
	LastHttpCallErrorMessage                      string                 `json:"last_http_call_error_message,omitempty" path:"last_http_call_error_message,omitempty" url:"last_http_call_error_message,omitempty"`
	LastHttpCallTime                              string                 `json:"last_http_call_time,omitempty" path:"last_http_call_time,omitempty" url:"last_http_call_time,omitempty"`
	LastHttpCallDurationMs                        int64                  `json:"last_http_call_duration_ms,omitempty" path:"last_http_call_duration_ms,omitempty" url:"last_http_call_duration_ms,omitempty"`
	MostRecentHttpCallSuccessTime                 string                 `json:"most_recent_http_call_success_time,omitempty" path:"most_recent_http_call_success_time,omitempty" url:"most_recent_http_call_success_time,omitempty"`
	ConnectionTestEntry                           string                 `json:"connection_test_entry,omitempty" path:"connection_test_entry,omitempty" url:"connection_test_entry,omitempty"`
	SplunkToken                                   string                 `json:"splunk_token,omitempty" path:"splunk_token,omitempty" url:"splunk_token,omitempty"`
	AzureOauthClientCredentialsClientSecret       string                 `json:"azure_oauth_client_credentials_client_secret,omitempty" path:"azure_oauth_client_credentials_client_secret,omitempty" url:"azure_oauth_client_credentials_client_secret,omitempty"`
	QradarPassword                                string                 `json:"qradar_password,omitempty" path:"qradar_password,omitempty" url:"qradar_password,omitempty"`
	SolarWindsToken                               string                 `json:"solar_winds_token,omitempty" path:"solar_winds_token,omitempty" url:"solar_winds_token,omitempty"`
	NewRelicApiKey                                string                 `json:"new_relic_api_key,omitempty" path:"new_relic_api_key,omitempty" url:"new_relic_api_key,omitempty"`
	DatadogApiKey                                 string                 `json:"datadog_api_key,omitempty" path:"datadog_api_key,omitempty" url:"datadog_api_key,omitempty"`
}

func (s SiemHttpDestination) Identifier() interface{} {
	return s.Id
}

type SiemHttpDestinationCollection []SiemHttpDestination

type SiemHttpDestinationGenericPayloadTypeEnum string

func (u SiemHttpDestinationGenericPayloadTypeEnum) String() string {
	return string(u)
}

func (u SiemHttpDestinationGenericPayloadTypeEnum) Enum() map[string]SiemHttpDestinationGenericPayloadTypeEnum {
	return map[string]SiemHttpDestinationGenericPayloadTypeEnum{
		"json_newline": SiemHttpDestinationGenericPayloadTypeEnum("json_newline"),
		"json_array":   SiemHttpDestinationGenericPayloadTypeEnum("json_array"),
	}
}

type SiemHttpDestinationDestinationTypeEnum string

func (u SiemHttpDestinationDestinationTypeEnum) String() string {
	return string(u)
}

func (u SiemHttpDestinationDestinationTypeEnum) Enum() map[string]SiemHttpDestinationDestinationTypeEnum {
	return map[string]SiemHttpDestinationDestinationTypeEnum{
		"generic":      SiemHttpDestinationDestinationTypeEnum("generic"),
		"splunk":       SiemHttpDestinationDestinationTypeEnum("splunk"),
		"azure_legacy": SiemHttpDestinationDestinationTypeEnum("azure_legacy"),
		"qradar":       SiemHttpDestinationDestinationTypeEnum("qradar"),
		"sumo":         SiemHttpDestinationDestinationTypeEnum("sumo"),
		"rapid7":       SiemHttpDestinationDestinationTypeEnum("rapid7"),
		"solar_winds":  SiemHttpDestinationDestinationTypeEnum("solar_winds"),
		"new_relic":    SiemHttpDestinationDestinationTypeEnum("new_relic"),
		"datadog":      SiemHttpDestinationDestinationTypeEnum("datadog"),
		"azure":        SiemHttpDestinationDestinationTypeEnum("azure"),
	}
}

type SiemHttpDestinationListParams struct {
	ListParams
}

type SiemHttpDestinationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SiemHttpDestinationCreateParams struct {
	Name                                    string                                    `url:"name,omitempty" json:"name,omitempty" path:"name"`
	AdditionalHeaders                       map[string]interface{}                    `url:"additional_headers,omitempty" json:"additional_headers,omitempty" path:"additional_headers"`
	SendingActive                           *bool                                     `url:"sending_active,omitempty" json:"sending_active,omitempty" path:"sending_active"`
	GenericPayloadType                      SiemHttpDestinationGenericPayloadTypeEnum `url:"generic_payload_type,omitempty" json:"generic_payload_type,omitempty" path:"generic_payload_type"`
	SplunkToken                             string                                    `url:"splunk_token,omitempty" json:"splunk_token,omitempty" path:"splunk_token"`
	AzureDcrImmutableId                     string                                    `url:"azure_dcr_immutable_id,omitempty" json:"azure_dcr_immutable_id,omitempty" path:"azure_dcr_immutable_id"`
	AzureStreamName                         string                                    `url:"azure_stream_name,omitempty" json:"azure_stream_name,omitempty" path:"azure_stream_name"`
	AzureOauthClientCredentialsTenantId     string                                    `url:"azure_oauth_client_credentials_tenant_id,omitempty" json:"azure_oauth_client_credentials_tenant_id,omitempty" path:"azure_oauth_client_credentials_tenant_id"`
	AzureOauthClientCredentialsClientId     string                                    `url:"azure_oauth_client_credentials_client_id,omitempty" json:"azure_oauth_client_credentials_client_id,omitempty" path:"azure_oauth_client_credentials_client_id"`
	AzureOauthClientCredentialsClientSecret string                                    `url:"azure_oauth_client_credentials_client_secret,omitempty" json:"azure_oauth_client_credentials_client_secret,omitempty" path:"azure_oauth_client_credentials_client_secret"`
	QradarUsername                          string                                    `url:"qradar_username,omitempty" json:"qradar_username,omitempty" path:"qradar_username"`
	QradarPassword                          string                                    `url:"qradar_password,omitempty" json:"qradar_password,omitempty" path:"qradar_password"`
	SolarWindsToken                         string                                    `url:"solar_winds_token,omitempty" json:"solar_winds_token,omitempty" path:"solar_winds_token"`
	NewRelicApiKey                          string                                    `url:"new_relic_api_key,omitempty" json:"new_relic_api_key,omitempty" path:"new_relic_api_key"`
	DatadogApiKey                           string                                    `url:"datadog_api_key,omitempty" json:"datadog_api_key,omitempty" path:"datadog_api_key"`
	SftpActionSendEnabled                   *bool                                     `url:"sftp_action_send_enabled,omitempty" json:"sftp_action_send_enabled,omitempty" path:"sftp_action_send_enabled"`
	FtpActionSendEnabled                    *bool                                     `url:"ftp_action_send_enabled,omitempty" json:"ftp_action_send_enabled,omitempty" path:"ftp_action_send_enabled"`
	WebDavActionSendEnabled                 *bool                                     `url:"web_dav_action_send_enabled,omitempty" json:"web_dav_action_send_enabled,omitempty" path:"web_dav_action_send_enabled"`
	SyncSendEnabled                         *bool                                     `url:"sync_send_enabled,omitempty" json:"sync_send_enabled,omitempty" path:"sync_send_enabled"`
	OutboundConnectionSendEnabled           *bool                                     `url:"outbound_connection_send_enabled,omitempty" json:"outbound_connection_send_enabled,omitempty" path:"outbound_connection_send_enabled"`
	AutomationSendEnabled                   *bool                                     `url:"automation_send_enabled,omitempty" json:"automation_send_enabled,omitempty" path:"automation_send_enabled"`
	ApiRequestSendEnabled                   *bool                                     `url:"api_request_send_enabled,omitempty" json:"api_request_send_enabled,omitempty" path:"api_request_send_enabled"`
	PublicHostingRequestSendEnabled         *bool                                     `url:"public_hosting_request_send_enabled,omitempty" json:"public_hosting_request_send_enabled,omitempty" path:"public_hosting_request_send_enabled"`
	EmailSendEnabled                        *bool                                     `url:"email_send_enabled,omitempty" json:"email_send_enabled,omitempty" path:"email_send_enabled"`
	ExavaultApiRequestSendEnabled           *bool                                     `url:"exavault_api_request_send_enabled,omitempty" json:"exavault_api_request_send_enabled,omitempty" path:"exavault_api_request_send_enabled"`
	SettingsChangeSendEnabled               *bool                                     `url:"settings_change_send_enabled,omitempty" json:"settings_change_send_enabled,omitempty" path:"settings_change_send_enabled"`
	DestinationType                         SiemHttpDestinationDestinationTypeEnum    `url:"destination_type" json:"destination_type" path:"destination_type"`
	DestinationUrl                          string                                    `url:"destination_url" json:"destination_url" path:"destination_url"`
}

type SiemHttpDestinationSendTestEntryParams struct {
	SiemHttpDestinationId                   int64                                     `url:"siem_http_destination_id,omitempty" json:"siem_http_destination_id,omitempty" path:"siem_http_destination_id"`
	DestinationType                         SiemHttpDestinationDestinationTypeEnum    `url:"destination_type,omitempty" json:"destination_type,omitempty" path:"destination_type"`
	DestinationUrl                          string                                    `url:"destination_url,omitempty" json:"destination_url,omitempty" path:"destination_url"`
	Name                                    string                                    `url:"name,omitempty" json:"name,omitempty" path:"name"`
	AdditionalHeaders                       map[string]interface{}                    `url:"additional_headers,omitempty" json:"additional_headers,omitempty" path:"additional_headers"`
	SendingActive                           *bool                                     `url:"sending_active,omitempty" json:"sending_active,omitempty" path:"sending_active"`
	GenericPayloadType                      SiemHttpDestinationGenericPayloadTypeEnum `url:"generic_payload_type,omitempty" json:"generic_payload_type,omitempty" path:"generic_payload_type"`
	SplunkToken                             string                                    `url:"splunk_token,omitempty" json:"splunk_token,omitempty" path:"splunk_token"`
	AzureDcrImmutableId                     string                                    `url:"azure_dcr_immutable_id,omitempty" json:"azure_dcr_immutable_id,omitempty" path:"azure_dcr_immutable_id"`
	AzureStreamName                         string                                    `url:"azure_stream_name,omitempty" json:"azure_stream_name,omitempty" path:"azure_stream_name"`
	AzureOauthClientCredentialsTenantId     string                                    `url:"azure_oauth_client_credentials_tenant_id,omitempty" json:"azure_oauth_client_credentials_tenant_id,omitempty" path:"azure_oauth_client_credentials_tenant_id"`
	AzureOauthClientCredentialsClientId     string                                    `url:"azure_oauth_client_credentials_client_id,omitempty" json:"azure_oauth_client_credentials_client_id,omitempty" path:"azure_oauth_client_credentials_client_id"`
	AzureOauthClientCredentialsClientSecret string                                    `url:"azure_oauth_client_credentials_client_secret,omitempty" json:"azure_oauth_client_credentials_client_secret,omitempty" path:"azure_oauth_client_credentials_client_secret"`
	QradarUsername                          string                                    `url:"qradar_username,omitempty" json:"qradar_username,omitempty" path:"qradar_username"`
	QradarPassword                          string                                    `url:"qradar_password,omitempty" json:"qradar_password,omitempty" path:"qradar_password"`
	SolarWindsToken                         string                                    `url:"solar_winds_token,omitempty" json:"solar_winds_token,omitempty" path:"solar_winds_token"`
	NewRelicApiKey                          string                                    `url:"new_relic_api_key,omitempty" json:"new_relic_api_key,omitempty" path:"new_relic_api_key"`
	DatadogApiKey                           string                                    `url:"datadog_api_key,omitempty" json:"datadog_api_key,omitempty" path:"datadog_api_key"`
	SftpActionSendEnabled                   *bool                                     `url:"sftp_action_send_enabled,omitempty" json:"sftp_action_send_enabled,omitempty" path:"sftp_action_send_enabled"`
	FtpActionSendEnabled                    *bool                                     `url:"ftp_action_send_enabled,omitempty" json:"ftp_action_send_enabled,omitempty" path:"ftp_action_send_enabled"`
	WebDavActionSendEnabled                 *bool                                     `url:"web_dav_action_send_enabled,omitempty" json:"web_dav_action_send_enabled,omitempty" path:"web_dav_action_send_enabled"`
	SyncSendEnabled                         *bool                                     `url:"sync_send_enabled,omitempty" json:"sync_send_enabled,omitempty" path:"sync_send_enabled"`
	OutboundConnectionSendEnabled           *bool                                     `url:"outbound_connection_send_enabled,omitempty" json:"outbound_connection_send_enabled,omitempty" path:"outbound_connection_send_enabled"`
	AutomationSendEnabled                   *bool                                     `url:"automation_send_enabled,omitempty" json:"automation_send_enabled,omitempty" path:"automation_send_enabled"`
	ApiRequestSendEnabled                   *bool                                     `url:"api_request_send_enabled,omitempty" json:"api_request_send_enabled,omitempty" path:"api_request_send_enabled"`
	PublicHostingRequestSendEnabled         *bool                                     `url:"public_hosting_request_send_enabled,omitempty" json:"public_hosting_request_send_enabled,omitempty" path:"public_hosting_request_send_enabled"`
	EmailSendEnabled                        *bool                                     `url:"email_send_enabled,omitempty" json:"email_send_enabled,omitempty" path:"email_send_enabled"`
	ExavaultApiRequestSendEnabled           *bool                                     `url:"exavault_api_request_send_enabled,omitempty" json:"exavault_api_request_send_enabled,omitempty" path:"exavault_api_request_send_enabled"`
	SettingsChangeSendEnabled               *bool                                     `url:"settings_change_send_enabled,omitempty" json:"settings_change_send_enabled,omitempty" path:"settings_change_send_enabled"`
}

type SiemHttpDestinationUpdateParams struct {
	Id                                      int64                                     `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name                                    string                                    `url:"name,omitempty" json:"name,omitempty" path:"name"`
	AdditionalHeaders                       map[string]interface{}                    `url:"additional_headers,omitempty" json:"additional_headers,omitempty" path:"additional_headers"`
	SendingActive                           *bool                                     `url:"sending_active,omitempty" json:"sending_active,omitempty" path:"sending_active"`
	GenericPayloadType                      SiemHttpDestinationGenericPayloadTypeEnum `url:"generic_payload_type,omitempty" json:"generic_payload_type,omitempty" path:"generic_payload_type"`
	SplunkToken                             string                                    `url:"splunk_token,omitempty" json:"splunk_token,omitempty" path:"splunk_token"`
	AzureDcrImmutableId                     string                                    `url:"azure_dcr_immutable_id,omitempty" json:"azure_dcr_immutable_id,omitempty" path:"azure_dcr_immutable_id"`
	AzureStreamName                         string                                    `url:"azure_stream_name,omitempty" json:"azure_stream_name,omitempty" path:"azure_stream_name"`
	AzureOauthClientCredentialsTenantId     string                                    `url:"azure_oauth_client_credentials_tenant_id,omitempty" json:"azure_oauth_client_credentials_tenant_id,omitempty" path:"azure_oauth_client_credentials_tenant_id"`
	AzureOauthClientCredentialsClientId     string                                    `url:"azure_oauth_client_credentials_client_id,omitempty" json:"azure_oauth_client_credentials_client_id,omitempty" path:"azure_oauth_client_credentials_client_id"`
	AzureOauthClientCredentialsClientSecret string                                    `url:"azure_oauth_client_credentials_client_secret,omitempty" json:"azure_oauth_client_credentials_client_secret,omitempty" path:"azure_oauth_client_credentials_client_secret"`
	QradarUsername                          string                                    `url:"qradar_username,omitempty" json:"qradar_username,omitempty" path:"qradar_username"`
	QradarPassword                          string                                    `url:"qradar_password,omitempty" json:"qradar_password,omitempty" path:"qradar_password"`
	SolarWindsToken                         string                                    `url:"solar_winds_token,omitempty" json:"solar_winds_token,omitempty" path:"solar_winds_token"`
	NewRelicApiKey                          string                                    `url:"new_relic_api_key,omitempty" json:"new_relic_api_key,omitempty" path:"new_relic_api_key"`
	DatadogApiKey                           string                                    `url:"datadog_api_key,omitempty" json:"datadog_api_key,omitempty" path:"datadog_api_key"`
	SftpActionSendEnabled                   *bool                                     `url:"sftp_action_send_enabled,omitempty" json:"sftp_action_send_enabled,omitempty" path:"sftp_action_send_enabled"`
	FtpActionSendEnabled                    *bool                                     `url:"ftp_action_send_enabled,omitempty" json:"ftp_action_send_enabled,omitempty" path:"ftp_action_send_enabled"`
	WebDavActionSendEnabled                 *bool                                     `url:"web_dav_action_send_enabled,omitempty" json:"web_dav_action_send_enabled,omitempty" path:"web_dav_action_send_enabled"`
	SyncSendEnabled                         *bool                                     `url:"sync_send_enabled,omitempty" json:"sync_send_enabled,omitempty" path:"sync_send_enabled"`
	OutboundConnectionSendEnabled           *bool                                     `url:"outbound_connection_send_enabled,omitempty" json:"outbound_connection_send_enabled,omitempty" path:"outbound_connection_send_enabled"`
	AutomationSendEnabled                   *bool                                     `url:"automation_send_enabled,omitempty" json:"automation_send_enabled,omitempty" path:"automation_send_enabled"`
	ApiRequestSendEnabled                   *bool                                     `url:"api_request_send_enabled,omitempty" json:"api_request_send_enabled,omitempty" path:"api_request_send_enabled"`
	PublicHostingRequestSendEnabled         *bool                                     `url:"public_hosting_request_send_enabled,omitempty" json:"public_hosting_request_send_enabled,omitempty" path:"public_hosting_request_send_enabled"`
	EmailSendEnabled                        *bool                                     `url:"email_send_enabled,omitempty" json:"email_send_enabled,omitempty" path:"email_send_enabled"`
	ExavaultApiRequestSendEnabled           *bool                                     `url:"exavault_api_request_send_enabled,omitempty" json:"exavault_api_request_send_enabled,omitempty" path:"exavault_api_request_send_enabled"`
	SettingsChangeSendEnabled               *bool                                     `url:"settings_change_send_enabled,omitempty" json:"settings_change_send_enabled,omitempty" path:"settings_change_send_enabled"`
	DestinationType                         SiemHttpDestinationDestinationTypeEnum    `url:"destination_type,omitempty" json:"destination_type,omitempty" path:"destination_type"`
	DestinationUrl                          string                                    `url:"destination_url,omitempty" json:"destination_url,omitempty" path:"destination_url"`
}

type SiemHttpDestinationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *SiemHttpDestination) UnmarshalJSON(data []byte) error {
	type siemHttpDestination SiemHttpDestination
	var v siemHttpDestination
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SiemHttpDestination(v)
	return nil
}

func (s *SiemHttpDestinationCollection) UnmarshalJSON(data []byte) error {
	type siemHttpDestinations SiemHttpDestinationCollection
	var v siemHttpDestinations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SiemHttpDestinationCollection(v)
	return nil
}

func (s *SiemHttpDestinationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
