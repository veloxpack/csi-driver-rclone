// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/retry"
	"golang.org/x/oauth2"
)

const (
	apiVersion    = 2
	defaultDomain = ".dropboxapi.com"
	hostAPI       = "api"
	hostContent   = "content"
	hostNotify    = "notify"
	sdkVersion    = "6.4.0"
	specVersion   = "b5f7ddd"
)

// Version returns the current SDK version and API Spec version
func Version() (string, string) {
	return sdkVersion, specVersion
}

// Tagged is used for tagged unions.
type Tagged struct {
	Tag string `json:".tag"`
}

// APIError is the base type for endpoint-specific errors.
type APIError struct {
	ErrorSummary string `json:"error_summary"`
}

func (e APIError) Error() string {
	return e.ErrorSummary
}

// SDKInternalError is returned when the Dropbox API responds with a status code
// that does not carry a typed error body (e.g. 500 Internal Server Error).
type SDKInternalError struct {
	StatusCode int
	Content    string
}

func (e SDKInternalError) Error() string {
	return fmt.Sprintf("Unexpected error: %v (code: %v)", e.Content, e.StatusCode)
}

// Config contains parameters for configuring the SDK.
type Config struct {
	// OAuth2 access token
	Token string
	// Logging level for SDK generated logs
	LogLevel LogLevel
	// Logging target for verbose SDK logging
	Logger *log.Logger
	// Used with APIs that support operations as another user
	AsMemberID string
	// Used with APIs that support operations as an admin
	AsAdminID string
	// Path relative to which action should be taken
	PathRoot string
	// Dropbox API domain. The default is ".dropboxapi.com".
	Domain string
	// HTTP client used for authenticated Dropbox API calls. If set, Client takes
	// precedence over TokenSource and Token.
	Client *http.Client
	// RetryPolicy controls automatic retries for Dropbox API calls. If nil, retries
	// are disabled.
	RetryPolicy *retry.Policy
	// No need to set -- for testing only
	HeaderGenerator func(hostType string, namespace string, route string) map[string]string
	// No need to set -- for testing only
	URLGenerator func(hostType string, namespace string, route string) string
	// OAuth2 token source used for authenticated Dropbox API calls. If Client is
	// nil and TokenSource is set, TokenSource takes precedence over Token.
	TokenSource oauth2.TokenSource
}

// LogLevel defines a type that can set the desired level of logging the SDK will generate.
type LogLevel uint

const (
	// LogOff will disable all SDK logging. This is the default log level
	LogOff LogLevel = iota * (1 << 8)
	// LogDebug will enable detailed SDK debug logs. It will log requests (including arguments),
	// response and body contents.
	LogDebug
	// LogInfo will log SDK request (not including arguments) and responses.
	LogInfo
)

func (l LogLevel) shouldLog(v LogLevel) bool {
	return l > v || l&v == v
}

