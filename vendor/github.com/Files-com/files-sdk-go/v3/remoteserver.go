package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type RemoteServer struct {
	Id                                      int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Disabled                                *bool  `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	AuthenticationMethod                    string `json:"authentication_method,omitempty" path:"authentication_method,omitempty" url:"authentication_method,omitempty"`
	Hostname                                string `json:"hostname,omitempty" path:"hostname,omitempty" url:"hostname,omitempty"`
	RemoteHomePath                          string `json:"remote_home_path,omitempty" path:"remote_home_path,omitempty" url:"remote_home_path,omitempty"`
	Name                                    string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Port                                    int64  `json:"port,omitempty" path:"port,omitempty" url:"port,omitempty"`
	BufferUploadsAlways                     *bool  `json:"buffer_uploads_always,omitempty" path:"buffer_uploads_always,omitempty" url:"buffer_uploads_always,omitempty"`
	MaxConnections                          int64  `json:"max_connections,omitempty" path:"max_connections,omitempty" url:"max_connections,omitempty"`
	PinToSiteRegion                         *bool  `json:"pin_to_site_region,omitempty" path:"pin_to_site_region,omitempty" url:"pin_to_site_region,omitempty"`
	PinnedRegion                            string `json:"pinned_region,omitempty" path:"pinned_region,omitempty" url:"pinned_region,omitempty"`
	S3Bucket                                string `json:"s3_bucket,omitempty" path:"s3_bucket,omitempty" url:"s3_bucket,omitempty"`
	S3Region                                string `json:"s3_region,omitempty" path:"s3_region,omitempty" url:"s3_region,omitempty"`
	AwsAccessKey                            string `json:"aws_access_key,omitempty" path:"aws_access_key,omitempty" url:"aws_access_key,omitempty"`
	ServerCertificate                       string `json:"server_certificate,omitempty" path:"server_certificate,omitempty" url:"server_certificate,omitempty"`
	ServerHostKey                           string `json:"server_host_key,omitempty" path:"server_host_key,omitempty" url:"server_host_key,omitempty"`
	ServerType                              string `json:"server_type,omitempty" path:"server_type,omitempty" url:"server_type,omitempty"`
	Ssl                                     string `json:"ssl,omitempty" path:"ssl,omitempty" url:"ssl,omitempty"`
	Username                                string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	GoogleCloudStorageBucket                string `json:"google_cloud_storage_bucket,omitempty" path:"google_cloud_storage_bucket,omitempty" url:"google_cloud_storage_bucket,omitempty"`
	GoogleCloudStorageProjectId             string `json:"google_cloud_storage_project_id,omitempty" path:"google_cloud_storage_project_id,omitempty" url:"google_cloud_storage_project_id,omitempty"`
	GoogleCloudStorageS3CompatibleAccessKey string `json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key,omitempty" url:"google_cloud_storage_s3_compatible_access_key,omitempty"`
	BackblazeB2S3Endpoint                   string `json:"backblaze_b2_s3_endpoint,omitempty" path:"backblaze_b2_s3_endpoint,omitempty" url:"backblaze_b2_s3_endpoint,omitempty"`
	BackblazeB2Bucket                       string `json:"backblaze_b2_bucket,omitempty" path:"backblaze_b2_bucket,omitempty" url:"backblaze_b2_bucket,omitempty"`
	WasabiBucket                            string `json:"wasabi_bucket,omitempty" path:"wasabi_bucket,omitempty" url:"wasabi_bucket,omitempty"`
	WasabiRegion                            string `json:"wasabi_region,omitempty" path:"wasabi_region,omitempty" url:"wasabi_region,omitempty"`
	WasabiAccessKey                         string `json:"wasabi_access_key,omitempty" path:"wasabi_access_key,omitempty" url:"wasabi_access_key,omitempty"`
	AuthStatus                              string `json:"auth_status,omitempty" path:"auth_status,omitempty" url:"auth_status,omitempty"`
	AuthAccountName                         string `json:"auth_account_name,omitempty" path:"auth_account_name,omitempty" url:"auth_account_name,omitempty"`
	OneDriveAccountType                     string `json:"one_drive_account_type,omitempty" path:"one_drive_account_type,omitempty" url:"one_drive_account_type,omitempty"`
	AzureBlobStorageAccount                 string `json:"azure_blob_storage_account,omitempty" path:"azure_blob_storage_account,omitempty" url:"azure_blob_storage_account,omitempty"`
	AzureBlobStorageContainer               string `json:"azure_blob_storage_container,omitempty" path:"azure_blob_storage_container,omitempty" url:"azure_blob_storage_container,omitempty"`
	AzureBlobStorageHierarchicalNamespace   *bool  `json:"azure_blob_storage_hierarchical_namespace,omitempty" path:"azure_blob_storage_hierarchical_namespace,omitempty" url:"azure_blob_storage_hierarchical_namespace,omitempty"`
	AzureBlobStorageDnsSuffix               string `json:"azure_blob_storage_dns_suffix,omitempty" path:"azure_blob_storage_dns_suffix,omitempty" url:"azure_blob_storage_dns_suffix,omitempty"`
	AzureFilesStorageAccount                string `json:"azure_files_storage_account,omitempty" path:"azure_files_storage_account,omitempty" url:"azure_files_storage_account,omitempty"`
	AzureFilesStorageShareName              string `json:"azure_files_storage_share_name,omitempty" path:"azure_files_storage_share_name,omitempty" url:"azure_files_storage_share_name,omitempty"`
	AzureFilesStorageDnsSuffix              string `json:"azure_files_storage_dns_suffix,omitempty" path:"azure_files_storage_dns_suffix,omitempty" url:"azure_files_storage_dns_suffix,omitempty"`
	S3CompatibleBucket                      string `json:"s3_compatible_bucket,omitempty" path:"s3_compatible_bucket,omitempty" url:"s3_compatible_bucket,omitempty"`
	S3CompatibleEndpoint                    string `json:"s3_compatible_endpoint,omitempty" path:"s3_compatible_endpoint,omitempty" url:"s3_compatible_endpoint,omitempty"`
	S3CompatibleRegion                      string `json:"s3_compatible_region,omitempty" path:"s3_compatible_region,omitempty" url:"s3_compatible_region,omitempty"`
	S3CompatibleAccessKey                   string `json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key,omitempty" url:"s3_compatible_access_key,omitempty"`
	EnableDedicatedIps                      *bool  `json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips,omitempty" url:"enable_dedicated_ips,omitempty"`
	FilesAgentPermissionSet                 string `json:"files_agent_permission_set,omitempty" path:"files_agent_permission_set,omitempty" url:"files_agent_permission_set,omitempty"`
	FilesAgentRoot                          string `json:"files_agent_root,omitempty" path:"files_agent_root,omitempty" url:"files_agent_root,omitempty"`
	FilesAgentApiToken                      string `json:"files_agent_api_token,omitempty" path:"files_agent_api_token,omitempty" url:"files_agent_api_token,omitempty"`
	FilesAgentVersion                       string `json:"files_agent_version,omitempty" path:"files_agent_version,omitempty" url:"files_agent_version,omitempty"`
	FilebaseBucket                          string `json:"filebase_bucket,omitempty" path:"filebase_bucket,omitempty" url:"filebase_bucket,omitempty"`
	FilebaseAccessKey                       string `json:"filebase_access_key,omitempty" path:"filebase_access_key,omitempty" url:"filebase_access_key,omitempty"`
	CloudflareBucket                        string `json:"cloudflare_bucket,omitempty" path:"cloudflare_bucket,omitempty" url:"cloudflare_bucket,omitempty"`
	CloudflareAccessKey                     string `json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key,omitempty" url:"cloudflare_access_key,omitempty"`
	CloudflareEndpoint                      string `json:"cloudflare_endpoint,omitempty" path:"cloudflare_endpoint,omitempty" url:"cloudflare_endpoint,omitempty"`
	DropboxTeams                            *bool  `json:"dropbox_teams,omitempty" path:"dropbox_teams,omitempty" url:"dropbox_teams,omitempty"`
	LinodeBucket                            string `json:"linode_bucket,omitempty" path:"linode_bucket,omitempty" url:"linode_bucket,omitempty"`
	LinodeAccessKey                         string `json:"linode_access_key,omitempty" path:"linode_access_key,omitempty" url:"linode_access_key,omitempty"`
	LinodeRegion                            string `json:"linode_region,omitempty" path:"linode_region,omitempty" url:"linode_region,omitempty"`
	SupportsVersioning                      *bool  `json:"supports_versioning,omitempty" path:"supports_versioning,omitempty" url:"supports_versioning,omitempty"`
	Password                                string `json:"password,omitempty" path:"password,omitempty" url:"password,omitempty"`
	PrivateKey                              string `json:"private_key,omitempty" path:"private_key,omitempty" url:"private_key,omitempty"`
	PrivateKeyPassphrase                    string `json:"private_key_passphrase,omitempty" path:"private_key_passphrase,omitempty" url:"private_key_passphrase,omitempty"`
	ResetAuthentication                     *bool  `json:"reset_authentication,omitempty" path:"reset_authentication,omitempty" url:"reset_authentication,omitempty"`
	SslCertificate                          string `json:"ssl_certificate,omitempty" path:"ssl_certificate,omitempty" url:"ssl_certificate,omitempty"`
	AwsSecretKey                            string `json:"aws_secret_key,omitempty" path:"aws_secret_key,omitempty" url:"aws_secret_key,omitempty"`
	AzureBlobStorageAccessKey               string `json:"azure_blob_storage_access_key,omitempty" path:"azure_blob_storage_access_key,omitempty" url:"azure_blob_storage_access_key,omitempty"`
	AzureBlobStorageSasToken                string `json:"azure_blob_storage_sas_token,omitempty" path:"azure_blob_storage_sas_token,omitempty" url:"azure_blob_storage_sas_token,omitempty"`
	AzureFilesStorageAccessKey              string `json:"azure_files_storage_access_key,omitempty" path:"azure_files_storage_access_key,omitempty" url:"azure_files_storage_access_key,omitempty"`
	AzureFilesStorageSasToken               string `json:"azure_files_storage_sas_token,omitempty" path:"azure_files_storage_sas_token,omitempty" url:"azure_files_storage_sas_token,omitempty"`
	BackblazeB2ApplicationKey               string `json:"backblaze_b2_application_key,omitempty" path:"backblaze_b2_application_key,omitempty" url:"backblaze_b2_application_key,omitempty"`
	BackblazeB2KeyId                        string `json:"backblaze_b2_key_id,omitempty" path:"backblaze_b2_key_id,omitempty" url:"backblaze_b2_key_id,omitempty"`
	CloudflareSecretKey                     string `json:"cloudflare_secret_key,omitempty" path:"cloudflare_secret_key,omitempty" url:"cloudflare_secret_key,omitempty"`
	FilebaseSecretKey                       string `json:"filebase_secret_key,omitempty" path:"filebase_secret_key,omitempty" url:"filebase_secret_key,omitempty"`
	GoogleCloudStorageCredentialsJson       string `json:"google_cloud_storage_credentials_json,omitempty" path:"google_cloud_storage_credentials_json,omitempty" url:"google_cloud_storage_credentials_json,omitempty"`
	GoogleCloudStorageS3CompatibleSecretKey string `json:"google_cloud_storage_s3_compatible_secret_key,omitempty" path:"google_cloud_storage_s3_compatible_secret_key,omitempty" url:"google_cloud_storage_s3_compatible_secret_key,omitempty"`
	LinodeSecretKey                         string `json:"linode_secret_key,omitempty" path:"linode_secret_key,omitempty" url:"linode_secret_key,omitempty"`
	S3CompatibleSecretKey                   string `json:"s3_compatible_secret_key,omitempty" path:"s3_compatible_secret_key,omitempty" url:"s3_compatible_secret_key,omitempty"`
	WasabiSecretKey                         string `json:"wasabi_secret_key,omitempty" path:"wasabi_secret_key,omitempty" url:"wasabi_secret_key,omitempty"`
}

