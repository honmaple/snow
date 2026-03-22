package parser

import (
	"github.com/honmaple/snow/utils"
	"strconv"
	"strings"
)

type (
	Heading struct {
		Id       string
		Level    int32
		Title    string
		Children []*Heading
	}
	Result struct {
		FrontMatter map[string]any
		Toc         []*Heading
		Summary     string
		Content     string
		RawContent  string
	}
)

func (r *Result) SetFrontMatter(k, v string) {
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
		oldv, ok := r.FrontMatter[k]
		if ok {
			r.FrontMatter[k] = utils.Merge(oldv, realVal)
		} else {
			r.FrontMatter[k] = realVal
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
		if oldv, ok := r.FrontMatter[key]; ok {
			r.FrontMatter[key] = utils.Merge(oldv, val)
		} else {
			r.FrontMatter[key] = val
		}
	}
}
