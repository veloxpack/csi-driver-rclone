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

// Package account : has no documentation (yet)
package account

import (
	"encoding/json"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
)

// AccountPhotoGetArg : has no documentation (yet)
type AccountPhotoGetArg struct {
	// DbxAccountId : Encoded ID of the user. Must start either with 'dbid:' or
	// 'dbaphid:'.
	DbxAccountId string `json:"dbx_account_id"`
	// Size : A string representing the size of the photo.
	Size string `json:"size"`
	// CircleCrop : True if the photo should be cropped and false otherwise.
	CircleCrop bool `json:"circle_crop"`
	// ExpectAccountPhoto : True if we expect account photo to exist.
	ExpectAccountPhoto bool `json:"expect_account_photo"`
}

// NewAccountPhotoGetArg returns a new AccountPhotoGetArg instance
func NewAccountPhotoGetArg(DbxAccountId string, Size string, CircleCrop bool, ExpectAccountPhoto bool) *AccountPhotoGetArg {
	s := new(AccountPhotoGetArg)
	s.DbxAccountId = DbxAccountId
	s.Size = Size
	s.CircleCrop = CircleCrop
	s.ExpectAccountPhoto = ExpectAccountPhoto
	return s
}

// AccountPhotoGetError : has no documentation (yet)
type AccountPhotoGetError struct {
	dropbox.Tagged
	// ThumbnailError : Indicates infrastructural failure.
	ThumbnailError *ThumbnailError `json:"thumbnail_error,omitempty"`
}

// Valid tag values for AccountPhotoGetError
const (
	AccountPhotoGetErrorThumbnailError              = "thumbnail_error"
	AccountPhotoGetErrorAccountPhotoMissing         = "account_photo_missing"
	AccountPhotoGetErrorExpectedAccountPhotoMissing = "expected_account_photo_missing"
	AccountPhotoGetErrorOther                       = "other"
)

// UnmarshalJSON deserializes into a AccountPhotoGetError instance
func (u *AccountPhotoGetError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// ThumbnailError : Indicates infrastructural failure.
		ThumbnailError *ThumbnailError `json:"thumbnail_error,omitempty"`
	}
	var w wrap
	var err error
	if err = json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "thumbnail_error":
		u.ThumbnailError = w.ThumbnailError

	}
	return nil
}

// AccountPhotoGetResult : has no documentation (yet)
type AccountPhotoGetResult struct {
	// ContentType : The data returned by get_photo.
	ContentType string `json:"content_type"`
}

// NewAccountPhotoGetResult returns a new AccountPhotoGetResult instance
func NewAccountPhotoGetResult(ContentType string) *AccountPhotoGetResult {
	s := new(AccountPhotoGetResult)
	s.ContentType = ContentType
	return s
}

// DeleteProfilePhotoArg : This struct is empty. The comment here is
// intentionally emitted to avoid indentation issues with Stone.
type DeleteProfilePhotoArg struct {
}

// NewDeleteProfilePhotoArg returns a new DeleteProfilePhotoArg instance
func NewDeleteProfilePhotoArg() *DeleteProfilePhotoArg {
	s := new(DeleteProfilePhotoArg)
	return s
}

// DeleteProfilePhotoError : This union is empty. The comment here is
// intentionally emitted to avoid indentation issues with Stone.
type DeleteProfilePhotoError struct {
	dropbox.Tagged
}

// Valid tag values for DeleteProfilePhotoError
const (
	DeleteProfilePhotoErrorOther = "other"
)

// DeleteProfilePhotoResult : This struct is empty. The comment here is
// intentionally emitted to avoid indentation issues with Stone.
type DeleteProfilePhotoResult struct {
}

// NewDeleteProfilePhotoResult returns a new DeleteProfilePhotoResult instance
func NewDeleteProfilePhotoResult() *DeleteProfilePhotoResult {
	s := new(DeleteProfilePhotoResult)
	return s
}

// PhotoSourceArg : has no documentation (yet)
type PhotoSourceArg struct {
	dropbox.Tagged
	// Base64Data : Image data in base64-encoded bytes.
	Base64Data string `json:"base64_data,omitempty"`
}

// Valid tag values for PhotoSourceArg
const (
	PhotoSourceArgBase64Data = "base64_data"
	PhotoSourceArgOther      = "other"
)

// UnmarshalJSON deserializes into a PhotoSourceArg instance
func (u *PhotoSourceArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Base64Data : Image data in base64-encoded bytes.
		Base64Data string `json:"base64_data,omitempty"`
	}
	var w wrap
	var err error
	if err = json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "base64_data":
		u.Base64Data = w.Base64Data

	}
	return nil
}

// SetProfilePhotoArg : has no documentation (yet)
type SetProfilePhotoArg struct {
	// Photo : Image to set as the user's new profile photo.
	Photo *PhotoSourceArg `json:"photo"`
}

// NewSetProfilePhotoArg returns a new SetProfilePhotoArg instance
func NewSetProfilePhotoArg(Photo *PhotoSourceArg) *SetProfilePhotoArg {
	s := new(SetProfilePhotoArg)
	s.Photo = Photo
	return s
}

// SetProfilePhotoError : has no documentation (yet)
type SetProfilePhotoError struct {
	dropbox.Tagged
}

// Valid tag values for SetProfilePhotoError
const (
	SetProfilePhotoErrorFileTypeError  = "file_type_error"
	SetProfilePhotoErrorFileSizeError  = "file_size_error"
	SetProfilePhotoErrorDimensionError = "dimension_error"
	SetProfilePhotoErrorThumbnailError = "thumbnail_error"
	SetProfilePhotoErrorTransientError = "transient_error"
	SetProfilePhotoErrorOther          = "other"
)

// SetProfilePhotoResult : has no documentation (yet)
type SetProfilePhotoResult struct {
	// ProfilePhotoUrl : URL for the photo representing the user, if one is set.
	ProfilePhotoUrl string `json:"profile_photo_url"`
}

// NewSetProfilePhotoResult returns a new SetProfilePhotoResult instance
func NewSetProfilePhotoResult(ProfilePhotoUrl string) *SetProfilePhotoResult {
	s := new(SetProfilePhotoResult)
	s.ProfilePhotoUrl = ProfilePhotoUrl
	return s
}

// ThumbnailError : has no documentation (yet)
type ThumbnailError struct {
	dropbox.Tagged
}

// Valid tag values for ThumbnailError
const (
	ThumbnailErrorPermanentFailure = "permanent_failure"
	ThumbnailErrorTemporaryFailure = "temporary_failure"
	ThumbnailErrorOther            = "other"
)
