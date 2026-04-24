package i18n

import (
	"fmt"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

type i18nNode struct {
	i18n   *I18n
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

	format := node.i18n.Translate(val.String(), lang)
	_, err = writer.WriteString(fmt.Sprintf(format, args...))
	return err
}

func i18nTagParser(i18n *I18n) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, error) {
		node := &i18nNode{args: make([]pongo2.IEvaluator, 0), i18n: i18n}

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

func i18nTranslate(i18n *I18n) func(map[string]any) any {
	return func(vars map[string]any) any {
		lang := ""
		if v, ok := vars["current_lang"]; ok {
			lang = v.(string)
		}
		// {{ i18n("cc") }}
		// {{ i18n("cc %d", 11) }}
		return func(id string, args ...any) string {
			format := i18n.Translate(id, lang)
			if len(args) == 0 {
				return format
			}
			return fmt.Sprintf(format, args...)
		}
	}
}

func init() {
	template.Register("i18n", func(ctx *core.Context, set template.TemplateSet) error {
		i18n, err := New(ctx)
		if err != nil {
			return err
		}

		tag := i18nTagParser(i18n)
		set.RegisterTag("T", tag)
		set.RegisterTag("i18n", tag)

		fn := i18nTranslate(i18n)
		set.RegisterTransient("_", fn)
		set.RegisterTransient("T", fn)
		set.RegisterTransient("i18n", fn)
		return nil
	})
}
