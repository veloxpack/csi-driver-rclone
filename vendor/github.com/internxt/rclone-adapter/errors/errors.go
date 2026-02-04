package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPError preserves HTTP response details
type HTTPError struct {
	Response  *http.Response
	Operation string
	Message   string
	Body      []byte
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (status %d)", e.Operation, e.Message, e.Response.StatusCode)
	}
	return fmt.Sprintf("%s: status %d", e.Operation, e.Response.StatusCode)
}

// Temporary implements net.Error interface
func (e *HTTPError) Temporary() bool {
	return e.Response.StatusCode >= 500
}

// StatusCode returns the HTTP status code
func (e *HTTPError) StatusCode() int {
	return e.Response.StatusCode
}

// NewHTTPError creates an HTTPError from a response
func NewHTTPError(resp *http.Response, operation string) error {
	body, _ := io.ReadAll(resp.Body)

	httpErr := &HTTPError{
		Response:  resp,
		Operation: operation,
		Body:      body,
	}

	var backendErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if json.Unmarshal(body, &backendErr) == nil {
		if backendErr.Message != "" {
			httpErr.Message = backendErr.Message
		} else if backendErr.Error != "" {
			httpErr.Message = backendErr.Error
		}
	} else {
		httpErr.Message = string(body)
	}

	return httpErr
}
