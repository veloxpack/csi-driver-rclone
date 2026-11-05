package lib

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type S3Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	HostId    string   `xml:"HostId"`
	RequestId string   `xml:"RequestId"`
}

func S3ErrorIsRequestHasExpired(err error) bool {
	var s3Error S3Error
	return errors.As(err, &s3Error) && s3Error.Message == "Request has expired"
}

// Deprecated: Use ClassifyS3Error and inspect the Class field instead.
func S3ErrorIsRequestTimeout(err error) bool {
	classified, ok := ClassifyS3Error(err)
	return ok && classified.Class == "request_timeout"
}

type S3ErrorClassification struct {
	Class     string
	Message   string
	Retryable bool
}

func ClassifyS3Error(err error) (S3ErrorClassification, bool) {
	var s3Error S3Error
	if !errors.As(err, &s3Error) {
		return S3ErrorClassification{}, false
	}
	return ClassifyS3ErrorCode(s3Error.Code), true
}

func ClassifyS3ErrorCode(code string) S3ErrorClassification {
	switch code {
	case "RequestTimeout":
		return S3ErrorClassification{
			Class:     "request_timeout",
			Message:   "Upload to backend storage timed out.",
			Retryable: true,
		}
	case "InternalError", "TemporarilyUnavailable", "ServiceUnavailable":
		return S3ErrorClassification{
			Class:     "service_unavailable",
			Message:   "Error returned by the remote service",
			Retryable: true,
		}
	case "SlowDown", "RequestLimitExceeded":
		return S3ErrorClassification{
			Class:     "rate_limit",
			Message:   "Rate limited by remote service",
			Retryable: true,
		}
	case "NotFound", "NoSuchKey", "NoSuchEntity":
		return S3ErrorClassification{Class: "not_found", Message: "Resource not found on the remote service"}
	case "StorageQuotaExceeded":
		return S3ErrorClassification{Class: "storage_limit", Message: "Storage limit reached for the account"}
	case "EntityTooLarge", "RequestEntityTooLarge":
		return S3ErrorClassification{Class: "resource_limit_exceeded", Message: "The object exceeded the maximum allowed size"}
	case "NoSuchBucket":
		return S3ErrorClassification{Class: "connect_validation", Message: "The specified bucket does not exist"}
	case "InvalidBucketName":
		return S3ErrorClassification{Class: "connect_validation", Message: "The specified bucket is not valid"}
	case "InvalidAccessKeyId":
		return S3ErrorClassification{Class: "connect_validation", Message: "The provided access key is invalid"}
	case "InvalidRequest":
		return S3ErrorClassification{Class: "upload_error", Message: "Invalid Request. The resource name or option is not supported by the remote service"}
	case "AccessDenied", "Forbidden":
		return S3ErrorClassification{Class: "permissions", Message: "Access denied by the remote service"}
	case "Unauthorized", "AuthorizationHeaderMalformed":
		return S3ErrorClassification{Class: "permissions", Message: "Permission denied by the remote service"}
	case "IllegalLocationConstraintException":
		return S3ErrorClassification{Class: "connect_validation", Message: "The specified region does not match the region of the bucket"}
	case "SignatureDoesNotMatch":
		return S3ErrorClassification{Class: "authentication", Message: "Signature does not match"}
	case "PermanentRedirect":
		return S3ErrorClassification{Class: "connect", Message: "Permanent redirect: the bucket you are attempting to access must be addressed using the specified endpoint"}
	case "PreconditionFailed":
		return S3ErrorClassification{Class: "if_match_condition_not_met", Message: "The file has changed since comparison or no longer exists. This can happen if the file was deleted or modified by another process."}
	case "InvalidObjectState":
		return S3ErrorClassification{Class: "archived_object", Message: "The object is in an archived storage class (such as Glacier) and must be restored before it can be accessed"}
	case "InvalidKey":
		return S3ErrorClassification{Class: "invalid_path", Message: "The specified path contains invalid characters"}
	case "AccountProblem", "PoorAccountStanding":
		return S3ErrorClassification{Class: "account_problem", Message: "There is a problem with your Amazon Web Services account that prevents the action from completing successfully"}
	case "SecurityPolicyViolated":
		return S3ErrorClassification{Class: "permissions", Message: "Request violates VPC Service Controls"}
	case "DatabaseTimeout":
		return S3ErrorClassification{Class: "database_timeout", Message: "Remote service timeout", Retryable: true}
	case "AccessDeniedException":
		return S3ErrorClassification{Class: "authentication", Message: "Access denied when assuming role. Verify the trust policy allows this principal."}
	case "ExpiredTokenException":
		return S3ErrorClassification{Class: "authentication", Message: "Session token has expired"}
	}

	return S3ErrorClassification{
		Class:   "unknown",
		Message: "Error returned by the remote service",
	}
}

func (s S3Error) Error() string {
	return fmt.Sprintf("%v - %v", s.Code, s.Message)
}

func (s S3Error) Empty() bool {
	return s.Message == "" && s.Code == ""
}
