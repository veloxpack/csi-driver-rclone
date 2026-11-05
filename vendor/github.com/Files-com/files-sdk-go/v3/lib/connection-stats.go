package lib

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

type Transport struct {
	*http.Transport
	*net.Dialer
	stats              *connectionStats
	statsMu            sync.Mutex
	dialerMu           sync.Mutex
	s3UploadHTTP1      *http.Transport
	s3UploadHTTP1Mutex sync.Mutex
}

type connectionStats struct {
	connections map[string]*int32
	mu          sync.RWMutex
}

func (t *Transport) GetConnectionStats() map[string]int {
	stats := t.connectionStats()
	stats.mu.RLock()
	defer stats.mu.RUnlock() // Keep the read lock for the entire function

	copiedMap := make(map[string]int)
	for key, value := range stats.connections {
		copiedMap[key] = int(atomic.LoadInt32(value))
	}
	return copiedMap
}

func (t *Transport) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := t.dialer()
	conn, err := dialer.DialContext(ctx, network, address)

	if err == nil {
		stats := t.connectionStats()
		stats.mu.Lock()
		counter, ok := stats.connections[address]
		if !ok {
			intCounter := int32(0)
			counter = &intCounter
			stats.connections[address] = counter
		}
		stats.mu.Unlock()
		atomic.AddInt32(counter, 1)
		return &Conn{Conn: conn, counter: counter}, err
	}

	return conn, err
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if shouldUseHTTP1ForS3Upload(req) {
		if transport := t.s3UploadHTTP1Transport(); transport != nil {
			return transport.RoundTrip(req)
		}
	}
	if t.Transport != nil {
		return t.Transport.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func (t *Transport) CloseIdleConnections() {
	if t.Transport != nil {
		t.Transport.CloseIdleConnections()
	}
	t.s3UploadHTTP1Mutex.Lock()
	transport := t.s3UploadHTTP1
	t.s3UploadHTTP1Mutex.Unlock()
	if transport != nil {
		transport.CloseIdleConnections()
	}
}

func (t *Transport) s3UploadHTTP1Transport() *http.Transport {
	t.s3UploadHTTP1Mutex.Lock()
	defer t.s3UploadHTTP1Mutex.Unlock()
	if t.s3UploadHTTP1 != nil {
		return t.s3UploadHTTP1
	}
	base := t.Transport
	if base == nil {
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil
		}
		base = defaultTransport
	}
	t.s3UploadHTTP1 = cloneHTTPTransportForS3UploadHTTP1(base, t.DialContext)
	return t.s3UploadHTTP1
}

func (t *Transport) dialer() *net.Dialer {
	t.dialerMu.Lock()
	defer t.dialerMu.Unlock()
	if t.Dialer == nil {
		t.Dialer = &net.Dialer{}
	}
	return t.Dialer
}

func (t *Transport) connectionStats() *connectionStats {
	t.statsMu.Lock()
	defer t.statsMu.Unlock()
	if t.stats == nil {
		t.stats = newConnectionStats()
	}
	return t.stats
}

func newConnectionStats() *connectionStats {
	return &connectionStats{connections: make(map[string]*int32)}
}

type Conn struct {
	net.Conn
	counter *int32
	sync.Once
}

func (c *Conn) Close() (err error) {
	err = c.Conn.Close()
	c.Do(func() { atomic.AddInt32(c.counter, -1) })

	return
}

func GetConnectionStatsFromClient(client *http.Client) (map[string]int, bool) {
	if client == nil {
		return nil, false
	}
	transport, ok := client.Transport.(*Transport)
	if !ok {
		return nil, false
	}
	return transport.GetConnectionStats(), true
}

func CloneHTTPClientWithMaxConnsPerHost(client *http.Client, maxConnsPerHost int) (*http.Client, bool) {
	if client == nil || maxConnsPerHost <= 0 {
		return client, false
	}

	cloned := *client
	switch transport := client.Transport.(type) {
	case *Transport:
		cloned.Transport = transport.cloneWithMaxConnsPerHost(maxConnsPerHost)
	case *http.Transport:
		cloned.Transport = cloneHTTPTransportWithMaxConnsPerHost(transport, maxConnsPerHost)
	case nil:
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return client, false
		}
		cloned.Transport = cloneHTTPTransportWithMaxConnsPerHost(defaultTransport, maxConnsPerHost)
	default:
		return client, false
	}
	return &cloned, true
}

func CloneHTTPClientWithExactMaxConnsPerHost(client *http.Client, maxConnsPerHost int) (*http.Client, bool) {
	if client == nil || maxConnsPerHost <= 0 {
		return client, false
	}

	cloned := *client
	switch transport := client.Transport.(type) {
	case *Transport:
		cloned.Transport = transport.cloneWithExactMaxConnsPerHost(maxConnsPerHost)
	case *http.Transport:
		cloned.Transport = cloneHTTPTransportWithExactMaxConnsPerHost(transport, maxConnsPerHost)
	case nil:
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return client, false
		}
		cloned.Transport = cloneHTTPTransportWithExactMaxConnsPerHost(defaultTransport, maxConnsPerHost)
	default:
		return client, false
	}
	return &cloned, true
}

