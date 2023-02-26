package utils

import (
	"strings"
	"time"
)

func Bool(value interface{}) bool {
	return value != nil && (value == true || value == "true")

}

func Compare(value interface{}, other interface{}) int {
	if value == other {
		return 0
	}
	if other == nil {
		return 1
	}
	if value == nil {
		return -1
	}
	switch v := value.(type) {
	case []string:
		var others []string
		otherv, ok := other.(string)
		if ok {
			others = strings.Split(otherv, ",")
		} else {
			others = other.([]string)
		}
		if len(v) != len(others) {
			return -1
		}
		for i, o := range others {
			if v[i] != o {
				return -1
			}
		}
		return 0
	case string:
		return strings.Compare(v, other.(string))
	case bool:
		if v {
			return 1
		}
		return -1
	case int:
		if v > other.(int) {
			return 1
		}
		return -1
	case time.Time:
		if v.Before(other.(time.Time)) {
			return 1
		}
		return -1
	}
	return 0
}
