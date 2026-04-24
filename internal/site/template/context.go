package template

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/spf13/viper"
)

type registry struct {
	ctx *core.Context
}

const (
	DAY90 = 90 * 24 * time.Hour
	DAY10 = 10 * 24 * time.Hour
	DAY7  = 7 * 24 * time.Hour
	DAY   = 24 * time.Hour
)

func newError(name string, err error) *pongo2.Error {
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

// call function and return nothing
func (r *registry) slient(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	return pongo2.AsValue(""), nil
}

func (r *registry) jsonify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	v := in.Interface()

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, newError("jsonify", err)
	}
	return pongo2.AsValue(string(buf)), nil
}

func (r *registry) parser(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	kind := "yaml"
	if param != nil {
		if t := param.String(); t != "" {
			kind = t
		}
	}

	conf := viper.New()
	conf.SetConfigType(kind)

	if err := conf.ReadConfig(strings.NewReader(in.String())); err != nil {
		return nil, newError("parser", err)
	}
	return pongo2.AsValue(conf.AllSettings()), nil
}

func (r *registry) absURL(vars map[string]any) pongo2.FilterFunction {
	lang := r.ctx.GetDefaultLanguage()
	if v, ok := vars["current_lang"]; ok {
		lang = v.(string)
	}

	lctx := r.ctx.For(lang)
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("absURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(lctx.GetURL(v)), nil
	}
}

func (r *registry) relURL(vars map[string]any) pongo2.FilterFunction {
	lang := r.ctx.GetDefaultLanguage()
	if v, ok := vars["current_lang"]; ok {
		lang = v.(string)
	}

	lctx := r.ctx.For(lang)
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("relURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(lctx.GetRelURL(v)), nil
	}
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

		set.RegisterFilter("parser", r.parser)
		set.RegisterFilter("slient", r.slient)
		set.RegisterFilter("jsonify", r.jsonify)

		set.RegisterTransient("config", r.config)
		set.RegisterTransientFilter("absURL", r.absURL)
		set.RegisterTransientFilter("relURL", r.relURL)
		return nil
	})
}
