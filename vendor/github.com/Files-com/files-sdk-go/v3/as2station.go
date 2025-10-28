package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type As2Station struct {
	Id                         int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                       string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Uri                        string `json:"uri,omitempty" path:"uri,omitempty" url:"uri,omitempty"`
	Domain                     string `json:"domain,omitempty" path:"domain,omitempty" url:"domain,omitempty"`
	HexPublicCertificateSerial string `json:"hex_public_certificate_serial,omitempty" path:"hex_public_certificate_serial,omitempty" url:"hex_public_certificate_serial,omitempty"`
	PublicCertificateMd5       string `json:"public_certificate_md5,omitempty" path:"public_certificate_md5,omitempty" url:"public_certificate_md5,omitempty"`
	PublicCertificate          string `json:"public_certificate,omitempty" path:"public_certificate,omitempty" url:"public_certificate,omitempty"`
	PrivateKeyMd5              string `json:"private_key_md5,omitempty" path:"private_key_md5,omitempty" url:"private_key_md5,omitempty"`
	PublicCertificateSubject   string `json:"public_certificate_subject,omitempty" path:"public_certificate_subject,omitempty" url:"public_certificate_subject,omitempty"`
	PublicCertificateIssuer    string `json:"public_certificate_issuer,omitempty" path:"public_certificate_issuer,omitempty" url:"public_certificate_issuer,omitempty"`
	PublicCertificateSerial    string `json:"public_certificate_serial,omitempty" path:"public_certificate_serial,omitempty" url:"public_certificate_serial,omitempty"`
	PublicCertificateNotBefore string `json:"public_certificate_not_before,omitempty" path:"public_certificate_not_before,omitempty" url:"public_certificate_not_before,omitempty"`
	PublicCertificateNotAfter  string `json:"public_certificate_not_after,omitempty" path:"public_certificate_not_after,omitempty" url:"public_certificate_not_after,omitempty"`
	PrivateKeyPasswordMd5      string `json:"private_key_password_md5,omitempty" path:"private_key_password_md5,omitempty" url:"private_key_password_md5,omitempty"`
	PrivateKey                 string `json:"private_key,omitempty" path:"private_key,omitempty" url:"private_key,omitempty"`
	PrivateKeyPassword         string `json:"private_key_password,omitempty" path:"private_key_password,omitempty" url:"private_key_password,omitempty"`
}

func (a As2Station) Identifier() interface{} {
	return a.Id
}

type As2StationCollection []As2Station

type As2StationListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type As2StationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type As2StationCreateParams struct {
	Name               string `url:"name" json:"name" path:"name"`
	PublicCertificate  string `url:"public_certificate" json:"public_certificate" path:"public_certificate"`
	PrivateKey         string `url:"private_key" json:"private_key" path:"private_key"`
	PrivateKeyPassword string `url:"private_key_password,omitempty" json:"private_key_password,omitempty" path:"private_key_password"`
}

type As2StationUpdateParams struct {
	Id                 int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name               string `url:"name,omitempty" json:"name,omitempty" path:"name"`
	PublicCertificate  string `url:"public_certificate,omitempty" json:"public_certificate,omitempty" path:"public_certificate"`
	PrivateKey         string `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassword string `url:"private_key_password,omitempty" json:"private_key_password,omitempty" path:"private_key_password"`
}

type As2StationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *As2Station) UnmarshalJSON(data []byte) error {
	type as2Station As2Station
	var v as2Station
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = As2Station(v)
	return nil
}

func (a *As2StationCollection) UnmarshalJSON(data []byte) error {
	type as2Stations As2StationCollection
	var v as2Stations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = As2StationCollection(v)
	return nil
}

func (a *As2StationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
