package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func SetField(obj interface{}, name string, value interface{}) error {
	field := reflect.ValueOf(obj).Elem().FieldByName(name)

	if !field.IsValid() {
		return fmt.Errorf("No such field: %s in %T", name, obj)
	}

	if !field.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}
	fieldType := field.Type()
	val := reflect.ValueOf(value)
	if fieldType != val.Type() {
		return fmt.Errorf("Provided value type didn't match obj field type")
	}
	field.Set(val)
	return nil
}

func GetField(obj interface{}, name string) (interface{}, error) {
	var value reflect.Value

	kind := reflect.TypeOf(obj).Kind()
	if kind != reflect.Struct && kind != reflect.Ptr {
		return nil, fmt.Errorf("Cannot use GetField on a non-struct interface")
	}
	if kind == reflect.Ptr {
		value = reflect.ValueOf(obj).Elem()
	} else {
		value = reflect.ValueOf(obj)
	}
	field := value.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("No such field: %s in %T", name, obj)
	}
	return field.Interface(), nil
}

func StringConcat(strs ...string) string {
	var b strings.Builder
	for _, str := range strs {
		b.WriteString(str)
	}
	return b.String()
}

func StringReplace(s string, vars map[string]string) string {
	if vars == nil || s == "" {
		return s
	}
	args := make([]string, 0)
	for k, v := range vars {
		args = append(args, k)
		args = append(args, v)
	}
	r := strings.NewReplacer(args...)
	return r.Replace(s)
}

// CheckInList ..
func CheckInList(f []string, v string) bool {
	for i := range f {
		if f[i] == v {
			return true
		}
	}
	return false
}

func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(s))
}

func SplitTrim(str string, split string) []string {
	result := make([]string, 0)
	for _, s := range strings.Split(str, ",") {
		result = append(result, strings.TrimSpace(s))
	}
	return result
}

func SplitPrefix(str string, split string) []string {
	result := make([]string, 0)
	for i := 0; i < len(str); i++ {
		for ; i < len(str) && str[i] != '/'; i++ {
		}
		result = append(result, str[:i])
	}
	return result
}
