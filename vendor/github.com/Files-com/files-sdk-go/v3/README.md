# Files.com Go Client

The Files.com Go SDK provides a direct, high performance integration to Files.com from applications written in Go.

Files.com is the cloud-native, next-gen MFT, SFTP, and secure file-sharing platform that replaces brittle legacy servers with one always-on, secure fabric. Automate mission-critical file flows—across any cloud, protocol, or partner—while supporting human collaboration and eliminating manual work.

With universal SFTP, AS2, HTTPS, and 50+ native connectors backed by military-grade encryption, Files.com unifies governance, visibility, and compliance in a single pane of glass.

The content included here should be enough to get started, but please visit our
[Developer Documentation Website](https://developers.files.com/go/) for the complete documentation.

## Introduction

The Files.com Go client library provides convenient access to all aspects of Files.com from applications written in the Go language.

Files.com customers use our Go client library for directly working with files and folders as well as performing management tasks such as adding/removing users, onboarding counterparties, retrieving information about automations and more.

### Installation

Make sure your project is using Go Modules (it will have a `go.mod` file in its
root if it already is):

``` shell
go mod init
```

Then, reference files-sdk-go in a Go program with `import`:

``` go
import (
    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/folder"
)
```

Run any of the normal `go` commands (`build`/`install`/`test`). The Go
toolchain will resolve and fetch the files module automatically.

### Files.com is Committed to Go

Go is a core language used by the Files.com team for internal development.  This library is directly used by the Files.com CLI app, Files.com Desktop App v6, the official Files.com Terraform integration, and the official Files.com RClone integration.

As such, this library is actively developed and should be expected to be highly performant.

Explore the [files-sdk-go](https://github.com/Files-com/files-sdk-go) code on GitHub.

### Getting Support

The Files.com Support team provides official support for all of our official Files.com integration tools.

To initiate a support conversation, you can send an [Authenticated Support Request](https://www.files.com/docs/overview/requesting-support) or simply send an E-Mail to support@files.com.

## Authentication

There are two ways to authenticate: API Key authentication and Session-based authentication.

### Authenticate with an API Key

Authenticating with an API key is the recommended authentication method for most scenarios, and is
the method used in the examples on this site.

To use an API Key, first generate an API key from the [web
interface](https://www.files.com/docs/sdk-and-apis/api-keys) or [via the API or an
SDK](/go/resources/developers/api-keys).

Note that when using a user-specific API key, if the user is an administrator, you will have full
access to the entire API. If the user is not an administrator, you will only be able to access files
that user can access, and no access will be granted to site administration functions in the API.

```go title="Example Request"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/folder"
)

// You can specify an API key in the GlobalConfig, and use that config when creating clients.
files_sdk.GlobalConfig.APIKey = "YOUR_API_KEY"
client := folder.Client{Config: files_sdk.GlobalConfig}
it, err := client.ListFor(files_sdk.FolderListForParams{})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

// Alternatively, you can specify the API key on a per-request basis using the Config struct.
config := files_sdk.Config{APIKey: "YOUR_API_KEY"}.Init()
client := folder.Client{Config: config}
it, err := client.ListFor(files_sdk.FolderListForParams{})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

// If the API Key is available in the `FILES_API_KEY` environment variable you do not need to create clients.
it, err := folder.ListFor(files_sdk.FolderListForParams{})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

Don't forget to replace the placeholder, `YOUR_API_KEY`, with your actual API key.

### Authenticate with a Session

You can also authenticate by creating a user session using the username and
password of an active user. If the user is an administrator, the session will have full access to
all capabilities of Files.com. Sessions created from regular user accounts will only be able to access files that
user can access, and no access will be granted to site administration functions.

Sessions use the exact same session timeout settings as web interface sessions. When a
session times out, simply create a new session and resume where you left off. This process is not
automatically handled by our SDKs because we do not want to store password information in memory without
your explicit consent.

#### Logging In

To create a session, create a `session Client` object that points to the subdomain of the Files.com site.

The `Create` method on the `session` client can then be used to create a `Session` object which can be used to authenticate SDK method calls.

```go title="Example Request"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/folder"
    "github.com/Files-com/files-sdk-go/v3/session"
)

sessionClient := session.Client{}
thisSession, err := sessionClient.Create(files_sdk.SessionCreateParams{Username: "USERNAME", Password: "PASSWORD"})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

config := files_sdk.Config{SessionId: thisSession.Id}.Init()
folderClient := folder.Client{Config: config}

it, err := folderClient.ListFor(files_sdk.FolderListForParams{})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

#### Using a Session

Once a session has been created, the `Session.Id` can be set in a `Config` object, which can then be used to authenticate `Client` objects.

```go title="Example Request"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/folder"
)

config := files_sdk.Config{SessionId: thisSession.Id}.Init()
folderClient := folder.Client{Config: config}

it, err := folderClient.ListFor(files_sdk.FolderListForParams{})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

#### Logging Out

User sessions can be ended by calling `Delete()` on the `Session` client.

```go title="Example Request"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/session"
)

sessionClient := session.Client{Config: files_sdk.Config{SessionId: thisSession.Id}.Init()}
err := sessionClient.Delete()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

## Configuration

Global configuration is performed by providing a `files_sdk.Config` object to the `Client`.

### Configuration Options

#### Base URL

Setting the base URL for the API is required if your site is configured to disable global acceleration.
This can also be set to use a mock server in development or CI.

```go title="Example setting"
import (
    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/file"
)

config := files_sdk.Config{
    EndpointOverride: "https://MY-SUBDOMAIN.files.com",
}.Init()
client := file.Client{Config: config}
```

## Sort and Filter

Several of the Files.com API resources have list operations that return multiple instances of the
resource. The List operations can be sorted and filtered.

### Sorting

To sort the returned data, pass in the ```SortBy``` method argument.

Each resource supports a unique set of valid sort fields and can only be sorted by one field at a
time.

The argument value is a Go ```map[string]interface{}``` map that has a key of the resource field
name to sort on and a value of either ```"asc"``` or ```"desc"``` to specify the sort order.

#### Special note about the List Folder Endpoint

For historical reasons, and to maintain compatibility
with a variety of other cloud-based MFT and EFSS services, Folders will always be listed before Files
when listing a Folder.  This applies regardless of the sorting parameters you provide.  These *will* be
used, after the initial sort application of Folders before Files.

```go title="Sort Example"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/user"
)

client := user.Client{Config: files_sdk.GlobalConfig}

// users sorted by username
parameters := files_sdk.UserListParams{SortBy: map[string]interface{}{"username":"asc"}}
userIterator, err := client.List(parameters)
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for userIterator.Next() {
  user := userIterator.User()
  fmt.Println(user.Username)
}
err = userIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

### Filtering

Filters apply selection criteria to the underlying query that returns the results. They can be
applied individually or combined with other filters, and the resulting data can be sorted by a
single field.

Each resource supports a unique set of valid filter fields, filter combinations, and combinations of
filters and sort fields.

For the filter type of ```Filter```, the passed in argument value is an initialized resource struct.
Struct fields are initialized with values to use in the filter.

For the other filter types, the passed in argument is a Go ```map[string]interface{}``` map that has
a key of the resource field name to filter on and a passed in value to use in the filter comparison.

#### Filter Types

| Filter | Type | Description |
| --------- | --------- | --------- |
| `Filter` | Exact | Find resources that have an exact field value match to a passed in value. (i.e., FIELD_VALUE = PASS_IN_VALUE). |
| `FilterPrefix` | Pattern | Find resources where the specified field is prefixed by the supplied value. This is applicable to values that are strings. |
| `FilterGt` | Range | Find resources that have a field value that is greater than the passed in value.  (i.e., FIELD_VALUE > PASS_IN_VALUE). |
| `FilterGteq` | Range | Find resources that have a field value that is greater than or equal to the passed in value.  (i.e., FIELD_VALUE >=  PASS_IN_VALUE). |
| `FilterLt` | Range | Find resources that have a field value that is less than the passed in value.  (i.e., FIELD_VALUE < PASS_IN_VALUE). |
| `FilterLteq` | Range | Find resources that have a field value that is less than or equal to the passed in value.  (i.e., FIELD_VALUE \<= PASS_IN_VALUE). |

```go title="Exact Filter Example"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/user"
)

