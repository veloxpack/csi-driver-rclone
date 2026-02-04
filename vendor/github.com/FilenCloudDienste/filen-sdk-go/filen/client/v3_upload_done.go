package client

import (
	"context"
)

// V3UploadDoneRequest represents the request structure for completing a file upload.
type V3UploadDoneRequest struct {
	V3UploadEmptyRequest
	Chunks    int64    `json:"chunks"`
	Rm        string `json:"rm"`
	UploadKey string `json:"uploadKey"`
}

// V3UploadDoneResponse represents the response structure from the upload done endpoint.
type V3UploadDoneResponse struct {
	Chunks int64 `json:"chunks"`
	Size   int64 `json:"size"`
}

// PostV3UploadDone calls /v3/upload/done to finalize a file upload.
// This is called after all chunks have been successfully uploaded to confirm completion.
func (c *Client) PostV3UploadDone(ctx context.Context, request V3UploadDoneRequest) (*V3UploadDoneResponse, error) {
	response := &V3UploadDoneResponse{}
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/upload/done"), request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
