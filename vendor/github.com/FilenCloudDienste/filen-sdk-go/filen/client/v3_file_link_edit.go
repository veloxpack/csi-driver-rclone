package client

import (
	"context"
	"encoding/hex"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"github.com/google/uuid"
)

// v3FileLinkEditRequest represents the request structure for editing a file link.
type v3FileLinkEditRequest struct {
	LinkUUID       string `json:"uuid"`
	FileUUID       string `json:"fileUUID"`
	Expiration     string `json:"expiration"`
	Password       string `json:"password"`
	PasswordHashed string `json:"passwordHashed"`
	DownloadBtn    bool   `json:"downloadBtn"`
	Type           string `json:"type"`
	Salt           string `json:"salt"`
}

// postV3FileLinkEdit is a helper function that calls /v3/file/link/edit to modify link settings.
// It's used internally by higher-level functions that create or modify file sharing links.
func (c *Client) postV3FileLinkEdit(ctx context.Context, request v3FileLinkEditRequest) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/file/link/edit"), request)
	return err
}

// PostV3FileLinkEditEnable calls /v3/file/link/edit to create a new public sharing link for a file.
// It returns the newly created link UUID or an error if the operation fails.
func (c *Client) PostV3FileLinkEditEnable(ctx context.Context, file types.File) (string, error) {
	linkUUID := uuid.NewString()
	err := c.postV3FileLinkEdit(ctx, v3FileLinkEditRequest{
		LinkUUID:       linkUUID,
		FileUUID:       file.UUID,
		Expiration:     "never",
		Password:       "empty",
		PasswordHashed: crypto.V2Hash([]byte("empty")),
		DownloadBtn:    false,
		Type:           "enable",
		Salt:           hex.EncodeToString(crypto.GenerateRandomBytes(128)),
	})
	if err != nil {
		return "", err
	}
	return linkUUID, nil
}
