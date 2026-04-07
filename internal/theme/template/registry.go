package template

import (
	"errors"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
)

var (
	RegisterTag    = pongo2.RegisterTag
	RegisterFilter = pongo2.RegisterFilter
	// 临时变量
	TransientFuncs     = make(map[string]TransientContextFunc)
	TransientVariables = make(map[string]any)

	GlobalTags      = make(map[string]func(*core.Context) pongo2.TagParser)
	GlobalFilters   = make(map[string]func(*core.Context) pongo2.FilterFunction)
	GlobalFuncs     = make(map[string]func(*core.Context) any)
	GlobalVariables = make(map[string]any)
)

type TransientContextFunc func(*core.Context, map[string]any) any

func Register(k string, v any) {
	GlobalVariables[k] = v
}

func RegisterContextFunc(k string, v func(*core.Context) any) {
	GlobalFuncs[k] = v
}

func RegisterContextTag(k string, v func(*core.Context) pongo2.TagParser) {
	GlobalTags[k] = v
}

func RegisterContextFilter(k string, v func(*core.Context) pongo2.FilterFunction) {
	GlobalFilters[k] = v
}

func RegisterTransient(k string, v any) {
	TransientVariables[k] = v
}

func RegisterTransientFunc(k string, v TransientContextFunc) {
	TransientFuncs[k] = v
}

func absURL(ctx *core.Context) pongo2.FilterFunction {
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("absURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(ctx.GetURL(v)), nil
	}
}

func relURL(ctx *core.Context) pongo2.FilterFunction {
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError("relURL", errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(ctx.GetRelURL(v)), nil
	}
}

func init() {
	Register("dict", dict)
	Register("slice", slice)

	RegisterFilter("parser", parser)
	RegisterFilter("slient", slient)
	RegisterFilter("jsonify", jsonify)

	RegisterContextFilter("absURL", absURL)
	RegisterContextFilter("relURL", relURL)

	RegisterTransient("scratch", newScratch)
	RegisterTransient("newScratch", newScratchFunc)

}
