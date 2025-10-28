//go:build !darwin

package file

import (
	"os"
)

func tmpDownloadPathOnNotExist(originalPath, tmpPath string) (string, error) {
	return tmpPath, nil
}

func finalizeTmpDownload(tmpName string, finalPath string) error {
	return os.Rename(tmpName, finalPath)
}

func removeTmpDownload(tmpName string) error {
	return os.Remove(tmpName)
}
