package lib

import "strings"

func UrlJoinNoEscape(paths ...string) string {
	var newPaths []string
	for _, p := range paths {
		if p == "" {
			continue
		}
		newPaths = append(newPaths, strings.TrimPrefix(strings.TrimSuffix(p, "/"), "/"))
	}

	return strings.Join(newPaths, "/")
}

func UrlLastSegment(path string) (rest string, lastSegment string) {
	segments := strings.Split(
		UrlJoinNoEscape(strings.Split(path, "/")...),
		"/",
	)
	if len(segments) == 0 {
		return "", ""
	}
	return UrlJoinNoEscape(segments[:len(segments)-1]...), segments[len(segments)-1]
}
