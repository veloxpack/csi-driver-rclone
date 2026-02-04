// Package client handles HTTP requests to the API and storage backends.
//
// API definitions are at https://gateway.filen.io/v3/docs.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rclone/rclone/fs/fshttp"
	"io"
	"net/http"
	"strings"
)

// UnauthorizedClient represents a client without authorization
// that can make requests to endpoints not requiring authentication.
type UnauthorizedClient struct {
	httpClient http.Client // cached request client
}

// Client extends UnauthorizedClient with API key authentication
// to access protected Filen API endpoints.
type Client struct {
	UnauthorizedClient
	APIKey string // the Filen API key
}

// New creates a new UnauthorizedClient with the provided context.
// The context is used to create the underlying HTTP client.
func New(ctx context.Context) *UnauthorizedClient {
	return &UnauthorizedClient{
		httpClient: *fshttp.NewClient(ctx),
	}
}

// Authorize creates an authorized Client from an UnauthorizedClient
// by adding the provided API key.
func (uc *UnauthorizedClient) Authorize(apiKey string) *Client {
	return &Client{
		UnauthorizedClient: *uc,
		APIKey:             apiKey,
	}
}

// NewWithAPIKey creates a new authorized Client with the provided context and API key.
// This is a convenience function that combines New and Authorize.
func NewWithAPIKey(ctx context.Context, apiKey string) *Client {
	return &Client{
		UnauthorizedClient: *New(ctx),
		APIKey:             apiKey,
	}
}

// RequestError carries information on a failed HTTP request.
// It implements the error interface and provides detailed information
// about where and why the request failed.
type RequestError struct {
	Message         string    // description of where the error occurred
	Method          string    // HTTP method of the request
	URL             *FilenURL // URL path of the request
	UnderlyingError error     // the underlying error
}

// Error returns a formatted error string for RequestError.
// It includes the HTTP method, URL, error message, and underlying error if present.
func (e *RequestError) Error() string {
	var builder strings.Builder
	builder.WriteString(e.Method)
	builder.WriteRune(' ')
	if e.URL.CachedUrl != "" {
		builder.WriteString(fmt.Sprintf("cached: %s", e.URL.CachedUrl))
	} else {
		builder.WriteString(e.URL.Path)
	}
	builder.WriteString(fmt.Sprintf(": %s", e.Message))
	if e.UnderlyingError != nil {
		builder.WriteString(fmt.Sprintf(" (%s)", e.UnderlyingError))
	}
	return builder.String()
}

// cannotSendError returns a RequestError from an error that occurred while sending an HTTP request.
// It formats the error message to indicate a request sending failure.
func cannotSendError(method string, url *FilenURL, err error) error {
	return &RequestError{
		Message:         "Cannot send request",
		Method:          method,
		URL:             url,
		UnderlyingError: err,
	}
}

// buildReaderRequest creates an HTTP request with the provided context, method, URL, and data.
// It returns the request or an error if the request cannot be created.
func (uc *UnauthorizedClient) buildReaderRequest(ctx context.Context, method string, url *FilenURL, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url.String(), data)
	if err != nil {
		return nil, &RequestError{"Cannot build requestData", method, url, err}
	}
	return req, nil
}

// buildReaderRequest creates an HTTP request with the provided context, method, URL, and data.
// It extends the unauthorized client method by adding the API key authorization header.
func (c *Client) buildReaderRequest(ctx context.Context, method string, url *FilenURL, data io.Reader) (*http.Request, error) {
	var request, err = c.UnauthorizedClient.buildReaderRequest(ctx, method, url, data)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.APIKey)
	return request, nil
}

// buildJSONRequest creates an HTTP request with JSON content type.
// It marshals the requestData to JSON and calls buildReaderRequest.
func (uc *UnauthorizedClient) buildJSONRequest(ctx context.Context, method string, url *FilenURL, requestData any) (*http.Request, error) {
	var marshalled []byte
	if requestData != nil {
		var err error
		marshalled, err = json.Marshal(requestData)
		if err != nil {
			return nil, &RequestError{fmt.Sprintf("Cannot unmarshal requestData body %#v", requestData), method, url, err}
		}
	}
	req, err := uc.buildReaderRequest(ctx, method, url, bytes.NewReader(marshalled))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// buildJSONRequest creates an HTTP request with JSON content type and API key authorization.
// It extends the unauthorized client method by adding the authorization header.
func (c *Client) buildJSONRequest(ctx context.Context, method string, url *FilenURL, requestData any) (*http.Request, error) {
	var request, err = c.UnauthorizedClient.buildJSONRequest(ctx, method, url, requestData)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.APIKey)
	return request, nil
}

// parseResponse reads and unmarshals an HTTP response body into an aPIResponse.
// It takes the HTTP method, path, and response as arguments.
// If the response body cannot be read or unmarshalled, it returns a RequestError.
// Otherwise, it returns the parsed aPIResponse.
func parseResponse(method string, url *FilenURL, response *http.Response) (*aPIResponse, error) {
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &RequestError{"Cannot read response body", method, url, err}
	}
	apiResponse := aPIResponse{}
	err = json.Unmarshal(resBody, &apiResponse)
	if err != nil {
		return nil, &RequestError{fmt.Sprintf("Cannot unmarshal response %s", string(resBody)), method, url, nil}
	}
	return &apiResponse, nil
}

