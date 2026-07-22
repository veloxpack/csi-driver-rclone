package files_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Files-com/files-sdk-go/v3/lib"
)

type ResponseError struct {
	Type           string `json:"type,omitempty"`
	Title          string `json:"title,omitempty"`
	ErrorMessage   string `json:"error,omitempty"`
	HttpCode       int    `json:"http-code,omitempty"`
	Data           `json:"-"`
	RawData        map[string]interface{} `json:"data,omitempty"`
	Errors         []ResponseError        `json:"errors,omitempty"`
	Instance       string                 `json:"instance,omitempty"`
	ModelErrors    map[string]interface{} `json:"model_errors,omitempty"`
	ModelErrorKeys map[string]interface{} `json:"model_error_keys,omitempty"`
}

type ResponseErrorType string
type ResponseErrorGroup string

func (e ResponseErrorType) Error() string {
	return string(e)
}

func (e ResponseErrorGroup) Error() string {
	return string(e)
}

const (
	ErrAgentUpgradeRequired                          ResponseErrorType = "bad-request/agent-upgrade-required"
	ErrAttachmentTooLarge                            ResponseErrorType = "bad-request/attachment-too-large"
	ErrCannotDownloadDirectory                       ResponseErrorType = "bad-request/cannot-download-directory"
	ErrCantMoveWithMultipleLocations                 ResponseErrorType = "bad-request/cant-move-with-multiple-locations"
	ErrDatetimeParse                                 ResponseErrorType = "bad-request/datetime-parse"
	ErrDestinationSame                               ResponseErrorType = "bad-request/destination-same"
	ErrDestinationSiteMismatch                       ResponseErrorType = "bad-request/destination-site-mismatch"
	ErrDoesNotSupportSorting                         ResponseErrorType = "bad-request/does-not-support-sorting"
	ErrFolderMustNotBeAFile                          ResponseErrorType = "bad-request/folder-must-not-be-a-file"
	ErrFoldersNotAllowed                             ResponseErrorType = "bad-request/folders-not-allowed"
	ErrInternalGeneralError                          ResponseErrorType = "bad-request/internal-general-error"
	ErrInvalidBody                                   ResponseErrorType = "bad-request/invalid-body"
	ErrInvalidCursor                                 ResponseErrorType = "bad-request/invalid-cursor"
	ErrInvalidCursorTypeForSort                      ResponseErrorType = "bad-request/invalid-cursor-type-for-sort"
	ErrInvalidEtags                                  ResponseErrorType = "bad-request/invalid-etags"
	ErrInvalidFilterAliasCombination                 ResponseErrorType = "bad-request/invalid-filter-alias-combination"
	ErrInvalidFilterField                            ResponseErrorType = "bad-request/invalid-filter-field"
	ErrInvalidFilterParam                            ResponseErrorType = "bad-request/invalid-filter-param"
	ErrInvalidFilterParamFormat                      ResponseErrorType = "bad-request/invalid-filter-param-format"
	ErrInvalidFilterParamValue                       ResponseErrorType = "bad-request/invalid-filter-param-value"
	ErrInvalidInputEncoding                          ResponseErrorType = "bad-request/invalid-input-encoding"
	ErrInvalidInterface                              ResponseErrorType = "bad-request/invalid-interface"
	ErrInvalidOauthProvider                          ResponseErrorType = "bad-request/invalid-oauth-provider"
	ErrInvalidPath                                   ResponseErrorType = "bad-request/invalid-path"
	ErrInvalidReturnToUrl                            ResponseErrorType = "bad-request/invalid-return-to-url"
	ErrInvalidSortField                              ResponseErrorType = "bad-request/invalid-sort-field"
	ErrInvalidSortFilterCombination                  ResponseErrorType = "bad-request/invalid-sort-filter-combination"
	ErrInvalidUploadOffset                           ResponseErrorType = "bad-request/invalid-upload-offset"
	ErrInvalidUploadPartGap                          ResponseErrorType = "bad-request/invalid-upload-part-gap"
	ErrInvalidUploadPartSize                         ResponseErrorType = "bad-request/invalid-upload-part-size"
	ErrInvalidWorkspaceIdHeader                      ResponseErrorType = "bad-request/invalid-workspace-id-header"
	ErrMethodNotAllowed                              ResponseErrorType = "bad-request/method-not-allowed"
	ErrMultipleSortParamsNotAllowed                  ResponseErrorType = "bad-request/multiple-sort-params-not-allowed"
	ErrNoValidInputParams                            ResponseErrorType = "bad-request/no-valid-input-params"
	ErrPartNumberTooLarge                            ResponseErrorType = "bad-request/part-number-too-large"
	ErrPathCannotHaveTrailingWhitespace              ResponseErrorType = "bad-request/path-cannot-have-trailing-whitespace"
	ErrReauthenticationNeededFields                  ResponseErrorType = "bad-request/reauthentication-needed-fields"
	ErrRequestParamsContainInvalidCharacter          ResponseErrorType = "bad-request/request-params-contain-invalid-character"
	ErrRequestParamsInvalid                          ResponseErrorType = "bad-request/request-params-invalid"
	ErrRequestParamsRequired                         ResponseErrorType = "bad-request/request-params-required"
	ErrSearchAllOnChildPath                          ResponseErrorType = "bad-request/search-all-on-child-path"
	ErrUnrecognizedSortIndex                         ResponseErrorType = "bad-request/unrecognized-sort-index"
	ErrUnsupportedCurrency                           ResponseErrorType = "bad-request/unsupported-currency"
	ErrUnsupportedHttpResponseFormat                 ResponseErrorType = "bad-request/unsupported-http-response-format"
	ErrUnsupportedMediaType                          ResponseErrorType = "bad-request/unsupported-media-type"
	ErrUserIdInvalid                                 ResponseErrorType = "bad-request/user-id-invalid"
	ErrUserIdOnUserEndpoint                          ResponseErrorType = "bad-request/user-id-on-user-endpoint"
	ErrUserRequired                                  ResponseErrorType = "bad-request/user-required"
	ErrAdditionalAuthenticationRequired              ResponseErrorType = "not-authenticated/additional-authentication-required"
	ErrApiKeySessionsNotSupported                    ResponseErrorType = "not-authenticated/api-key-sessions-not-supported"
	ErrAuthenticationRequired                        ResponseErrorType = "not-authenticated/authentication-required"
	ErrBundleRegistrationCodeFailed                  ResponseErrorType = "not-authenticated/bundle-registration-code-failed"
	ErrFilesAgentTokenFailed                         ResponseErrorType = "not-authenticated/files-agent-token-failed"
	ErrInboxRegistrationCodeFailed                   ResponseErrorType = "not-authenticated/inbox-registration-code-failed"
	ErrInvalidCredentials                            ResponseErrorType = "not-authenticated/invalid-credentials"
	ErrInvalidOauth                                  ResponseErrorType = "not-authenticated/invalid-oauth"
	ErrInvalidOrExpiredCode                          ResponseErrorType = "not-authenticated/invalid-or-expired-code"
	ErrInvalidSession                                ResponseErrorType = "not-authenticated/invalid-session"
	ErrInvalidUsernameOrPassword                     ResponseErrorType = "not-authenticated/invalid-username-or-password"
	ErrLockedOut                                     ResponseErrorType = "not-authenticated/locked-out"
	ErrLockoutRegionMismatch                         ResponseErrorType = "not-authenticated/lockout-region-mismatch"
	ErrOneTimePasswordIncorrect                      ResponseErrorType = "not-authenticated/one-time-password-incorrect"
	ErrTwoFactorAuthenticationError                  ResponseErrorType = "not-authenticated/two-factor-authentication-error"
	ErrTwoFactorAuthenticationSetupExpired           ResponseErrorType = "not-authenticated/two-factor-authentication-setup-expired"
	ErrApiKeyIsDisabled                              ResponseErrorType = "not-authorized/api-key-is-disabled"
	ErrApiKeyIsPathRestricted                        ResponseErrorType = "not-authorized/api-key-is-path-restricted"
	ErrApiKeyOnlyForDesktopApp                       ResponseErrorType = "not-authorized/api-key-only-for-desktop-app"
	ErrApiKeyOnlyForMobileApp                        ResponseErrorType = "not-authorized/api-key-only-for-mobile-app"
	ErrApiKeyOnlyForOfficeIntegration                ResponseErrorType = "not-authorized/api-key-only-for-office-integration"
	ErrBillingOrSiteAdminPermissionRequired          ResponseErrorType = "not-authorized/billing-or-site-admin-permission-required"
	ErrBillingPermissionRequired                     ResponseErrorType = "not-authorized/billing-permission-required"
	ErrBundleMaximumUsesReached                      ResponseErrorType = "not-authorized/bundle-maximum-uses-reached"
	ErrBundlePermissionRequired                      ResponseErrorType = "not-authorized/bundle-permission-required"
	ErrCannotLoginWhileUsingKey                      ResponseErrorType = "not-authorized/cannot-login-while-using-key"
	ErrCantActForOtherUser                           ResponseErrorType = "not-authorized/cant-act-for-other-user"
	ErrContactAdminForPasswordChangeHelp             ResponseErrorType = "not-authorized/contact-admin-for-password-change-help"
	ErrFilesAgentFailedAuthorization                 ResponseErrorType = "not-authorized/files-agent-failed-authorization"
	ErrFolderAdminOrBillingPermissionRequired        ResponseErrorType = "not-authorized/folder-admin-or-billing-permission-required"
	ErrFolderAdminPermissionRequired                 ResponseErrorType = "not-authorized/folder-admin-permission-required"
	ErrFullPermissionRequired                        ResponseErrorType = "not-authorized/full-permission-required"
	ErrHistoryPermissionRequired                     ResponseErrorType = "not-authorized/history-permission-required"
	ErrInsufficientPermissionForParams               ResponseErrorType = "not-authorized/insufficient-permission-for-params"
	ErrInsufficientPermissionForSite                 ResponseErrorType = "not-authorized/insufficient-permission-for-site"
	ErrMoverAccessDenied                             ResponseErrorType = "not-authorized/mover-access-denied"
	ErrMoverPackageRequired                          ResponseErrorType = "not-authorized/mover-package-required"
	ErrMustAuthenticateWithApiKey                    ResponseErrorType = "not-authorized/must-authenticate-with-api-key"
	ErrNeedAdminPermissionForInbox                   ResponseErrorType = "not-authorized/need-admin-permission-for-inbox"
	ErrNonAdminsMustQueryByFolderOrPath              ResponseErrorType = "not-authorized/non-admins-must-query-by-folder-or-path"
	ErrNotAllowedToCreateBundle                      ResponseErrorType = "not-authorized/not-allowed-to-create-bundle"
	ErrNotEnqueuableSync                             ResponseErrorType = "not-authorized/not-enqueuable-sync"
	ErrPasswordChangeNotRequired                     ResponseErrorType = "not-authorized/password-change-not-required"
	ErrPasswordChangeRequired                        ResponseErrorType = "not-authorized/password-change-required"
	ErrPaymentMethodError                            ResponseErrorType = "not-authorized/payment-method-error"
	ErrPreviewOnlyPermissionCannotDownload           ResponseErrorType = "not-authorized/preview-only-permission-cannot-download"
	ErrReadOnlySession                               ResponseErrorType = "not-authorized/read-only-session"
	ErrReadPermissionRequired                        ResponseErrorType = "not-authorized/read-permission-required"
	ErrReauthenticationFailed                        ResponseErrorType = "not-authorized/reauthentication-failed"
	ErrReauthenticationFailedFinal                   ResponseErrorType = "not-authorized/reauthentication-failed-final"
	ErrReauthenticationNeededAction                  ResponseErrorType = "not-authorized/reauthentication-needed-action"
	ErrRecaptchaFailed                               ResponseErrorType = "not-authorized/recaptcha-failed"
	ErrSelfManagedRequired                           ResponseErrorType = "not-authorized/self-managed-required"
	ErrSiteAdminRequired                             ResponseErrorType = "not-authorized/site-admin-required"
	ErrSiteFilesAreImmutable                         ResponseErrorType = "not-authorized/site-files-are-immutable"
	ErrTwoFactorAuthenticationRequired               ResponseErrorType = "not-authorized/two-factor-authentication-required"
	ErrUserIdWithoutSiteAdmin                        ResponseErrorType = "not-authorized/user-id-without-site-admin"
	ErrWriteAndBundlePermissionRequired              ResponseErrorType = "not-authorized/write-and-bundle-permission-required"
	ErrWritePermissionRequired                       ResponseErrorType = "not-authorized/write-permission-required"
	ErrApiKeyNotFound                                ResponseErrorType = "not-found/api-key-not-found"
	ErrBundlePathNotFound                            ResponseErrorType = "not-found/bundle-path-not-found"
	ErrBundleRegistrationNotFound                    ResponseErrorType = "not-found/bundle-registration-not-found"
	ErrCodeNotFound                                  ResponseErrorType = "not-found/code-not-found"
	ErrFileNotFound                                  ResponseErrorType = "not-found/file-not-found"
	ErrFileUploadNotFound                            ResponseErrorType = "not-found/file-upload-not-found"
	ErrGroupNotFound                                 ResponseErrorType = "not-found/group-not-found"
	ErrInboxNotFound                                 ResponseErrorType = "not-found/inbox-not-found"
	ErrNestedNotFound                                ResponseErrorType = "not-found/nested-not-found"
	ErrPlanNotFound                                  ResponseErrorType = "not-found/plan-not-found"
	ErrSiteNotFound                                  ResponseErrorType = "not-found/site-not-found"
	ErrUserNotFound                                  ResponseErrorType = "not-found/user-not-found"
	ErrAgentUnavailable                              ResponseErrorType = "processing-failure/agent-unavailable"
	ErrAlreadyCompleted                              ResponseErrorType = "processing-failure/already-completed"
	ErrAutomationCannotBeRunManually                 ResponseErrorType = "processing-failure/automation-cannot-be-run-manually"
	ErrBehaviorNotAllowedOnRemoteServer              ResponseErrorType = "processing-failure/behavior-not-allowed-on-remote-server"
	ErrBufferedUploadDisabledForThisDestination      ResponseErrorType = "processing-failure/buffered-upload-disabled-for-this-destination"
	ErrBundleOnlyAllowsPreviews                      ResponseErrorType = "processing-failure/bundle-only-allows-previews"
	ErrBundleOperationRequiresSubfolder              ResponseErrorType = "processing-failure/bundle-operation-requires-subfolder"
	ErrConfigurationLockedPath                       ResponseErrorType = "processing-failure/configuration-locked-path"
	ErrCouldNotCreateParent                          ResponseErrorType = "processing-failure/could-not-create-parent"
	ErrDestinationExists                             ResponseErrorType = "processing-failure/destination-exists"
	ErrDestinationFolderLimited                      ResponseErrorType = "processing-failure/destination-folder-limited"
	ErrDestinationParentConflict                     ResponseErrorType = "processing-failure/destination-parent-conflict"
	ErrDestinationParentDoesNotExist                 ResponseErrorType = "processing-failure/destination-parent-does-not-exist"
	ErrExceededRuntimeLimit                          ResponseErrorType = "processing-failure/exceeded-runtime-limit"
	ErrExpectationAlreadyHasOpenWindow               ResponseErrorType = "processing-failure/expectation-already-has-open-window"
	ErrExpectationNotManualTrigger                   ResponseErrorType = "processing-failure/expectation-not-manual-trigger"
	ErrExpiredPrivateKey                             ResponseErrorType = "processing-failure/expired-private-key"
	ErrExpiredPublicKey                              ResponseErrorType = "processing-failure/expired-public-key"
	ErrExportFailure                                 ResponseErrorType = "processing-failure/export-failure"
	ErrExportNotReady                                ResponseErrorType = "processing-failure/export-not-ready"
	ErrFailedToChangePassword                        ResponseErrorType = "processing-failure/failed-to-change-password"
	ErrFileLocked                                    ResponseErrorType = "processing-failure/file-locked"
	ErrFileNotUploaded                               ResponseErrorType = "processing-failure/file-not-uploaded"
	ErrFilePendingProcessing                         ResponseErrorType = "processing-failure/file-pending-processing"
	ErrFileProcessingError                           ResponseErrorType = "processing-failure/file-processing-error"
	ErrFileTooBigToDecrypt                           ResponseErrorType = "processing-failure/file-too-big-to-decrypt"
	ErrFileTooBigToEncrypt                           ResponseErrorType = "processing-failure/file-too-big-to-encrypt"
	ErrFileUploadedToWrongRegion                     ResponseErrorType = "processing-failure/file-uploaded-to-wrong-region"
	ErrFilenameTooLong                               ResponseErrorType = "processing-failure/filename-too-long"
	ErrFolderLocked                                  ResponseErrorType = "processing-failure/folder-locked"
	ErrFolderNotEmpty                                ResponseErrorType = "processing-failure/folder-not-empty"
	ErrHistoryUnavailable                            ResponseErrorType = "processing-failure/history-unavailable"
	ErrInvalidBundleCode                             ResponseErrorType = "processing-failure/invalid-bundle-code"
	ErrInvalidFileType                               ResponseErrorType = "processing-failure/invalid-file-type"
	ErrInvalidFilename                               ResponseErrorType = "processing-failure/invalid-filename"
	ErrInvalidPriorityColor                          ResponseErrorType = "processing-failure/invalid-priority-color"
	ErrInvalidRange                                  ResponseErrorType = "processing-failure/invalid-range"
	ErrInvalidSite                                   ResponseErrorType = "processing-failure/invalid-site"
	ErrInvalidZipFile                                ResponseErrorType = "processing-failure/invalid-zip-file"
	ErrMetadataNotSupportedOnRemotes                 ResponseErrorType = "processing-failure/metadata-not-supported-on-remotes"
	ErrModelSaveError                                ResponseErrorType = "processing-failure/model-save-error"
	ErrMultipleProcessingErrors                      ResponseErrorType = "processing-failure/multiple-processing-errors"
	ErrPathTooLong                                   ResponseErrorType = "processing-failure/path-too-long"
	ErrRecipientAlreadyShared                        ResponseErrorType = "processing-failure/recipient-already-shared"
	ErrRemoteServerError                             ResponseErrorType = "processing-failure/remote-server-error"
	ErrResourceBelongsToParentSite                   ResponseErrorType = "processing-failure/resource-belongs-to-parent-site"
	ErrResourceLocked                                ResponseErrorType = "processing-failure/resource-locked"
	ErrSubfolderLocked                               ResponseErrorType = "processing-failure/subfolder-locked"
	ErrSyncInProgress                                ResponseErrorType = "processing-failure/sync-in-progress"
	ErrTwoFactorAuthenticationCodeAlreadySent        ResponseErrorType = "processing-failure/two-factor-authentication-code-already-sent"
	ErrTwoFactorAuthenticationCountryBlacklisted     ResponseErrorType = "processing-failure/two-factor-authentication-country-blacklisted"
	ErrTwoFactorAuthenticationGeneralError           ResponseErrorType = "processing-failure/two-factor-authentication-general-error"
	ErrTwoFactorAuthenticationMethodUnsupportedError ResponseErrorType = "processing-failure/two-factor-authentication-method-unsupported-error"
	ErrTwoFactorAuthenticationUnsubscribedRecipient  ResponseErrorType = "processing-failure/two-factor-authentication-unsubscribed-recipient"
	ErrUpdatesNotAllowedForRemotes                   ResponseErrorType = "processing-failure/updates-not-allowed-for-remotes"
	ErrDuplicateShareRecipient                       ResponseErrorType = "rate-limited/duplicate-share-recipient"
	ErrReauthenticationRateLimited                   ResponseErrorType = "rate-limited/reauthentication-rate-limited"
	ErrTooManyConcurrentLogins                       ResponseErrorType = "rate-limited/too-many-concurrent-logins"
	ErrTooManyConcurrentRequests                     ResponseErrorType = "rate-limited/too-many-concurrent-requests"
	ErrTooManyLoginAttempts                          ResponseErrorType = "rate-limited/too-many-login-attempts"
	ErrTooManyRequests                               ResponseErrorType = "rate-limited/too-many-requests"
	ErrTooManyShares                                 ResponseErrorType = "rate-limited/too-many-shares"
	ErrAutomationsUnavailable                        ResponseErrorType = "service-unavailable/automations-unavailable"
	ErrMigrationInProgress                           ResponseErrorType = "service-unavailable/migration-in-progress"
	ErrSiteDisabled                                  ResponseErrorType = "service-unavailable/site-disabled"
	ErrUploadsUnavailable                            ResponseErrorType = "service-unavailable/uploads-unavailable"
	ErrAccountAlreadyExists                          ResponseErrorType = "site-configuration/account-already-exists"
	ErrAccountOverdue                                ResponseErrorType = "site-configuration/account-overdue"
	ErrNoAccountForSite                              ResponseErrorType = "site-configuration/no-account-for-site"
	ErrSiteWasRemoved                                ResponseErrorType = "site-configuration/site-was-removed"
	ErrTrialExpired                                  ResponseErrorType = "site-configuration/trial-expired"
	ErrTrialLocked                                   ResponseErrorType = "site-configuration/trial-locked"
	ErrUserRequestsEnabledRequired                   ResponseErrorType = "site-configuration/user-requests-enabled-required"
	ErrDownloadRequestExpired                        ResponseErrorType = "download_request_expired"
	ErrUploadRequestExpired                          ResponseErrorType = "upload_request_expired"
)

