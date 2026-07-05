package shortcode

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/spf13/cast"
)

type Params map[string]any

func (p Params) Pop(key string) any {
	value := p[key]
	delete(p, key)
	return value
}

func (p Params) Get(key string) any {
	return p[key]
}

func (p Params) String() string {
	if len(p) == 0 {
		return ""
	}

	keys := make([]string, 0, len(p))
	for key := range p {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	attrs := make([]string, 0, len(keys))
	for _, key := range keys {
		value := cast.ToString(p[key])
		if value == "" {
			attrs = append(attrs, html.EscapeString(key))
			continue
		}
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, html.EscapeString(key), html.EscapeString(value)))
	}
	return strings.Join(attrs, " ")
}
