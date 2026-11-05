package file

import (
	files_sdk "github.com/Files-com/files-sdk-go/v3"
	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

func (c *Client) UploadToRemoteServer(remoteServerID int64, destinationPath string, opts ...UploadOption) error {
	return c.Upload(append(opts, UploadWithDestinationPath(lib.UnderscoreDestinationPath("RemoteServers", remoteServerID, destinationPath)))...)
}

func UploadToRemoteServer(remoteServerID int64, destinationPath string, opts ...UploadOption) error {
	return (&Client{}).UploadToRemoteServer(remoteServerID, destinationPath, opts...)
}

func (c *Client) CopyToRemoteServer(params files_sdk.FileCopyParams, remoteServerID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("RemoteServers", remoteServerID, destinationPath)
	return c.Copy(params, opts...)
}

func CopyToRemoteServer(params files_sdk.FileCopyParams, remoteServerID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).CopyToRemoteServer(params, remoteServerID, destinationPath, opts...)
}

func (c *Client) MoveToRemoteServer(params files_sdk.FileMoveParams, remoteServerID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("RemoteServers", remoteServerID, destinationPath)
	return c.Move(params, opts...)
}

func MoveToRemoteServer(params files_sdk.FileMoveParams, remoteServerID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).MoveToRemoteServer(params, remoteServerID, destinationPath, opts...)
}

func (c *Client) UploadToSnapshot(snapshotID int64, destinationPath string, opts ...UploadOption) error {
	return c.Upload(append(opts, UploadWithDestinationPath(lib.UnderscoreDestinationPath("Snapshots", snapshotID, destinationPath)))...)
}

func UploadToSnapshot(snapshotID int64, destinationPath string, opts ...UploadOption) error {
	return (&Client{}).UploadToSnapshot(snapshotID, destinationPath, opts...)
}

func (c *Client) CopyToSnapshot(params files_sdk.FileCopyParams, snapshotID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("Snapshots", snapshotID, destinationPath)
	return c.Copy(params, opts...)
}

func CopyToSnapshot(params files_sdk.FileCopyParams, snapshotID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).CopyToSnapshot(params, snapshotID, destinationPath, opts...)
}

func (c *Client) MoveToSnapshot(params files_sdk.FileMoveParams, snapshotID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("Snapshots", snapshotID, destinationPath)
	return c.Move(params, opts...)
}

func MoveToSnapshot(params files_sdk.FileMoveParams, snapshotID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).MoveToSnapshot(params, snapshotID, destinationPath, opts...)
}

func (c *Client) UploadToChildSite(siteID int64, destinationPath string, opts ...UploadOption) error {
	return c.Upload(append(opts, UploadWithDestinationPath(lib.UnderscoreDestinationPath("Sites", siteID, destinationPath)))...)
}

func UploadToChildSite(siteID int64, destinationPath string, opts ...UploadOption) error {
	return (&Client{}).UploadToChildSite(siteID, destinationPath, opts...)
}

func (c *Client) CopyToChildSite(params files_sdk.FileCopyParams, siteID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("Sites", siteID, destinationPath)
	return c.Copy(params, opts...)
}

func CopyToChildSite(params files_sdk.FileCopyParams, siteID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).CopyToChildSite(params, siteID, destinationPath, opts...)
}

func (c *Client) MoveToChildSite(params files_sdk.FileMoveParams, siteID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	params.Destination = lib.UnderscoreDestinationPath("Sites", siteID, destinationPath)
	return c.Move(params, opts...)
}

func MoveToChildSite(params files_sdk.FileMoveParams, siteID int64, destinationPath string, opts ...files_sdk.RequestResponseOption) (files_sdk.FileAction, error) {
	return (&Client{}).MoveToChildSite(params, siteID, destinationPath, opts...)
}
