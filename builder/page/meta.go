package page

import (
	"strconv"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/spf13/cast"
)

type Meta map[string]any

func (m Meta) load(other map[string]any) {
	for k, v := range other {
		m[k] = v
	}
}

func (m Meta) clone() Meta {
	return utils.DeepCopy(m)
}

func (m Meta) Get(k string) any {
	return m[k]
}

func (m Meta) GetInt(k string) int {
	return cast.ToInt(m[k])
}

func (m Meta) GetBool(k string) bool {
	return cast.ToBool(m[k])
}

func (m Meta) GetString(k string) string {
	return cast.ToString(m[k])
}

func (m Meta) GetSlice(k string) []string {
	return cast.ToStringSlice(m[k])
}

func (m Meta) GetStringMap(k string) map[string]any {
	return cast.ToStringMap(m[k])
}

func (m Meta) Set(k, v string) {
	var realVal any

	k = strings.ToLower(k)
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '[' && v[len(v)-1] == ']' {
		realVal = utils.SplitTrim(v[1:len(v)-1], ",")
	} else if b, err := strconv.Atoi(v); err == nil {
		realVal = b
	} else if b, err := strconv.ParseBool(v); err == nil {
		realVal = b
	} else {
		realVal = v
	}

	ss := utils.SplitTrim(k, ".")
	if len(ss) == 1 {
		oldv, ok := m[k]
		if ok {
			m[k] = utils.Merge(oldv, realVal)
		} else {
			m[k] = realVal
		}
		return
	}
	var result map[string]any
	for i := len(ss) - 1; i >= 0; i-- {
		if i == len(ss)-1 {
			result = map[string]any{
				ss[i]: realVal,
			}
		} else {
			result = map[string]any{
				ss[i]: result,
			}
		}
	}
	for key, val := range result {
		if oldv, ok := m[key]; ok {
			m[key] = utils.Merge(oldv, val)
		} else {
			m[key] = val
		}
	}
}
