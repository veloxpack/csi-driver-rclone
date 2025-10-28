package files_sdk

import (
	"context"
	"io"
	"net/http"
)

type requestResponseOption struct {
	*http.Request
	*http.Response
	context.Context
}

type RequestResponseOption func(*requestResponseOption) error

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

func RequestHeadersOption(headers *http.Header) RequestResponseOption {
	return RequestOption(func(req *http.Request) error {
		for k, v := range *headers {
			req.Header.Set(k, v[0])
		}
		return nil
	})
}

func WithContext(ctx context.Context) RequestResponseOption {
	return func(opt *requestResponseOption) error {
		if opt.Request != nil && ctx != nil {
			opt.Request.WithContext(ctx)
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
	for _, opt := range opts {
		err := opt(&requestResponseOption{Request: request})
		if err != nil {
			return &http.Response{}, err
		}
	}

	resp, err := config.Do(request)
	if err != nil {
		return resp, err
	}

	for _, opt := range opts {
		err := opt(&requestResponseOption{Response: resp})
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
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
