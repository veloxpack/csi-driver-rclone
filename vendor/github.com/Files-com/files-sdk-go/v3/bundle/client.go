package bundle

import (
	files_sdk "github.com/Files-com/files-sdk-go/v3"
	lib "github.com/Files-com/files-sdk-go/v3/lib"
	listquery "github.com/Files-com/files-sdk-go/v3/listquery"
)

type Client struct {
	files_sdk.Config
}

type Iter struct {
	*files_sdk.Iter
	*Client
}

func (i *Iter) Reload(opts ...files_sdk.RequestResponseOption) files_sdk.IterI {
	return &Iter{Iter: i.Iter.Reload(opts...).(*files_sdk.Iter), Client: i.Client}
}

func (i *Iter) Bundle() files_sdk.Bundle {
	return i.Current().(files_sdk.Bundle)
}

func (i *Iter) LoadResource(identifier interface{}, opts ...files_sdk.RequestResponseOption) (interface{}, error) {
	params := files_sdk.BundleFindParams{}
	if id, ok := identifier.(int64); ok {
		params.Id = id
	}
	return i.Client.Find(params, opts...)
}

func (c *Client) List(params files_sdk.BundleListParams, opts ...files_sdk.RequestResponseOption) (*Iter, error) {
	i := &Iter{Iter: &files_sdk.Iter{}, Client: c}
	path, err := lib.BuildPath("/bundles", params)
	if err != nil {
		return i, err
	}
	i.ListParams = &params
	list := files_sdk.BundleCollection{}
	i.Query = listquery.Build(c.Config, path, &list, opts...)
	return i, nil
}

func List(params files_sdk.BundleListParams, opts ...files_sdk.RequestResponseOption) (*Iter, error) {
	return (&Client{}).List(params, opts...)
}

func (c *Client) Find(params files_sdk.BundleFindParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "GET", Path: "/bundles/{id}", Params: params, Entity: &bundle}, opts...)
	return
}

func Find(params files_sdk.BundleFindParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	return (&Client{}).Find(params, opts...)
}

func (c *Client) Create(params files_sdk.BundleCreateParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "POST", Path: "/bundles", Params: params, Entity: &bundle}, opts...)
	return
}

func Create(params files_sdk.BundleCreateParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	return (&Client{}).Create(params, opts...)
}

func (c *Client) Share(params files_sdk.BundleShareParams, opts ...files_sdk.RequestResponseOption) (err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "POST", Path: "/bundles/{id}/share", Params: params, Entity: nil}, opts...)
	return
}

func Share(params files_sdk.BundleShareParams, opts ...files_sdk.RequestResponseOption) (err error) {
	return (&Client{}).Share(params, opts...)
}

func (c *Client) Update(params files_sdk.BundleUpdateParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "PATCH", Path: "/bundles/{id}", Params: params, Entity: &bundle}, opts...)
	return
}

func Update(params files_sdk.BundleUpdateParams, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	return (&Client{}).Update(params, opts...)
}

func (c *Client) UpdateWithMap(params map[string]interface{}, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "PATCH", Path: "/bundles/{id}", Params: params, Entity: &bundle}, opts...)
	return
}

func UpdateWithMap(params map[string]interface{}, opts ...files_sdk.RequestResponseOption) (bundle files_sdk.Bundle, err error) {
	return (&Client{}).UpdateWithMap(params, opts...)
}

func (c *Client) Delete(params files_sdk.BundleDeleteParams, opts ...files_sdk.RequestResponseOption) (err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "DELETE", Path: "/bundles/{id}", Params: params, Entity: nil}, opts...)
	return
}

func Delete(params files_sdk.BundleDeleteParams, opts ...files_sdk.RequestResponseOption) (err error) {
	return (&Client{}).Delete(params, opts...)
}
