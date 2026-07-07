package template

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/honmaple/snow/internal/core"
)

type registry struct {
	ctx *core.Context
}

func NewFilterError(name string, err error) *pongo2.Error {
	return &pongo2.Error{
		Sender:    "filter:" + name,
		OrigError: err,
	}
}

// {% set newDict = dict("title", "h1", "weight", 0) %}
func (r *registry) dict(args ...any) map[string]any {
	m := make(map[string]any)

	for i := 0; i < len(args); i = i + 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		if i+1 == len(args) {
			m[key] = nil
		} else {
			m[key] = args[i+1]
		}
	}
	return m
}

// {% set newSlice = slice("item1", "item2") %}
func (r *registry) slice(args ...any) []any {
	m := make([]any, len(args))
	for i, arg := range args {
		m[i] = arg
	}
	return m
}

func (r *registry) startsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// call function and return nothing
func (r *registry) slient(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	return pongo2.AsValue(""), nil
}

func (r *registry) jsonify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	v := in.Interface()

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, NewFilterError("jsonify", err)
	}
	return pongo2.AsValue(string(buf)), nil
}

func (r *registry) trimline(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	text := in.String()
	if text == "" {
		return pongo2.AsValue(""), nil
	}
	lines := strings.Split(text, "\n")

	start := 0
	end := len(lines)

	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start == end {
		return pongo2.AsValue(""), nil
	}

	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return pongo2.AsValue(strings.Join(lines[start:end], "\n")), nil
}

func (r *registry) dedent(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	text := in.String()
	if text == "" {
		return pongo2.AsValue(""), nil
	}

	min := -1

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		newline := strings.TrimLeft(line, " ")
		if newline == "" {
			continue
		}
		if indent := len(line) - len(newline); min == -1 || indent < min {
			min = indent
		}
	}
	if min == -1 {
		return pongo2.AsValue(text), nil
	}
	prefix := strings.Repeat(" ", min)
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, prefix)
	}
	return pongo2.AsValue(strings.Join(lines, "\n")), nil
}

func (r *registry) unmarshal(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	data := in.String()
	if data == "" {
		return nil, NewFilterError("unmarshal", errors.New("filter input argument is empty"))
	}
	var (
		err    error
		result any
	)
	format := param.String()
	switch format {
	case "json":
		err = json.Unmarshal([]byte(data), &result)
	case "yaml":
		err = yaml.Unmarshal([]byte(data), &result)
	case "toml":
		err = toml.Unmarshal([]byte(data), &result)
	default:
		result = data
	}
	if err != nil {
		return nil, NewFilterError("unmarshal", err)
	}
	return pongo2.AsValue(result), nil
}

func (r *registry) absURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
	v, ok := in.Interface().(string)
	if !ok {
		return nil, NewFilterError("absURL", errors.New("filter input argument must be of type 'string'"))
	}
	return pongo2.AsValue(r.ctx.GetURL(v)), nil
}

func (r *registry) relURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
	v, ok := in.Interface().(string)
	if !ok {
		return nil, NewFilterError("relURL", errors.New("filter input argument must be of type 'string'"))
	}
	return pongo2.AsValue(r.ctx.GetRelURL(v)), nil
}

func (r *registry) config(vars map[string]any) any {
	lang := r.ctx.GetDefaultLanguage()
	if v, ok := vars["current_lang"]; ok {
		lang = v.(string)
	}
	return r.ctx.For(lang).Config.AllSettings()
}

func init() {
	Register("default", func(ctx *core.Context, set TemplateSet) error {
		r := &registry{ctx: ctx}

		set.Register("dict", r.dict)
		set.Register("slice", r.slice)
		set.Register("startsWith", r.startsWith)

		set.RegisterFilter("slient", r.slient)
		set.RegisterFilter("jsonify", r.jsonify)
		set.RegisterFilter("dedent", r.dedent)
		set.RegisterFilter("trimline", r.trimline)
		set.RegisterFilter("unmarshal", r.unmarshal)

		set.RegisterFilter("absURL", r.absURL)
		set.RegisterFilter("relURL", r.relURL)

		set.RegisterTransient("config", r.config)
		return nil
	})
}