func (r RemoteServer) Identifier() interface{} {
	return r.Id
}

type RemoteServerCollection []RemoteServer

type RemoteServerFilesAgentPermissionSetEnum string

func (u RemoteServerFilesAgentPermissionSetEnum) String() string {
	return string(u)
}

func (u RemoteServerFilesAgentPermissionSetEnum) Enum() map[string]RemoteServerFilesAgentPermissionSetEnum {
	return map[string]RemoteServerFilesAgentPermissionSetEnum{
		"read_write": RemoteServerFilesAgentPermissionSetEnum("read_write"),
		"read_only":  RemoteServerFilesAgentPermissionSetEnum("read_only"),
		"write_only": RemoteServerFilesAgentPermissionSetEnum("write_only"),
	}
}

type RemoteServerOneDriveAccountTypeEnum string

func (u RemoteServerOneDriveAccountTypeEnum) String() string {
	return string(u)
}

func (u RemoteServerOneDriveAccountTypeEnum) Enum() map[string]RemoteServerOneDriveAccountTypeEnum {
	return map[string]RemoteServerOneDriveAccountTypeEnum{
		"personal":       RemoteServerOneDriveAccountTypeEnum("personal"),
		"business_other": RemoteServerOneDriveAccountTypeEnum("business_other"),
	}
}

