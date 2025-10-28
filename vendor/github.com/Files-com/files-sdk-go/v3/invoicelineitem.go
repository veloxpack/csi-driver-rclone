package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type InvoiceLineItem struct {
	Id                    int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Amount                string     `json:"amount,omitempty" path:"amount,omitempty" url:"amount,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Description           string     `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	Type                  string     `json:"type,omitempty" path:"type,omitempty" url:"type,omitempty"`
	ServiceEndAt          *time.Time `json:"service_end_at,omitempty" path:"service_end_at,omitempty" url:"service_end_at,omitempty"`
	ServiceStartAt        *time.Time `json:"service_start_at,omitempty" path:"service_start_at,omitempty" url:"service_start_at,omitempty"`
	Plan                  string     `json:"plan,omitempty" path:"plan,omitempty" url:"plan,omitempty"`
	Site                  string     `json:"site,omitempty" path:"site,omitempty" url:"site,omitempty"`
	PrepaidBytes          int64      `json:"prepaid_bytes,omitempty" path:"prepaid_bytes,omitempty" url:"prepaid_bytes,omitempty"`
	PrepaidBytesExpireAt  *time.Time `json:"prepaid_bytes_expire_at,omitempty" path:"prepaid_bytes_expire_at,omitempty" url:"prepaid_bytes_expire_at,omitempty"`
	PrepaidBytesUsed      int64      `json:"prepaid_bytes_used,omitempty" path:"prepaid_bytes_used,omitempty" url:"prepaid_bytes_used,omitempty"`
	PrepaidBytesAvailable int64      `json:"prepaid_bytes_available,omitempty" path:"prepaid_bytes_available,omitempty" url:"prepaid_bytes_available,omitempty"`
}

func (i InvoiceLineItem) Identifier() interface{} {
	return i.Id
}

type InvoiceLineItemCollection []InvoiceLineItem

func (i *InvoiceLineItem) UnmarshalJSON(data []byte) error {
	type invoiceLineItem InvoiceLineItem
	var v invoiceLineItem
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = InvoiceLineItem(v)
	return nil
}

func (i *InvoiceLineItemCollection) UnmarshalJSON(data []byte) error {
	type invoiceLineItems InvoiceLineItemCollection
	var v invoiceLineItems
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = InvoiceLineItemCollection(v)
	return nil
}

func (i *InvoiceLineItemCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
