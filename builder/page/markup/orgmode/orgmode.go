package orgmode

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

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

func (m *orgmode) Read(r io.Reader) (page.Meta, error) {
	var (
		content   bytes.Buffer
		summary   bytes.Buffer
		meta      = make(page.Meta)
		scanner   = bufio.NewScanner(r)
		isMeta    = true
		isSummary = true
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
	buf := content.Bytes()
	if summary.Len() == 0 {
		meta["summary"] = m.HTML(buf, false, true)
	} else {
		meta["summary"] = m.HTML(summary.Bytes(), false, false)
	}
	meta["content"] = m.HTML(buf, true, false)
	return meta, nil
}

func (m *orgmode) HTML(data []byte, showToc bool, summary bool) string {
	rd := render.HTML{
		Toc:            showToc,
		Document:       org.New(bytes.NewBuffer(data)),
		RenderNodeFunc: m.renderNode,
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
