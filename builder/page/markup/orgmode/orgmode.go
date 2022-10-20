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
	"github.com/honmaple/snow/config"
)

var (
	ORGMODE_MORE = "#+more"
	ORGMODE_META = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
)

func highlightCodeBlock(source, lang string) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("friendly"), it)
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}

type orgmode struct {
	conf config.Config
}

func (m *orgmode) Exts() []string {
	return []string{".org"}
}

func (m *orgmode) Read(r io.Reader) (map[string]string, error) {
	var (
		content    bytes.Buffer
		summary    bytes.Buffer
		summeryEnd = false
		metaEnd    = false
		meta       = make(map[string]string)
		scanner    = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		line := scanner.Text()
		if !metaEnd {
			if match := ORGMODE_META.FindStringSubmatch(line); match != nil {
				if match[1] == "PROPERTY" {
					m := strings.SplitN(match[3], " ", 2)
					k := strings.ToLower(m[0])
					v := ""
					if len(m) > 1 {
						v = strings.TrimSpace(m[1])
					}
					meta[k] = v
				} else {
					meta[strings.ToLower(match[1])] = strings.TrimSpace(match[3])
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
	meta["summary"] = m.HTML(&summary)
	meta["content"] = m.HTML(&content)
	return meta, nil
}

func (m *orgmode) HTML(r io.Reader) string {
	rd := render.HTML{
		Toc:      false,
		Document: org.New(r),
		// Highlight: highlightCodeBlock,
	}
	return rd.String()
}

func New(conf config.Config) *orgmode {
	return &orgmode{conf}
}
