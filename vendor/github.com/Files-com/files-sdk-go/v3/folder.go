package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Folder struct {
	Path                               string                 `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	CreatedById                        int64                  `json:"created_by_id,omitempty" path:"created_by_id,omitempty" url:"created_by_id,omitempty"`
	CreatedByApiKeyId                  int64                  `json:"created_by_api_key_id,omitempty" path:"created_by_api_key_id,omitempty" url:"created_by_api_key_id,omitempty"`
	CreatedByAs2IncomingMessageId      int64                  `json:"created_by_as2_incoming_message_id,omitempty" path:"created_by_as2_incoming_message_id,omitempty" url:"created_by_as2_incoming_message_id,omitempty"`
	CreatedByAutomationId              int64                  `json:"created_by_automation_id,omitempty" path:"created_by_automation_id,omitempty" url:"created_by_automation_id,omitempty"`
	CreatedByBundleRegistrationId      int64                  `json:"created_by_bundle_registration_id,omitempty" path:"created_by_bundle_registration_id,omitempty" url:"created_by_bundle_registration_id,omitempty"`
	CreatedByInboxId                   int64                  `json:"created_by_inbox_id,omitempty" path:"created_by_inbox_id,omitempty" url:"created_by_inbox_id,omitempty"`
	CreatedByRemoteServerId            int64                  `json:"created_by_remote_server_id,omitempty" path:"created_by_remote_server_id,omitempty" url:"created_by_remote_server_id,omitempty"`
	CreatedByRemoteServerSyncId        int64                  `json:"created_by_remote_server_sync_id,omitempty" path:"created_by_remote_server_sync_id,omitempty" url:"created_by_remote_server_sync_id,omitempty"`
	CustomMetadata                     map[string]interface{} `json:"custom_metadata,omitempty" path:"custom_metadata,omitempty" url:"custom_metadata,omitempty"`
	DisplayName                        string                 `json:"display_name,omitempty" path:"display_name,omitempty" url:"display_name,omitempty"`
	Type                               string                 `json:"type,omitempty" path:"type,omitempty" url:"type,omitempty"`
	Size                               int64                  `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
	CreatedAt                          *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	LastModifiedById                   int64                  `json:"last_modified_by_id,omitempty" path:"last_modified_by_id,omitempty" url:"last_modified_by_id,omitempty"`
	LastModifiedByApiKeyId             int64                  `json:"last_modified_by_api_key_id,omitempty" path:"last_modified_by_api_key_id,omitempty" url:"last_modified_by_api_key_id,omitempty"`
	LastModifiedByAutomationId         int64                  `json:"last_modified_by_automation_id,omitempty" path:"last_modified_by_automation_id,omitempty" url:"last_modified_by_automation_id,omitempty"`
	LastModifiedByBundleRegistrationId int64                  `json:"last_modified_by_bundle_registration_id,omitempty" path:"last_modified_by_bundle_registration_id,omitempty" url:"last_modified_by_bundle_registration_id,omitempty"`
	LastModifiedByRemoteServerId       int64                  `json:"last_modified_by_remote_server_id,omitempty" path:"last_modified_by_remote_server_id,omitempty" url:"last_modified_by_remote_server_id,omitempty"`
	LastModifiedByRemoteServerSyncId   int64                  `json:"last_modified_by_remote_server_sync_id,omitempty" path:"last_modified_by_remote_server_sync_id,omitempty" url:"last_modified_by_remote_server_sync_id,omitempty"`
	Mtime                              *time.Time             `json:"mtime,omitempty" path:"mtime,omitempty" url:"mtime,omitempty"`
	ProvidedMtime                      *time.Time             `json:"provided_mtime,omitempty" path:"provided_mtime,omitempty" url:"provided_mtime,omitempty"`
	Crc32                              string                 `json:"crc32,omitempty" path:"crc32,omitempty" url:"crc32,omitempty"`
	Md5                                string                 `json:"md5,omitempty" path:"md5,omitempty" url:"md5,omitempty"`
	Sha1                               string                 `json:"sha1,omitempty" path:"sha1,omitempty" url:"sha1,omitempty"`
	Sha256                             string                 `json:"sha256,omitempty" path:"sha256,omitempty" url:"sha256,omitempty"`
	MimeType                           string                 `json:"mime_type,omitempty" path:"mime_type,omitempty" url:"mime_type,omitempty"`
	Region                             string                 `json:"region,omitempty" path:"region,omitempty" url:"region,omitempty"`
	Permissions                        string                 `json:"permissions,omitempty" path:"permissions,omitempty" url:"permissions,omitempty"`
	SubfoldersLocked                   *bool                  `json:"subfolders_locked?,omitempty" path:"subfolders_locked?,omitempty" url:"subfolders_locked?,omitempty"`
	IsLocked                           *bool                  `json:"is_locked,omitempty" path:"is_locked,omitempty" url:"is_locked,omitempty"`
	DownloadUri                        string                 `json:"download_uri,omitempty" path:"download_uri,omitempty" url:"download_uri,omitempty"`
	PriorityColor                      string                 `json:"priority_color,omitempty" path:"priority_color,omitempty" url:"priority_color,omitempty"`
	PreviewId                          int64                  `json:"preview_id,omitempty" path:"preview_id,omitempty" url:"preview_id,omitempty"`
	Preview                            Preview                `json:"preview,omitempty" path:"preview,omitempty" url:"preview,omitempty"`
	MkdirParents                       *bool                  `json:"mkdir_parents,omitempty" path:"mkdir_parents,omitempty" url:"mkdir_parents,omitempty"`
}

