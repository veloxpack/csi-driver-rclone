package lib

import (
	"encoding/json"
	"errors"
)

type ErrorWithOriginalResponse struct {
	error
	originalResponse interface{}
}

func (u ErrorWithOriginalResponse) OriginalResponse() interface{} {
	return u.originalResponse
}

func (u ErrorWithOriginalResponse) ProcessError(data []byte, err error, t interface{}) error {
	var unmarshalError *json.UnmarshalTypeError
	ok := errors.As(err, &unmarshalError)

	if ok {
		ignoreErr := json.Unmarshal(data, &t)
		if ignoreErr == nil {
			return ErrorWithOriginalResponse{error: unmarshalError, originalResponse: t}
		}
	}
	var errorWithOriginalResponse ErrorWithOriginalResponse
	ok = errors.As(err, &errorWithOriginalResponse)
	if ok {
		ignoreErr := json.Unmarshal(data, &t)
		if ignoreErr == nil {
			return ErrorWithOriginalResponse{error: errorWithOriginalResponse.error, originalResponse: t}
		}
	}
	return err
}
