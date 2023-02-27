package utils

import (
	"reflect"
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

func Merge(v0, v1 interface{}) interface{} {
	val0 := reflect.ValueOf(v0)
	val1 := reflect.ValueOf(v1)
	// if val0.Kind() != val1.Kind() {
	//	return v0
	// }
	switch val0.Kind() {
	case reflect.Slice, reflect.Array:
		switch val1.Kind() {
		case reflect.Slice, reflect.Array:
			m := make([]interface{}, 0)
			for i := 0; i < val0.Len(); i++ {
				m = append(m, val0.Index(i).Interface())
			}
			for i := 0; i < val1.Len(); i++ {
				m = append(m, val1.Index(i).Interface())
			}
			return m
		default:
			return v1
		}
	case reflect.Map:
		switch val1.Kind() {
		case reflect.Map:
			m := make(map[string]interface{})
			for _, key := range val0.MapKeys() {
				m[key.String()] = val0.MapIndex(key).Interface()
			}
			for _, key := range val1.MapKeys() {
				if v, ok := m[key.String()]; ok {
					m[key.String()] = Merge(v, val1.MapIndex(key).Interface())
				} else {
					m[key.String()] = val1.MapIndex(key).Interface()
				}
			}
			return m
		default:
			return v1
		}
	}
	return v1
}
