package files_sdk

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

type requestResponseOption struct {
	*http.Request
	*http.Response
	context.Context
}

// RequestResponseOption customizes an individual SDK request or response.
// WithContext is the normal public use. Request/header mutation options are
// escape hatches for custom HTTP behavior, not the primary way to authenticate
// or scope Files.com API calls; use Config.APIKey, Config.SessionId, and
// Config.WorkspaceId for that. Request mutation options are not guaranteed
// through every higher-level helper path.
type RequestResponseOption func(*requestResponseOption) error

// RequestOption customizes SDK HTTP requests.
func RequestOption(call func(req *http.Request) error) RequestResponseOption {
	return func(opt *requestResponseOption) error {
		if opt.Request != nil {
			return call(opt.Request)
		}
		return nil
	}
}

func ResponseOption(call func(req *http.Response) error) RequestResponseOption {
	return func(opt *requestResponseOption) error {
		if opt.Response != nil {
			return call(opt.Response)
		}
		return nil
	}
}

// RequestHeadersOption customizes SDK HTTP request headers.
func RequestHeadersOption(headers *http.Header) RequestResponseOption {
	return RequestOption(func(req *http.Request) error {
		for k, v := range *headers {
			req.Header.Set(k, v[0])
		}
		return nil
	})
}

// WithWorkspaceId sets the workspace header for a request.
//
// Deprecated: use Config.WorkspaceId; request/header mutations, including API-key headers, are not guaranteed through every helper.
func WithWorkspaceId(workspaceId int64) RequestResponseOption {
	return RequestOption(func(req *http.Request) error {
		req.Header.Set(workspaceIdHeader, strconv.FormatInt(workspaceId, 10))
		return nil
	})
}

// WithoutWorkspaceId removes the workspace header from a request.
//
// Deprecated: use Config without WorkspaceId; request/header mutations, including API-key headers, are not guaranteed through every helper.
func WithoutWorkspaceId() RequestResponseOption {
	return RequestOption(func(req *http.Request) error {
		req.Header.Del(workspaceIdHeader)
		return nil
	})
}

// WithContext sets the context for an SDK request.
func WithContext(ctx context.Context) RequestResponseOption {
	return func(opt *requestResponseOption) error {
		if opt.Request != nil && ctx != nil {
			opt.Request = opt.Request.WithContext(ctx)
		} else {
			opt.Context = ctx
		}
		return nil
	}
}

func ResponseBodyOption(opt func(io.ReadCloser) error) RequestResponseOption {
	return ResponseOption(func(res *http.Response) error {
		return opt(res.Body)
	})
}

func WrapRequestOptions(config Config, request *http.Request, opts ...RequestResponseOption) (*http.Response, error) {
	// Apply request options
	modifiedRequest, err := BuildRequest(request, opts...)
	if err != nil {
		return &http.Response{}, err
	}

	// Execute the request
	response, err := config.Do(modifiedRequest)
	if err != nil {
		return response, err
	}

	return BuildResponse(response, opts...)
}

func ContextOption(opts []RequestResponseOption) context.Context {
	params := &requestResponseOption{}
	for _, opt := range opts {
		opt(params)
	}
	if params.Context == nil {
		params.Context = context.Background()
	}
	return params.Context
}

// BuildRequest applies request options to an HTTP request and returns the modified request.
// This is useful for tests and code that need to build requests with options applied.
func BuildRequest(request *http.Request, opts ...RequestResponseOption) (*http.Request, error) {
	optionRequestResponse := &requestResponseOption{Request: request}

	for _, opt := range opts {
		if err := opt(optionRequestResponse); err != nil {
			return nil, err
		}
	}

	return optionRequestResponse.Request, nil
}

// BuildResponse applies response options to an HTTP response and returns the modified response.
// This is useful for tests and code that need to process responses with options applied.
func BuildResponse(response *http.Response, opts ...RequestResponseOption) (*http.Response, error) {
	optionRequestResponse := &requestResponseOption{Response: response}

	for _, opt := range opts {
		if err := opt(optionRequestResponse); err != nil {
			return nil, err
		}
	}

	return optionRequestResponse.Response, nil
}
