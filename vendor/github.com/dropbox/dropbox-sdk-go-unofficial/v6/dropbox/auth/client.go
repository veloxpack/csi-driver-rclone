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

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
)

// Client interface describes all routes in this namespace
type Client interface {
	// TokenFromOauth1 : Creates an OAuth 2.0 access token from the supplied
	// OAuth 1.0 access token.
	// Deprecated:
	TokenFromOauth1(arg *TokenFromOAuth1Arg) (res *TokenFromOAuth1Result, err error)
	// TokenRevoke : Disables the access token used to authenticate the call. If
	// there is a corresponding refresh token for the access token, this
	// disables that refresh token, as well as any other access tokens for that
	// refresh token.
	TokenRevoke() (err error)
}

// ContextClient interface describes all routes in this namespace with context support
type ContextClient interface {
	Client
	// TokenFromOauth1Context : Creates an OAuth 2.0 access token from the
	// supplied OAuth 1.0 access token.
	// Deprecated:
	TokenFromOauth1Context(ctx context.Context, arg *TokenFromOAuth1Arg) (res *TokenFromOAuth1Result, err error)
	// TokenRevokeContext : Disables the access token used to authenticate the
	// call. If there is a corresponding refresh token for the access token,
	// this disables that refresh token, as well as any other access tokens for
	// that refresh token.
	TokenRevokeContext(ctx context.Context) (err error)
}

type apiImpl dropbox.Context

// TokenFromOauth1APIError is an error-wrapper for the token/from_oauth1 route
type TokenFromOauth1APIError struct {
	dropbox.APIError
	EndpointError *TokenFromOAuth1Error `json:"error"`
}

// TokenFromOauth1Context : Creates an OAuth 2.0 access token from the supplied
// OAuth 1.0 access token.
// Deprecated:
func (dbx *apiImpl) TokenFromOauth1Context(ctx context.Context, arg *TokenFromOAuth1Arg) (res *TokenFromOAuth1Result, err error) {
	log.Printf("WARNING: API `TokenFromOauth1` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "auth",
		Route:        "token/from_oauth1",
		Auth:         "app",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TokenFromOauth1APIError
		err = ParseError(err, &appErr)
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

func (dbx *apiImpl) TokenFromOauth1(arg *TokenFromOAuth1Arg) (res *TokenFromOAuth1Result, err error) {
	return dbx.TokenFromOauth1Context(context.Background(), arg)
}

// TokenRevokeAPIError is an error-wrapper for the token/revoke route
type TokenRevokeAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// TokenRevokeContext : Disables the access token used to authenticate the call.
// If there is a corresponding refresh token for the access token, this disables
// that refresh token, as well as any other access tokens for that refresh
// token.
func (dbx *apiImpl) TokenRevokeContext(ctx context.Context) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "auth",
		Route:        "token/revoke",
		Auth:         "user",
		Style:        "rpc",
		Arg:          nil,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TokenRevokeAPIError
		err = ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) TokenRevoke() (err error) {
	return dbx.TokenRevokeContext(context.Background())
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
