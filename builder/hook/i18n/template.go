package i18n

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/builder/theme/template"
)

type i18nNode struct {
	expr pongo2.IEvaluator
	args []pongo2.IEvaluator
	i18n *i18n
}

func (node *i18nNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	val, err := node.expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	args := make([]interface{}, len(node.args))
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
	format := node.i18n.Tran(val.String(), lang)
	writer.WriteString(fmt.Sprintf(format, args...))
	return nil
}

func (self *i18n) nodeParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &i18nNode{args: make([]pongo2.IEvaluator, 0), i18n: self}

	// {% i18n "aaaaa" %}
	// {% i18n "aaaaa %s %d" "ccc" 108 %}
	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'i18n' requires at least one argument.", nil)
	}

	expr, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	node.expr = expr

	for arguments.Remaining() > 0 {
		expr, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.args = append(node.args, expr)
	}
	return node, nil
}

func (self *i18n) tranFunc(ctx map[string]interface{}) interface{} {
	lang := ""
	if v, ok := ctx["current_lang"]; ok {
		lang = v.(string)
	}
	// {{ i18n("cc") }}
	// {{ i18n("cc %d", 11) }}
	return func(id string, args ...interface{}) string {
		format := self.Tran(id, lang)
		if len(args) == 0 {
			return format
		}
		return fmt.Sprintf(format, args...)
	}
}

func (self *i18n) register() {
	template.RegisterTag("i18n", self.nodeParser)
	template.RegisterTag("T", self.nodeParser)

	template.RegisterFunc("i18n", self.tranFunc)
	template.RegisterFunc("T", self.tranFunc)
	template.RegisterFunc("_", self.tranFunc)
}
