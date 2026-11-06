package buckets

import (
	"path/filepath"
	"strings"
)

// SplitNameExt splits a file name into the (plainName, type) pair Drive stores
// it under.
func SplitNameExt(fileName string) (plainName, fileType string) {
	base := filepath.Base(fileName)
	ext := filepath.Ext(base)
	if ext == base {
		return base, ""
	}
	return strings.TrimSuffix(base, ext), strings.TrimPrefix(ext, ".")
}

// JoinNameExt rebuilds the file name a stored (plainName, type) pair displays
// as, the way every Drive client renders one.
func JoinNameExt(plainName, fileType string) string {
	if fileType == "" {
		return plainName
	}
	return plainName + "." + fileType
}