client := user.Client{Config: files_sdk.GlobalConfig}

// non admin users
filter_value := true;
parameters := files_sdk.UserListParams{
    Filter: files_sdk.User{NotSiteAdmin: &filter_value}
}
userIterator, err := client.List(parameters)
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for userIterator.Next() {
  user := userIterator.User()
  fmt.Println(user.Username)
}
err = userIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

```go title="Range Filter Example"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/user"
)

client := user.Client{Config: files_sdk.GlobalConfig};

// users who haven't logged in since 2024-01-01
parameters := files_sdk.UserListParams{
    FilterLt: map[string]interface{}{"last_login_at": "2024-01-01"}
}
userIterator, err := client.List(parameters)
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for userIterator.Next() {
  user := userIterator.User()
  fmt.Println(user.Username)
}
err = userIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

```go title="Pattern Filter Example"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/user"
)

client := user.Client{Config: files_sdk.GlobalConfig};

// users whose usernames start with 'test'
parameters := files_sdk.UserListParams{
    FilterPrefix: map[string]interface{}{"username": "test"}
}
userIterator, err := client.List(parameters)
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for userIterator.Next() {
  user := userIterator.User()
  fmt.Println(user.Username)
}
err = userIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

```go title="Combination Filter with Sort Example"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/user"
)

client := user.Client{Config: files_sdk.GlobalConfig};

// users whose usernames start with 'test' and are not admins
filterValue := true;
parameters := files_sdk.UserListParams{
    FilterPrefix: map[string]interface{}{"username": "test"},
    Filter:       files_sdk.User{NotSiteAdmin: &filterValue},
    SortBy:       map[string]interface{}{"username": "asc"}
}
userIterator, err := client.List(parameters)
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for userIterator.Next() {
  user := userIterator.User()
  fmt.Println(user.Username)
}
err = userIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

## Paths

Working with paths in Files.com involves several important considerations. Understanding how path comparisons are applied helps developers ensure consistency and accuracy across all interactions with the platform.
<div></div>

### Capitalization

Files.com compares paths in a **case-insensitive** manner. This means path segments are treated as equivalent regardless of letter casing.

For example, all of the following resolve to the same internal path:

| Path Variant                          | Interpreted As              |
|---------------------------------------|------------------------------|
| `Documents/Reports/Q1.pdf`            | `documents/reports/q1.pdf`  |
| `documents/reports/q1.PDF`            | `documents/reports/q1.pdf`  |
| `DOCUMENTS/REPORTS/Q1.PDF`            | `documents/reports/q1.pdf`  |

This behavior applies across:
- API requests
- Folder and file lookup operations
- Automations and workflows

See also: [Case Sensitivity Documentation](https://www.files.com/docs/files-and-folders/case-sensitivity/)

### Slashes

All path parameters in Files.com (API, SDKs, CLI, automations, integrations) must **omit leading and trailing slashes**. Paths are always treated as **absolute and slash-delimited**, so only internal `/` separators are used and never at the start or end of the string.

####  Path Slash Examples
| Path                              | Valid? | Notes                         |
|-----------------------------------|--------|-------------------------------|
| `folder/subfolder/file.txt`       |   ✅   | Correct, internal separators only |
| `/folder/subfolder/file.txt`      |   ❌   | Leading slash not allowed     |
| `folder/subfolder/file.txt/`      |   ❌   | Trailing slash not allowed    |
| `//folder//file.txt`              |   ❌   | Duplicate separators not allowed |

