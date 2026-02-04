package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// V3UploadResponse represents the response structure from the upload endpoint.
type V3UploadResponse struct {
	Bucket string `json:"bucket"`
	Region string `json:"region"`
}

// PostV3Upload uploads a file chunk to the Filen storage backend.
// It handles the direct binary upload to the ingest servers and returns storage metadata.
func (c *Client) PostV3Upload(ctx context.Context, uuid string, chunkIdx int, parentUUID string, uploadKey string, data []byte) (*V3UploadResponse, error) {
	// build request
	dataHash := hex.EncodeToString(crypto.RunSHA512(data))
	url := &FilenURL{
		Type: URLTypeIngest,
		Path: fmt.Sprintf("/v3/upload?uuid=%s&index=%v&parent=%s&uploadKey=%s&hash=%s",
			uuid, chunkIdx, parentUUID, uploadKey, dataHash),
	}
	method := "POST"
	// Can't use the standard Client.RequestData because our request body is raw bytes
	req, err := c.buildReaderRequest(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	response, err := handleRequest(req, &c.httpClient, method, url)
	if err != nil {
		return nil, err
	}

	if !response.Status {
		return nil, errors.New("Cannot upload file chunk: " + response.Message)
	}

	uploadResponse := &V3UploadResponse{}
	err = response.IntoData(uploadResponse)
	if err != nil {
		return nil, err
	}
	return uploadResponse, nil
}
