package orgmode

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/honmaple/org-golang"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/config"
)

var (
	ORGMODE_MORE = regexp.MustCompile(`^(?i:#\+more)\s*$`)
	ORGMODE_META = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
)

type orgmode struct {
	conf config.Config
}

func (m *orgmode) highlightCodeBlock(source, lang string) string {
	theme := m.conf.GetHighlightStyle()
	if theme == "" {
		return source
	}

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

	// lexer = chroma.Coalesce(lexer)
	it, _ := lexer.Tokenise(nil, source)
	_ = html.New().Format(&w, style, it)
	return w.String()
	// return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}

func (m *orgmode) Read(r io.Reader) (page.Meta, error) {
	var (
		content   bytes.Buffer
		summary   bytes.Buffer
		meta      = make(page.Meta)
		scanner   = bufio.NewScanner(r)
		isMeta    = false
		isSummary = false
	)
	for scanner.Scan() {
		line := scanner.Text()
		if isMeta {
			if match := ORGMODE_META.FindStringSubmatch(line); match != nil {
				if match[1] == "PROPERTY" {
					s := strings.SplitN(match[3], " ", 2)
					k := strings.ToLower(s[0])
					v := ""
					if len(s) > 1 {
						v = strings.TrimSpace(s[1])
					}
					meta.Set(m.conf, k, v)
				} else {
					meta.Set(m.conf, strings.ToLower(match[1]), strings.TrimSpace(match[3]))
				}
				continue
			}
		}
		isMeta = false
		if isSummary && ORGMODE_MORE.MatchString(line) {
			summary.WriteString(content.String())
			isSummary = false
		}
		content.WriteString(line)
		content.WriteString("\n")
	}
	if summary.Len() == 0 {
		meta["summary"] = m.HTML(&content, true)
	} else {
		meta["summary"] = m.HTML(&summary, true)
	}
	meta["content"] = m.HTML(&content, false)
	return meta, nil
}

func (m *orgmode) HTML(r io.Reader, summary bool) string {
	rd := render.HTML{
		Toc:       !summary,
		Document:  org.New(r),
		Highlight: m.highlightCodeBlock,
	}
	if summary {
		return m.conf.GetSummary(rd.String())
	}
	return rd.String()
}

func New(conf config.Config) page.Reader {
	return &orgmode{conf}
}

func init() {
	page.Register(".org", New)
}