const (
	ErrBadRequest         ResponseErrorGroup = "bad-request"
	ErrNotAuthenticated   ResponseErrorGroup = "not-authenticated"
	ErrNotAuthorized      ResponseErrorGroup = "not-authorized"
	ErrNotFound           ResponseErrorGroup = "not-found"
	ErrProcessingFailure  ResponseErrorGroup = "processing-failure"
	ErrRateLimited        ResponseErrorGroup = "rate-limited"
	ErrServiceUnavailable ResponseErrorGroup = "service-unavailable"
	ErrSiteConfiguration  ResponseErrorGroup = "site-configuration"
)

// DestinationExists is deprecated; use ErrDestinationExists.
const DestinationExists = string(ErrDestinationExists)

// DownloadRequestExpired is deprecated; use ErrDownloadRequestExpired.
const DownloadRequestExpired = string(ErrDownloadRequestExpired)

// UploadRequestExpired is deprecated; use ErrUploadRequestExpired.
const UploadRequestExpired = string(ErrUploadRequestExpired)

func ResponseErrorTypeOf(err error) (ResponseErrorType, bool) {
	var responseError ResponseError
	if ok := errors.As(err, &responseError); ok && responseError.Type != "" {
		return ResponseErrorType(responseError.Type), true
	}

	var responseErrorPtr *ResponseError
	if ok := errors.As(err, &responseErrorPtr); ok && responseErrorPtr != nil && responseErrorPtr.Type != "" {
		return ResponseErrorType(responseErrorPtr.Type), true
	}

	return "", false
}

