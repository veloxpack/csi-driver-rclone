//go:build darwin

package file

import (
	"os"
	"path/filepath"
)

// tmpDownloadPathOnNotExist create .download folder a common pattern on macOS
func tmpDownloadPathOnNotExist(originalPath, tmpPath string) (string, error) {
	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		return "", err
	}
	_, fileName := filepath.Split(originalPath)
	return filepath.Join(tmpPath, fileName), nil
}

func finalizeTmpDownload(tmpName string, finalPath string) error {
	err := os.Rename(tmpName, finalPath)
	if err != nil {
		return err
	}
	downloadPackage, _ := filepath.Split(tmpName)
	return os.Remove(downloadPackage)
}

func removeTmpDownload(tmpName string) error {
	err := os.Remove(tmpName)
	if err != nil {
		return err
	}
	downloadPackage, _ := filepath.Split(tmpName)
	return os.Remove(downloadPackage)
}
