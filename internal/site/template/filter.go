package template

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/flosch/pongo2/v7"
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
func dict(args ...any) map[string]any {
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
func slice(args ...any) []any {
	m := make([]any, len(args))
	for i, arg := range args {
		m[i] = arg
	}
	return m
}

// call function and return nothing
func slient(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	return pongo2.AsValue(""), nil
}

func jsonify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	v := in.Interface()

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, newError("jsonify", err)
	}
	return pongo2.AsValue(string(buf)), nil
}

func parser(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
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