type RemoteServerServerCertificateEnum string

func (u RemoteServerServerCertificateEnum) String() string {
	return string(u)
}

func (u RemoteServerServerCertificateEnum) Enum() map[string]RemoteServerServerCertificateEnum {
	return map[string]RemoteServerServerCertificateEnum{
		"require_match": RemoteServerServerCertificateEnum("require_match"),
		"allow_any":     RemoteServerServerCertificateEnum("allow_any"),
	}
}

type RemoteServerServerTypeEnum string

func (u RemoteServerServerTypeEnum) String() string {
	return string(u)
}

func (u RemoteServerServerTypeEnum) Enum() map[string]RemoteServerServerTypeEnum {
	return map[string]RemoteServerServerTypeEnum{
		"ftp":                  RemoteServerServerTypeEnum("ftp"),
		"sftp":                 RemoteServerServerTypeEnum("sftp"),
		"s3":                   RemoteServerServerTypeEnum("s3"),
		"google_cloud_storage": RemoteServerServerTypeEnum("google_cloud_storage"),
		"webdav":               RemoteServerServerTypeEnum("webdav"),
		"wasabi":               RemoteServerServerTypeEnum("wasabi"),
		"backblaze_b2":         RemoteServerServerTypeEnum("backblaze_b2"),
		"one_drive":            RemoteServerServerTypeEnum("one_drive"),
		"box":                  RemoteServerServerTypeEnum("box"),
		"dropbox":              RemoteServerServerTypeEnum("dropbox"),
		"google_drive":         RemoteServerServerTypeEnum("google_drive"),
		"azure":                RemoteServerServerTypeEnum("azure"),
		"sharepoint":           RemoteServerServerTypeEnum("sharepoint"),
		"s3_compatible":        RemoteServerServerTypeEnum("s3_compatible"),
		"azure_files":          RemoteServerServerTypeEnum("azure_files"),
		"files_agent":          RemoteServerServerTypeEnum("files_agent"),
		"filebase":             RemoteServerServerTypeEnum("filebase"),
		"cloudflare":           RemoteServerServerTypeEnum("cloudflare"),
		"linode":               RemoteServerServerTypeEnum("linode"),
	}
}