func (f Folder) Identifier() interface{} {
	return f.Path
}

type FolderCollection []Folder

type FolderListForParams struct {
	Path                    string                              `url:"-,omitempty" json:"-,omitempty" path:"path"`
	PreviewSize             string                              `url:"preview_size,omitempty" json:"preview_size,omitempty" path:"preview_size"`
	SortBy                  map[string]interface{}              `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Search                  string                              `url:"search,omitempty" json:"search,omitempty" path:"search"`
	SearchCustomMetadataKey string                              `url:"search_custom_metadata_key,omitempty" json:"search_custom_metadata_key,omitempty" path:"search_custom_metadata_key"`
	SearchAll               *bool                               `url:"search_all,omitempty" json:"search_all,omitempty" path:"search_all"`
	WithPreviews            *bool                               `url:"with_previews,omitempty" json:"with_previews,omitempty" path:"with_previews"`
	WithPriorityColor       *bool                               `url:"with_priority_color,omitempty" json:"with_priority_color,omitempty" path:"with_priority_color"`
	Type                    string                              `url:"type,omitempty" json:"type,omitempty" path:"type"`
	ModifiedAtDatetime      *time.Time                          `url:"modified_at_datetime,omitempty" json:"modified_at_datetime,omitempty" path:"modified_at_datetime"`
	ConcurrencyManager      lib.ConcurrencyManagerWithSubWorker `url:"-" required:"false" json:"-"`
	ListParams
}

type FolderCreateParams struct {
	Path          string     `url:"-,omitempty" json:"-,omitempty" path:"path"`
	MkdirParents  *bool      `url:"mkdir_parents,omitempty" json:"mkdir_parents,omitempty" path:"mkdir_parents"`
	ProvidedMtime *time.Time `url:"provided_mtime,omitempty" json:"provided_mtime,omitempty" path:"provided_mtime"`
}

func (f *Folder) ToFile() (File, error) {
	bodyBytes, err := json.Marshal(f)
	if err != nil {
		return File{}, err
	}
	file := File{}
	file.UnmarshalJSON(bodyBytes)
	return file, nil
}

func (f Folder) IsDir() bool {
	return f.Type == "directory"
}

func (f *Folder) UnmarshalJSON(data []byte) error {
	type folder Folder
	var v folder
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = Folder(v)
	return nil
}

func (f *FolderCollection) UnmarshalJSON(data []byte) error {
	type folders FolderCollection
	var v folders
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FolderCollection(v)
	return nil
}

func (f *FolderCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
