package listquery

import (
	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

type List interface {
	UnmarshalJSON(data []byte) error
	ToSlice() *[]interface{}
}

func Build(config files_sdk.Config, path string, list List, opts ...files_sdk.RequestResponseOption) func(params lib.Values, laterOpts ...files_sdk.RequestResponseOption) (*[]interface{}, string, error) {
	return func(params lib.Values, laterOpts ...files_sdk.RequestResponseOption) (*[]interface{}, string, error) {
		defaultValue := make([]interface{}, 0)
		data, res, err := files_sdk.Call("GET", config, path, params, append(opts, laterOpts...)...)
		defer func() {
			if res != nil && res.Body != nil {
				res.Body.Close()
			}
		}()
		if err != nil {
			return &defaultValue, "", err
		}

		if err := lib.ResponseErrors(res, lib.NonOkError, lib.NonJSONError); err != nil {
			return &defaultValue, "", err
		}

		if err := list.UnmarshalJSON(*data); err != nil {
			return &defaultValue, res.Header.Get("X-Files-Cursor"), err
		}
		return list.ToSlice(), res.Header.Get("X-Files-Cursor"), nil
	}
}
