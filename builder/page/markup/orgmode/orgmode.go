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
	ORGMODE_MORE = "#+more"
	ORGMODE_META = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
)

type orgmode struct {
	conf config.Config
}

func (m *orgmode) highlightCodeBlock(source, lang string) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("monokai"), it)
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}

func (m *orgmode) Read(r io.Reader) (page.Meta, error) {
	var (
		content    bytes.Buffer
		summary    bytes.Buffer
		summeryEnd = false
		metaEnd    = false
		meta       = make(page.Meta)
		scanner    = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		line := scanner.Text()
		if !metaEnd {
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
		metaEnd = true
		if line == ORGMODE_MORE {
			summeryEnd = true
		}
		if !summeryEnd {
			summary.WriteString(line)
			summary.WriteString("\n")
		}
		content.WriteString(line)
		content.WriteString("\n")
	}
	if summary.Len() == content.Len() {
		meta["summary"] = m.conf.GetSummary(m.HTML(&summary))
	} else {
		meta["summary"] = m.HTML(&summary)
	}
	meta["content"] = m.HTML(&content)
	return meta, nil
}

func (m *orgmode) HTML(r io.Reader) string {
	rd := render.HTML{
		Toc:       false,
		Document:  org.New(r),
		Highlight: m.highlightCodeBlock,
	}
	return rd.String()
}

func New(conf config.Config) page.Reader {
	return &orgmode{conf}
}

func init() {
	page.Register(".org", New)
}
