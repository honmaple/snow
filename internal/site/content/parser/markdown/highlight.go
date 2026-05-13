package markdown

import (
	"bytes"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type highlightExtension struct {
	opt       *Option
	formatter *html.Formatter
}

func (r *highlightExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(r, 200),
	))
}

func (r *highlightExtension) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *highlightExtension) render(w util.BufWriter, lang string, source string) error {
	var lexer chroma.Lexer

	if lang != "" {
		lexer = lexers.Get(lang)
	} else {
		lexer = lexers.Analyse(source)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get(r.opt.Style)
	if style == nil {
		style = styles.Fallback
	}

	it, _ := lexer.Tokenise(nil, source)
	return r.formatter.Format(w, style, it)
}

func (r *highlightExtension) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.CodeBlock)

	var buf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(source))
	}
	return ast.WalkContinue, r.render(w, "", buf.String())
}

func (r *highlightExtension) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)

	var buf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(source))
	}
	lang := string(n.Language(source))
	return ast.WalkContinue, r.render(w, lang, buf.String())
}

func NewHighlightExtension(opt *Option) goldmark.Extender {
	opts := make([]html.Option, 0)
	if opt.ShowLineNumbers {
		opts = append(opts, html.WithLineNumbers(true))
	}
	r := &highlightExtension{
		opt:       opt,
		formatter: html.New(opts...),
	}
	return r
}
