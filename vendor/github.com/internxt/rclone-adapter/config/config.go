package config

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/internxt/rclone-adapter/endpoints"
)

const (
	DefaultChunkSize        = 30 * 1024 * 1024
	DefaultMultipartMinSize = 100 * 1024 * 1024
	DefaultMaxConcurrency   = 6
	MaxThumbnailSourceSize  = 50 * 1024 * 1024
	ClientName              = "rclone-adapter"

	DefaultFileLimitsTTL = 1 * time.Hour
)

type Config struct {
	Token              string            `json:"token,omitempty"`
	RootFolderID       string            `json:"root_folder_id,omitempty"`
	Bucket             string            `json:"bucket,omitempty"`
	Mnemonic           string            `json:"mnemonic,omitempty"`
	BasicAuthHeader    string            `json:"basic_auth_header,omitempty"`
	HTTPClient         *http.Client      `json:"-"` // Centralized HTTP client with proper timeouts
	Endpoints          *endpoints.Config `json:"-"` // Centralized API endpoint management
	FileLimits         *FileLimitsCache  `json:"-"` // Cached account upload limits
	SkipHashValidation bool              `json:"skip_hash_validation,omitempty"`
}

// FileLimitsCache holds the account's MaxUploadFileSize with a TTL refresh
type FileLimitsCache struct {
	mu                sync.Mutex
	maxUploadFileSize int64
	fetchedAt         time.Time
	populated         bool
	ttl               time.Duration
}

// NewFileLimitsCache returns a cache with the default TTL.
func NewFileLimitsCache() *FileLimitsCache {
	return &FileLimitsCache{ttl: DefaultFileLimitsTTL}
}

// GetOrFetch returns the cached MaxUploadFileSize, refreshing via fetchFn
// when the cache is empty or stale
func (c *FileLimitsCache) GetOrFetch(ctx context.Context, fetchFn func(context.Context) (int64, error)) (maxUploadFileSize int64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.populated && time.Since(c.fetchedAt) < c.ttl {
		return c.maxUploadFileSize, true
	}
	v, err := fetchFn(ctx)
	if err != nil {
		return 0, false
	}
	c.maxUploadFileSize = v
	c.fetchedAt = time.Now()
	c.populated = true
	return c.maxUploadFileSize, true
}

// Invalidate clears the cache so the next GetOrFetch will refetch.
func (c *FileLimitsCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.populated = false
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
	if c.FileLimits == nil {
		c.FileLimits = NewFileLimitsCache()
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
