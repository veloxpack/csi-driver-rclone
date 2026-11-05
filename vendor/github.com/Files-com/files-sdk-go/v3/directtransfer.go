package files_sdk

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/hashicorp/go-retryablehttp"
)

// The connect budgets are package variables instead of constants so tests can
// shorten them; production code treats them as constants.
var (
	// directTransferConnectTimeout caps the total time spent connecting
	// to all direct candidates before the caller falls back to the proxy URL.
	directTransferConnectTimeout = 5 * time.Second
	// directTransferPerCandidateTimeout caps the connect/handshake time
	// for each candidate address so a single stalled candidate cannot consume
	// the whole connect budget and starve the remaining candidates.
	directTransferPerCandidateTimeout   = 2 * time.Second
	directTransferResponseHeaderTimeout = 60 * time.Second
)

var (
	ErrDirectTransferUnavailable     = errors.New("direct transfer unavailable")
	ErrDirectTransferResponseStarted = errors.New("direct transfer response processing started")
)

type directTransferClientCacheKey struct {
	serverName string
	caPEM      string
	candidates string
}

type directTransferClientCache struct {
	mu         sync.Mutex
	key        directTransferClientCacheKey
	httpClient *http.Client
}

type directTransferClientCacheContextKey struct{}

// WithDirectTransferClientCache scopes direct HTTP clients to one transfer.
func WithDirectTransferClientCache(ctx context.Context) (context.Context, func()) {
	cache := &directTransferClientCache{}
	return context.WithValue(ctx, directTransferClientCacheContextKey{}, cache), cache.close
}

func (c *directTransferClientCache) client(key directTransferClientCacheKey, build func() (*http.Client, error)) (*http.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.httpClient != nil {
		if c.key != key {
			return nil, ErrDirectTransferUnavailable
		}
		return c.httpClient, nil
	}
	client, err := build()
	if err == nil {
		c.key = key
		c.httpClient = client
	}
	return client, err
}

func (c *directTransferClientCache) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
		c.httpClient = nil
	}
}

var directTransferBlockedIPPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:2::/48"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func DirectConnectionInfoPresent(info DirectConnectionInfo) bool {
	return info.Version == 1 &&
		info.ServerName != "" &&
		len(info.Addresses) > 0 &&
		info.DirectUri != "" &&
		info.CaPem != ""
}

func DirectTransferRetryableClient(ctx context.Context, config Config, info DirectConnectionInfo) (string, *retryablehttp.Client, error) {
	directURL, httpClient, err := directTransferHTTPClient(ctx, config, info)
	if err != nil {
		return "", nil, err
	}

	return directURL, lib.DefaultRetryableHttp(config.Logger, httpClient), nil
}

func WrapDirectTransferOptions(config Config, info DirectConnectionInfo, request *http.Request, opts ...RequestResponseOption) (*http.Response, error) {
	if config.DisableDirectTransfers || !DirectConnectionInfoPresent(info) {
		return nil, ErrDirectTransferUnavailable
	}

	directAttemptRequest := request.Clone(request.Context())
	modifiedRequest, err := BuildRequest(directAttemptRequest, opts...)
	if err != nil {
		return nil, err
	}

	directURL, httpClient, err := directTransferHTTPClient(modifiedRequest.Context(), config, info)
	if err != nil {
		return nil, err
	}

	directRequest, err := cloneRequestForDirectTransfer(modifiedRequest, directURL)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(directRequest)
	if err != nil {
		return response, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		lib.CloseBody(response)
		return response, fmt.Errorf("%w: status %d", ErrDirectTransferUnavailable, response.StatusCode)
	}

	processedResponse, err := BuildResponse(response, opts...)
	if err != nil {
		lib.CloseBody(response)
		return response, errors.Join(ErrDirectTransferResponseStarted, err)
	}
	return processedResponse, nil
}

func directTransferHTTPClient(ctx context.Context, config Config, info DirectConnectionInfo) (string, *http.Client, error) {
	if config.DisableDirectTransfers || !DirectConnectionInfoPresent(info) {
		return "", nil, ErrDirectTransferUnavailable
	}

	directURL, err := directTransferDirectURL(info)
	if err != nil {
		return "", nil, err
	}

	candidates, err := directTransferCandidateAddresses(config, info)
	if err != nil {
		return "", nil, err
	}
	build := func() (*http.Client, error) {
		roots, err := directTransferCertPool(info.CaPem)
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    roots,
			ServerName: info.ServerName,
		}
		transport := &http.Transport{
			DialTLSContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return directTransferDialTLSCandidates(ctx, network, candidates, tlsConfig)
			},
			ResponseHeaderTimeout: directTransferResponseHeaderTimeout,
			IdleConnTimeout:       30 * time.Second,
			// The Agent direct listener is an HTTP/1.1 endpoint.
			TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
		}

		return &http.Client{
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}, nil
	}

	if cache, ok := ctx.Value(directTransferClientCacheContextKey{}).(*directTransferClientCache); ok {
		client, err := cache.client(directTransferClientCacheKey{
			serverName: info.ServerName,
			caPEM:      info.CaPem,
			candidates: strings.Join(candidates, ","),
		}, build)
		return directURL, client, err
	}

	client, err := build()
	return directURL, client, err
}

