package lib

import (
	"strconv"
	"strings"
)

func NormalizeAPIPath(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ReplaceAll(part, "\x00", "")
		part = strings.ReplaceAll(part, "\\", "/")
		for _, segment := range strings.Split(part, "/") {
			if segment == "" || segment == "." || segment == ".." {
				continue
			}
			cleaned = append(cleaned, segment)
		}
	}
	return strings.Join(cleaned, "/")
}

func UnderscoreDestinationPath(root string, id int64, relativePath string) string {
	return NormalizeAPIPath("_", root, strconv.FormatInt(id, 10), relativePath)
}
