package lib

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

type ResponseError struct {
	StatusCode int
	err        error
}

func (r ResponseError) Error() string {
	return fmt.Sprintf("status code: %v - %v", r.StatusCode, r.err.Error())
}

func ResponseErrors(res *http.Response, errorFunc ...func(res *http.Response) error) error {
	for _, errFunc := range errorFunc {
		if err := errFunc(res); err != nil {
			return err
		}
	}
	return nil
}

func IsNonOkStatus(res *http.Response) bool {
	return !IsOkStatus(res)
}

func IsOkStatus(res *http.Response) bool {
	if res.StatusCode < 300 && res.StatusCode > 199 {
		return true
	}

	return false
}

func NonOkError(res *http.Response) error {
	if IsNonOkStatus(res) {
		return errorFromBodyDefault(res)
	}

	return nil
}

func NonOkErrorCustom(callbacks ...func(error) error) func(res *http.Response) error {
	return func(res *http.Response) error {
		if IsNonOkStatus(res) {
			return errorFromBody(res, callbacks)
		}
		return nil
	}
}

func NotStatus(status int) func(res *http.Response) error {
	return func(res *http.Response) error {
		if res.StatusCode != status {
			return errorFromBodyDefault(res)
		}

		return nil
	}
}

func IsStatus(status int) func(res *http.Response) error {
	return func(res *http.Response) error {
		if res.StatusCode == status {
			return errorFromBodyDefault(res)
		}

		return nil
	}
}

func IsHTML(res *http.Response) bool {
	return strings.Contains(res.Header.Get("Content-type"), "text/html")
}

func IsJSON(res *http.Response) bool {
	return strings.Contains(res.Header.Get("Content-type"), "application/json")
}

func IsXML(res *http.Response) bool {
	return strings.Contains(res.Header.Get("Content-type"), "application/xml") || strings.Contains(res.Header.Get("Content-type"), "text/xml")
}

func S3XMLError(res *http.Response) error {
	if IsNonOkStatus(res) && IsXML(res) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var s3Err S3Error
		if err := xml.Unmarshal(body, &s3Err); err != nil {
			return err
		}
		if s3Err.Empty() {
			return ResponseError{StatusCode: res.StatusCode, err: fmt.Errorf(strings.ReplaceAll(string(body), "\n", " "))}
		}
		return s3Err
	}
	return nil
}

func NonJSONError(res *http.Response) error {
	if !IsJSON(res) {
		if res.StatusCode == 204 {
			return nil
		}
		return errorFromBodyDefault(res)
	}

	return nil
}

func CloseBody(res *http.Response) {
	if res != nil && res.Body != nil {
		res.Body.Close()
	}
}

func errorFromBodyDefault(res *http.Response) error {
	var callbacks []func(error) error
	return errorFromBody(res, callbacks)
}

func errorFromBody(res *http.Response, callbacks []func(error) error) error {
	if IsHTML(res) {
		return ResponseError{StatusCode: res.StatusCode, err: fmt.Errorf(http.StatusText(res.StatusCode))}
	}
	var body []byte
	var err error
	if res.ContentLength == -1 {
		body = make([]byte, 512)
	} else {
		body = make([]byte, int(math.Min(float64(res.ContentLength), float64(512))))
	}
	_, err = io.ReadFull(res.Body, body)
	defer CloseBody(res)
	if err == nil || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		err = ResponseError{StatusCode: res.StatusCode, err: fmt.Errorf(strings.ReplaceAll(string(body), "\n", " "))}
	} else {
		err = ResponseError{StatusCode: res.StatusCode, err: fmt.Errorf(http.StatusText(res.StatusCode))}
	}

	for _, callback := range callbacks {
		err = callback(err)
	}

	return err
}
