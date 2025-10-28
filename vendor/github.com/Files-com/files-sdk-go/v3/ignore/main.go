package ignore

import (
	"fmt"
	"runtime"
	"strings"

	_ "embed"

	ignore "github.com/sabhiram/go-gitignore"
)

//go:embed data/Windows.gitignore
var WindowsGitignore []byte

//go:embed data/macOS.gitignore
var macOSGitignore []byte

//go:embed data/Linux.gitignore
var LinuxGitignore []byte

//go:embed data/common.gitignore
var commonGitignore []byte

func New(overrides ...string) (*ignore.GitIgnore, error) {
	if len(overrides) > 0 {
		return ignore.CompileIgnoreLines(overrides...), nil
	}

	os := runtime.GOOS
	switch os {
	case "windows":
		return ignore.CompileIgnoreLines(format(WindowsGitignore)...), nil
	case "darwin":
		return ignore.CompileIgnoreLines(format(macOSGitignore)...), nil
	case "linux":
		return ignore.CompileIgnoreLines(format(LinuxGitignore)...), nil
	default:
		return &ignore.GitIgnore{}, fmt.Errorf("unknown os %s", os)
	}
}

func format(b []byte) []string {
	return append(strings.Split(string(commonGitignore), "\n"), strings.Split(string(b), "\n")...)
}