func (c *Config) doLog(l LogLevel, format string, v ...interface{}) {
	if !c.LogLevel.shouldLog(l) {
		return
	}

	if c.Logger != nil {
		c.Logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// LogDebug emits a debug level SDK log if config's log level is at least LogDebug
func (c *Config) LogDebug(format string, v ...interface{}) {
	c.doLog(LogDebug, format, v...)
}

// LogInfo emits an info level SDK log if config's log level is at least LogInfo
func (c *Config) LogInfo(format string, v ...interface{}) {
	c.doLog(LogInfo, format, v...)
}

// Ergonomic methods to set namespace relative to which action should be taken
func (c Config) WithNamespaceID(nsID string) Config {
	c.PathRoot = fmt.Sprintf(`{".tag": "namespace_id", "namespace_id": "%s"}`, nsID)
	return c
}

func (c Config) WithRoot(nsID string) Config {
	c.PathRoot = fmt.Sprintf(`{".tag": "root", "root": "%s"}`, nsID)
	return c
}

// Context is the base client context used to implement per-namespace clients.
type Context struct {
	Config          Config
	Client          *http.Client
	NoAuthClient    *http.Client
	HeaderGenerator func(hostType string, namespace string, route string) map[string]string
	URLGenerator    func(hostType string, namespace string, route string) string
}

type Request struct {
	Host      string
	Namespace string
	Route     string
	Style     string
	Auth      string

	Arg          interface{}
	ExtraHeaders map[string]string
}

// Execute executes a Dropbox API request with a background context.
func (c *Context) Execute(req Request, body io.Reader) ([]byte, io.ReadCloser, error) {
	return c.ExecuteContext(context.Background(), req, body)
}

// ExecuteContext executes a Dropbox API request with the provided context.
// Retries use Config.RetryPolicy when set.
func (c *Context) ExecuteContext(ctx context.Context, req Request, body io.Reader) ([]byte, io.ReadCloser, error) {
	var policy retry.Policy
	retryConfigured := c.Config.RetryPolicy != nil
	if retryConfigured {
		policy = c.Config.RetryPolicy.Normalized()
	}
	bodyOffset, bodySeeker, bodyRetryable := retryableBody(body)
	canRetry := retryConfigured && bodyRetryable

	client := c.Client
	if req.Auth == "noauth" {
		client = c.NoAuthClient
	}

	for attempt := 0; ; attempt++ {
		if attempt > 0 && bodySeeker != nil {
			if _, err := bodySeeker.Seek(bodyOffset, io.SeekStart); err != nil {
				return nil, nil, err
			}
		}

		httpReq, err := c.newHTTPRequest(ctx, req, body)
		if err != nil {
			return nil, nil, err
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, nil, err
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
			switch req.Style {
			case "rpc", "upload":
				if resp.Body == nil {
					return nil, nil, errors.New("expected body in RPC response, got nil")
				}

				b, err := io.ReadAll(resp.Body)
				if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
				if err != nil {
					return nil, nil, err
				}

				return b, nil, nil
			case "download":
				b := []byte(resp.Header.Get("Dropbox-API-Result"))
				return b, resp.Body, nil
			}
		}

		b, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		bodyErr := readErr
		if bodyErr == nil {
			bodyErr = closeErr
		}
		if bodyErr != nil {
			if canRetry {
				retryBody := b
				if readErr != nil {
					retryBody = nil
				}
				if wait, ok := policy.Delay(resp, retryBody, attempt); ok {
					if err := retry.Sleep(ctx, wait); err != nil {
						return nil, nil, err
					}
					continue
				}
			}
			return nil, nil, bodyErr
		}

		if canRetry {
			if wait, ok := policy.Delay(resp, b, attempt); ok {
				if err := retry.Sleep(ctx, wait); err != nil {
					return nil, nil, err
				}
				continue
			}
		}

		return nil, nil, SDKInternalError{
			StatusCode: resp.StatusCode,
			Content:    string(b),
		}
	}
}

func (c *Context) newHTTPRequest(ctx context.Context, req Request, body io.Reader) (*http.Request, error) {
	url := c.URLGenerator(req.Host, req.Namespace, req.Route)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		httpReq.Body = io.NopCloser(body)
	}

	for k, v := range req.ExtraHeaders {
		httpReq.Header.Add(k, v)
	}

	for k, v := range c.HeaderGenerator(req.Host, req.Namespace, req.Route) {
		httpReq.Header.Add(k, v)
	}

	if httpReq.Header.Get("Host") != "" {
		httpReq.Host = httpReq.Header.Get("Host")
	}

	if req.Auth == "noauth" {
		httpReq.Header.Del("Authorization")
	}
	if req.Auth != "team" && c.Config.AsMemberID != "" {
		httpReq.Header.Add("Dropbox-API-Select-User", c.Config.AsMemberID)
	}
	if req.Auth != "team" && c.Config.AsAdminID != "" {
		httpReq.Header.Add("Dropbox-API-Select-Admin", c.Config.AsAdminID)
	}
	if c.Config.PathRoot != "" {
		httpReq.Header.Add("Dropbox-API-Path-Root", c.Config.PathRoot)
	}

	if req.Arg != nil {
		serializedArg, err := json.Marshal(req.Arg)
		if err != nil {
			return nil, err
		}

		switch req.Style {
		case "rpc":
			if body != nil {
				return nil, errors.New("RPC style requests can not have body")
			}

			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Body = io.NopCloser(bytes.NewReader(serializedArg))
			httpReq.ContentLength = int64(len(serializedArg))
		case "upload", "download":
			httpReq.Header.Set("Dropbox-API-Arg", string(serializedArg))
			httpReq.Header.Set("Content-Type", "application/octet-stream")
		}
	}

	return httpReq, nil
}

func retryableBody(body io.Reader) (int64, io.Seeker, bool) {
	if body == nil {
		return 0, nil, true
	}
	seeker, ok := body.(io.Seeker)
	if !ok {
		return 0, nil, false
	}
	offset, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, nil, false
	}
	return offset, seeker, true
}

// NewContext returns a new Context with the given Config.
func NewContext(c Config) Context {
	domain := c.Domain
	if domain == "" {
		domain = defaultDomain
	}

	client := c.Client
	if client == nil {
		if c.TokenSource != nil {
			client = oauth2.NewClient(context.Background(), c.TokenSource)
		} else {
			var conf = &oauth2.Config{Endpoint: OAuthEndpoint(domain)}
			tok := &oauth2.Token{AccessToken: c.Token}
			client = conf.Client(context.Background(), tok)
		}
	}

	noAuthClient := c.Client
	if noAuthClient == nil {
		noAuthClient = &http.Client{}
	}

	headerGenerator := c.HeaderGenerator
	if headerGenerator == nil {
		headerGenerator = func(hostType string, namespace string, route string) map[string]string {
			return map[string]string{}
		}
	}

	urlGenerator := c.URLGenerator
	if urlGenerator == nil {
		hostMap := map[string]string{
			hostAPI:     hostAPI + domain,
			hostContent: hostContent + domain,
			hostNotify:  hostNotify + domain,
		}
		urlGenerator = func(hostType string, namespace string, route string) string {
			fqHost := hostMap[hostType]
			return fmt.Sprintf("https://%s/%d/%s/%s", fqHost, apiVersion, namespace, route)
		}
	}

	return Context{c, client, noAuthClient, headerGenerator, urlGenerator}
}

// OAuthEndpoint constructs an `oauth2.Endpoint` for the given domain
func OAuthEndpoint(domain string) oauth2.Endpoint {
	if domain == "" {
		domain = defaultDomain
	}
	authURL := fmt.Sprintf("https://meta%s/oauth2/authorize", domain)
	tokenURL := fmt.Sprintf("https://api%s/oauth2/token", domain)
	if domain == defaultDomain {
		authURL = "https://www.dropbox.com/oauth2/authorize"
	}
	return oauth2.Endpoint{AuthURL: authURL, TokenURL: tokenURL}
}

// HTTPHeaderSafeJSON encode the JSON passed in b []byte passed in in
// a way that is suitable for HTTP headers.
//
// See: https://www.dropbox.com/developers/reference/json-encoding
func HTTPHeaderSafeJSON(b []byte) string {
	var s strings.Builder
	s.Grow(len(b))
	for _, r := range string(b) {
		if r >= 0x007f {
			fmt.Fprintf(&s, "\\u%04x", r)
		} else {
			s.WriteRune(r)
		}
	}
	return s.String()
}
