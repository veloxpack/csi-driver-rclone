package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PublicIpAddress struct {
	IpAddress   string `json:"ip_address,omitempty" path:"ip_address,omitempty" url:"ip_address,omitempty"`
	ServerName  string `json:"server_name,omitempty" path:"server_name,omitempty" url:"server_name,omitempty"`
	FtpEnabled  *bool  `json:"ftp_enabled,omitempty" path:"ftp_enabled,omitempty" url:"ftp_enabled,omitempty"`
	SftpEnabled *bool  `json:"sftp_enabled,omitempty" path:"sftp_enabled,omitempty" url:"sftp_enabled,omitempty"`
}

// Identifier no path or id

type PublicIpAddressCollection []PublicIpAddress

func (p *PublicIpAddress) UnmarshalJSON(data []byte) error {
	type publicIpAddress PublicIpAddress
	var v publicIpAddress
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PublicIpAddress(v)
	return nil
}

func (p *PublicIpAddressCollection) UnmarshalJSON(data []byte) error {
	type publicIpAddresss PublicIpAddressCollection
	var v publicIpAddresss
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PublicIpAddressCollection(v)
	return nil
}

func (p *PublicIpAddressCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
