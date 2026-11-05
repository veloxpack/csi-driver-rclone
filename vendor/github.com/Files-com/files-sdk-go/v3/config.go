package files_sdk

import (
	"cmp"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Files-com/files-sdk-go/v3/lib"
	libLog "github.com/Files-com/files-sdk-go/v3/lib/logpath"
	"github.com/hashicorp/go-retryablehttp"
)

var VERSION = "3.3.194"
var defaultUserAgent = fmt.Sprintf("%v %v", UserAgent, strings.TrimSpace(VERSION))

const (
	UserAgent   = "Files.com Go SDK"
	DefaultSite = "app"
	APIPath     = "/api/rest/v1"
)

const (
	apiKeyHeader           = "X-FilesAPI-Key"
	sessionIdHeader        = "X-FilesAPI-Auth"
	reauthenticationHeader = "X-Files-Reauthentication"
	workspaceIdHeader      = "X-Files-Workspace-Id"
)

var GlobalConfig Config

func init() {
	GlobalConfig = Config{}.Init()
}

// Config controls Files.com API authentication, workspace scoping, and transport behavior.
type Config struct {
	APIKey    string `header:"X-FilesAPI-Key" json:"api_key"`
	SessionId string `header:"X-FilesAPI-Auth" json:"session_id"`
	// WorkspaceId scopes Files.com API requests when non-zero.
	WorkspaceId      int64  `header:"X-Files-Workspace-Id" json:"workspace_id,omitempty"`
	Language         string `header:"Accept-Language"`
	Subdomain        string `json:"subdomain"`
	EndpointOverride string `json:"endpoint_override"`
	*retryablehttp.Client
	AdditionalHeaders map[string]string `json:"additional_headers"`
	lib.Logger
	Debug        bool   `json:"debug"`
	UserAgent    string `json:"user_agents"`
	Environment  `json:"environment"`
	FeatureFlags map[string]bool `json:"feature_flags"`
	// DisableDirectTransfers forces upload and download requests to use proxied transfer paths.
	DisableDirectTransfers bool `json:"disable_direct_transfers"`
}

func (c Config) Init() Config {
	if c.Logger == nil {
		c.Logger = lib.NullLogger{}
	}
	if c.Client == nil {
		c.Client = lib.DefaultRetryableHttp(c)
	}

	if c.FeatureFlags == nil {
		c.FeatureFlags = FeatureFlags()
	}

	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}

	return c
}

func (c Config) Endpoint() string {
	if c.EndpointOverride != "" && !strings.HasPrefix(c.EndpointOverride, "https://") && !strings.HasPrefix(c.EndpointOverride, "http://") {
		c.EndpointOverride = "https://" + c.EndpointOverride
	}

	return lib.DefaultString(
		c.EndpointOverride,
		strings.Replace(c.Environment.Endpoint(), "{{SUBDOMAIN}}", lib.DefaultString(c.Subdomain, DefaultSite), 1),
	)
}

func (c Config) Do(req *http.Request) (*http.Response, error) {
	return c.redirectSafeClient().StandardClient().Do(req)
}

func (c Config) SetCustomClient(client *http.Client) Config {
	c.Client = lib.DefaultRetryableHttp(c, client)
	return c
}

func (c Config) InDebug() bool {
	return c.Debug || (os.Getenv("FILES_SDK_DEBUG") != "")
}

func (c Config) LogPath(path string, args map[string]interface{}) {
	c.Logger.Printf(libLog.New(path, args))
}

func (c Config) RootPath() string {
	return c.Endpoint() + APIPath
}

func (c Config) GetAPIKey() string {
	if c.SessionId != "" && c.APIKey == "" {
		return ""
	}
	return lib.DefaultString(c.APIKey, os.Getenv("FILES_API_KEY"))
}

func (c Config) SetUserAgentHeader(headers *http.Header) {
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", cmp.Or(c.UserAgent, defaultUserAgent))
	}
}

