package lib

import (
	"os"
	"path/filepath"
	"strings"
)

type Path struct {
	Path      string
	Separator string
}

const (
	URLPathSeparator = "/"
)

func (p Path) PathSeparator() string {
	if p.Separator == "" {
		return string(os.PathSeparator)
	}

	return p.Separator
}

func (p Path) Pop() string {
	_, last := filepath.Split(strings.TrimSuffix(p.Path, p.PathSeparator()))
	return last
}

func (p Path) EndingSlash() bool {
	if p.Path == "" {
		return false
	}
	return p.Path[len(p.Path)-1:] == p.PathSeparator()
}

func (p Path) PruneStartingSlash() Path {
	if p.Path == "" {
		return p
	}

	if p.Path[0:1] == p.PathSeparator() {
		return Path{Path: p.Path[1:]}
	}
	return p
}

func (p Path) PruneEndingSlash() Path {
	if !p.EndingSlash() {
		return Path{Path: p.Path}
	}

	return Path{Path: p.Path[0 : len(p.Path)-1]}
}

func (p Path) ConvertEmptyToRoot() Path {
	if p.Path == "" {
		return Path{Path: "."}
	}

	return p
}

func (p Path) Join(str ...string) Path {
	return Path{Path: filepath.Join(p.String(), filepath.Join(str...))}
}

func (p Path) String() string {
	return p.Path
}

func (p Path) NormalizePathSystemForAPI() Path {
	return NewUrlPath(UrlJoinNoEscape(strings.Split(filepath.Clean(p.Path), p.PathSeparator())...)).PruneStartingSlash()
}

func (p Path) SwitchPathSeparator(separator string) Path {
	return Path{Path: strings.Join(strings.Split(p.Path, p.PathSeparator()), separator), Separator: separator}
}

func NewUrlPath(path string) Path {
	return Path{
		Path:      path,
		Separator: URLPathSeparator,
	}
}
