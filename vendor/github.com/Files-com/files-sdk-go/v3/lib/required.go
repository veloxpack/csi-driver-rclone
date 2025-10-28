package lib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func CheckRequired(iStruct interface{}) error {
	var errors = make([]string, 0)
	val := reflect.ValueOf(iStruct)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if iStruct == nil {
		return nil
	}

	if val.Kind() == reflect.Map {
		return nil
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("CheckRequired expects struct input. Got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		tag := sf.Tag.Get("required")
		value := val.FieldByName(sf.Name)
		jsonValue, err := json.Marshal(value.Interface())
		if err != nil {
			return err
		}

		if tag == "true" && (string(jsonValue) == "null" || string(jsonValue) == JSONEmptyValue(sf.Type)) {
			errors = append(
				errors,
				CreateError(iStruct, sf.Name).Error(),
			)
		}
	}
	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	return nil
}

func CreateError(i interface{}, name string) error {
	structName := reflect.TypeOf(i).Name()
	return fmt.Errorf("missing required field: %v{}.%v", structName, name)
}

func JSONEmptyValue(v reflect.Type) string {
	switch v.Kind() {
	case reflect.Map, reflect.Struct:
		return "{}"
	case reflect.Slice:
		return "[]"
	case reflect.Int, reflect.Int64:
		return "0"
	case reflect.String:
		return `""`
	default:
		return ""
	}
}
