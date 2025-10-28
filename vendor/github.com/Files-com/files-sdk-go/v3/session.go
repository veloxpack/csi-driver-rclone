package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Session struct {
	Id                  string `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Language            string `json:"language,omitempty" path:"language,omitempty" url:"language,omitempty"`
	ReadOnly            *bool  `json:"read_only,omitempty" path:"read_only,omitempty" url:"read_only,omitempty"`
	SftpInsecureCiphers *bool  `json:"sftp_insecure_ciphers,omitempty" path:"sftp_insecure_ciphers,omitempty" url:"sftp_insecure_ciphers,omitempty"`
	Username            string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	Password            string `json:"password,omitempty" path:"password,omitempty" url:"password,omitempty"`
	Otp                 string `json:"otp,omitempty" path:"otp,omitempty" url:"otp,omitempty"`
	PartialSessionId    string `json:"partial_session_id,omitempty" path:"partial_session_id,omitempty" url:"partial_session_id,omitempty"`
}

func (s Session) Identifier() interface{} {
	return s.Id
}

type SessionCollection []Session

type SessionCreateParams struct {
	Username         string `url:"username,omitempty" json:"username,omitempty" path:"username"`
	Password         string `url:"password,omitempty" json:"password,omitempty" path:"password"`
	Otp              string `url:"otp,omitempty" json:"otp,omitempty" path:"otp"`
	PartialSessionId string `url:"partial_session_id,omitempty" json:"partial_session_id,omitempty" path:"partial_session_id"`
}

func (s *Session) UnmarshalJSON(data []byte) error {
	type session Session
	var v session
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Session(v)
	return nil
}

func (s *SessionCollection) UnmarshalJSON(data []byte) error {
	type sessions SessionCollection
	var v sessions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SessionCollection(v)
	return nil
}

func (s *SessionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
