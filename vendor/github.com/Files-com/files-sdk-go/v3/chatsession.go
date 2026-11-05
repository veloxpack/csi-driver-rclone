package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ChatSession struct {
	Id           string        `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Title        string        `json:"title,omitempty" path:"title,omitempty" url:"title,omitempty"`
	UserId       int64         `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	AiTaskId     int64         `json:"ai_task_id,omitempty" path:"ai_task_id,omitempty" url:"ai_task_id,omitempty"`
	WorkspaceId  int64         `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	LastActiveAt *time.Time    `json:"last_active_at,omitempty" path:"last_active_at,omitempty" url:"last_active_at,omitempty"`
	CreatedAt    *time.Time    `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Messages     []ChatMessage `json:"messages,omitempty" path:"messages,omitempty" url:"messages,omitempty"`
}

func (c ChatSession) Identifier() interface{} {
	return c.Id
}

type ChatSessionCollection []ChatSession

type ChatSessionListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type ChatSessionFindParams struct {
	Id string `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (c *ChatSession) UnmarshalJSON(data []byte) error {
	type chatSession ChatSession
	var v chatSession
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*c = ChatSession(v)
	return nil
}

func (c *ChatSessionCollection) UnmarshalJSON(data []byte) error {
	type chatSessions ChatSessionCollection
	var v chatSessions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*c = ChatSessionCollection(v)
	return nil
}

func (c *ChatSessionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*c))
	for i, v := range *c {
		ret[i] = v
	}

	return &ret
}
