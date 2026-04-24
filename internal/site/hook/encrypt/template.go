package encrypt

import (
	"errors"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

func encryptFilter(ctx *core.Context) pongo2.FilterFunction {
	return func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
		plaintext, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:encrypt",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}

		password := ""
		if param == nil {
			password = ctx.Config.GetString("hooks.encrypt.password")
		} else {
			password = param.String()
		}

		if password == "" {
			return nil, &pongo2.Error{
				Sender:    "filter:encrypt",
				OrigError: errors.New("password is required"),
			}
		}

		e := &Encrypt{ctx: ctx}

		text, err := e.encrypt(plaintext, password)
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "filter:encrypt",
				OrigError: err,
			}
		}
		return pongo2.AsValue(text), nil
	}
}

func init() {
	template.Register("encrypt", func(ctx *core.Context, set template.TemplateSet) error {
		set.RegisterFilter("encrypt", encryptFilter(ctx))
		return nil
	})
}
