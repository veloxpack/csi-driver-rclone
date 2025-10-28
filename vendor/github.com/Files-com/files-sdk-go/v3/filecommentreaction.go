package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FileCommentReaction struct {
	Id            int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Emoji         string `json:"emoji,omitempty" path:"emoji,omitempty" url:"emoji,omitempty"`
	UserId        int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	FileCommentId int64  `json:"file_comment_id,omitempty" path:"file_comment_id,omitempty" url:"file_comment_id,omitempty"`
}

func (f FileCommentReaction) Identifier() interface{} {
	return f.Id
}

type FileCommentReactionCollection []FileCommentReaction

type FileCommentReactionCreateParams struct {
	UserId        int64  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	FileCommentId int64  `url:"file_comment_id" json:"file_comment_id" path:"file_comment_id"`
	Emoji         string `url:"emoji" json:"emoji" path:"emoji"`
}

type FileCommentReactionDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (f *FileCommentReaction) UnmarshalJSON(data []byte) error {
	type fileCommentReaction FileCommentReaction
	var v fileCommentReaction
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FileCommentReaction(v)
	return nil
}

func (f *FileCommentReactionCollection) UnmarshalJSON(data []byte) error {
	type fileCommentReactions FileCommentReactionCollection
	var v fileCommentReactions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileCommentReactionCollection(v)
	return nil
}

func (f *FileCommentReactionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
