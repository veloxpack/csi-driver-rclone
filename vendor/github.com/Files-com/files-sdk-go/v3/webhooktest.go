package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type WebhookTest struct {
	Code            int64                  `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	Message         string                 `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	Status          string                 `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Data            Auto                   `json:"data,omitempty" path:"data,omitempty" url:"data,omitempty"`
	Success         *bool                  `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	Url             string                 `json:"url,omitempty" path:"url,omitempty" url:"url,omitempty"`
	Method          string                 `json:"method,omitempty" path:"method,omitempty" url:"method,omitempty"`
	Encoding        string                 `json:"encoding,omitempty" path:"encoding,omitempty" url:"encoding,omitempty"`
	Headers         map[string]interface{} `json:"headers,omitempty" path:"headers,omitempty" url:"headers,omitempty"`
	Body            map[string]interface{} `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	RawBody         string                 `json:"raw_body,omitempty" path:"raw_body,omitempty" url:"raw_body,omitempty"`
	FileAsBody      *bool                  `json:"file_as_body,omitempty" path:"file_as_body,omitempty" url:"file_as_body,omitempty"`
	FileFormField   string                 `json:"file_form_field,omitempty" path:"file_form_field,omitempty" url:"file_form_field,omitempty"`
	Action          string                 `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	UseDedicatedIps *bool                  `json:"use_dedicated_ips,omitempty" path:"use_dedicated_ips,omitempty" url:"use_dedicated_ips,omitempty"`
}

// Identifier no path or id

type WebhookTestCollection []WebhookTest

type WebhookTestCreateParams struct {
	Url             string                 `url:"url" json:"url" path:"url"`
	Method          string                 `url:"method,omitempty" json:"method,omitempty" path:"method"`
	Encoding        string                 `url:"encoding,omitempty" json:"encoding,omitempty" path:"encoding"`
	Headers         map[string]interface{} `url:"headers,omitempty" json:"headers,omitempty" path:"headers"`
	Body            map[string]interface{} `url:"body,omitempty" json:"body,omitempty" path:"body"`
	RawBody         string                 `url:"raw_body,omitempty" json:"raw_body,omitempty" path:"raw_body"`
	FileAsBody      *bool                  `url:"file_as_body,omitempty" json:"file_as_body,omitempty" path:"file_as_body"`
	FileFormField   string                 `url:"file_form_field,omitempty" json:"file_form_field,omitempty" path:"file_form_field"`
	Action          string                 `url:"action,omitempty" json:"action,omitempty" path:"action"`
	UseDedicatedIps *bool                  `url:"use_dedicated_ips,omitempty" json:"use_dedicated_ips,omitempty" path:"use_dedicated_ips"`
}

func (w *WebhookTest) UnmarshalJSON(data []byte) error {
	type webhookTest WebhookTest
	var v webhookTest
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*w = WebhookTest(v)
	return nil
}

func (w *WebhookTestCollection) UnmarshalJSON(data []byte) error {
	type webhookTests WebhookTestCollection
	var v webhookTests
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*w = WebhookTestCollection(v)
	return nil
}

func (w *WebhookTestCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*w))
	for i, v := range *w {
		ret[i] = v
	}

	return &ret
}
