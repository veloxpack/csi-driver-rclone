package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type MessageReaction struct {
	Id     int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Emoji  string `json:"emoji,omitempty" path:"emoji,omitempty" url:"emoji,omitempty"`
	UserId int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (m MessageReaction) Identifier() interface{} {
	return m.Id
}

type MessageReactionCollection []MessageReaction

type MessageReactionListParams struct {
	UserId    int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	MessageId int64 `url:"message_id" json:"message_id" path:"message_id"`
	ListParams
}

type MessageReactionFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type MessageReactionCreateParams struct {
	UserId int64  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Emoji  string `url:"emoji" json:"emoji" path:"emoji"`
}

type MessageReactionDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (m *MessageReaction) UnmarshalJSON(data []byte) error {
	type messageReaction MessageReaction
	var v messageReaction
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*m = MessageReaction(v)
	return nil
}

func (m *MessageReactionCollection) UnmarshalJSON(data []byte) error {
	type messageReactions MessageReactionCollection
	var v messageReactions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*m = MessageReactionCollection(v)
	return nil
}

func (m *MessageReactionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*m))
	for i, v := range *m {
		ret[i] = v
	}

	return &ret
}
