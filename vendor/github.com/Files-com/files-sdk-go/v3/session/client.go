package session

import (
	files_sdk "github.com/Files-com/files-sdk-go/v3"
	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Client struct {
	files_sdk.Config
}

func (c *Client) Create(params files_sdk.SessionCreateParams, opts ...files_sdk.RequestResponseOption) (session files_sdk.Session, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "POST", Path: "/sessions", Params: params, Entity: &session}, opts...)
	return
}

func Create(params files_sdk.SessionCreateParams, opts ...files_sdk.RequestResponseOption) (session files_sdk.Session, err error) {
	return (&Client{}).Create(params, opts...)
}

func (c *Client) Delete(opts ...files_sdk.RequestResponseOption) (err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "DELETE", Path: "/sessions", Entity: nil}, opts...)
	return
}

func Delete(opts ...files_sdk.RequestResponseOption) (err error) {
	return (&Client{}).Delete(opts...)
}
