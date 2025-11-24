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

// New creates a new GitIgnore instance based on the current operating system's default ignore
// patterns, with optional overrides. If overrides are provided, they will be used instead of
// the default patterns. If no overrides are provided, the function will use the default ignore
// patterns for the current operating system.
func New(overrides ...string) (*ignore.GitIgnore, error) {
	if overrides == nil {
		osIgnoreLines, err := osIgnoreLines()
		if err != nil {
			return nil, err
		}
		return ignore.CompileIgnoreLines(osIgnoreLines...), nil
	}
	return ignore.CompileIgnoreLines(overrides...), nil
}

// NewWithAllowList creates a new GitIgnore instance based on the OS specific defaults,
// but removes items in the allow list from the ignored patterns. If no allow list is provided,
// it will use the default ignore patterns for the current operating system, which is the same
// behavior as calling New() with no arguments.
func NewWithAllowList(allowed ...string) (*ignore.GitIgnore, error) {
	if len(allowed) == 0 {
		return New()
	}
	osIgnoreLines, err := osIgnoreLines()
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of allowed patterns
	allowedMap := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	// Filter out any osIgnoreLines that are in the allowed list
	filtered := make([]string, 0, len(osIgnoreLines))
	for _, line := range osIgnoreLines {
		if _, ok := allowedMap[line]; !ok {
			filtered = append(filtered, line)
		}
	}
	return ignore.CompileIgnoreLines(filtered...), nil
}

// NewWithDenyList creates a new GitIgnore instance based on the OS specific defaults,
// and adds items in the deny list to the ignored patterns. If no deny list is provided,
// it will use the default ignore patterns for the current operating system, which is the same
// behavior as calling New() with no arguments.
func NewWithDenyList(denied ...string) (*ignore.GitIgnore, error) {
	if len(denied) == 0 {
		return New()
	}
	lines, err := osIgnoreLines()
	if err != nil {
		return nil, err
	}
	lines = append(lines, denied...)
	unique := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		unique[line] = struct{}{}
	}
	var uniqueLines []string
	for line := range unique {
		uniqueLines = append(uniqueLines, line)
	}
	return ignore.CompileIgnoreLines(uniqueLines...), nil
}

func osIgnoreLines() ([]string, error) {
	os := runtime.GOOS
	switch os {
	case "windows":
		return format(WindowsGitignore), nil
	case "darwin":
		return format(macOSGitignore), nil
	case "linux":
		return format(LinuxGitignore), nil
	default:
		return nil, fmt.Errorf("unknown os %s", os)
	}
}

func format(b []byte) []string {
	return append(strings.Split(string(commonGitignore), "\n"), strings.Split(string(b), "\n")...)
}
