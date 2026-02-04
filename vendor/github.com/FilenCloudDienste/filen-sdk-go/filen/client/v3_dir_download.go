package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3DirDownloadRequest represents the request structure for downloading directory information.
type v3DirDownloadRequest struct {
	UUID      string `json:"uuid"`
	SkipCache string `json:"skipCache"`
}

// v3DirDownloadLinkedRequest represents the request structure for downloading shared directory information.
type v3DirDownloadLinkedRequest struct {
	UUID       string `json:"uuid"`
	SkipCache  string `json:"skipCache"`
	ParentUUID string `json:"parent"`
	Password   string `json:"password"`
}

// V3DirDownloadResponse represents the response structure from the directory download endpoint.
// It contains information needed to download files and navigate folders.
type V3DirDownloadResponse struct {
	Files []struct {
		UUID     string                       `json:"uuid"`
		Bucket   string                       `json:"bucket"`
		Region   string                       `json:"region"`
		Chunks   int64                        `json:"chunks"`
		Parent   string                       `json:"parent"`
		Metadata crypto.EncryptedString       `json:"metadata"`
		Version  crypto.FileEncryptionVersion `json:"version"`

		// optional
		Name       string `json:"name"`
		Size       string `json:"size"`
		MimeType   string `json:"mime"`
		ChunksSize int    `json:"chunksSize"` // no idea what this is
		Timestamp  int    `json:"timestamp"`
		Favorited  bool   `json:"favorited"`
	} `json:"files"`
	Folders []struct {
		UUID     string                 `json:"uuid"`
		Metadata crypto.EncryptedString `json:"name"` // name is actually the metadata
		Parent   string                 `json:"parent"`

		// optional
		Timestamp int    `json:"timestamp"`
		Color     string `json:"color"`
		Favorited bool   `json:"favorited"`
	} `json:"folders"`
}

// postV3DirDownload is a helper function for directory download endpoints.
// It handles the common request/response processing for various download endpoints.
func (c *Client) postV3DirDownload(ctx context.Context, endpoint string, req v3DirDownloadRequest) (*V3DirDownloadResponse, error) {
	var resp V3DirDownloadResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL(endpoint), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PostV3DirDownload calls /v3/dir/download to retrieve directory download information.
// This provides the necessary metadata to download files from a directory.
func (c *Client) PostV3DirDownload(ctx context.Context, uuid string) (*V3DirDownloadResponse, error) {
	return c.postV3DirDownload(ctx, "/v3/dir/download", v3DirDownloadRequest{
		UUID: uuid,
	})
}

// PostV3DirDownloadShared calls /v3/dir/download/shared to retrieve information about a shared directory.
// This endpoint is used for directories shared directly with the user.
func (c *Client) PostV3DirDownloadShared(ctx context.Context, uuid string) (*V3DirDownloadResponse, error) {
	return c.postV3DirDownload(ctx, "/v3/dir/download/shared", v3DirDownloadRequest{
		UUID: uuid,
	})
}

// PostV3DirDownloadLinked calls an endpoint to download a directory shared via a public link.
// Not yet implemented.
func (c *Client) PostV3DirDownloadLinked(ctx context.Context, uuid, linkUUID, linkHasPassword, linkSalt string) (*V3DirDownloadResponse, error) {
	panic("unimplemented")
}
