package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type MessageComment struct {
	Id        int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Body      string                   `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	Reactions []map[string]interface{} `json:"reactions,omitempty" path:"reactions,omitempty" url:"reactions,omitempty"`
	UserId    int64                    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (m MessageComment) Identifier() interface{} {
	return m.Id
}

type MessageCommentCollection []MessageComment

type MessageCommentListParams struct {
	UserId    int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy    map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	MessageId int64                  `url:"message_id" json:"message_id" path:"message_id"`
	ListParams
}

type MessageCommentFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type MessageCommentCreateParams struct {
	UserId int64  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Body   string `url:"body" json:"body" path:"body"`
}

type MessageCommentUpdateParams struct {
	Id   int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Body string `url:"body" json:"body" path:"body"`
}

type MessageCommentDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (m *MessageComment) UnmarshalJSON(data []byte) error {
	type messageComment MessageComment
	var v messageComment
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*m = MessageComment(v)
	return nil
}

func (m *MessageCommentCollection) UnmarshalJSON(data []byte) error {
	type messageComments MessageCommentCollection
	var v messageComments
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*m = MessageCommentCollection(v)
	return nil
}

func (m *MessageCommentCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*m))
	for i, v := range *m {
		ret[i] = v
	}

	return &ret
}
