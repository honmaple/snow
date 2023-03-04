package template

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/flosch/pongo2/v6"
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
func (t *template) slient(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(""), nil
}

func (t *template) jsonify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v := in.Interface()

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, newError("jsonify", err)
	}
	return pongo2.AsValue(string(buf)), nil
}

func (t *template) absURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	v, ok := in.Interface().(string)
	if !ok {
		return nil, newError("absURL", errors.New("filter input argument must be of type 'string'"))
	}
	path := ""
	if param.Len() == 0 {
		path = t.conf.GetRelURL(v, t.conf.Site.Language)
	} else {
		path = t.conf.GetRelURL(v, param.String())
	}
	return pongo2.AsValue(t.conf.GetURL(path)), nil
}

func (t *template) relURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	v, ok := in.Interface().(string)
	if !ok {
		return nil, newError("relURL", errors.New("filter input argument must be of type 'string'"))
	}
	if param.Len() == 0 {
		return pongo2.AsValue(t.conf.GetRelURL(v, t.conf.Site.Language)), nil
	}
	return pongo2.AsValue(t.conf.GetRelURL(v, param.String())), nil
}
