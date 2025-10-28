package keyvalue

import (
	"fmt"
	"strings"
)

func New(args map[string]interface{}) string {
	return KeyValue{Args: args}.String()
}

type KeyValue struct {
	Args map[string]interface{}
}

func (a KeyValue) String() string {
	var argStr []string
	for k, v := range a.Args {
		argStr = append(argStr, fmt.Sprintf("%v: %v", k, v))
	}

	return strings.Join(argStr, ", ")
}