<div></div>

### Unicode Normalization

Files.com normalizes all paths using [Unicode NFC (Normalization Form C)](https://www.unicode.org/reports/tr15/#Norm_Forms) before comparison. This ensures consistency across different representations of the same characters.

For example, the following two paths are treated as equivalent after NFC normalization:

| Input                                  | Normalized Form       |
|----------------------------------------|------------------------|
| `uploads/\u0065\u0301.txt`             | `uploads/é.txt`        |
| `docs/Café/Report.txt`                 | `docs/Café/Report.txt` |

- All input must be UTF‑8 encoded.
- Precomposed and decomposed characters are unified.
- This affects search, deduplication, and comparisons across SDKs.

<div></div>

## Foreign Language Support

The Files.com Go SDK supports localized responses by using the `Language` attribute on the `Config` struct.
When configured, this guides the API in selecting a preferred language for applicable response content.

Language support currently applies to select human-facing fields only, such as notification messages
and error descriptions.

If the specified language is not supported or the value is omitted, the API defaults to English.

```shell title="Example Request"
import (
	files_sdk "github.com/Files-com/files-sdk-go/v3"
)

files_sdk.GlobalConfig.Language = "es";
```

## Errors

The Files.com Go SDK will return errors from function/method calls using the standard Go error handling pattern.

The returned errors fall into basic categories:

1.  `error` - errors that originate in the SDK or standard libraries.
2.  `ResponseError` - errors that occur due to the response from the Files.com API.

The `error` type are errors that implement the `type error` interface and the error specifics can be accessed with the `Error()` method.

`ResponseError` also implements the `type error` interface but is a custom error with additional data.

The additional data includes:

- `Type` - the type of error returned by the Files.com API
- `Title` - a description of the error returned by the Files.com API
- `ErrorMessage` - additional error information

```go title="Example Error Handling"
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/session"
)

thisSession, err := session.Create(files_sdk.SessionCreateParams{ Username: "USERNAME", Password: "BADPASSWORD" })
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

### ResponseError Types

ResponseError errors have additional data returned from the Files.com API to help determine the cause of the error.

| Type | Title |
| --------- | --------- |
| `bad-request` | Bad Request |
| `bad-request/agent-upgrade-required` | Agent Upgrade Required |
| `bad-request/attachment-too-large` | Attachment Too Large |
| `bad-request/cannot-download-directory` | Cannot Download Directory |
| `bad-request/cant-move-with-multiple-locations` | Cant Move With Multiple Locations |
| `bad-request/datetime-parse` | Datetime Parse |
| `bad-request/destination-same` | Destination Same |
| `bad-request/does-not-support-sorting` | Does Not Support Sorting |
| `bad-request/folder-must-not-be-a-file` | Folder Must Not Be A File |
| `bad-request/folders-not-allowed` | Folders Not Allowed |
| `bad-request/invalid-body` | Invalid Body |
| `bad-request/invalid-cursor` | Invalid Cursor |
| `bad-request/invalid-cursor-type-for-sort` | Invalid Cursor Type For Sort |
| `bad-request/invalid-etags` | Invalid Etags |
| `bad-request/invalid-filter-alias-combination` | Invalid Filter Alias Combination |
| `bad-request/invalid-filter-field` | Invalid Filter Field |
| `bad-request/invalid-filter-param` | Invalid Filter Param |
| `bad-request/invalid-filter-param-format` | Invalid Filter Param Format |
| `bad-request/invalid-filter-param-value` | Invalid Filter Param Value |
| `bad-request/invalid-input-encoding` | Invalid Input Encoding |
| `bad-request/invalid-interface` | Invalid Interface |
| `bad-request/invalid-oauth-provider` | Invalid Oauth Provider |
| `bad-request/invalid-path` | Invalid Path |
| `bad-request/invalid-return-to-url` | Invalid Return To Url |
| `bad-request/invalid-sort-field` | Invalid Sort Field |
| `bad-request/invalid-sort-filter-combination` | Invalid Sort Filter Combination |
| `bad-request/invalid-upload-offset` | Invalid Upload Offset |
| `bad-request/invalid-upload-part-gap` | Invalid Upload Part Gap |
| `bad-request/invalid-upload-part-size` | Invalid Upload Part Size |
| `bad-request/method-not-allowed` | Method Not Allowed |
| `bad-request/multiple-sort-params-not-allowed` | Multiple Sort Params Not Allowed |
| `bad-request/no-valid-input-params` | No Valid Input Params |
| `bad-request/part-number-too-large` | Part Number Too Large |
| `bad-request/path-cannot-have-trailing-whitespace` | Path Cannot Have Trailing Whitespace |
| `bad-request/reauthentication-needed-fields` | Reauthentication Needed Fields |
| `bad-request/request-params-contain-invalid-character` | Request Params Contain Invalid Character |
| `bad-request/request-params-invalid` | Request Params Invalid |
| `bad-request/request-params-required` | Request Params Required |
| `bad-request/search-all-on-child-path` | Search All On Child Path |
| `bad-request/unrecognized-sort-index` | Unrecognized Sort Index |
| `bad-request/unsupported-currency` | Unsupported Currency |
| `bad-request/unsupported-http-response-format` | Unsupported Http Response Format |
| `bad-request/unsupported-media-type` | Unsupported Media Type |
| `bad-request/user-id-invalid` | User Id Invalid |
| `bad-request/user-id-on-user-endpoint` | User Id On User Endpoint |
| `bad-request/user-required` | User Required |
| `not-authenticated/additional-authentication-required` | Additional Authentication Required |
| `not-authenticated/api-key-sessions-not-supported` | Api Key Sessions Not Supported |
| `not-authenticated/authentication-required` | Authentication Required |
| `not-authenticated/bundle-registration-code-failed` | Bundle Registration Code Failed |
| `not-authenticated/files-agent-token-failed` | Files Agent Token Failed |
| `not-authenticated/inbox-registration-code-failed` | Inbox Registration Code Failed |
| `not-authenticated/invalid-credentials` | Invalid Credentials |
| `not-authenticated/invalid-oauth` | Invalid Oauth |
| `not-authenticated/invalid-or-expired-code` | Invalid Or Expired Code |
| `not-authenticated/invalid-session` | Invalid Session |
| `not-authenticated/invalid-username-or-password` | Invalid Username Or Password |
| `not-authenticated/locked-out` | Locked Out |
| `not-authenticated/lockout-region-mismatch` | Lockout Region Mismatch |
| `not-authenticated/one-time-password-incorrect` | One Time Password Incorrect |
| `not-authenticated/two-factor-authentication-error` | Two Factor Authentication Error |
| `not-authenticated/two-factor-authentication-setup-expired` | Two Factor Authentication Setup Expired |
| `not-authorized/api-key-is-disabled` | Api Key Is Disabled |
| `not-authorized/api-key-is-path-restricted` | Api Key Is Path Restricted |
| `not-authorized/api-key-only-for-desktop-app` | Api Key Only For Desktop App |
| `not-authorized/api-key-only-for-mobile-app` | Api Key Only For Mobile App |
| `not-authorized/api-key-only-for-office-integration` | Api Key Only For Office Integration |
| `not-authorized/billing-or-site-admin-permission-required` | Billing Or Site Admin Permission Required |
| `not-authorized/billing-permission-required` | Billing Permission Required |
| `not-authorized/bundle-maximum-uses-reached` | Bundle Maximum Uses Reached |
| `not-authorized/bundle-permission-required` | Bundle Permission Required |
| `not-authorized/cannot-login-while-using-key` | Cannot Login While Using Key |
| `not-authorized/cant-act-for-other-user` | Cant Act For Other User |
| `not-authorized/contact-admin-for-password-change-help` | Contact Admin For Password Change Help |
| `not-authorized/files-agent-failed-authorization` | Files Agent Failed Authorization |
| `not-authorized/folder-admin-or-billing-permission-required` | Folder Admin Or Billing Permission Required |
| `not-authorized/folder-admin-permission-required` | Folder Admin Permission Required |
| `not-authorized/full-permission-required` | Full Permission Required |
| `not-authorized/history-permission-required` | History Permission Required |
| `not-authorized/insufficient-permission-for-params` | Insufficient Permission For Params |
| `not-authorized/insufficient-permission-for-site` | Insufficient Permission For Site |
| `not-authorized/mover-access-denied` | Mover Access Denied |
| `not-authorized/mover-package-required` | Mover Package Required |
| `not-authorized/must-authenticate-with-api-key` | Must Authenticate With Api Key |
| `not-authorized/need-admin-permission-for-inbox` | Need Admin Permission For Inbox |
| `not-authorized/non-admins-must-query-by-folder-or-path` | Non Admins Must Query By Folder Or Path |
| `not-authorized/not-allowed-to-create-bundle` | Not Allowed To Create Bundle |
| `not-authorized/not-enqueuable-sync` | Not Enqueuable Sync |
| `not-authorized/password-change-not-required` | Password Change Not Required |
| `not-authorized/password-change-required` | Password Change Required |
| `not-authorized/payment-method-error` | Payment Method Error |
| `not-authorized/read-only-session` | Read Only Session |
| `not-authorized/read-permission-required` | Read Permission Required |
| `not-authorized/reauthentication-failed` | Reauthentication Failed |
| `not-authorized/reauthentication-failed-final` | Reauthentication Failed Final |
| `not-authorized/reauthentication-needed-action` | Reauthentication Needed Action |
| `not-authorized/recaptcha-failed` | Recaptcha Failed |
| `not-authorized/self-managed-required` | Self Managed Required |
| `not-authorized/site-admin-required` | Site Admin Required |
| `not-authorized/site-files-are-immutable` | Site Files Are Immutable |
| `not-authorized/two-factor-authentication-required` | Two Factor Authentication Required |
| `not-authorized/user-id-without-site-admin` | User Id Without Site Admin |
| `not-authorized/write-and-bundle-permission-required` | Write And Bundle Permission Required |
| `not-authorized/write-permission-required` | Write Permission Required |
| `not-found` | Not Found |
| `not-found/api-key-not-found` | Api Key Not Found |
| `not-found/bundle-path-not-found` | Bundle Path Not Found |
| `not-found/bundle-registration-not-found` | Bundle Registration Not Found |
| `not-found/code-not-found` | Code Not Found |
| `not-found/file-not-found` | File Not Found |
| `not-found/file-upload-not-found` | File Upload Not Found |
| `not-found/group-not-found` | Group Not Found |
| `not-found/inbox-not-found` | Inbox Not Found |
| `not-found/nested-not-found` | Nested Not Found |
| `not-found/plan-not-found` | Plan Not Found |
| `not-found/site-not-found` | Site Not Found |
| `not-found/user-not-found` | User Not Found |
| `processing-failure` | Processing Failure |
| `processing-failure/agent-unavailable` | Agent Unavailable |
| `processing-failure/already-completed` | Already Completed |
| `processing-failure/automation-cannot-be-run-manually` | Automation Cannot Be Run Manually |
| `processing-failure/behavior-not-allowed-on-remote-server` | Behavior Not Allowed On Remote Server |
| `processing-failure/bundle-only-allows-previews` | Bundle Only Allows Previews |
| `processing-failure/bundle-operation-requires-subfolder` | Bundle Operation Requires Subfolder |
| `processing-failure/could-not-create-parent` | Could Not Create Parent |
| `processing-failure/destination-exists` | Destination Exists |
| `processing-failure/destination-folder-limited` | Destination Folder Limited |
| `processing-failure/destination-parent-conflict` | Destination Parent Conflict |
| `processing-failure/destination-parent-does-not-exist` | Destination Parent Does Not Exist |
| `processing-failure/exceeded-runtime-limit` | Exceeded Runtime Limit |
| `processing-failure/expired-private-key` | Expired Private Key |
| `processing-failure/expired-public-key` | Expired Public Key |
| `processing-failure/export-failure` | Export Failure |
| `processing-failure/export-not-ready` | Export Not Ready |
| `processing-failure/failed-to-change-password` | Failed To Change Password |
| `processing-failure/file-locked` | File Locked |
| `processing-failure/file-not-uploaded` | File Not Uploaded |
| `processing-failure/file-pending-processing` | File Pending Processing |
| `processing-failure/file-processing-error` | File Processing Error |
| `processing-failure/file-too-big-to-decrypt` | File Too Big To Decrypt |
| `processing-failure/file-too-big-to-encrypt` | File Too Big To Encrypt |
| `processing-failure/file-uploaded-to-wrong-region` | File Uploaded To Wrong Region |
| `processing-failure/filename-too-long` | Filename Too Long |
| `processing-failure/folder-locked` | Folder Locked |
| `processing-failure/folder-not-empty` | Folder Not Empty |
| `processing-failure/history-unavailable` | History Unavailable |
| `processing-failure/invalid-bundle-code` | Invalid Bundle Code |
| `processing-failure/invalid-file-type` | Invalid File Type |
| `processing-failure/invalid-filename` | Invalid Filename |
| `processing-failure/invalid-priority-color` | Invalid Priority Color |
| `processing-failure/invalid-range` | Invalid Range |
| `processing-failure/invalid-site` | Invalid Site |
| `processing-failure/metadata-not-supported-on-remotes` | Metadata Not Supported On Remotes |
| `processing-failure/model-save-error` | Model Save Error |
| `processing-failure/multiple-processing-errors` | Multiple Processing Errors |
| `processing-failure/path-too-long` | Path Too Long |
| `processing-failure/recipient-already-shared` | Recipient Already Shared |
| `processing-failure/remote-server-error` | Remote Server Error |
| `processing-failure/resource-belongs-to-parent-site` | Resource Belongs To Parent Site |
| `processing-failure/resource-locked` | Resource Locked |
| `processing-failure/subfolder-locked` | Subfolder Locked |
| `processing-failure/sync-in-progress` | Sync In Progress |
| `processing-failure/two-factor-authentication-code-already-sent` | Two Factor Authentication Code Already Sent |
| `processing-failure/two-factor-authentication-country-blacklisted` | Two Factor Authentication Country Blacklisted |
| `processing-failure/two-factor-authentication-general-error` | Two Factor Authentication General Error |
| `processing-failure/two-factor-authentication-method-unsupported-error` | Two Factor Authentication Method Unsupported Error |
| `processing-failure/two-factor-authentication-unsubscribed-recipient` | Two Factor Authentication Unsubscribed Recipient |
| `processing-failure/updates-not-allowed-for-remotes` | Updates Not Allowed For Remotes |
| `rate-limited/duplicate-share-recipient` | Duplicate Share Recipient |
| `rate-limited/reauthentication-rate-limited` | Reauthentication Rate Limited |
| `rate-limited/too-many-concurrent-logins` | Too Many Concurrent Logins |
| `rate-limited/too-many-concurrent-requests` | Too Many Concurrent Requests |
| `rate-limited/too-many-login-attempts` | Too Many Login Attempts |
| `rate-limited/too-many-requests` | Too Many Requests |
| `rate-limited/too-many-shares` | Too Many Shares |
| `service-unavailable/automations-unavailable` | Automations Unavailable |
| `service-unavailable/migration-in-progress` | Migration In Progress |
| `service-unavailable/site-disabled` | Site Disabled |
| `service-unavailable/uploads-unavailable` | Uploads Unavailable |
| `site-configuration/account-already-exists` | Account Already Exists |
| `site-configuration/account-overdue` | Account Overdue |
| `site-configuration/no-account-for-site` | No Account For Site |
| `site-configuration/site-was-removed` | Site Was Removed |
| `site-configuration/trial-expired` | Trial Expired |
| `site-configuration/trial-locked` | Trial Locked |
| `site-configuration/user-requests-enabled-required` | User Requests Enabled Required |

## Pagination

Certain API operations return lists of objects. When the number of objects in the list is large,
the API will paginate the results.

The Files.com Go SDK automatically paginates through lists of objects by default.

```go title="Example Request" hasDataFormatSelector
import (
    "fmt"
    "errors"

    files_sdk "github.com/Files-com/files-sdk-go/v3"
    "github.com/Files-com/files-sdk-go/v3/folder"
)

fileIterator, err := folder.ListFor(files_sdk.FolderListForParams{Path: "path"})
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}

for fileIterator.Next() {
    file := fileIterator.file()
}
err = fileIterator.Err()
if err != nil {
    var respErr files_sdk.ResponseError
    if errors.As(err, &respErr) {
        fmt.Println("Response Error Occurred (" + respErr.Type + "): " + respErr.ErrorMessage)
    } else {
        fmt.Printf("Unexpected Error: %s\n", err.Error())
    }
}
```

## Mock Server

Files.com publishes a Files.com API server, which is useful for testing your use of the Files.com
SDKs and other direct integrations against the Files.com API in an integration test environment.

It is a Ruby app that operates as a minimal server for the purpose of testing basic network
operations and JSON encoding for your SDK or API client. It does not maintain state and it does not
deeply inspect your submissions for correctness.

Eventually we will add more features intended for integration testing, such as the ability to
intentionally provoke errors.

Download the server as a Docker image via [Docker Hub](https://hub.docker.com/r/filescom/files-mock-server).

The Source Code is also available on [GitHub](https://github.com/Files-com/files-mock-server).

A README is available on the GitHub link.
