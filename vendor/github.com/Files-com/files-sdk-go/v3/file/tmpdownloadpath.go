package file

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const TempDownloadExtension = "download"

// tmpDownloadPath Generates a unique temporary download path for a given file path by appending a ".download" extension and, if necessary, additional identifiers to avoid name conflicts.
func tmpDownloadPath(path string, tempPath string) (string, error) {
	var index int
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

	if tempPath != "" {
		_, fileName := filepath.Split(path)
		path = filepath.Join(tempPath, fileName)
	}

	for {
		var tempPath, uniqueness string
		if index == 0 {
			tempPath = fmt.Sprintf("%v.%v", path, TempDownloadExtension)
		} else if index > 25 {
			return "", fmt.Errorf("unable to create a unique temporary path after 25 attempts, consider deleting existing .%v files; attempted path: %v", TempDownloadExtension, path)
		} else {
			if index > 10 {
				for i := 0; i < 4; i++ {
					uniqueness += string(rune(randGenerator.Intn(26) + 'a'))
				}
			} else {
				uniqueness = fmt.Sprintf("%v", index)
			}
			tempPath = fmt.Sprintf("%v (%v).%v", path, uniqueness, TempDownloadExtension)
		}

		if _, err := os.Stat(tempPath); os.IsNotExist(err) {
			return tmpDownloadPathOnNotExist(path, tempPath)
		}
		index++
	}
}
