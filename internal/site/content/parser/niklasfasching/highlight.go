package niklasfasching

import (
	"fmt"
	"html"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	contentparser "github.com/honmaple/snow/internal/site/content/parser"
)

type Renderer struct {
	opt       *Option
	formatter *contentparser.HTMLFormatter
}

func (r *Renderer) highlightCodeBlock(source, lang string, inline bool, params map[string]string) string {
	if r.opt.Style == "" || r.opt.Style == "none" {
		return fmt.Sprintf("<pre>\n%s\n</pre>", html.EscapeString(source))
	}

	var w strings.Builder
	var lexer chroma.Lexer

	if lang == "" || lang == "text" || lang == "example" {
		lexer = lexers.Analyse(source)
	} else {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get(r.opt.Style)
	if style == nil {
		style = styles.Fallback
	}

	it, _ := lexer.Tokenise(nil, source)
	r.formatter.Format(&w, style, it)
	return w.String()
}

func NewRenderer(opt *Option) *Renderer {
	return &Renderer{
		opt:       opt,
		formatter: contentparser.NewHTMLFormatter(opt.MarkupOption),
	}
}