type RemoteServerSslEnum string

func (u RemoteServerSslEnum) String() string {
	return string(u)
}

func (u RemoteServerSslEnum) Enum() map[string]RemoteServerSslEnum {
	return map[string]RemoteServerSslEnum{
		"if_available":     RemoteServerSslEnum("if_available"),
		"require":          RemoteServerSslEnum("require"),
		"require_implicit": RemoteServerSslEnum("require_implicit"),
		"never":            RemoteServerSslEnum("never"),
	}
}

type RemoteServerListParams struct {
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       RemoteServer           `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

type RemoteServerFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type RemoteServerFindConfigurationFileParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type RemoteServerCreateParams struct {
	Password                                string                                  `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PrivateKey                              string                                  `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassphrase                    string                                  `url:"private_key_passphrase,omitempty" json:"private_key_passphrase,omitempty" path:"private_key_passphrase"`
	ResetAuthentication                     *bool                                   `url:"reset_authentication,omitempty" json:"reset_authentication,omitempty" path:"reset_authentication"`
	SslCertificate                          string                                  `url:"ssl_certificate,omitempty" json:"ssl_certificate,omitempty" path:"ssl_certificate"`
	AwsSecretKey                            string                                  `url:"aws_secret_key,omitempty" json:"aws_secret_key,omitempty" path:"aws_secret_key"`
	AzureBlobStorageAccessKey               string                                  `url:"azure_blob_storage_access_key,omitempty" json:"azure_blob_storage_access_key,omitempty" path:"azure_blob_storage_access_key"`
	AzureBlobStorageSasToken                string                                  `url:"azure_blob_storage_sas_token,omitempty" json:"azure_blob_storage_sas_token,omitempty" path:"azure_blob_storage_sas_token"`
	AzureFilesStorageAccessKey              string                                  `url:"azure_files_storage_access_key,omitempty" json:"azure_files_storage_access_key,omitempty" path:"azure_files_storage_access_key"`
	AzureFilesStorageSasToken               string                                  `url:"azure_files_storage_sas_token,omitempty" json:"azure_files_storage_sas_token,omitempty" path:"azure_files_storage_sas_token"`
	BackblazeB2ApplicationKey               string                                  `url:"backblaze_b2_application_key,omitempty" json:"backblaze_b2_application_key,omitempty" path:"backblaze_b2_application_key"`
	BackblazeB2KeyId                        string                                  `url:"backblaze_b2_key_id,omitempty" json:"backblaze_b2_key_id,omitempty" path:"backblaze_b2_key_id"`
	CloudflareSecretKey                     string                                  `url:"cloudflare_secret_key,omitempty" json:"cloudflare_secret_key,omitempty" path:"cloudflare_secret_key"`
	FilebaseSecretKey                       string                                  `url:"filebase_secret_key,omitempty" json:"filebase_secret_key,omitempty" path:"filebase_secret_key"`
	GoogleCloudStorageCredentialsJson       string                                  `url:"google_cloud_storage_credentials_json,omitempty" json:"google_cloud_storage_credentials_json,omitempty" path:"google_cloud_storage_credentials_json"`
	GoogleCloudStorageS3CompatibleSecretKey string                                  `url:"google_cloud_storage_s3_compatible_secret_key,omitempty" json:"google_cloud_storage_s3_compatible_secret_key,omitempty" path:"google_cloud_storage_s3_compatible_secret_key"`
	LinodeSecretKey                         string                                  `url:"linode_secret_key,omitempty" json:"linode_secret_key,omitempty" path:"linode_secret_key"`
	S3CompatibleSecretKey                   string                                  `url:"s3_compatible_secret_key,omitempty" json:"s3_compatible_secret_key,omitempty" path:"s3_compatible_secret_key"`
	WasabiSecretKey                         string                                  `url:"wasabi_secret_key,omitempty" json:"wasabi_secret_key,omitempty" path:"wasabi_secret_key"`
	AwsAccessKey                            string                                  `url:"aws_access_key,omitempty" json:"aws_access_key,omitempty" path:"aws_access_key"`
	AzureBlobStorageAccount                 string                                  `url:"azure_blob_storage_account,omitempty" json:"azure_blob_storage_account,omitempty" path:"azure_blob_storage_account"`
	AzureBlobStorageContainer               string                                  `url:"azure_blob_storage_container,omitempty" json:"azure_blob_storage_container,omitempty" path:"azure_blob_storage_container"`
	AzureBlobStorageDnsSuffix               string                                  `url:"azure_blob_storage_dns_suffix,omitempty" json:"azure_blob_storage_dns_suffix,omitempty" path:"azure_blob_storage_dns_suffix"`
	AzureBlobStorageHierarchicalNamespace   *bool                                   `url:"azure_blob_storage_hierarchical_namespace,omitempty" json:"azure_blob_storage_hierarchical_namespace,omitempty" path:"azure_blob_storage_hierarchical_namespace"`
	AzureFilesStorageAccount                string                                  `url:"azure_files_storage_account,omitempty" json:"azure_files_storage_account,omitempty" path:"azure_files_storage_account"`
	AzureFilesStorageDnsSuffix              string                                  `url:"azure_files_storage_dns_suffix,omitempty" json:"azure_files_storage_dns_suffix,omitempty" path:"azure_files_storage_dns_suffix"`
	AzureFilesStorageShareName              string                                  `url:"azure_files_storage_share_name,omitempty" json:"azure_files_storage_share_name,omitempty" path:"azure_files_storage_share_name"`
	BackblazeB2Bucket                       string                                  `url:"backblaze_b2_bucket,omitempty" json:"backblaze_b2_bucket,omitempty" path:"backblaze_b2_bucket"`
	BackblazeB2S3Endpoint                   string                                  `url:"backblaze_b2_s3_endpoint,omitempty" json:"backblaze_b2_s3_endpoint,omitempty" path:"backblaze_b2_s3_endpoint"`
	BufferUploadsAlways                     *bool                                   `url:"buffer_uploads_always,omitempty" json:"buffer_uploads_always,omitempty" path:"buffer_uploads_always"`
	CloudflareAccessKey                     string                                  `url:"cloudflare_access_key,omitempty" json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key"`
	CloudflareBucket                        string                                  `url:"cloudflare_bucket,omitempty" json:"cloudflare_bucket,omitempty" path:"cloudflare_bucket"`
	CloudflareEndpoint                      string                                  `url:"cloudflare_endpoint,omitempty" json:"cloudflare_endpoint,omitempty" path:"cloudflare_endpoint"`
	DropboxTeams                            *bool                                   `url:"dropbox_teams,omitempty" json:"dropbox_teams,omitempty" path:"dropbox_teams"`
	EnableDedicatedIps                      *bool                                   `url:"enable_dedicated_ips,omitempty" json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips"`
	FilebaseAccessKey                       string                                  `url:"filebase_access_key,omitempty" json:"filebase_access_key,omitempty" path:"filebase_access_key"`
	FilebaseBucket                          string                                  `url:"filebase_bucket,omitempty" json:"filebase_bucket,omitempty" path:"filebase_bucket"`
	FilesAgentPermissionSet                 RemoteServerFilesAgentPermissionSetEnum `url:"files_agent_permission_set,omitempty" json:"files_agent_permission_set,omitempty" path:"files_agent_permission_set"`
	FilesAgentRoot                          string                                  `url:"files_agent_root,omitempty" json:"files_agent_root,omitempty" path:"files_agent_root"`
	FilesAgentVersion                       string                                  `url:"files_agent_version,omitempty" json:"files_agent_version,omitempty" path:"files_agent_version"`
	GoogleCloudStorageBucket                string                                  `url:"google_cloud_storage_bucket,omitempty" json:"google_cloud_storage_bucket,omitempty" path:"google_cloud_storage_bucket"`
	GoogleCloudStorageProjectId             string                                  `url:"google_cloud_storage_project_id,omitempty" json:"google_cloud_storage_project_id,omitempty" path:"google_cloud_storage_project_id"`
	GoogleCloudStorageS3CompatibleAccessKey string                                  `url:"google_cloud_storage_s3_compatible_access_key,omitempty" json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key"`
	Hostname                                string                                  `url:"hostname,omitempty" json:"hostname,omitempty" path:"hostname"`
	LinodeAccessKey                         string                                  `url:"linode_access_key,omitempty" json:"linode_access_key,omitempty" path:"linode_access_key"`
	LinodeBucket                            string                                  `url:"linode_bucket,omitempty" json:"linode_bucket,omitempty" path:"linode_bucket"`
	LinodeRegion                            string                                  `url:"linode_region,omitempty" json:"linode_region,omitempty" path:"linode_region"`
	MaxConnections                          int64                                   `url:"max_connections,omitempty" json:"max_connections,omitempty" path:"max_connections"`
	Name                                    string                                  `url:"name,omitempty" json:"name,omitempty" path:"name"`
	OneDriveAccountType                     RemoteServerOneDriveAccountTypeEnum     `url:"one_drive_account_type,omitempty" json:"one_drive_account_type,omitempty" path:"one_drive_account_type"`
	PinToSiteRegion                         *bool                                   `url:"pin_to_site_region,omitempty" json:"pin_to_site_region,omitempty" path:"pin_to_site_region"`
	Port                                    int64                                   `url:"port,omitempty" json:"port,omitempty" path:"port"`
	S3Bucket                                string                                  `url:"s3_bucket,omitempty" json:"s3_bucket,omitempty" path:"s3_bucket"`
	S3CompatibleAccessKey                   string                                  `url:"s3_compatible_access_key,omitempty" json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key"`
	S3CompatibleBucket                      string                                  `url:"s3_compatible_bucket,omitempty" json:"s3_compatible_bucket,omitempty" path:"s3_compatible_bucket"`
	S3CompatibleEndpoint                    string                                  `url:"s3_compatible_endpoint,omitempty" json:"s3_compatible_endpoint,omitempty" path:"s3_compatible_endpoint"`
	S3CompatibleRegion                      string                                  `url:"s3_compatible_region,omitempty" json:"s3_compatible_region,omitempty" path:"s3_compatible_region"`
	S3Region                                string                                  `url:"s3_region,omitempty" json:"s3_region,omitempty" path:"s3_region"`
	ServerCertificate                       RemoteServerServerCertificateEnum       `url:"server_certificate,omitempty" json:"server_certificate,omitempty" path:"server_certificate"`
	ServerHostKey                           string                                  `url:"server_host_key,omitempty" json:"server_host_key,omitempty" path:"server_host_key"`
	ServerType                              RemoteServerServerTypeEnum              `url:"server_type,omitempty" json:"server_type,omitempty" path:"server_type"`
	Ssl                                     RemoteServerSslEnum                     `url:"ssl,omitempty" json:"ssl,omitempty" path:"ssl"`
	Username                                string                                  `url:"username,omitempty" json:"username,omitempty" path:"username"`
	WasabiAccessKey                         string                                  `url:"wasabi_access_key,omitempty" json:"wasabi_access_key,omitempty" path:"wasabi_access_key"`
	WasabiBucket                            string                                  `url:"wasabi_bucket,omitempty" json:"wasabi_bucket,omitempty" path:"wasabi_bucket"`
	WasabiRegion                            string                                  `url:"wasabi_region,omitempty" json:"wasabi_region,omitempty" path:"wasabi_region"`
}

