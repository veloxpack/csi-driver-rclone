package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ChatMessage struct {
	Id        int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Role      string     `json:"role,omitempty" path:"role,omitempty" url:"role,omitempty"`
	Content   string     `json:"content,omitempty" path:"content,omitempty" url:"content,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

func (c ChatMessage) Identifier() interface{} {
	return c.Id
}

type ChatMessageCollection []ChatMessage

func (c *ChatMessage) UnmarshalJSON(data []byte) error {
	type chatMessage ChatMessage
	var v chatMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*c = ChatMessage(v)
	return nil
}

func (c *ChatMessageCollection) UnmarshalJSON(data []byte) error {
	type chatMessages ChatMessageCollection
	var v chatMessages
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*c = ChatMessageCollection(v)
	return nil
}

func (c *ChatMessageCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*c))
	for i, v := range *c {
		ret[i] = v
	}

	return &ret
}