func (c Config) SetHeaders(headers *http.Header) {
	c.setHeadersForURL(headers, c.RootPath())
}

func (c Config) SetHeadersForRequest(req *http.Request) {
	c.setHeadersForURL(&req.Header, req.URL.String())
}

func (c Config) setHeadersForURL(headers *http.Header, rawURL string) {
	headers.Set("User-Agent", cmp.Or(c.UserAgent, defaultUserAgent))
	apiKey := c.GetAPIKey()
	if apiKey != "" {
		headers.Set(apiKeyHeader, apiKey)
	} else if c.SessionId != "" {
		headers.Set(sessionIdHeader, c.SessionId)
	}
	if c.Language != "" {
		headers.Set("Accept-Language", c.Language)
	}
	if c.WorkspaceId != 0 {
		headers.Set(workspaceIdHeader, strconv.FormatInt(c.WorkspaceId, 10))
	}
	for key, value := range c.AdditionalHeaders {
		headers.Set(key, value)
	}
	if !c.shouldSendAuthHeaders(rawURL) {
		clearAuthHeaders(headers)
	}
}

func (c Config) redirectSafeClient() *retryablehttp.Client {
	retrySource := c.Client
	if retrySource == nil {
		initialized := c.Init()
		retrySource = initialized.Client
	}

	httpClient := http.Client{}
	if retrySource.HTTPClient != nil {
		httpClient = *retrySource.HTTPClient
	}
	originalCheckRedirect := httpClient.CheckRedirect
	// Go copies custom headers during redirects, including cross-origin redirects.
	// Re-apply URL-aware headers to each redirected request so Files auth is
	// stripped if a same-origin download URL redirects to a storage provider.
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		c.SetHeadersForRequest(req)
		if originalCheckRedirect != nil {
			return originalCheckRedirect(req, via)
		}
		return nil
	}

	return &retryablehttp.Client{
		HTTPClient:      &httpClient,
		Logger:          retrySource.Logger,
		RetryWaitMin:    retrySource.RetryWaitMin,
		RetryWaitMax:    retrySource.RetryWaitMax,
		RetryMax:        retrySource.RetryMax,
		RequestLogHook:  retrySource.RequestLogHook,
		ResponseLogHook: retrySource.ResponseLogHook,
		CheckRetry:      retrySource.CheckRetry,
		Backoff:         retrySource.Backoff,
		ErrorHandler:    retrySource.ErrorHandler,
		PrepareRetry:    retrySource.PrepareRetry,
	}
}

func (c Config) shouldSendAuthHeaders(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	destination, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if destination.Host == "" {
		return true
	}

	endpoint, err := url.Parse(c.Endpoint())
	if err != nil {
		return false
	}

	return strings.EqualFold(destination.Scheme, endpoint.Scheme) && normalizedURLHost(destination) == normalizedURLHost(endpoint)
}

func clearAuthHeaders(headers *http.Header) {
	headers.Del(apiKeyHeader)
	headers.Del(sessionIdHeader)
	headers.Del(reauthenticationHeader)
	headers.Del(workspaceIdHeader)
}

func normalizedURLHost(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	// Treat default ports as equivalent to an omitted port for origin comparison.
	if port == "" || (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		return host
	}
	return net.JoinHostPort(host, port)
}

func (c Config) FeatureFlag(flag string) bool {
	value, ok := c.FeatureFlags[flag]
	if !ok {
		panic(fmt.Sprintf("feature flag `%v` is not a valid flag", flag))
	}
	return value
}

const (
	FeatureFlagAdaptiveUploadV2        = "adaptive-upload-v2"
	FeatureFlagUploadV2ChecksumTrailer = "upload-v2-checksum-trailer"
)

func FeatureFlags() map[string]bool {
	return map[string]bool{
		FeatureFlagAdaptiveUploadV2:        false,
		FeatureFlagUploadV2ChecksumTrailer: false,
		"incremental-updates":              false,
	}
}
