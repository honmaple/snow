package utils

import (
	"fmt"
	"strings"
	"time"
)

var DateFormats = []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}

func ParseTime(value string) (time.Time, error) {
	for _, f := range DateFormats {
		// 需要和yaml时间解析保持一致
		if date, err := time.Parse(f, value); err == nil {
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

func dropQuote(str string) string {
	str = strings.TrimSpace(str)
	if len(str) < 2 || str[0] != '"' || str[len(str)-1] != '"' {
		return str
	}
	if str[len(str)-2] == '\\' {
		return str
	}
	return str[1 : len(str)-1]
}

func SplitTrim(str string, split string) []string {
	result := make([]string, 0)

	last, idx := 0, 0

	l := len(split)
	for idx < len(str) {
		if str[idx] == '"' && (idx == 0 || str[idx-1] != '\\') {
			tmp := idx + 1
			for ; tmp < len(str); tmp++ {
				if str[tmp] == '"' && str[tmp-1] != '\\' {
					break
				}
			}
			if tmp < len(str) {
				idx = tmp + 1
			}
		}
		if idx+l >= len(str) {
			result = append(result, dropQuote(str[last:]))
			break
		}
		if str[idx:idx+l] == split {
			result = append(result, dropQuote(str[last:idx]))
			idx = idx + l
			last = idx
		} else {
			idx++
		}
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
