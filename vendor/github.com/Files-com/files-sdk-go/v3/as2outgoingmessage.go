package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type As2OutgoingMessage struct {
	Id                          int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	As2PartnerId                int64                  `json:"as2_partner_id,omitempty" path:"as2_partner_id,omitempty" url:"as2_partner_id,omitempty"`
	As2StationId                int64                  `json:"as2_station_id,omitempty" path:"as2_station_id,omitempty" url:"as2_station_id,omitempty"`
	Uuid                        string                 `json:"uuid,omitempty" path:"uuid,omitempty" url:"uuid,omitempty"`
	HttpHeaders                 map[string]interface{} `json:"http_headers,omitempty" path:"http_headers,omitempty" url:"http_headers,omitempty"`
	ProcessingResult            string                 `json:"processing_result,omitempty" path:"processing_result,omitempty" url:"processing_result,omitempty"`
	ProcessingResultDescription string                 `json:"processing_result_description,omitempty" path:"processing_result_description,omitempty" url:"processing_result_description,omitempty"`
	Mic                         string                 `json:"mic,omitempty" path:"mic,omitempty" url:"mic,omitempty"`
	MicSha256                   string                 `json:"mic_sha_256,omitempty" path:"mic_sha_256,omitempty" url:"mic_sha_256,omitempty"`
	As2To                       string                 `json:"as2_to,omitempty" path:"as2_to,omitempty" url:"as2_to,omitempty"`
	As2From                     string                 `json:"as2_from,omitempty" path:"as2_from,omitempty" url:"as2_from,omitempty"`
	Date                        string                 `json:"date,omitempty" path:"date,omitempty" url:"date,omitempty"`
	MessageId                   string                 `json:"message_id,omitempty" path:"message_id,omitempty" url:"message_id,omitempty"`
	BodySize                    string                 `json:"body_size,omitempty" path:"body_size,omitempty" url:"body_size,omitempty"`
	AttachmentFilename          string                 `json:"attachment_filename,omitempty" path:"attachment_filename,omitempty" url:"attachment_filename,omitempty"`
	CreatedAt                   *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	HttpResponseCode            string                 `json:"http_response_code,omitempty" path:"http_response_code,omitempty" url:"http_response_code,omitempty"`
	HttpResponseHeaders         map[string]interface{} `json:"http_response_headers,omitempty" path:"http_response_headers,omitempty" url:"http_response_headers,omitempty"`
	HttpTransmissionDuration    string                 `json:"http_transmission_duration,omitempty" path:"http_transmission_duration,omitempty" url:"http_transmission_duration,omitempty"`
	MdnReceived                 *bool                  `json:"mdn_received,omitempty" path:"mdn_received,omitempty" url:"mdn_received,omitempty"`
	MdnValid                    *bool                  `json:"mdn_valid,omitempty" path:"mdn_valid,omitempty" url:"mdn_valid,omitempty"`
	MdnSignatureVerified        *bool                  `json:"mdn_signature_verified,omitempty" path:"mdn_signature_verified,omitempty" url:"mdn_signature_verified,omitempty"`
	MdnMessageIdMatched         *bool                  `json:"mdn_message_id_matched,omitempty" path:"mdn_message_id_matched,omitempty" url:"mdn_message_id_matched,omitempty"`
	MdnMicMatched               *bool                  `json:"mdn_mic_matched,omitempty" path:"mdn_mic_matched,omitempty" url:"mdn_mic_matched,omitempty"`
	MdnProcessingSuccess        *bool                  `json:"mdn_processing_success,omitempty" path:"mdn_processing_success,omitempty" url:"mdn_processing_success,omitempty"`
	RawUri                      string                 `json:"raw_uri,omitempty" path:"raw_uri,omitempty" url:"raw_uri,omitempty"`
	SmimeUri                    string                 `json:"smime_uri,omitempty" path:"smime_uri,omitempty" url:"smime_uri,omitempty"`
	SmimeSignedUri              string                 `json:"smime_signed_uri,omitempty" path:"smime_signed_uri,omitempty" url:"smime_signed_uri,omitempty"`
	EncryptedUri                string                 `json:"encrypted_uri,omitempty" path:"encrypted_uri,omitempty" url:"encrypted_uri,omitempty"`
	MdnResponseUri              string                 `json:"mdn_response_uri,omitempty" path:"mdn_response_uri,omitempty" url:"mdn_response_uri,omitempty"`
}

func (a As2OutgoingMessage) Identifier() interface{} {
	return a.Id
}

type As2OutgoingMessageCollection []As2OutgoingMessage

type As2OutgoingMessageListParams struct {
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       As2OutgoingMessage     `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	As2PartnerId int64                  `url:"as2_partner_id,omitempty" json:"as2_partner_id,omitempty" path:"as2_partner_id"`
	ListParams
}

func (a *As2OutgoingMessage) UnmarshalJSON(data []byte) error {
	type as2OutgoingMessage As2OutgoingMessage
	var v as2OutgoingMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = As2OutgoingMessage(v)
	return nil
}

func (a *As2OutgoingMessageCollection) UnmarshalJSON(data []byte) error {
	type as2OutgoingMessages As2OutgoingMessageCollection
	var v as2OutgoingMessages
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = As2OutgoingMessageCollection(v)
	return nil
}

func (a *As2OutgoingMessageCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
