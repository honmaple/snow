package orgmode

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/honmaple/org-golang/parser"
	"github.com/honmaple/org-golang/render"
)

type Renderer struct {
	opt       *Option
	formatter *html.Formatter
}

func (e *Renderer) highlightCodeBlock(source, lang string) string {
	var w strings.Builder
	var lexer chroma.Lexer

	if lang == "" || lang == "example" {
		lexer = lexers.Analyse(source)
	} else {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get(e.opt.Style)
	if style == nil {
		style = styles.Fallback
	}

	it, _ := lexer.Tokenise(nil, source)
	e.formatter.Format(&w, style, it)
	return w.String()
}

func (e *Renderer) RenderNode(r render.Renderer, n parser.Node) string {
	if e.opt.Style != "" && e.opt.Style != "none" {
		switch node := n.(type) {
		case *parser.Block:
			if node.Type == "SRC" || node.Type == "EXAMPLE" {
				lang := ""
				if len(node.Parameters) > 0 {
					lang = node.Parameters[0]
				}
				text := render.DedentString(r.RenderNodes(node.Children, "\n"))
				return e.highlightCodeBlock(text, lang)
			}
		}
	}
	return r.RenderNode(n, true)
}

func NewRenderer(opt *Option) *Renderer {
	opts := make([]html.Option, 0)
	if opt.ShowLineNumbers {
		opts = append(opts, html.WithLineNumbers(true))
	}
	r := &Renderer{
		opt:       opt,
		formatter: html.New(opts...),
	}
	return r
}
