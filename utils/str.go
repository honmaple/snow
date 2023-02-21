package utils

import (
	"fmt"
	"strings"
	"time"
)

var DateFormats = []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}

func ParseTime(value string) (time.Time, error) {
	for _, f := range DateFormats {
		if date, err := time.ParseInLocation(f, value, time.Local); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("date format error")
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

func CheckInList(f []string, v string) bool {
	for i := range f {
		if f[i] == v {
			return true
		}
	}
	return false
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