func IsErrorType(err error, responseType ResponseErrorType) bool {
	errType, ok := ResponseErrorTypeOf(err)
	return ok && errType == responseType
}

func IsAnyErrorType(err error, responseTypes ...ResponseErrorType) bool {
	errType, ok := ResponseErrorTypeOf(err)
	if !ok {
		return false
	}

	for _, responseType := range responseTypes {
		if errType == responseType {
			return true
		}
	}
	return false
}

func IsErrorGroup(err error, responseGroup ResponseErrorGroup) bool {
	errType, ok := ResponseErrorTypeOf(err)
	if !ok {
		return false
	}
	return responseErrorGroupForType(errType) == responseGroup
}

func IsExpired(err error) bool {
	return IsAnyErrorType(err, ErrDownloadRequestExpired, ErrUploadRequestExpired)
}

func IsExist(err error) bool {
	return errors.Is(err, ErrDestinationExists)
}

func IsNotExist(err error) bool {
	return IsNotFound(err)
}

func IsNotAuthenticated(err error) bool {
	return IsAuthenticationError(err)
}

func IsBadRequest(err error) bool {
	return IsErrorGroup(err, ErrBadRequest)
}

func IsAuthenticationError(err error) bool {
	return IsErrorGroup(err, ErrNotAuthenticated)
}

