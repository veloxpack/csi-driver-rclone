package files_sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/hashicorp/go-retryablehttp"
	"moul.io/http2curl/v2"
)

func Resource(config Config, resource lib.Resource, opt ...RequestResponseOption) error {
	out, err := resource.Out()
	if err != nil {
		return err
	}
	data, res, err := Call(resource.Method, config, out.Path, out.Values, opt...)
	defer lib.CloseBody(res)
	if err != nil {
		return err
	}
	if res.StatusCode == 204 {
		return err
	}

	return out.Entity.UnmarshalJSON(*data)
}

func Call(method string, config Config, resource string, params lib.Values, opts ...RequestResponseOption) (*[]byte, *http.Response, error) {
	defaultHeaders := &http.Header{}
	config.SetHeaders(defaultHeaders)
	callParams := &CallParams{
		Method:  method,
		Config:  config,
		Uri:     config.RootPath() + resource,
		Params:  params,
		Headers: defaultHeaders,
	}
	request, err := buildRequest(callParams)
	if err != nil {
		return nil, &http.Response{}, err
	}
	response, err := WrapRequestOptions(config, request, opts...)
	if err != nil {
		return nil, response, err
	}
	data, res, err := ParseResponse(response, resource)
	var responseError ResponseError
	ok := errors.As(err, &responseError)
	if ok {
		err = responseError
	}
	return data, res, err
}

func ParseResponse(res *http.Response, resource string) (*[]byte, *http.Response, error) {
	defaultValue := make([]byte, 0)
	if res.StatusCode == 204 {
		return &defaultValue, res, nil
	}
	nonOkError := lib.NonOkErrorCustom(func(err error) error {
		return fmt.Errorf("%v - %v", resource, err)
	})

	if err := lib.ResponseErrors(res, APIError(), nonOkError); err != nil {
		return &defaultValue, res, err
	}
	data, err := io.ReadAll(res.Body)
	return &data, res, err
}

type CallParams struct {
	Method  string
	Config  Config
	Uri     string
	Params  lib.Values
	BodyIo  io.ReadCloser
	Headers *http.Header
	context.Context
}

func CallRaw(params *CallParams) (*http.Response, error) {
	request, err := buildRequest(params)
	if err != nil {
		return &http.Response{}, err
	}
	retryRequest := &retryablehttp.Request{Request: request}
	retryRequest.Body = request.Body
	return params.Config.Client.Do(retryRequest)
}

func buildRequest(opts *CallParams) (*http.Request, error) {
	var bodyIsJson bool
	if opts.Headers == nil {
		opts.Headers = &http.Header{}
	}

	var req *http.Request
	var err error
	if opts.Context != nil {
		req, err = http.NewRequestWithContext(opts.Context, opts.Method, opts.Uri, nil)
	} else {
		req, err = http.NewRequest(opts.Method, opts.Uri, nil)
	}

	if err != nil {
		return &http.Request{}, err
	}
	req.Header = *opts.Headers
	if req.Header.Get("Content-Length") != "" {
		c, _ := strconv.ParseInt(req.Header.Get("Content-Length"), 10, 64)
		req.ContentLength = c
	}

	switch opts.Method {
	case "GET", "HEAD", "DELETE":
		if opts.Params != nil {
			values, err := opts.Params.ToValues()
			if err != nil {
				return nil, err
			}
			req.URL.RawQuery = values.Encode()
		}
	default:
		if opts.BodyIo == nil {
			bodyIsJson = true
			jsonBody, err := opts.Params.ToJSON()
			if err != nil {
				return &http.Request{}, err
			}
			req.Body = io.NopCloser(jsonBody)
			req.Header.Add("Content-Type", "application/json")
		} else {
			if req.ContentLength != 0 {
				req.Body = opts.BodyIo
			}
		}
	}

	if opts.Config.InDebug() {
		defer debugLog(bodyIsJson, req, opts.Config, opts.Params)
	}
	return req, nil
}

func debugLog(bodyIsJson bool, req *http.Request, config Config, params lib.Values) {
	clonedReq := req.Clone(context.Background())
	clonedReq.Body = nil
	if bodyIsJson {
		jsonBody, err := params.ToJSON()
		if err != nil {
			panic(err)
		}
		clonedReq.Body = io.NopCloser(jsonBody)
	}
	command, err := http2curl.GetCurlCommand(clonedReq)
	if err != nil {
		panic(err)
	}
	config.Printf(" %v", command)
}
