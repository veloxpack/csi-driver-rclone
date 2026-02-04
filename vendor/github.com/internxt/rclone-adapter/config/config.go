package config

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/internxt/rclone-adapter/endpoints"
)

const (
	DefaultChunkSize        = 30 * 1024 * 1024
	DefaultMultipartMinSize = 100 * 1024 * 1024
	DefaultMaxConcurrency   = 6
	MaxThumbnailSourceSize  = 50 * 1024 * 1024
	ClientName              = "rclone-adapter"
)

type Config struct {
	Token              string            `json:"token,omitempty"`
	RootFolderID       string            `json:"root_folder_id,omitempty"`
	Bucket             string            `json:"bucket,omitempty"`
	Mnemonic           string            `json:"mnemonic,omitempty"`
	BasicAuthHeader    string            `json:"basic_auth_header,omitempty"`
	HTTPClient         *http.Client      `json:"-"` // Centralized HTTP client with proper timeouts
	Endpoints          *endpoints.Config `json:"-"` // Centralized API endpoint management
	SkipHashValidation bool              `json:"skip_hash_validation,omitempty"`
}

func NewDefaultToken(token string) *Config {
	cfg := &Config{
		Token: token,
	}
	cfg.ApplyDefaults()
	return cfg
}

// ApplyDefaults sets default values for any unset configuration fields.
// This is useful for test configurations to ensure they have properly configured HTTPClient with custom transport.
func (c *Config) ApplyDefaults() {
	if c.HTTPClient == nil {
		c.HTTPClient = newHTTPClient()
	}
	if c.Endpoints == nil {
		c.Endpoints = endpoints.Default()
	}
}

// clientHeaderTransport wraps http.RoundTripper to automatically add the internxt-client header
type clientHeaderTransport struct {
	base http.RoundTripper
}

func (t *clientHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.validateSecurity(req); err != nil {
		return nil, err
	}

	req.Header.Set("internxt-client", ClientName)
	return t.base.RoundTrip(req)
}

func (t *clientHeaderTransport) validateSecurity(req *http.Request) error {
	_, _, isBasic := req.BasicAuth()
	if !isBasic {
		return nil
	}

	if req.URL.Scheme != "https" && !isLoopback(req.URL.Hostname()) {
		return errors.New("security error: Basic Auth requires HTTPS")
	}

	return nil
}

func isLoopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// newHTTPClient: properly configured HTTP client with sensible timeouts
func newHTTPClient() *http.Client {
	baseTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		ForceAttemptHTTP2:     true,
	}

	return &http.Client{
		Timeout:   5 * time.Minute,
		Transport: &clientHeaderTransport{base: baseTransport},
	}
}
