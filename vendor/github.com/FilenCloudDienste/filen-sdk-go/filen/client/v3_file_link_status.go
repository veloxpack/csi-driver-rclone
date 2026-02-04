package client

import "context"

// V3FileLinkStatusResponse represents the response structure from the file link status endpoint.
type V3FileLinkStatusResponse struct {
	LinkUUID       string `json:"uuid"`
	Enabled        bool   `json:"enabled"`
	Expiration     int    `json:"expiration"`
	ExpirationText string `json:"expirationText"`
	DownloadBtn    int    `json:"downloadBtn"`
	Password       string `json:"password"`
}

// v3FileLinkStatusRequest represents the request structure for checking a file link's status.
type v3FileLinkStatusRequest struct {
	UUID string `json:"uuid"`
}

// V3FileLinkStatus calls /v3/file/link/status to check if a file sharing link exists and get its details.
// The status includes information like expiration, password protection, and download button settings.
func (c *Client) V3FileLinkStatus(ctx context.Context, uuid string) (*V3FileLinkStatusResponse, error) {
	var response V3FileLinkStatusResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/file/link/status"), &v3FileLinkStatusRequest{
		UUID: uuid,
	}, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