func directTransferDirectURL(info DirectConnectionInfo) (string, error) {
	if !DirectConnectionInfoPresent(info) {
		return "", ErrDirectTransferUnavailable
	}

	transferURL, err := url.Parse(info.DirectUri)
	if err != nil {
		return "", err
	}
	if transferURL.Scheme != "https" {
		return "", fmt.Errorf("%w: direct_uri must use https scheme", ErrDirectTransferUnavailable)
	}
	if !strings.EqualFold(transferURL.Hostname(), info.ServerName) {
		return "", fmt.Errorf("%w: direct_uri host does not match server_name", ErrDirectTransferUnavailable)
	}
	transferURL.User = nil
	return transferURL.String(), nil
}

func directTransferCertPool(caPEM string) (*x509.CertPool, error) {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(caPEM)) {
		return nil, fmt.Errorf("%w: invalid ca_pem", ErrDirectTransferUnavailable)
	}
	return roots, nil
}

func directTransferCandidateAddresses(config Config, info DirectConnectionInfo) ([]string, error) {
	allowPrivateCandidates := config.Environment != Production
	candidates := make([]string, 0, len(info.Addresses))
	for _, rawAddress := range info.Addresses {
		candidate, ok := directTransferCandidateAddress(rawAddress, allowPrivateCandidates)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return nil, ErrDirectTransferUnavailable
	}
	return candidates, nil
}

func directTransferCandidateAddress(rawAddress string, allowPrivate bool) (string, bool) {
	locator, err := url.Parse(rawAddress)
	if err != nil || locator.Scheme != "tcp" || locator.Host == "" || locator.User != nil || locator.Path != "" || locator.RawQuery != "" || locator.Fragment != "" {
		return "", false
	}

	host, port, err := net.SplitHostPort(locator.Host)
	if err != nil {
		return "", false
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", false
	}

	ip, err := netip.ParseAddr(host)
	if err != nil || !directTransferCandidateIP(ip, allowPrivate) {
		return "", false
	}

	return net.JoinHostPort(ip.String(), port), true
}

func directTransferCandidateIP(ip netip.Addr, allowPrivate bool) bool {
	if !ip.IsValid() {
		return false
	}
	if allowPrivate && (ip.IsLoopback() || ip.IsPrivate()) {
		return true
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	for _, prefix := range directTransferBlockedIPPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func directTransferDialTLSCandidates(ctx context.Context, network string, candidates []string, tlsConfig *tls.Config) (net.Conn, error) {
	if len(candidates) == 0 {
		return nil, ErrDirectTransferUnavailable
	}

	connectCtx, cancel := context.WithTimeout(ctx, directTransferConnectTimeout)
	defer cancel()

	var errs []error
	dialer := &net.Dialer{}
	for _, candidate := range candidates {
		if ctxErr := connectCtx.Err(); ctxErr != nil {
			errs = append(errs, ctxErr)
			break
		}

		candidateCtx, cancelCandidate := context.WithTimeout(connectCtx, directTransferPerCandidateTimeout)
		conn, err := dialer.DialContext(candidateCtx, network, candidate)
		if err == nil {
			tlsConn := tls.Client(conn, tlsConfig.Clone())
			err = tlsConn.HandshakeContext(candidateCtx)
			if err == nil {
				cancelCandidate()
				return tlsConn, nil
			}
			_ = conn.Close()
		}
		cancelCandidate()
		errs = append(errs, fmt.Errorf("%s: %w", candidate, err))
	}

	if len(errs) == 0 {
		return nil, ErrDirectTransferUnavailable
	}
	return nil, errors.Join(errs...)
}

func cloneRequestForDirectTransfer(request *http.Request, rawURL string) (*http.Request, error) {
	directURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	cloned := request.Clone(request.Context())
	cloned.URL = directURL
	cloned.Host = directURL.Host
	cloned.Header = *DirectTransferRequestHeaders(&request.Header)
	return cloned, nil
}

// DirectTransferRequestHeaders returns a header copy safe for direct requests.
func DirectTransferRequestHeaders(headers *http.Header) *http.Header {
	cloned := http.Header{}
	if headers != nil {
		cloned = headers.Clone()
	}
	clearAuthHeaders(&cloned)
	cloned.Del("Authorization")
	cloned.Del("Cookie")
	return &cloned
}