func (t *Transport) cloneWithMaxConnsPerHost(maxConnsPerHost int) *Transport {
	var base *http.Transport
	if t.Transport != nil {
		base = t.Transport
	} else if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		base = defaultTransport
	} else {
		base = &http.Transport{}
	}

	cloned := &Transport{
		Transport: cloneHTTPTransportWithMaxConnsPerHost(base, maxConnsPerHost),
		Dialer:    t.dialer(),
		stats:     t.connectionStats(),
	}
	cloned.Transport.DialContext = cloned.DialContext
	cloned.s3UploadHTTP1 = cloneHTTPTransportForS3UploadHTTP1(cloned.Transport, cloned.DialContext)
	return cloned
}

func (t *Transport) cloneWithExactMaxConnsPerHost(maxConnsPerHost int) *Transport {
	var base *http.Transport
	if t.Transport != nil {
		base = t.Transport
	} else if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		base = defaultTransport
	} else {
		base = &http.Transport{}
	}

	cloned := &Transport{
		Transport: cloneHTTPTransportWithExactMaxConnsPerHost(base, maxConnsPerHost),
		Dialer:    t.dialer(),
		stats:     t.connectionStats(),
	}
	cloned.Transport.DialContext = cloned.DialContext
	cloned.s3UploadHTTP1 = cloneHTTPTransportForS3UploadHTTP1(cloned.Transport, cloned.DialContext)
	return cloned
}

func cloneHTTPTransportWithMaxConnsPerHost(transport *http.Transport, maxConnsPerHost int) *http.Transport {
	cloned := transport.Clone()
	applyHTTPTransportMaxConnsPerHost(cloned, maxConnsPerHost)
	return cloned
}

func cloneHTTPTransportWithExactMaxConnsPerHost(transport *http.Transport, maxConnsPerHost int) *http.Transport {
	cloned := transport.Clone()
	applyHTTPTransportExactMaxConnsPerHost(cloned, maxConnsPerHost)
	return cloned
}

func applyHTTPTransportMaxConnsPerHost(transport *http.Transport, maxConnsPerHost int) {
	if maxConnsPerHost <= 0 {
		return
	}
	if transport.MaxConnsPerHost == 0 || transport.MaxConnsPerHost < maxConnsPerHost {
		transport.MaxConnsPerHost = maxConnsPerHost
	}
	if transport.MaxIdleConns < maxConnsPerHost {
		transport.MaxIdleConns = maxConnsPerHost
	}
	if transport.MaxIdleConnsPerHost < maxConnsPerHost {
		transport.MaxIdleConnsPerHost = maxConnsPerHost
	}
}

func applyHTTPTransportExactMaxConnsPerHost(transport *http.Transport, maxConnsPerHost int) {
	if maxConnsPerHost <= 0 {
		return
	}
	transport.MaxConnsPerHost = maxConnsPerHost
	if transport.MaxIdleConns < maxConnsPerHost {
		transport.MaxIdleConns = maxConnsPerHost
	}
	if transport.MaxIdleConnsPerHost < maxConnsPerHost {
		transport.MaxIdleConnsPerHost = maxConnsPerHost
	}
}

func cloneHTTPTransportForS3UploadHTTP1(transport *http.Transport, dialContext func(context.Context, string, string) (net.Conn, error)) *http.Transport {
	cloned := transport.Clone()
	cloned.ForceAttemptHTTP2 = false
	cloned.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	if cloned.TLSClientConfig != nil {
		cloned.TLSClientConfig = cloned.TLSClientConfig.Clone()
		cloned.TLSClientConfig.NextProtos = withoutHTTP2NextProto(cloned.TLSClientConfig.NextProtos)
	}
	cloned.DialContext = dialContext
	return cloned
}

func withoutHTTP2NextProto(protocols []string) []string {
	if len(protocols) == 0 {
		return protocols
	}
	filtered := make([]string, 0, len(protocols))
	for _, protocol := range protocols {
		if !strings.EqualFold(protocol, "h2") {
			filtered = append(filtered, protocol)
		}
	}
	return filtered
}

func shouldUseHTTP1ForS3Upload(req *http.Request) bool {
	if req == nil || req.URL == nil || !isUploadHTTPMethod(req.Method) {
		return false
	}
	if isS3EndpointHost(req.URL.Hostname()) {
		return true
	}
	return hasAWSPresignedQuery(req.URL.Query())
}

func isUploadHTTPMethod(method string) bool {
	return method == http.MethodPut || method == http.MethodPost
}

func isS3EndpointHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "" {
		return false
	}
	if host == "s3.amazonaws.com" || strings.HasSuffix(host, ".s3.amazonaws.com") {
		return true
	}
	if strings.HasPrefix(host, "s3.") && strings.HasSuffix(host, ".amazonaws.com") {
		return true
	}
	if strings.Contains(host, ".s3.") && strings.HasSuffix(host, ".amazonaws.com") {
		return true
	}
	if strings.HasPrefix(host, "s3-") && strings.HasSuffix(host, ".amazonaws.com") {
		return true
	}
	if strings.Contains(host, ".s3-") && strings.HasSuffix(host, ".amazonaws.com") {
		return true
	}
	return false
}

func hasAWSPresignedQuery(values url.Values) bool {
	if len(values) == 0 {
		return false
	}
	for key := range values {
		switch strings.ToLower(key) {
		case "x-amz-algorithm", "x-amz-credential", "x-amz-signature":
			return true
		}
	}
	return false
}
