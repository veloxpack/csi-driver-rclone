package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type RemoteServerCredential struct {
	Id                                      int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId                             int64  `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                                    string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description                             string `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	ServerType                              string `json:"server_type,omitempty" path:"server_type,omitempty" url:"server_type,omitempty"`
	AwsAccessKey                            string `json:"aws_access_key,omitempty" path:"aws_access_key,omitempty" url:"aws_access_key,omitempty"`
	S3AssumeRoleArn                         string `json:"s3_assume_role_arn,omitempty" path:"s3_assume_role_arn,omitempty" url:"s3_assume_role_arn,omitempty"`
	S3AssumeRoleDurationSeconds             int64  `json:"s3_assume_role_duration_seconds,omitempty" path:"s3_assume_role_duration_seconds,omitempty" url:"s3_assume_role_duration_seconds,omitempty"`
	S3AssumeRoleExternalId                  string `json:"s3_assume_role_external_id,omitempty" path:"s3_assume_role_external_id,omitempty" url:"s3_assume_role_external_id,omitempty"`
	GoogleCloudStorageS3CompatibleAccessKey string `json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key,omitempty" url:"google_cloud_storage_s3_compatible_access_key,omitempty"`
	WasabiAccessKey                         string `json:"wasabi_access_key,omitempty" path:"wasabi_access_key,omitempty" url:"wasabi_access_key,omitempty"`
	S3CompatibleAccessKey                   string `json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key,omitempty" url:"s3_compatible_access_key,omitempty"`
	FilebaseAccessKey                       string `json:"filebase_access_key,omitempty" path:"filebase_access_key,omitempty" url:"filebase_access_key,omitempty"`
	CloudflareAccessKey                     string `json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key,omitempty" url:"cloudflare_access_key,omitempty"`
	LinodeAccessKey                         string `json:"linode_access_key,omitempty" path:"linode_access_key,omitempty" url:"linode_access_key,omitempty"`
	Username                                string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	Password                                string `json:"password,omitempty" path:"password,omitempty" url:"password,omitempty"`
	PrivateKey                              string `json:"private_key,omitempty" path:"private_key,omitempty" url:"private_key,omitempty"`
	PrivateKeyPassphrase                    string `json:"private_key_passphrase,omitempty" path:"private_key_passphrase,omitempty" url:"private_key_passphrase,omitempty"`
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

func (r RemoteServerCredential) Identifier() interface{} {
	return r.Id
}

type RemoteServerCredentialCollection []RemoteServerCredential

type RemoteServerCredentialServerTypeEnum string

func (u RemoteServerCredentialServerTypeEnum) String() string {
	return string(u)
}

func (u RemoteServerCredentialServerTypeEnum) Enum() map[string]RemoteServerCredentialServerTypeEnum {
	return map[string]RemoteServerCredentialServerTypeEnum{
		"ftp":                  RemoteServerCredentialServerTypeEnum("ftp"),
		"sftp":                 RemoteServerCredentialServerTypeEnum("sftp"),
		"s3":                   RemoteServerCredentialServerTypeEnum("s3"),
		"google_cloud_storage": RemoteServerCredentialServerTypeEnum("google_cloud_storage"),
		"webdav":               RemoteServerCredentialServerTypeEnum("webdav"),
		"wasabi":               RemoteServerCredentialServerTypeEnum("wasabi"),
		"backblaze_b2":         RemoteServerCredentialServerTypeEnum("backblaze_b2"),
		"one_drive":            RemoteServerCredentialServerTypeEnum("one_drive"),
		"box":                  RemoteServerCredentialServerTypeEnum("box"),
		"dropbox":              RemoteServerCredentialServerTypeEnum("dropbox"),
		"google_drive":         RemoteServerCredentialServerTypeEnum("google_drive"),
		"azure":                RemoteServerCredentialServerTypeEnum("azure"),
		"sharepoint":           RemoteServerCredentialServerTypeEnum("sharepoint"),
		"s3_compatible":        RemoteServerCredentialServerTypeEnum("s3_compatible"),
		"azure_files":          RemoteServerCredentialServerTypeEnum("azure_files"),
		"files_agent":          RemoteServerCredentialServerTypeEnum("files_agent"),
		"filebase":             RemoteServerCredentialServerTypeEnum("filebase"),
		"cloudflare":           RemoteServerCredentialServerTypeEnum("cloudflare"),
		"linode":               RemoteServerCredentialServerTypeEnum("linode"),
	}
}

type RemoteServerCredentialListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

type RemoteServerCredentialFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type RemoteServerCredentialCreateParams struct {
	Name                                    string                               `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description                             string                               `url:"description,omitempty" json:"description,omitempty" path:"description"`
	ServerType                              RemoteServerCredentialServerTypeEnum `url:"server_type,omitempty" json:"server_type,omitempty" path:"server_type"`
	AwsAccessKey                            string                               `url:"aws_access_key,omitempty" json:"aws_access_key,omitempty" path:"aws_access_key"`
	S3AssumeRoleArn                         string                               `url:"s3_assume_role_arn,omitempty" json:"s3_assume_role_arn,omitempty" path:"s3_assume_role_arn"`
	S3AssumeRoleDurationSeconds             int64                                `url:"s3_assume_role_duration_seconds,omitempty" json:"s3_assume_role_duration_seconds,omitempty" path:"s3_assume_role_duration_seconds"`
	CloudflareAccessKey                     string                               `url:"cloudflare_access_key,omitempty" json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key"`
	FilebaseAccessKey                       string                               `url:"filebase_access_key,omitempty" json:"filebase_access_key,omitempty" path:"filebase_access_key"`
	GoogleCloudStorageS3CompatibleAccessKey string                               `url:"google_cloud_storage_s3_compatible_access_key,omitempty" json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key"`
	LinodeAccessKey                         string                               `url:"linode_access_key,omitempty" json:"linode_access_key,omitempty" path:"linode_access_key"`
	S3CompatibleAccessKey                   string                               `url:"s3_compatible_access_key,omitempty" json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key"`
	Username                                string                               `url:"username,omitempty" json:"username,omitempty" path:"username"`
	WasabiAccessKey                         string                               `url:"wasabi_access_key,omitempty" json:"wasabi_access_key,omitempty" path:"wasabi_access_key"`
	Password                                string                               `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PrivateKey                              string                               `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassphrase                    string                               `url:"private_key_passphrase,omitempty" json:"private_key_passphrase,omitempty" path:"private_key_passphrase"`
	AwsSecretKey                            string                               `url:"aws_secret_key,omitempty" json:"aws_secret_key,omitempty" path:"aws_secret_key"`
	AzureBlobStorageAccessKey               string                               `url:"azure_blob_storage_access_key,omitempty" json:"azure_blob_storage_access_key,omitempty" path:"azure_blob_storage_access_key"`
	AzureBlobStorageSasToken                string                               `url:"azure_blob_storage_sas_token,omitempty" json:"azure_blob_storage_sas_token,omitempty" path:"azure_blob_storage_sas_token"`
	AzureFilesStorageAccessKey              string                               `url:"azure_files_storage_access_key,omitempty" json:"azure_files_storage_access_key,omitempty" path:"azure_files_storage_access_key"`
	AzureFilesStorageSasToken               string                               `url:"azure_files_storage_sas_token,omitempty" json:"azure_files_storage_sas_token,omitempty" path:"azure_files_storage_sas_token"`
	BackblazeB2ApplicationKey               string                               `url:"backblaze_b2_application_key,omitempty" json:"backblaze_b2_application_key,omitempty" path:"backblaze_b2_application_key"`
	BackblazeB2KeyId                        string                               `url:"backblaze_b2_key_id,omitempty" json:"backblaze_b2_key_id,omitempty" path:"backblaze_b2_key_id"`
	CloudflareSecretKey                     string                               `url:"cloudflare_secret_key,omitempty" json:"cloudflare_secret_key,omitempty" path:"cloudflare_secret_key"`
	FilebaseSecretKey                       string                               `url:"filebase_secret_key,omitempty" json:"filebase_secret_key,omitempty" path:"filebase_secret_key"`
	GoogleCloudStorageCredentialsJson       string                               `url:"google_cloud_storage_credentials_json,omitempty" json:"google_cloud_storage_credentials_json,omitempty" path:"google_cloud_storage_credentials_json"`
	GoogleCloudStorageS3CompatibleSecretKey string                               `url:"google_cloud_storage_s3_compatible_secret_key,omitempty" json:"google_cloud_storage_s3_compatible_secret_key,omitempty" path:"google_cloud_storage_s3_compatible_secret_key"`
	LinodeSecretKey                         string                               `url:"linode_secret_key,omitempty" json:"linode_secret_key,omitempty" path:"linode_secret_key"`
	S3CompatibleSecretKey                   string                               `url:"s3_compatible_secret_key,omitempty" json:"s3_compatible_secret_key,omitempty" path:"s3_compatible_secret_key"`
	WasabiSecretKey                         string                               `url:"wasabi_secret_key,omitempty" json:"wasabi_secret_key,omitempty" path:"wasabi_secret_key"`
	WorkspaceId                             int64                                `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type RemoteServerCredentialUpdateParams struct {
	Id                                      int64                                `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name                                    string                               `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description                             string                               `url:"description,omitempty" json:"description,omitempty" path:"description"`
	ServerType                              RemoteServerCredentialServerTypeEnum `url:"server_type,omitempty" json:"server_type,omitempty" path:"server_type"`
	AwsAccessKey                            string                               `url:"aws_access_key,omitempty" json:"aws_access_key,omitempty" path:"aws_access_key"`
	S3AssumeRoleArn                         string                               `url:"s3_assume_role_arn,omitempty" json:"s3_assume_role_arn,omitempty" path:"s3_assume_role_arn"`
	S3AssumeRoleDurationSeconds             int64                                `url:"s3_assume_role_duration_seconds,omitempty" json:"s3_assume_role_duration_seconds,omitempty" path:"s3_assume_role_duration_seconds"`
	CloudflareAccessKey                     string                               `url:"cloudflare_access_key,omitempty" json:"cloudflare_access_key,omitempty" path:"cloudflare_access_key"`
	FilebaseAccessKey                       string                               `url:"filebase_access_key,omitempty" json:"filebase_access_key,omitempty" path:"filebase_access_key"`
	GoogleCloudStorageS3CompatibleAccessKey string                               `url:"google_cloud_storage_s3_compatible_access_key,omitempty" json:"google_cloud_storage_s3_compatible_access_key,omitempty" path:"google_cloud_storage_s3_compatible_access_key"`
	LinodeAccessKey                         string                               `url:"linode_access_key,omitempty" json:"linode_access_key,omitempty" path:"linode_access_key"`
	S3CompatibleAccessKey                   string                               `url:"s3_compatible_access_key,omitempty" json:"s3_compatible_access_key,omitempty" path:"s3_compatible_access_key"`
	Username                                string                               `url:"username,omitempty" json:"username,omitempty" path:"username"`
	WasabiAccessKey                         string                               `url:"wasabi_access_key,omitempty" json:"wasabi_access_key,omitempty" path:"wasabi_access_key"`
	Password                                string                               `url:"password,omitempty" json:"password,omitempty" path:"password"`
	PrivateKey                              string                               `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassphrase                    string                               `url:"private_key_passphrase,omitempty" json:"private_key_passphrase,omitempty" path:"private_key_passphrase"`
	AwsSecretKey                            string                               `url:"aws_secret_key,omitempty" json:"aws_secret_key,omitempty" path:"aws_secret_key"`
	AzureBlobStorageAccessKey               string                               `url:"azure_blob_storage_access_key,omitempty" json:"azure_blob_storage_access_key,omitempty" path:"azure_blob_storage_access_key"`
	AzureBlobStorageSasToken                string                               `url:"azure_blob_storage_sas_token,omitempty" json:"azure_blob_storage_sas_token,omitempty" path:"azure_blob_storage_sas_token"`
	AzureFilesStorageAccessKey              string                               `url:"azure_files_storage_access_key,omitempty" json:"azure_files_storage_access_key,omitempty" path:"azure_files_storage_access_key"`
	AzureFilesStorageSasToken               string                               `url:"azure_files_storage_sas_token,omitempty" json:"azure_files_storage_sas_token,omitempty" path:"azure_files_storage_sas_token"`
	BackblazeB2ApplicationKey               string                               `url:"backblaze_b2_application_key,omitempty" json:"backblaze_b2_application_key,omitempty" path:"backblaze_b2_application_key"`
	BackblazeB2KeyId                        string                               `url:"backblaze_b2_key_id,omitempty" json:"backblaze_b2_key_id,omitempty" path:"backblaze_b2_key_id"`
	CloudflareSecretKey                     string                               `url:"cloudflare_secret_key,omitempty" json:"cloudflare_secret_key,omitempty" path:"cloudflare_secret_key"`
	FilebaseSecretKey                       string                               `url:"filebase_secret_key,omitempty" json:"filebase_secret_key,omitempty" path:"filebase_secret_key"`
	GoogleCloudStorageCredentialsJson       string                               `url:"google_cloud_storage_credentials_json,omitempty" json:"google_cloud_storage_credentials_json,omitempty" path:"google_cloud_storage_credentials_json"`
	GoogleCloudStorageS3CompatibleSecretKey string                               `url:"google_cloud_storage_s3_compatible_secret_key,omitempty" json:"google_cloud_storage_s3_compatible_secret_key,omitempty" path:"google_cloud_storage_s3_compatible_secret_key"`
	LinodeSecretKey                         string                               `url:"linode_secret_key,omitempty" json:"linode_secret_key,omitempty" path:"linode_secret_key"`
	S3CompatibleSecretKey                   string                               `url:"s3_compatible_secret_key,omitempty" json:"s3_compatible_secret_key,omitempty" path:"s3_compatible_secret_key"`
	WasabiSecretKey                         string                               `url:"wasabi_secret_key,omitempty" json:"wasabi_secret_key,omitempty" path:"wasabi_secret_key"`
}

type RemoteServerCredentialDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (r *RemoteServerCredential) UnmarshalJSON(data []byte) error {
	type remoteServerCredential RemoteServerCredential
	var v remoteServerCredential
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = RemoteServerCredential(v)
	return nil
}

func (r *RemoteServerCredentialCollection) UnmarshalJSON(data []byte) error {
	type remoteServerCredentials RemoteServerCredentialCollection
	var v remoteServerCredentials
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RemoteServerCredentialCollection(v)
	return nil
}

func (r *RemoteServerCredentialCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
