package parser

import (
	"errors"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

// {{ "*content*" | parser:"orgmode" }}
// {{ "*content*" | parser:"markdown" }}
func templateFilter(ctx *core.Context) pongo2.FilterFunction {
	parser := New(ctx).(*parserImpl)
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		content := in.String()
		if content == "" {
			return nil, template.NewFilterError("parser", errors.New("filter input argument is empty"))
		}
		format := param.String()
		if format == "" {
			return nil, template.NewFilterError("parser", errors.New("format is empty"))
		}

		result, err := parser.ParseString(content, param.String())
		if err != nil {
			return nil, template.NewFilterError("parser", err)
		}
		return pongo2.AsValue(result.Content), nil
	}
}

func init() {
	template.Register("parser", func(ctx *core.Context, set template.TemplateSet) error {
		set.RegisterFilter("parser", templateFilter(ctx))
		return nil
	})
}
