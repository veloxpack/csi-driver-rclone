package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
)

// v3dirContentRequest represents the request structure for the directory content endpoint.
type v3dirContentRequest struct {
	UUID string `json:"uuid"`
}

// V3DirContentResponse represents the response structure from the directory content endpoint.
// It contains lists of files and folders within the requested directory.
type V3DirContentResponse struct {
	Uploads []struct {
		UUID      string                       `json:"uuid"`
		Metadata  crypto.EncryptedString       `json:"metadata"`
		Rm        string                       `json:"rm"`
		Timestamp int                          `json:"timestamp"`
		Chunks    int64                        `json:"chunks"`
		Size      int64                        `json:"size"`
		Bucket    string                       `json:"bucket"`
		Region    string                       `json:"region"`
		Parent    string                       `json:"parent"`
		Version   crypto.FileEncryptionVersion `json:"version"`
		Favorited int                          `json:"favorited"`
	} `json:"uploads"`
	Folders []struct {
		UUID      string                 `json:"uuid"`
		Metadata  crypto.EncryptedString `json:"name"` // name is actually the metadata
		Parent    string                 `json:"parent"`
		Color     types.DirColor         `json:"color"`
		Timestamp int                    `json:"timestamp"`
		Favorited int                    `json:"favorited"`
		IsSync    int                    `json:"is_sync"`
		IsDefault int                    `json:"is_default"`
	} `json:"folders"`
}

// PostV3DirContent calls /v3/dir/content to retrieve the contents of a directory.
// It returns files and folders within the specified directory UUID.
func (c *Client) PostV3DirContent(ctx context.Context, uuid string) (*V3DirContentResponse, error) {
	directoryContent := &V3DirContentResponse{}
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/content"), v3dirContentRequest{
		UUID: uuid,
	}, directoryContent)
	return directoryContent, err
}
