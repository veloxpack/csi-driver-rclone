package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type IpAddress struct {
	Id             string   `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	AssociatedWith string   `json:"associated_with,omitempty" path:"associated_with,omitempty" url:"associated_with,omitempty"`
	GroupId        int64    `json:"group_id,omitempty" path:"group_id,omitempty" url:"group_id,omitempty"`
	IpAddresses    []string `json:"ip_addresses,omitempty" path:"ip_addresses,omitempty" url:"ip_addresses,omitempty"`
}

func (i IpAddress) Identifier() interface{} {
	return i.Id
}

type IpAddressCollection []IpAddress

type IpAddressListParams struct {
	ListParams
}

type IpAddressGetSmartfileReservedParams struct {
	ListParams
}

type IpAddressGetExavaultReservedParams struct {
	ListParams
}

type IpAddressGetReservedParams struct {
	ListParams
}

func (i *IpAddress) UnmarshalJSON(data []byte) error {
	type ipAddress IpAddress
	var v ipAddress
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = IpAddress(v)
	return nil
}

func (i *IpAddressCollection) UnmarshalJSON(data []byte) error {
	type ipAddresss IpAddressCollection
	var v ipAddresss
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = IpAddressCollection(v)
	return nil
}

func (i *IpAddressCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
