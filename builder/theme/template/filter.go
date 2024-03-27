package template

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/config"
	"github.com/spf13/viper"
)

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
func dict(args ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})

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
func slice(args ...interface{}) []interface{} {
	m := make([]interface{}, len(args))
	for i, arg := range args {
		m[i] = arg
	}
	return m
}

// call function and return nothing
func slient(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(""), nil
}

func jsonify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v := in.Interface()

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, newError("jsonify", err)
	}
	return pongo2.AsValue(string(buf)), nil
}

func parser(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
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

func absURL(conf config.Config) pongo2.FilterFunction {
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("absURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(conf.GetURL(v)), nil
	}
}

func relURL(conf config.Config) pongo2.FilterFunction {
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("relURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(conf.GetRelURL(v)), nil
	}
}

func newConfig(conf config.Config) func(map[string]interface{}) interface{} {
	langs := make(map[string]interface{})
	for lang, c := range conf.Languages {
		langs[lang] = c.AllSettings()
	}
	return func(ctx map[string]interface{}) interface{} {
		lang := ctx["current_lang"]
		if lang == nil {
			return langs[conf.Site.Language]
		}
		return langs[lang.(string)]
	}
}
