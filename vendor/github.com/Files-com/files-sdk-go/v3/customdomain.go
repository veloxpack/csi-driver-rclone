package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type CustomDomain struct {
	Id               int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Domain           string     `json:"domain,omitempty" path:"domain,omitempty" url:"domain,omitempty"`
	Destination      string     `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	DnsStatus        string     `json:"dns_status,omitempty" path:"dns_status,omitempty" url:"dns_status,omitempty"`
	SslCertificateId int64      `json:"ssl_certificate_id,omitempty" path:"ssl_certificate_id,omitempty" url:"ssl_certificate_id,omitempty"`
	BrickManaged     *bool      `json:"brick_managed,omitempty" path:"brick_managed,omitempty" url:"brick_managed,omitempty"`
	FolderBehaviorId int64      `json:"folder_behavior_id,omitempty" path:"folder_behavior_id,omitempty" url:"folder_behavior_id,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (c CustomDomain) Identifier() interface{} {
	return c.Id
}

type CustomDomainCollection []CustomDomain

type CustomDomainDestinationEnum string

func (u CustomDomainDestinationEnum) String() string {
	return string(u)
}

func (u CustomDomainDestinationEnum) Enum() map[string]CustomDomainDestinationEnum {
	return map[string]CustomDomainDestinationEnum{
		"site_alias":     CustomDomainDestinationEnum("site_alias"),
		"public_hosting": CustomDomainDestinationEnum("public_hosting"),
		"s3_endpoint":    CustomDomainDestinationEnum("s3_endpoint"),
		"unassigned":     CustomDomainDestinationEnum("unassigned"),
	}
}

type CustomDomainListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type CustomDomainFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type CustomDomainCreateParams struct {
	Destination      CustomDomainDestinationEnum `url:"destination,omitempty" json:"destination,omitempty" path:"destination"`
	FolderBehaviorId int64                       `url:"folder_behavior_id,omitempty" json:"folder_behavior_id,omitempty" path:"folder_behavior_id"`
	SslCertificateId int64                       `url:"ssl_certificate_id,omitempty" json:"ssl_certificate_id,omitempty" path:"ssl_certificate_id"`
	Domain           string                      `url:"domain" json:"domain" path:"domain"`
}

type CustomDomainUpdateParams struct {
	Id               int64                       `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Destination      CustomDomainDestinationEnum `url:"destination,omitempty" json:"destination,omitempty" path:"destination"`
	FolderBehaviorId int64                       `url:"folder_behavior_id,omitempty" json:"folder_behavior_id,omitempty" path:"folder_behavior_id"`
	SslCertificateId int64                       `url:"ssl_certificate_id,omitempty" json:"ssl_certificate_id,omitempty" path:"ssl_certificate_id"`
	Domain           string                      `url:"domain,omitempty" json:"domain,omitempty" path:"domain"`
}

type CustomDomainDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (c *CustomDomain) UnmarshalJSON(data []byte) error {
	type customDomain CustomDomain
	var v customDomain
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*c = CustomDomain(v)
	return nil
}

func (c *CustomDomainCollection) UnmarshalJSON(data []byte) error {
	type customDomains CustomDomainCollection
	var v customDomains
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*c = CustomDomainCollection(v)
	return nil
}

func (c *CustomDomainCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*c))
	for i, v := range *c {
		ret[i] = v
	}

	return &ret
}
