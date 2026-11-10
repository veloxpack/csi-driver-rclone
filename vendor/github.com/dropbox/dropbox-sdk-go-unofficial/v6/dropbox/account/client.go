// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package account

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
)

// Client interface describes all routes in this namespace
type Client interface {
	// DeleteProfilePhoto : Deletes the current user's profile photo.
	DeleteProfilePhoto(arg *DeleteProfilePhotoArg) (res *DeleteProfilePhotoResult, err error)
	// GetPhoto : This lovely endpoint gets the account photo of a given user.
	GetPhoto(arg *AccountPhotoGetArg) (res *AccountPhotoGetResult, content io.ReadCloser, err error)
	// SetProfilePhoto : Sets a user's profile photo.
	SetProfilePhoto(arg *SetProfilePhotoArg) (res *SetProfilePhotoResult, err error)
}

// ContextClient interface describes all routes in this namespace with context support
type ContextClient interface {
	Client
	// DeleteProfilePhotoContext : Deletes the current user's profile photo.
	DeleteProfilePhotoContext(ctx context.Context, arg *DeleteProfilePhotoArg) (res *DeleteProfilePhotoResult, err error)
	// GetPhotoContext : This lovely endpoint gets the account photo of a given
	// user.
	GetPhotoContext(ctx context.Context, arg *AccountPhotoGetArg) (res *AccountPhotoGetResult, content io.ReadCloser, err error)
	// SetProfilePhotoContext : Sets a user's profile photo.
	SetProfilePhotoContext(ctx context.Context, arg *SetProfilePhotoArg) (res *SetProfilePhotoResult, err error)
}

type apiImpl dropbox.Context

// DeleteProfilePhotoAPIError is an error-wrapper for the delete_profile_photo route
type DeleteProfilePhotoAPIError struct {
	dropbox.APIError
	EndpointError *DeleteProfilePhotoError `json:"error"`
}

// DeleteProfilePhotoContext : Deletes the current user's profile photo.
func (dbx *apiImpl) DeleteProfilePhotoContext(ctx context.Context, arg *DeleteProfilePhotoArg) (res *DeleteProfilePhotoResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "account",
		Route:        "delete_profile_photo",
		Auth:         "user",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DeleteProfilePhotoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) DeleteProfilePhoto(arg *DeleteProfilePhotoArg) (res *DeleteProfilePhotoResult, err error) {
	return dbx.DeleteProfilePhotoContext(context.Background(), arg)
}

// GetPhotoAPIError is an error-wrapper for the get_photo route
type GetPhotoAPIError struct {
	dropbox.APIError
	EndpointError *AccountPhotoGetError `json:"error"`
}

// GetPhotoContext : This lovely endpoint gets the account photo of a given
// user.
func (dbx *apiImpl) GetPhotoContext(ctx context.Context, arg *AccountPhotoGetArg) (res *AccountPhotoGetResult, content io.ReadCloser, err error) {
	req := dropbox.Request{
		Host:         "content",
		Namespace:    "account",
		Route:        "get_photo",
		Auth:         "user",
		Style:        "download",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GetPhotoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	content = respBody
	return
}

func (dbx *apiImpl) GetPhoto(arg *AccountPhotoGetArg) (res *AccountPhotoGetResult, content io.ReadCloser, err error) {
	return dbx.GetPhotoContext(context.Background(), arg)
}

// SetProfilePhotoAPIError is an error-wrapper for the set_profile_photo route
type SetProfilePhotoAPIError struct {
	dropbox.APIError
	EndpointError *SetProfilePhotoError `json:"error"`
}

// SetProfilePhotoContext : Sets a user's profile photo.
func (dbx *apiImpl) SetProfilePhotoContext(ctx context.Context, arg *SetProfilePhotoArg) (res *SetProfilePhotoResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "account",
		Route:        "set_profile_photo",
		Auth:         "user",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr SetProfilePhotoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) SetProfilePhoto(arg *SetProfilePhotoArg) (res *SetProfilePhotoResult, err error) {
	return dbx.SetProfilePhotoContext(context.Background(), arg)
}

// NewContext returns a ContextClient implementation for this namespace
func NewContext(c dropbox.Config) ContextClient {
	ctx := apiImpl(dropbox.NewContext(c))
	return &ctx
}

// New returns a Client implementation for this namespace
func New(c dropbox.Config) Client {
	return NewContext(c)
}
