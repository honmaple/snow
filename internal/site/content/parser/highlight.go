package parser

import (
	"fmt"

	"github.com/alecthomas/chroma/v2/formatters/html"
)

type OnlyPreWrapper struct{}

func (OnlyPreWrapper) Start(code bool, styleAttr string) string {
	return fmt.Sprintf(`<pre%s>`, styleAttr)
}

func (OnlyPreWrapper) End(code bool) string {
	return `</pre>`
}

type HTMLFormatter = html.Formatter

func NewHTMLFormatter(opt MarkupOption) *html.Formatter {
	opts := make([]html.Option, 0)
	if opt.ShowLineNumbers {
		opts = append(opts, html.WithLineNumbers(true))
	}
	if opt.PreventPreCode {
		opts = append(opts, html.WithPreWrapper(OnlyPreWrapper{}))
	}
	return html.New(opts...)
}