// Post local changes, check in, and download configuration file (used by some Remote Server integrations, such as the Files.com Agent)
type RemoteServerConfigurationFileParams struct {
	Id            int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	ApiToken      string `url:"api_token,omitempty" json:"api_token,omitempty" path:"api_token"`
	PermissionSet string `url:"permission_set,omitempty" json:"permission_set,omitempty" path:"permission_set"`
	Root          string `url:"root,omitempty" json:"root,omitempty" path:"root"`
	Hostname      string `url:"hostname,omitempty" json:"hostname,omitempty" path:"hostname"`
	Port          int64  `url:"port,omitempty" json:"port,omitempty" path:"port"`
	Status        string `url:"status,omitempty" json:"status,omitempty" path:"status"`
	ConfigVersion string `url:"config_version,omitempty" json:"config_version,omitempty" path:"config_version"`
	PrivateKey    string `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PublicKey     string `url:"public_key,omitempty" json:"public_key,omitempty" path:"public_key"`
	ServerHostKey string `url:"server_host_key,omitempty" json:"server_host_key,omitempty" path:"server_host_key"`
	Subdomain     string `url:"subdomain,omitempty" json:"subdomain,omitempty" path:"subdomain"`
}

type RemoteServerUpdateParams struct {
	Id                                      int64                                   `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Password                                string                                  `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PrivateKey                              string                                  `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassphrase                    string                                  `url:"private_key_passphrase,omitempty" json:"private_key_passphrase,omitempty" path:"private_key_passphrase"`
	ResetAuthentication                     *bool                                   `url:"reset_authentication,omitempty" json:"reset_authentication,omitempty" path:"reset_authentication"`
	SslCertificate                          string                                  `url:"ssl_certificate,omitempty" json:"ssl_certificate,omitempty" path:"ssl_certificate"`
	AwsSecretKey                            string                                  `url:"aws_secret_key,omitempty" json:"aws_secret_key,omitempty" path:"aws_secret_key"`
	AzureBlobStorageAccessKey               string                                  `url:"azure_blob_storage_access_key,omitempty" json:"azure_blob_storage_access_key,omitempty" path:"azure_blob_storage_access_key"`
	AzureBlobStorageSasToken                string                                  `url:"azure_blob_storage_sas_token,omitempty" json:"azure_blob_storage_sas_token,omitempty" path:"azure_blob_storage_sas_token"`
	AzureFilesStorageAccessKey              string                                  `url:"azure_files_storage_access_key,omitempty" json:"azure_files_storage_access_key,omitempty" path:"azure_files_storage_access_key"`
	AzureFilesStorageSasToken               string                                  `url:"azure_files_storage_sas_token,omitempty" json:"azure_files_storage_sas_token,omitempty" path:"azure_files_storage_sas_token"`
	BackblazeB2ApplicationKey               string                                  `url:"backblaze_b2_application_key,omitempty" json:"backblaze_b2_application_key,omitempty" path:"backblaze_b2_application_key"`
	BackblazeB2KeyId                        string                                  `url:"backblaze_b2_key_id,omitempty" json:"backblaze_b2_key_id,omitempty" path:"backblaze_b2_key_id"`
	CloudflareSecretKey                     string                                  `url:"cloudflare_secret_key,omitempty" json:"cloudflare_secret_key,omitempty" path:"cloudflare_secret_key"`
	FilebaseSecretKey                       string                                  `url:"filebase_secret_key,omitempty" json:"filebase_secret_key,omitempty" path:"filebase_secret_key"`
	GoogleCloudStorageCredentialsJson       string                                  `url:"google_cloud_storage_credentials_json,omitempty" json:"google_cloud_storage_credentials_json,omitempty" path:"google_cloud_storage_credentials_json"`
	GoogleCloudStorageS3CompatibleSecretKey string                                  `url:"google_cloud_storage_s3_compatible_secret_key,omitempty" json:"google_cloud_storage_s3_compatible_secret_key,omitempty" path:"google_cloud_storage_s3_compatible_secret_key"`
	LinodeSecretKey                         string                                  `url:"linode_secret_key,omitempty" json:"linode_secret_key,omitempty" path:"linode_secret_key"`
	S3CompatibleSecretKey                   string                                  `url:"s3_compatible_secret_key,omitempty" json:"s3_compatible_secret_key,omitempty" path:"s3_compatible_secret_key"`
	WasabiSecretKey                         string                                  `url:"wasabi_secret_key,omitempty" json:"wasabi_secret_key,omitempty" path:"wasabi_secret_key"`
	AwsAccessKey                            string                                  `url:"aws_access_key,omitempty" json:"aws_access_key,omitempty" path:"aws_access_key"`
	AzureBlobStorageAccount                 string                                  `url:"azure_blob_storage_account,omitempty" json:"azure_blob_storage_account,omitempty" path:"azure_blob_storage_account"`
	AzureBlobStorageContainer               string                                  `url:"azure_blob_storage_container,omitempty" json:"azure_blob_storage_container,omitempty" path:"azure_blob_storage_container"`
	AzureBlobStorageDnsSuffix               string                                  `url:"azure_blob_storage_dns_suffix,omitempty" json:"azure_blob_storage_dns_suffix,omitempty" path:"azure_blob_storage_dns_suffix"`
	AzureBlobStorageHierarchicalNamespace   *bool                                   `url:"azure_blob_storage_hierarchical_namespace,omitempty" json:"azure_blob_storage_hierarchical_namespace,omitempty" path:"azure_blob_storage_hierarchical_namespace"`
	AzureFilesStorageAccount                string                                  `url:"azure_files_storage_account,omitempty" json:"azure_files_storage_account,omitempty" path:"azure_files_storage_account"`
	AzureFilesStorageDnsSuffix              string                                  `url:"azure_files_storage_dns_suffix,omitempty" json:"azure_files_storage_dns_suffix,omitempty" path:"azure_files_storage_dns_suffix"`
	AzureFilesStorageShareName              string                                  `url:"azure_files_storage_share_name,omitempty" json:"azure_files_storage_share_name,omitempty" path:"azure_files_storage_share_name"`
	BackblazeB2Bucket                       string                                  `url:"backblaze_b2_bucket,omitempty" json:"backblaze_b2_bucket,omitempty" path:"backblaze_b2_bucket"`
	BackblazeB2S3Endpoint                   string                                  `url:"backblaze_b2_s3_endpoint,omitempty" json:"backblaze_b2_s3_endpoint,omitempty" path:"backblaze_b2_s3_endpoint"`
	BufferUploadsAlways                     *bool                                   `url:"buffer_uploads_always,omitempty" json:"buffer_uploads_always,omitempty" path:"buffer_uploads_always"`
	CloudflareAccessKey                     string                                  `url:"cloudflare_access_key,omitempty" json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key"`
	CloudflareBucket                        string                                  `url:"cloudflare_bucket,omitempty" json:"cloudflare_bucket,omitempty" path:"cloudflare_bucket"`
	CloudflareEndpoint                      string                                  `url:"cloudflare_endpoint,omitempty" json:"cloudflare_endpoint,omitempty" path:"cloudflare_endpoint"`
	DropboxTeams                            *bool                                   `url:"dropbox_teams,omitempty" json:"dropbox_teams,omitempty" path:"dropbox_teams"`
	EnableDedicatedIps                      *bool                                   `url:"enable_dedicated_ips,omitempty" json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips"`
	FilebaseAccessKey                       string                                  `url:"filebase_access_key,omitempty" json:"filebase_access_key,omitempty" path:"filebase_access_key"`
	FilebaseBucket                          string                                  `url:"filebase_bucket,omitempty" json:"filebase_bucket,omitempty" path:"filebase_bucket"`
	FilesAgentPermissionSet                 RemoteServerFilesAgentPermissionSetEnum `url:"files_agent_permission_set,omitempty" json:"files_agent_permission_set,omitempty" path:"files_agent_permission_set"`
	FilesAgentRoot                          string                                  `url:"files_agent_root,omitempty" json:"files_agent_root,omitempty" path:"files_agent_root"`
	FilesAgentVersion                       string                                  `url:"files_agent_version,omitempty" json:"files_agent_version,omitempty" path:"files_agent_version"`
	GoogleCloudStorageBucket                string                                  `url:"google_cloud_storage_bucket,omitempty" json:"google_cloud_storage_bucket,omitempty" path:"google_cloud_storage_bucket"`
	GoogleCloudStorageProjectId             string                                  `url:"google_cloud_storage_project_id,omitempty" json:"google_cloud_storage_project_id,omitempty" path:"google_cloud_storage_project_id"`
	GoogleCloudStorageS3CompatibleAccessKey string                                  `url:"google_cloud_storage_s3_compatible_access_key,omitempty" json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key"`
	Hostname                                string                                  `url:"hostname,omitempty" json:"hostname,omitempty" path:"hostname"`
	LinodeAccessKey                         string                                  `url:"linode_access_key,omitempty" json:"linode_access_key,omitempty" path:"linode_access_key"`
	LinodeBucket                            string                                  `url:"linode_bucket,omitempty" json:"linode_bucket,omitempty" path:"linode_bucket"`
	LinodeRegion                            string                                  `url:"linode_region,omitempty" json:"linode_region,omitempty" path:"linode_region"`
	MaxConnections                          int64                                   `url:"max_connections,omitempty" json:"max_connections,omitempty" path:"max_connections"`
	Name                                    string                                  `url:"name,omitempty" json:"name,omitempty" path:"name"`
	OneDriveAccountType                     RemoteServerOneDriveAccountTypeEnum     `url:"one_drive_account_type,omitempty" json:"one_drive_account_type,omitempty" path:"one_drive_account_type"`
	PinToSiteRegion                         *bool                                   `url:"pin_to_site_region,omitempty" json:"pin_to_site_region,omitempty" path:"pin_to_site_region"`
	Port                                    int64                                   `url:"port,omitempty" json:"port,omitempty" path:"port"`
	S3Bucket                                string                                  `url:"s3_bucket,omitempty" json:"s3_bucket,omitempty" path:"s3_bucket"`
	S3CompatibleAccessKey                   string                                  `url:"s3_compatible_access_key,omitempty" json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key"`
	S3CompatibleBucket                      string                                  `url:"s3_compatible_bucket,omitempty" json:"s3_compatible_bucket,omitempty" path:"s3_compatible_bucket"`
	S3CompatibleEndpoint                    string                                  `url:"s3_compatible_endpoint,omitempty" json:"s3_compatible_endpoint,omitempty" path:"s3_compatible_endpoint"`
	S3CompatibleRegion                      string                                  `url:"s3_compatible_region,omitempty" json:"s3_compatible_region,omitempty" path:"s3_compatible_region"`
	S3Region                                string                                  `url:"s3_region,omitempty" json:"s3_region,omitempty" path:"s3_region"`
	ServerCertificate                       RemoteServerServerCertificateEnum       `url:"server_certificate,omitempty" json:"server_certificate,omitempty" path:"server_certificate"`
	ServerHostKey                           string                                  `url:"server_host_key,omitempty" json:"server_host_key,omitempty" path:"server_host_key"`
	ServerType                              RemoteServerServerTypeEnum              `url:"server_type,omitempty" json:"server_type,omitempty" path:"server_type"`
	Ssl                                     RemoteServerSslEnum                     `url:"ssl,omitempty" json:"ssl,omitempty" path:"ssl"`
	Username                                string                                  `url:"username,omitempty" json:"username,omitempty" path:"username"`
	WasabiAccessKey                         string                                  `url:"wasabi_access_key,omitempty" json:"wasabi_access_key,omitempty" path:"wasabi_access_key"`
	WasabiBucket                            string                                  `url:"wasabi_bucket,omitempty" json:"wasabi_bucket,omitempty" path:"wasabi_bucket"`
	WasabiRegion                            string                                  `url:"wasabi_region,omitempty" json:"wasabi_region,omitempty" path:"wasabi_region"`
}

type RemoteServerDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (r *RemoteServer) UnmarshalJSON(data []byte) error {
	type remoteServer RemoteServer
	var v remoteServer
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = RemoteServer(v)
	return nil
}

func (r *RemoteServerCollection) UnmarshalJSON(data []byte) error {
	type remoteServers RemoteServerCollection
	var v remoteServers
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RemoteServerCollection(v)
	return nil
}

func (r *RemoteServerCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
