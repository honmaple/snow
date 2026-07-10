package orgmode

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/honmaple/org-golang/parser"
	"github.com/honmaple/org-golang/render"
	contentparser "github.com/honmaple/snow/internal/site/content/parser"
)

type Renderer struct {
	opt       *Option
	formatter *contentparser.HTMLFormatter
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

func (e *Renderer) RenderKeyword(r render.Renderer, n *parser.Keyword) string {
	if strings.ToUpper(n.Key) == "HTML" {
		return n.Value
	}
	return r.RenderKeyword(n)
}

func (e *Renderer) RenderBlock(r render.Renderer, n *parser.Block) string {
	switch n.Type {
	case "SRC", "EXAMPLE":
		if e.opt.Style != "" && e.opt.Style != "none" {
			lang := ""
			if len(n.Parameters) > 0 {
				lang = n.Parameters[0]
			}
			text := render.DedentString(r.RenderNodes(n.Children, "\n"))
			return e.highlightCodeBlock(text, lang)
		}
	case "SHORTCODE":
		if len(n.Parameters) > 0 {
			return fmt.Sprintf("<shortcode %[1]s>\n%[2]s\n</shortcode>", strings.Join(n.Parameters, " "), r.RenderNodes(n.Children, "\n"))
		}
	}
	return r.RenderBlock(n)
}

func (e *Renderer) RenderNode(r render.Renderer, n parser.Node) string {
	switch node := n.(type) {
	case *parser.Block:
		return e.RenderBlock(r, node)
	case *parser.Keyword:
		return e.RenderKeyword(r, node)
	default:
		return r.RenderNode(n, true)
	}
}

func NewRenderer(opt *Option) *Renderer {
	r := &Renderer{
		opt:       opt,
		formatter: contentparser.NewHTMLFormatter(opt.MarkupOption),
	}
	return r
}
