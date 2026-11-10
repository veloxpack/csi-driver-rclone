package files

import (
	"fmt"
)

// SetRange configures arg to download bytes starting at offset until the end
// of the file. It sets the HTTP Range header to "bytes=<offset>-".
func SetRange(arg *DownloadArg, offset int64) error {
	if arg == nil {
		return fmt.Errorf("download argument is nil")
	}

	if offset < 0 {
		return fmt.Errorf("range offset must be non-negative")
	}

	if arg.ExtraHeaders == nil {
		arg.ExtraHeaders = map[string]string{}
	}

	arg.ExtraHeaders["Range"] = fmt.Sprintf("bytes=%d-", offset)
	return nil
}

// SetRangeLength configures arg to download length bytes starting at offset.
// It sets the HTTP Range header to "bytes=<offset>-<end>".
func SetRangeLength(arg *DownloadArg, offset int64, length int64) error {
	if arg == nil {
		return fmt.Errorf("download argument is nil")
	}

	if offset < 0 {
		return fmt.Errorf("range offset must be non-negative")
	}

	if length <= 0 {
		return fmt.Errorf("range length must be positive")
	}

	end := offset + length - 1
	if end < offset {
		return fmt.Errorf("range end overflow")
	}

	if arg.ExtraHeaders == nil {
		arg.ExtraHeaders = make(map[string]string)
	}

	arg.ExtraHeaders["Range"] = fmt.Sprintf("bytes=%d-%d", offset, end)
	return nil
}
