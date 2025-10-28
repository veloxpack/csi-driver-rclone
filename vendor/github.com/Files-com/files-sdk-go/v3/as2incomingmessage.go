package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type As2IncomingMessage struct {
	Id                          int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	As2PartnerId                int64                  `json:"as2_partner_id,omitempty" path:"as2_partner_id,omitempty" url:"as2_partner_id,omitempty"`
	As2StationId                int64                  `json:"as2_station_id,omitempty" path:"as2_station_id,omitempty" url:"as2_station_id,omitempty"`
	Uuid                        string                 `json:"uuid,omitempty" path:"uuid,omitempty" url:"uuid,omitempty"`
	ContentType                 string                 `json:"content_type,omitempty" path:"content_type,omitempty" url:"content_type,omitempty"`
	HttpHeaders                 map[string]interface{} `json:"http_headers,omitempty" path:"http_headers,omitempty" url:"http_headers,omitempty"`
	ProcessingResult            string                 `json:"processing_result,omitempty" path:"processing_result,omitempty" url:"processing_result,omitempty"`
	ProcessingResultDescription string                 `json:"processing_result_description,omitempty" path:"processing_result_description,omitempty" url:"processing_result_description,omitempty"`
	Mic                         string                 `json:"mic,omitempty" path:"mic,omitempty" url:"mic,omitempty"`
	MicAlgo                     string                 `json:"mic_algo,omitempty" path:"mic_algo,omitempty" url:"mic_algo,omitempty"`
	As2To                       string                 `json:"as2_to,omitempty" path:"as2_to,omitempty" url:"as2_to,omitempty"`
	As2From                     string                 `json:"as2_from,omitempty" path:"as2_from,omitempty" url:"as2_from,omitempty"`
	MessageId                   string                 `json:"message_id,omitempty" path:"message_id,omitempty" url:"message_id,omitempty"`
	Subject                     string                 `json:"subject,omitempty" path:"subject,omitempty" url:"subject,omitempty"`
	Date                        string                 `json:"date,omitempty" path:"date,omitempty" url:"date,omitempty"`
	BodySize                    string                 `json:"body_size,omitempty" path:"body_size,omitempty" url:"body_size,omitempty"`
	AttachmentFilename          string                 `json:"attachment_filename,omitempty" path:"attachment_filename,omitempty" url:"attachment_filename,omitempty"`
	Ip                          string                 `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	CreatedAt                   *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	HttpResponseCode            string                 `json:"http_response_code,omitempty" path:"http_response_code,omitempty" url:"http_response_code,omitempty"`
	HttpResponseHeaders         map[string]interface{} `json:"http_response_headers,omitempty" path:"http_response_headers,omitempty" url:"http_response_headers,omitempty"`
	RecipientSerial             string                 `json:"recipient_serial,omitempty" path:"recipient_serial,omitempty" url:"recipient_serial,omitempty"`
	HexRecipientSerial          string                 `json:"hex_recipient_serial,omitempty" path:"hex_recipient_serial,omitempty" url:"hex_recipient_serial,omitempty"`
	RecipientIssuer             string                 `json:"recipient_issuer,omitempty" path:"recipient_issuer,omitempty" url:"recipient_issuer,omitempty"`
	MessageReceived             *bool                  `json:"message_received,omitempty" path:"message_received,omitempty" url:"message_received,omitempty"`
	MessageDecrypted            *bool                  `json:"message_decrypted,omitempty" path:"message_decrypted,omitempty" url:"message_decrypted,omitempty"`
	MessageSignatureVerified    *bool                  `json:"message_signature_verified,omitempty" path:"message_signature_verified,omitempty" url:"message_signature_verified,omitempty"`
	MessageProcessingSuccess    *bool                  `json:"message_processing_success,omitempty" path:"message_processing_success,omitempty" url:"message_processing_success,omitempty"`
	MessageMdnReturned          *bool                  `json:"message_mdn_returned,omitempty" path:"message_mdn_returned,omitempty" url:"message_mdn_returned,omitempty"`
	EncryptedUri                string                 `json:"encrypted_uri,omitempty" path:"encrypted_uri,omitempty" url:"encrypted_uri,omitempty"`
	SmimeSignedUri              string                 `json:"smime_signed_uri,omitempty" path:"smime_signed_uri,omitempty" url:"smime_signed_uri,omitempty"`
	SmimeUri                    string                 `json:"smime_uri,omitempty" path:"smime_uri,omitempty" url:"smime_uri,omitempty"`
	RawUri                      string                 `json:"raw_uri,omitempty" path:"raw_uri,omitempty" url:"raw_uri,omitempty"`
	MdnResponseUri              string                 `json:"mdn_response_uri,omitempty" path:"mdn_response_uri,omitempty" url:"mdn_response_uri,omitempty"`
}

func (a As2IncomingMessage) Identifier() interface{} {
	return a.Id
}

type As2IncomingMessageCollection []As2IncomingMessage

type As2IncomingMessageListParams struct {
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       As2IncomingMessage     `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	As2PartnerId int64                  `url:"as2_partner_id,omitempty" json:"as2_partner_id,omitempty" path:"as2_partner_id"`
	ListParams
}

func (a *As2IncomingMessage) UnmarshalJSON(data []byte) error {
	type as2IncomingMessage As2IncomingMessage
	var v as2IncomingMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = As2IncomingMessage(v)
	return nil
}

func (a *As2IncomingMessageCollection) UnmarshalJSON(data []byte) error {
	type as2IncomingMessages As2IncomingMessageCollection
	var v as2IncomingMessages
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = As2IncomingMessageCollection(v)
	return nil
}

func (a *As2IncomingMessageCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
