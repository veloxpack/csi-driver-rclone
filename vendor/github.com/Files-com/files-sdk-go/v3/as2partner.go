package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type As2Partner struct {
	Id                         int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	As2StationId               int64                  `json:"as2_station_id,omitempty" path:"as2_station_id,omitempty" url:"as2_station_id,omitempty"`
	Name                       string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Uri                        string                 `json:"uri,omitempty" path:"uri,omitempty" url:"uri,omitempty"`
	ServerCertificate          string                 `json:"server_certificate,omitempty" path:"server_certificate,omitempty" url:"server_certificate,omitempty"`
	HttpAuthUsername           string                 `json:"http_auth_username,omitempty" path:"http_auth_username,omitempty" url:"http_auth_username,omitempty"`
	AdditionalHttpHeaders      map[string]interface{} `json:"additional_http_headers,omitempty" path:"additional_http_headers,omitempty" url:"additional_http_headers,omitempty"`
	DefaultMimeType            string                 `json:"default_mime_type,omitempty" path:"default_mime_type,omitempty" url:"default_mime_type,omitempty"`
	MdnValidationLevel         string                 `json:"mdn_validation_level,omitempty" path:"mdn_validation_level,omitempty" url:"mdn_validation_level,omitempty"`
	SignatureValidationLevel   string                 `json:"signature_validation_level,omitempty" path:"signature_validation_level,omitempty" url:"signature_validation_level,omitempty"`
	EnableDedicatedIps         *bool                  `json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips,omitempty" url:"enable_dedicated_ips,omitempty"`
	HexPublicCertificateSerial string                 `json:"hex_public_certificate_serial,omitempty" path:"hex_public_certificate_serial,omitempty" url:"hex_public_certificate_serial,omitempty"`
	PublicCertificate          string                 `json:"public_certificate,omitempty" path:"public_certificate,omitempty" url:"public_certificate,omitempty"`
	PublicCertificateMd5       string                 `json:"public_certificate_md5,omitempty" path:"public_certificate_md5,omitempty" url:"public_certificate_md5,omitempty"`
	PublicCertificateSubject   string                 `json:"public_certificate_subject,omitempty" path:"public_certificate_subject,omitempty" url:"public_certificate_subject,omitempty"`
	PublicCertificateIssuer    string                 `json:"public_certificate_issuer,omitempty" path:"public_certificate_issuer,omitempty" url:"public_certificate_issuer,omitempty"`
	PublicCertificateSerial    string                 `json:"public_certificate_serial,omitempty" path:"public_certificate_serial,omitempty" url:"public_certificate_serial,omitempty"`
	PublicCertificateNotBefore string                 `json:"public_certificate_not_before,omitempty" path:"public_certificate_not_before,omitempty" url:"public_certificate_not_before,omitempty"`
	PublicCertificateNotAfter  string                 `json:"public_certificate_not_after,omitempty" path:"public_certificate_not_after,omitempty" url:"public_certificate_not_after,omitempty"`
	HttpAuthPassword           string                 `json:"http_auth_password,omitempty" path:"http_auth_password,omitempty" url:"http_auth_password,omitempty"`
}

func (a As2Partner) Identifier() interface{} {
	return a.Id
}

type As2PartnerCollection []As2Partner

type As2PartnerMdnValidationLevelEnum string

func (u As2PartnerMdnValidationLevelEnum) String() string {
	return string(u)
}

func (u As2PartnerMdnValidationLevelEnum) Enum() map[string]As2PartnerMdnValidationLevelEnum {
	return map[string]As2PartnerMdnValidationLevelEnum{
		"none":   As2PartnerMdnValidationLevelEnum("none"),
		"weak":   As2PartnerMdnValidationLevelEnum("weak"),
		"normal": As2PartnerMdnValidationLevelEnum("normal"),
		"strict": As2PartnerMdnValidationLevelEnum("strict"),
		"auto":   As2PartnerMdnValidationLevelEnum("auto"),
	}
}

type As2PartnerSignatureValidationLevelEnum string

func (u As2PartnerSignatureValidationLevelEnum) String() string {
	return string(u)
}

func (u As2PartnerSignatureValidationLevelEnum) Enum() map[string]As2PartnerSignatureValidationLevelEnum {
	return map[string]As2PartnerSignatureValidationLevelEnum{
		"normal": As2PartnerSignatureValidationLevelEnum("normal"),
		"none":   As2PartnerSignatureValidationLevelEnum("none"),
		"auto":   As2PartnerSignatureValidationLevelEnum("auto"),
	}
}

