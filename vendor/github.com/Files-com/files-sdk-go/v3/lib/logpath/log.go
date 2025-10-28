package logpath

import (
	"fmt"

	"github.com/Files-com/files-sdk-go/v3/lib/keyvalue"
)

func New(path string, args map[string]interface{}) string {
	return LogPath{Path: path, Args: args}.String()
}

type LogPath struct {
	Path string
	Args map[string]interface{}
}

func (l LogPath) String() string {
	return fmt.Sprintf("%v - %v", l.Path, keyvalue.New(l.Args))
}
