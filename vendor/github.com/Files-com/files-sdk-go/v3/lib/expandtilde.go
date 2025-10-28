package lib

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandTilde(path string) string {
	dirname, err := os.UserHomeDir()
	if err == nil {
		if path == "~" {
			// In case of "~", which won't be caught by the "else if"
			path = dirname
		} else if strings.HasPrefix(path, "~/") {
			// Use strings.HasPrefix so we don't match paths like
			// "/something/~/something/"
			path = filepath.Join(dirname, path[2:])
		}
	}

	return path
}