type As2PartnerServerCertificateEnum string

func (u As2PartnerServerCertificateEnum) String() string {
	return string(u)
}

func (u As2PartnerServerCertificateEnum) Enum() map[string]As2PartnerServerCertificateEnum {
	return map[string]As2PartnerServerCertificateEnum{
		"require_match": As2PartnerServerCertificateEnum("require_match"),
		"allow_any":     As2PartnerServerCertificateEnum("allow_any"),
	}
}

type As2PartnerListParams struct {
	ListParams
}

type As2PartnerFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type As2PartnerCreateParams struct {
	EnableDedicatedIps       *bool                                  `url:"enable_dedicated_ips,omitempty" json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips"`
	HttpAuthUsername         string                                 `url:"http_auth_username,omitempty" json:"http_auth_username,omitempty" path:"http_auth_username"`
	HttpAuthPassword         string                                 `url:"http_auth_password,omitempty" json:"http_auth_password,omitempty" path:"http_auth_password"`
	MdnValidationLevel       As2PartnerMdnValidationLevelEnum       `url:"mdn_validation_level,omitempty" json:"mdn_validation_level,omitempty" path:"mdn_validation_level"`
	SignatureValidationLevel As2PartnerSignatureValidationLevelEnum `url:"signature_validation_level,omitempty" json:"signature_validation_level,omitempty" path:"signature_validation_level"`
	ServerCertificate        As2PartnerServerCertificateEnum        `url:"server_certificate,omitempty" json:"server_certificate,omitempty" path:"server_certificate"`
	DefaultMimeType          string                                 `url:"default_mime_type,omitempty" json:"default_mime_type,omitempty" path:"default_mime_type"`
	AdditionalHttpHeaders    map[string]interface{}                 `url:"additional_http_headers,omitempty" json:"additional_http_headers,omitempty" path:"additional_http_headers"`
	As2StationId             int64                                  `url:"as2_station_id" json:"as2_station_id" path:"as2_station_id"`
	Name                     string                                 `url:"name" json:"name" path:"name"`
	Uri                      string                                 `url:"uri" json:"uri" path:"uri"`
	PublicCertificate        string                                 `url:"public_certificate" json:"public_certificate" path:"public_certificate"`
}

type As2PartnerUpdateParams struct {
	Id                       int64                                  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	EnableDedicatedIps       *bool                                  `url:"enable_dedicated_ips,omitempty" json:"enable_dedicated_ips,omitempty" path:"enable_dedicated_ips"`
	HttpAuthUsername         string                                 `url:"http_auth_username,omitempty" json:"http_auth_username,omitempty" path:"http_auth_username"`
	HttpAuthPassword         string                                 `url:"http_auth_password,omitempty" json:"http_auth_password,omitempty" path:"http_auth_password"`
	MdnValidationLevel       As2PartnerMdnValidationLevelEnum       `url:"mdn_validation_level,omitempty" json:"mdn_validation_level,omitempty" path:"mdn_validation_level"`
	SignatureValidationLevel As2PartnerSignatureValidationLevelEnum `url:"signature_validation_level,omitempty" json:"signature_validation_level,omitempty" path:"signature_validation_level"`
	ServerCertificate        As2PartnerServerCertificateEnum        `url:"server_certificate,omitempty" json:"server_certificate,omitempty" path:"server_certificate"`
	DefaultMimeType          string                                 `url:"default_mime_type,omitempty" json:"default_mime_type,omitempty" path:"default_mime_type"`
	AdditionalHttpHeaders    map[string]interface{}                 `url:"additional_http_headers,omitempty" json:"additional_http_headers,omitempty" path:"additional_http_headers"`
	Name                     string                                 `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Uri                      string                                 `url:"uri,omitempty" json:"uri,omitempty" path:"uri"`
	PublicCertificate        string                                 `url:"public_certificate,omitempty" json:"public_certificate,omitempty" path:"public_certificate"`
}

type As2PartnerDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *As2Partner) UnmarshalJSON(data []byte) error {
	type as2Partner As2Partner
	var v as2Partner
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = As2Partner(v)
	return nil
}

func (a *As2PartnerCollection) UnmarshalJSON(data []byte) error {
	type as2Partners As2PartnerCollection
	var v as2Partners
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = As2PartnerCollection(v)
	return nil
}

func (a *As2PartnerCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
