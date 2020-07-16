package markup

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type Markup struct {
}

func highlightCodeBlock(source, lang string, inline bool) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("friendly"), it)
	if inline {
		return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
	}
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}

func (m *Markup) Read(file string) (map[string]string, error) {
	ext := filepath.Ext(file)[1:]
	switch ext {
	case ".org":
		return m.orgmode(file)
	case ".md":
		return m.markdown(file)
	default:
		return nil, errors.New(fmt.Sprintf("no reader for %s: %s", ext, file))
	}
	return nil, nil
}
