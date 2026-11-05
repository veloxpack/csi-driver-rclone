package lib

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Files-com/files-sdk-go/v3/downloadurl"
	"github.com/hashicorp/go-retryablehttp"
)

var DefaultClient *http.Client

func init() {
	DefaultClient = defaultPooledClient()
}

func DefaultRetryableHttp(logger Logger, client ...*http.Client) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logger
	if len(client) == 1 {
		retryClient.HTTPClient = client[0]
	} else {
		retryClient.HTTPClient = DefaultClient
	}
	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if resp != nil {
			if sleep, ok := retryAfterDuration(resp.Header.Get("Retry-After"), time.Now()); ok {
				return sleep
			}
		}

		// Current uses do not improve with adding backoff it will only make things worse, things will become expired.
		return time.Duration(0)
	}
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		ok, err := retryablehttp.DefaultRetryPolicy(ctx, resp, err)

		if ok && resp != nil {
			// Don't waste time retrying an expired URL
			if !requestURLValidForRetry(resp) {
				return false, err
			}
		}
		return ok, err
	}
	retryClient.RetryMax = 3
	return retryClient
}

func UploadRetryableHttp(base *retryablehttp.Client, canReplay func() bool) *retryablehttp.Client {
	return UploadRetryableHttpWithObserver(base, canReplay, nil)
}

type UploadRetryAttempt struct {
	StatusCode int
	Retryable  bool
	RetryAfter time.Duration
	Err        error
}

func UploadRetryableHttpWithObserver(base *retryablehttp.Client, canReplay func() bool, observer func(UploadRetryAttempt)) *retryablehttp.Client {
	if base == nil {
		base = DefaultRetryableHttp(nil)
	}

	retryClient := *base
	baseCheckRetry := retryClient.CheckRetry
	if baseCheckRetry == nil {
		baseCheckRetry = retryablehttp.DefaultRetryPolicy
	}
	baseBackoff := retryClient.Backoff
	if baseBackoff == nil {
		baseBackoff = retryablehttp.DefaultBackoff
	}
	// CheckRetry and Backoff run sequentially for one request;
	// upload code builds this wrapper per upload part so retryClass is not shared across requests.
	retryClass := ""

	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		retryClass = ""
		classified := peekS3XMLClass(resp)
		ok, retryErr := baseCheckRetry(ctx, resp, err)
		if (ok || classified.Retryable) && !requestURLValidForRetry(resp) {
			return false, retryErr
		}
		if !ok && classified.Retryable {
			ok = true
		}
		if ok && canReplay != nil && !canReplay() {
			return false, retryErr
		}
		if ok {
			retryClass = classified.Class
		}
		if observer != nil && (resp != nil || err != nil) {
			attempt := UploadRetryAttempt{Retryable: ok, Err: err}
			if resp != nil {
				attempt.StatusCode = resp.StatusCode
				attempt.RetryAfter, _ = retryAfterDuration(resp.Header.Get("Retry-After"), time.Now())
			}
			observer(attempt)
		}
		return ok, retryErr
	}

	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if d := baseBackoff(min, max, attemptNum, resp); d > 0 {
			return d
		}
		if retryClass == "rate_limit" {
			return uploadRetryRateLimitBackoff(attemptNum)
		}
		return 0
	}

	return &retryClient
}

func peekS3XMLClass(resp *http.Response) S3ErrorClassification {
	if resp == nil || resp.Body == nil || !isS3XMLRetryStatus(resp.StatusCode) || !IsXML(resp) {
		return S3ErrorClassification{}
	}

	peek, err := readAndRestoreResponseBody(resp, 4*1024)
	if err != nil {
		return S3ErrorClassification{}
	}

	var s3Err S3Error
	if err := xml.Unmarshal(peek, &s3Err); err != nil || s3Err.Code == "" {
		return S3ErrorClassification{}
	}
	return ClassifyS3ErrorCode(s3Err.Code)
}

func readAndRestoreResponseBody(resp *http.Response, limit int64) ([]byte, error) {
	originalBody := resp.Body
	peek, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, err
	}
	resp.Body = replayReadCloser{
		Reader: io.MultiReader(bytes.NewReader(peek), originalBody),
		Closer: originalBody,
	}
	return peek, nil
}

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func isS3XMLRetryStatus(status int) bool {
	switch status {
	case http.StatusBadRequest, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
		return true
	default:
		return false
	}
}

func requestURLValidForRetry(resp *http.Response) bool {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return true
	}
	url, err := downloadurl.NewFromURL(resp.Request.URL)
	if err != nil || url.IsZero() {
		return true
	}
	_, valid := url.Valid(time.Millisecond * 100)
	return valid
}

func uploadRetryRateLimitBackoff(attemptNum int) time.Duration {
	if attemptNum < 1 {
		attemptNum = 1
	}

	maxDelay := time.Second << (attemptNum - 1)
	const capDelay = 4 * time.Second
	if maxDelay > capDelay {
		maxDelay = capDelay
	}

	minDelay := maxDelay / 2
	return minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay)))
}

func retryAfterDuration(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Second * time.Duration(seconds), true
	}

	retryAt, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	if retryAt.Before(now) {
		return 0, true
	}
	return retryAt.Sub(now), true
}

func defaultPooledClient() *http.Client {
	return &http.Client{
		Transport: DefaultPooledTransport(),
		// Don't use 'Timeout' since it applies to the entire request/response.
	}
}

func DefaultPooledTransport() *Transport {
	transport := &Transport{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			ResponseHeaderTimeout: 60 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   75,
			MaxConnsPerHost:       75,
			ForceAttemptHTTP2:     true,
		},
		stats: newConnectionStats(),
		Dialer: &net.Dialer{
			Timeout: 30 * time.Second,
		},
	}

	transport.Transport.DialContext = transport.DialContext

	return transport
}
