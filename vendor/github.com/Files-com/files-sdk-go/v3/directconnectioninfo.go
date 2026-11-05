package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type DirectConnectionInfo struct {
	Version    int64    `json:"version,omitempty" path:"version,omitempty" url:"version,omitempty"`
	ServerName string   `json:"server_name,omitempty" path:"server_name,omitempty" url:"server_name,omitempty"`
	Addresses  []string `json:"addresses,omitempty" path:"addresses,omitempty" url:"addresses,omitempty"`
	DirectUri  string   `json:"direct_uri,omitempty" path:"direct_uri,omitempty" url:"direct_uri,omitempty"`
	CaPem      string   `json:"ca_pem,omitempty" path:"ca_pem,omitempty" url:"ca_pem,omitempty"`
}

// Identifier no path or id

type DirectConnectionInfoCollection []DirectConnectionInfo

func (d *DirectConnectionInfo) UnmarshalJSON(data []byte) error {
	type directConnectionInfo DirectConnectionInfo
	var v directConnectionInfo
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*d = DirectConnectionInfo(v)
	return nil
}

func (d *DirectConnectionInfoCollection) UnmarshalJSON(data []byte) error {
	type directConnectionInfos DirectConnectionInfoCollection
	var v directConnectionInfos
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*d = DirectConnectionInfoCollection(v)
	return nil
}

func (d *DirectConnectionInfoCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*d))
	for i, v := range *d {
		ret[i] = v
	}

	return &ret
}
