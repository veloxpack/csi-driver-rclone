package lib

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

func PathEscape(path string) (string, error) {
	if path == "nil" {
		return "", nil
	}
	pathParts := strings.Split(path, "/")
	newParts := make([]string, len(pathParts))

	for i, part := range pathParts {
		newParts[i] = url.PathEscape(part)
	}
	var err error
	if len(newParts) > 1 {
		path, err = url.JoinPath(newParts[0], newParts[1:]...)
		if err != nil {
			return path, err
		}
	} else {
		path = newParts[0]
	}

	return NewUrlPath(path).PruneStartingSlash().String(), nil
}

func BuildPath(resourcePath string, values interface{}) (string, error) {
	// Regular expression to find placeholders
	r := regexp.MustCompile(`\{(.+?)\}`) // Use non-greedy match to capture individual placeholders

	// Iterate over placeholders and replace them
	var unreplacedPlaceholders []string
	matches := r.FindAllSubmatch([]byte(resourcePath), -1)
	for _, match := range matches {
		// Extract placeholder name
		placeholder := string(match[1])

		var value interface{}
		var exists bool
		if m, ok := values.(map[string]interface{}); ok {
			value, exists = m[placeholder]
		} else if pathValue, err := findTag(values, "path", placeholder); err == nil {
			exists = true
			value = pathValue
		} else if jsonValue, err := findTag(values, "json", placeholder); err == nil {
			exists = true
			value = jsonValue
		}

		if !exists {
			// path is allowed to be empty because that can represent the root path.
			if placeholder == "path" {
				resourcePath = strings.Replace(resourcePath, string(match[0]), "", 1)
			} else {
				unreplacedPlaceholders = append(unreplacedPlaceholders, placeholder)
			}
			continue
		}

		// Convert value to string
		var stringValue string
		switch v := value.(type) {
		case string:
			var err error
			if placeholder == "path" {
				stringValue, err = PathEscape(v)
				if err != nil {
					return "", err
				}
			}
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		// Replace placeholder in resourcePath
		resourcePath = strings.Replace(resourcePath, string(match[0]), stringValue, 1)
	}

	// Check if there are unreplaced placeholders
	if len(unreplacedPlaceholders) > 0 {
		return "", fmt.Errorf("placeholders %v were not replaced", unreplacedPlaceholders)
	}

	return resourcePath, nil
}

func findTag(s interface{}, tag string, value string) (interface{}, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected struct or pointer to struct, got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tag := typ.Field(i).Tag.Get(tag)
		if tag != "" && tag == value {
			fieldVal := val.Field(i)
			if !fieldVal.IsZero() {
				return fieldVal.Interface(), nil
			}
		}
	}

	return "", fmt.Errorf("`%v` tag not found in struct", tag)
}
