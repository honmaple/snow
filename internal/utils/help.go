package utils

import (
	"cmp"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

func Bool(value any) bool {
	return value != nil && (value == true || value == "true")

}

func Sort(key string, f func(string, int, int) int) func(int, int) bool {
	ks := strings.Split(key, ",")
	return func(i, j int) bool {
		for _, k := range ks {
			k = strings.TrimSpace(k)
			if strings.HasSuffix(strings.ToUpper(k), " DESC") {
				k = k[:len(k)-5]
				if result := f(k, i, j); result != 0 {
					return result > 0
				}
				continue
			}
			if strings.HasSuffix(strings.ToUpper(k), " ASC") {
				k = k[:len(k)-4]
			}
			if result := f(k, i, j); result != 0 {
				return result < 0
			}
		}
		return f("-", i, j) <= 0
	}
}

func isNumber(value any) bool {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func compareStringSlices(value []string, other any) int {
	var others []string
	switch v := other.(type) {
	case []string:
		others = v
	case string:
		others = strings.Split(v, ",")
	default:
		return cmp.Compare(strings.Join(value, ","), cast.ToString(other))
	}
	for i := 0; i < len(value) && i < len(others); i++ {
		if result := cmp.Compare(value[i], others[i]); result != 0 {
			return result
		}
	}
	return cmp.Compare(len(value), len(others))
}

func Compare(value any, other any) int {
	if other == nil {
		return 1
	}
	if value == nil {
		return -1
	}
	if reflect.DeepEqual(value, other) {
		return 0
	}
	if isNumber(value) && isNumber(other) {
		return cmp.Compare(cast.ToFloat64(value), cast.ToFloat64(other))
	}
	switch v := value.(type) {
	case []string:
		return compareStringSlices(v, other)
	case string:
		return cmp.Compare(v, cast.ToString(other))
	case bool:
		if o, ok := other.(bool); ok {
			return cmp.Compare(cast.ToInt(v), cast.ToInt(o))
		}
		return cmp.Compare(cast.ToString(v), cast.ToString(other))
	case time.Time:
		if o, ok := other.(time.Time); ok {
			return v.Compare(o)
		}
		return cmp.Compare(v.String(), cast.ToString(other))
	}
	return cmp.Compare(cast.ToString(value), cast.ToString(other))
}

func Merge(v0, v1 any) any {
	val0 := reflect.ValueOf(v0)
	val1 := reflect.ValueOf(v1)
	// if val0.Kind() != val1.Kind() {
	//	return v0
	// }
	switch val0.Kind() {
	case reflect.Slice, reflect.Array:
		switch val1.Kind() {
		case reflect.Slice, reflect.Array:
			m := make([]any, 0)
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
			m := make(map[string]any)
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

func DeepCopy(m map[string]any) map[string]any {
	newm := make(map[string]any)
	for k, v := range m {
		mval, ok := v.(map[string]any)
		if ok {
			newm[k] = DeepCopy(mval)
			continue
		}
		newm[k] = v
	}
	return newm
}