func IsAuthorizationError(err error) bool {
	return IsErrorGroup(err, ErrNotAuthorized)
}

func IsNotFound(err error) bool {
	return IsErrorGroup(err, ErrNotFound)
}

func IsProcessingFailure(err error) bool {
	return IsErrorGroup(err, ErrProcessingFailure)
}

func IsRateLimited(err error) bool {
	return IsErrorGroup(err, ErrRateLimited)
}

func IsServiceUnavailable(err error) bool {
	return IsErrorGroup(err, ErrServiceUnavailable)
}

func IsSiteConfiguration(err error) bool {
	return IsErrorGroup(err, ErrSiteConfiguration)
}

type SignRequest struct {
	Version   string `json:"version"`
	KeyHandle string `json:"keyHandle"`
}

type U2fSignRequests struct {
	AppId       string      `json:"app_id"`
	Challenge   string      `json:"challenge"`
	SignRequest SignRequest `json:"sign_request"`
}

type Data struct {
	U2fSIgnRequests               []U2fSignRequests `json:"u2f_sign_requests,omitempty"`
	PartialSessionId              string            `json:"partial_session_id,omitempty"`
	TwoFactorAuthenticationMethod []string          `json:"two_factor_authentication_methods,omitempty"`
	Host                          string            `json:"host,omitempty"`
	// Download Request Status
	BytesTransferred int64      `json:"bytes_transferred,omitempty"`
	Status           string     `json:"status,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	TouchedAt        *time.Time `json:"touched_at,omitempty"`
}

func (e ResponseError) Error() string {
	if e.Title == "" {
		return e.ErrorMessage
	}
	return fmt.Sprintf("%v - `%v`", e.Title, e.ErrorMessage)
}

func (e ResponseError) IsNil() bool {
	return e.ErrorMessage == ""
}

func (e ResponseError) Is(err error) bool {
	switch target := err.(type) {
	case ResponseErrorType:
		return ResponseErrorType(e.Type) == target
	case ResponseErrorGroup:
		return responseErrorGroupForType(ResponseErrorType(e.Type)) == target
	case ResponseError:
		return target.Type == "" || e.Type == target.Type
	case *ResponseError:
		return target != nil && (target.Type == "" || e.Type == target.Type)
	default:
		return false
	}
}

func responseErrorGroupForType(responseType ResponseErrorType) ResponseErrorGroup {
	tokens := strings.SplitN(string(responseType), "/", 2)
	return ResponseErrorGroup(tokens[0])
}

func (e ResponseError) MarshalJSON() ([]byte, error) {
	type re ResponseError
	var v re
	v = re(e)

	rawDataJson, err := json.Marshal(v.Data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawDataJson, &v.RawData); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func (e *ResponseError) UnmarshalJSON(data []byte) error {
	type re ResponseError
	var v re

	if err := json.Unmarshal(data, &v); err != nil {
		var jsonError *json.UnmarshalTypeError
		if ok := errors.As(err, &jsonError); ok && jsonError.Field == "" {
			if jsonError.Value == "string" {
				var str string
				json.Unmarshal(data, &str)
				v.ErrorMessage = str
			} else if jsonError.Value != "array" {
				return err
			}
		} else if ok && jsonError.Field == "http-code" {
			tmp := make(map[string]interface{})
			json.Unmarshal(data, &tmp)
			intVar, _ := strconv.Atoi(tmp["http-code"].(string))
			v.HttpCode = intVar
		} else {
			return err
		}

		var jsonSyntaxErr *json.SyntaxError
		if ok := errors.As(err, &jsonSyntaxErr); ok && jsonSyntaxErr.Error() == "invalid character '<' looking for beginning of value" {
			return fmt.Errorf(string(data))
		}
	}

	rawDataJson, err := json.Marshal(v.RawData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawDataJson, &v.Data); err != nil {
		return err
	}

	*e = ResponseError(v)
	return nil
}

func APIError(callbacks ...func(ResponseError) ResponseError) func(res *http.Response) error {
	return func(res *http.Response) error {
		if lib.IsNonOkStatus(res) && lib.IsHTML(res) && res.Header.Get("X-Request-Id") != "" && res.Header.Get("Server") == "nginx" {
			return fmt.Errorf("files.com Server error - request id: %v", res.Header.Get("X-Request-Id"))
		}

		if lib.IsNonOkStatus(res) && lib.IsJSON(res) {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return lib.NonOkError(res)
			}

			re := ResponseError{}

			err = re.UnmarshalJSON(data)
			if err != nil {
				return lib.NonOkError(res)
			}

			if re.IsNil() {
				return lib.NonOkError(res)
			}
			for _, callback := range callbacks {
				re = callback(re)
			}
			return re
		}
		return nil
	}
}