// handleRequest sends an HTTP request and processes the response.
// It takes a http.Request object, the associated http.Client, and the method and path
// as parameters. It returns an aPIResponse containing the parsed response data, or a
// RequestError if the request fails or the response cannot be parsed.
func handleRequest(request *http.Request, httpClient *http.Client, method string, url *FilenURL) (*aPIResponse, error) {
	//startTime := time.Now()
	res, err := httpClient.Do(request)
	if err != nil {
		return nil, cannotSendError(method, url, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	apiRes, err := parseResponse(method, url, res)
	if err != nil {
		return nil, err
	}
	err = apiRes.CheckError()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Request %s %s took %s\n", method, url, time.Since(startTime))
	return apiRes, nil
}

// convertIntoResponseData unmarshals the response data into the provided output data structure.
// It returns a RequestError if the unmarshalling process fails.
func convertIntoResponseData(method string, url *FilenURL, response *aPIResponse, outData any) error {
	err := response.IntoData(outData)
	if err != nil {
		return &RequestError{
			Message:         fmt.Sprintf("Cannot unmarshal response data %#v", response.Data),
			Method:          method,
			URL:             url,
			UnderlyingError: err,
		}
	}
	return nil
}

// Request sends an HTTP request to the Filen API without authorization.
// It takes the context, HTTP method, URL, and request data as parameters.
// It returns the API response or an error if the request fails.
func (uc *UnauthorizedClient) Request(ctx context.Context, method string, url *FilenURL, requestData any) (*aPIResponse, error) {
	request, err := uc.buildJSONRequest(ctx, method, url, requestData)
	if err != nil {
		return nil, err
	}
	return handleRequest(request, &uc.httpClient, method, url)
}

// RequestData sends an HTTP request to the Filen API without authorization and unmarshals
// the response data into the provided output structure.
// It takes the context, HTTP method, URL, request data, and output data structure as parameters.
// It returns the API response or an error if the request fails or unmarshalling fails.
func (uc *UnauthorizedClient) RequestData(ctx context.Context, method string, url *FilenURL, requestData any, outData any) (*aPIResponse, error) {
	response, err := uc.Request(ctx, method, url, requestData)
	if err != nil {
		return nil, err
	}
	err = convertIntoResponseData(method, url, response, outData)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Request sends an HTTP request to the Filen API with authorization.
// It takes the context, HTTP method, URL, and request data as parameters.
// It returns the API response or an error if the request fails.
func (c *Client) Request(ctx context.Context, method string, url *FilenURL, requestData any) (*aPIResponse, error) {
	request, err := c.buildJSONRequest(ctx, method, url, requestData)
	if err != nil {
		return nil, err
	}
	return handleRequest(request, &c.httpClient, method, url)
}

// RequestData sends an HTTP request to the Filen API with authorization and unmarshals
// the response data into the provided output structure.
// It takes the context, HTTP method, URL, request data, and output data structure as parameters.
// It returns the API response or an error if the request fails or unmarshalling fails.
func (c *Client) RequestData(ctx context.Context, method string, url *FilenURL, requestData any, outData any) (*aPIResponse, error) {
	response, err := c.Request(ctx, method, url, requestData)
	if err != nil {
		return nil, err
	}
	err = convertIntoResponseData(method, url, response, outData)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api

// aPIResponse represents a response from the Filen API.
// It contains the status, message, code, and data returned by the API.
type aPIResponse struct {
	Status  bool            `json:"status"`  // whether the request was successful
	Message string          `json:"message"` // additional information
	Code    string          `json:"code"`    // a status code
	Data    json.RawMessage `json:"data"`    // response body, or nil
}

// CheckError checks if the API response indicates an error.
// It returns an error if the response status is false.
func (res *aPIResponse) CheckError() error {
	if !res.Status {
		return fmt.Errorf("response error: %s %s", res.Message, res.Code)
	}
	return nil
}

// String returns a string representation of the API response.
// It includes the status, message, code, and data.
func (res *aPIResponse) String() string {
	return fmt.Sprintf("ApiResponse{status: %t, message: %s, code: %s, data: %s}", res.Status, res.Message, res.Code, res.Data)
}

// IntoData unmarshals the response body into the provided data structure.
//
// If the response does not contain a body, an error is returned.
// If the unmarshalling process fails, the error is returned.
func (res *aPIResponse) IntoData(data any) error {
	if res.Data == nil {
		return errors.New(fmt.Sprintf("No data in response %s", res))
	}
	err := json.Unmarshal(res.Data, data)
	if err != nil {
		return err
	}
	return nil
}

// file chunks

// DownloadFileChunk downloads a file chunk from the Filen storage backend.
// It takes the context, file UUID, region, bucket, and chunk index as parameters.
// It returns the chunk data or an error if the download fails.
func (c *Client) DownloadFileChunk(ctx context.Context, uuid string, region string, bucket string, chunkIdx int64) ([]byte, error) {
	url := &FilenURL{
		Type: URLTypeEgest,
		Path: fmt.Sprintf("/%s/%s/%s/%v", region, bucket, uuid, chunkIdx),
	}

	// Can't use the standard Client.RequestData because the response body is raw bytes
	request, err := c.buildJSONRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Do(request)
	if err != nil {
		return nil, cannotSendError("GET", url, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
