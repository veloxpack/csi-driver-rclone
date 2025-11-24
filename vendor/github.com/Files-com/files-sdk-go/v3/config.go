package files_sdk

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Files-com/files-sdk-go/v3/lib"
	libLog "github.com/Files-com/files-sdk-go/v3/lib/logpath"
	"github.com/hashicorp/go-retryablehttp"
)

var VERSION = "3.2.264"

const (
	UserAgent   = "Files.com Go SDK"
	DefaultSite = "app"
	APIPath     = "/api/rest/v1"
)

var GlobalConfig Config

func init() {
	GlobalConfig = Config{}.Init()
}

type Config struct {
	APIKey           string `header:"X-FilesAPI-Key" json:"api_key"`
	SessionId        string `header:"X-FilesAPI-Auth" json:"session_id"`
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
		c.UserAgent = fmt.Sprintf("%v %v", UserAgent, strings.TrimSpace(VERSION))
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
	return c.Client.StandardClient().Do(req)
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
	return lib.DefaultString(c.APIKey, os.Getenv("FILES_API_KEY"))
}

func (c Config) SetHeaders(headers *http.Header) {
	headers.Set("User-Agent", c.UserAgent)
	if c.GetAPIKey() != "" {
		headers.Set("X-FilesAPI-Key", c.GetAPIKey())
	} else if c.SessionId != "" {
		headers.Set("X-FilesAPI-Auth", c.SessionId)
	}
	if c.Language != "" {
		headers.Set("Accept-Language", c.Language)
	}
	for key, value := range c.AdditionalHeaders {
		headers.Set(key, value)
	}
}

func (c Config) FeatureFlag(flag string) bool {
	value, ok := c.FeatureFlags[flag]
	if !ok {
		panic(fmt.Sprintf("feature flag `%v` is not a valid flag", flag))
	}
	return value
}

func FeatureFlags() map[string]bool {
	return map[string]bool{
		"incremental-updates": false,
	}
}
