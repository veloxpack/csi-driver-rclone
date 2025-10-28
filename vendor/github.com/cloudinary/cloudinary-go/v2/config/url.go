package config

import "github.com/cloudinary/cloudinary-go/v2/internal/signature"

// URL defines the configuration applied when generating Cloudinary URLs.
//
// https://cloudinary.com/documentation/how_to_integrate_cloudinary#get_familiar_with_the_cloudinary_console
type URL struct {
	Domain             string `schema:"-" default:"cloudinary.com"`
	SubDomain          string `schema:"-" default:"res"`
	SharedHost         string `schema:"-" default:"res.cloudinary.com"`
	CName              string `schema:"cname"`
	SecureCName        string `schema:"secure_cname"`
	Secure             bool   `schema:"secure" default:"true"`
	CDNSubDomain       bool   `schema:"cdn_sub_domain"`
	SecureCDNSubDomain bool   `schema:"secure_cdn_sub_domain"`
	PrivateCDN         bool   `schema:"private_cdn"`
	SignURL            bool   `schema:"sign_url"`
	LongURLSignature   bool   `schema:"long_url_signature"`
	Shorten            bool   `schema:"shorten"`
	UseRootPath        bool   `schema:"use_root_path"`
	ForceVersion       bool   `schema:"force_version" default:"true"`
	Analytics          bool   `schema:"analytics" default:"true"`
}

// Protocol returns URL protocol (http or https).
func (uc URL) Protocol() string {
	if uc.Secure {
		return "https"
	}

	return "http"
}

// GetSignatureLength returns the length of the URL signature.
func (uc URL) GetSignatureLength() uint8 {
	if uc.LongURLSignature {
		return signature.Long
	}

	return signature.Short
}
