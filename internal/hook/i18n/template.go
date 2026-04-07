package i18n

import (
	"errors"
	"fmt"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/theme/template"
	"github.com/honmaple/snow/internal/core"
	"sync"
)

type i18nNode struct {
	args   []pongo2.IEvaluator
	format pongo2.IEvaluator
}

func (node *i18nNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) error {
	val, err := node.format.Evaluate(ctx)
	if err != nil {
		return err
	}
	args := make([]any, len(node.args))
	for i, arg := range node.args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return err
		}
		args[i] = val.Interface()
	}
	lang := ""
	if v, ok := ctx.Public["current_lang"]; ok {
		lang = v.(string)
	}

	i18n, ok := ctx.Public["__i18n__"].(*I18n)
	if !ok {
		return errors.New("can't found i18n")
	}

	format := i18n.Translate(val.String(), lang)
	_, err = writer.WriteString(fmt.Sprintf(format, args...))
	return err
}

func i18nTagParser(ctx *core.Context) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, error) {
		node := &i18nNode{args: make([]pongo2.IEvaluator, 0)}

		// {% i18n "aaaaa" %}
		// {% i18n "aaaaa %s %d" "ccc" 108 %}
		if arguments.Count() == 0 {
			return nil, arguments.Error("Tag 'i18n' requires at least one argument.", nil)
		}

		format, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.format = format

		for arguments.Remaining() > 0 {
			arg, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.args = append(node.args, arg)
		}
		return node, nil
	}
}

func i18nTranslate(ctx *core.Context, vars map[string]any) any {
	lang := ""
	if v, ok := vars["current_lang"]; ok {
		lang = v.(string)
	}
	// {{ i18n("cc") }}
	// {{ i18n("cc %d", 11) }}
	return func(id string, args ...any) string {
		i18n, ok := vars["__i18n__"].(*I18n)
		if !ok {
			return fmt.Sprintf(id, args...)
		}
		format := i18n.Translate(id, lang)
		if len(args) == 0 {
			return format
		}
		return fmt.Sprintf(format, args...)
	}
}

var (
	internalOnce sync.Once
	internalI18n *I18n
)

func getI18n() *I18n {
	internalOnce.Do(func() {
		internalI18n = &I18n{}
	})
	return internalI18n
}

func init() {
	template.RegisterContextTag("T", i18nTagParser)
	template.RegisterContextTag("i18n", i18nTagParser)

	template.RegisterTransientFunc("_", i18nTranslate)
	template.RegisterTransientFunc("T", i18nTranslate)
	template.RegisterTransientFunc("i18n", i18nTranslate)
}
