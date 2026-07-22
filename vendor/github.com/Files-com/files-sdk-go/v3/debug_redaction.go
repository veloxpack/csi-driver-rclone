package files_sdk

import "net/http"

const fullyRedactedDebugValue = "<redacted>"
const apiKeyDebugVisiblePrefixLength = 4
const apiKeyDebugMaskSuffix = "****************"

func maskAPIKeyForDebug(value string) string {
	if len(value) <= apiKeyDebugVisiblePrefixLength {
		if value == "" {
			return ""
		}
		return fullyRedactedDebugValue
	}

	return value[:apiKeyDebugVisiblePrefixLength] + apiKeyDebugMaskSuffix
}

func redactDebugValue(value string) string {
	if value == "" {
		return ""
	}

	return fullyRedactedDebugValue
}

func sanitizeDebugRequest(req *http.Request) *http.Request {
	sanitizedReq := req.Clone(req.Context())
	sanitizedReq.Header = req.Header.Clone()

	if value := sanitizedReq.Header.Get("X-FilesAPI-Key"); value != "" {
		sanitizedReq.Header.Set("X-FilesAPI-Key", maskAPIKeyForDebug(value))
	}
	if value := sanitizedReq.Header.Get("X-FilesAPI-Auth"); value != "" {
		sanitizedReq.Header.Set("X-FilesAPI-Auth", redactDebugValue(value))
	}
	if value := sanitizedReq.Header.Get("X-Files-Reauthentication"); value != "" {
		sanitizedReq.Header.Set("X-Files-Reauthentication", redactDebugValue(value))
	}

	return sanitizedReq
}
