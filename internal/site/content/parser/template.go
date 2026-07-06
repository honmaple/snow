package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

// {{ "*content*" | parser:"orgmode" }}
// {{ "*content*" | parser:"markdown" }}
// {{ "*content*" | parser:".md" }}
// {{ "*content*" | parser:".org" }}
func (d *parserImpl) parseString(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
	content := in.String()
	if content == "" {
		return nil, template.NewFilterError("parser", errors.New("filter input argument is empty"))
	}
	format := param.String()
	if format == "" {
		return pongo2.AsValue(content), nil
	}

	var markup MarkupParser
	if strings.HasPrefix(format, ".") {
		markup = d.extMap[format]
	} else {
		markup = d.formatMap[format]
	}
	if markup == nil {
		return nil, fmt.Errorf("no %s parser", format)
	}

	result, err := markup.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, template.NewFilterError("parser", err)
	}
	return pongo2.AsValue(result.Content), nil
}

func init() {
	template.Register("parser", func(ctx *core.Context, set template.TemplateSet) error {
		parser := New(ctx).(*parserImpl)

		set.RegisterFilter("parser", parser.parseString)
		return nil
	})
}
