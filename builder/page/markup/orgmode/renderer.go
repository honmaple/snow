package orgmode

import (
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/honmaple/org-golang/parser"
	"github.com/honmaple/org-golang/render"
)

func (m *orgmode) highlightCodeBlock(source, lang string) string {
	theme := m.conf.GetHighlightStyle()

	var w strings.Builder
	var lexer chroma.Lexer

	if lang != "" {
		lexer = lexers.Get(lang)
	} else {
		lexer = lexers.Analyse(source)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}

	it, _ := lexer.Tokenise(nil, source)
	_ = html.New().Format(&w, style, it)
	return w.String()
}

func (m *orgmode) renderNode(r render.Renderer, n parser.Node) string {
	switch node := n.(type) {
	case *parser.Block:
		if node.Type == "SRC" || node.Type == "EXAMPLE" {
			lang := ""
			if len(node.Parameters) > 0 {
				lang = node.Parameters[0]
			}
			text := render.DedentString(r.RenderNodes(node.Children, "\n"))
			return m.highlightCodeBlock(text, lang)
		}
	}
	return r.RenderNode(n, true)
}
