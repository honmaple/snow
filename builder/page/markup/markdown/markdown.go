package markdown

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/config"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v3"
)

var (
	MAKRDOWN_LINE = "---" // 兼容hugo
	MAKRDOWN_MORE = "<!--more-->"
	MAKRDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

type markdown struct {
	conf config.Config
}

func (m *markdown) Read(r io.Reader) (page.Meta, error) {
	var (
		summary    bytes.Buffer
		content    bytes.Buffer
		summeryEnd = false
		metaEnd    = false
		meta       = make(page.Meta)
		scanner    = bufio.NewScanner(r)
	)
	isYAML := true
	for scanner.Scan() {
		line := scanner.Text()
		if isYAML && line == MAKRDOWN_LINE {
			var b bytes.Buffer
			for scanner.Scan() {
				l := scanner.Text()
				if l == MAKRDOWN_LINE || l == "" {
					break
				}
				b.WriteString(l)
				b.WriteString("\n")
			}
			if err := yaml.Unmarshal(b.Bytes(), &meta); err != nil {
				return nil, err
			}
			meta.Fix()
			isYAML = false
			continue
		}
		if !metaEnd {
			if match := MAKRDOWN_META.FindStringSubmatch(line); match != nil {
				meta.Set(m.conf, strings.ToLower(match[1]), strings.TrimSpace(match[3]))
				continue
			}
		}
		metaEnd = true
		if line == MAKRDOWN_MORE {
			summeryEnd = true
		}
		if !summeryEnd {
			summary.WriteString(line)
			summary.WriteString("\n")
		}
		content.WriteString(line)
		content.WriteString("\n")
	}
	meta["summary"] = m.HTML(summary.Bytes())
	meta["content"] = m.HTML(content.Bytes())
	return meta, nil
}

func (s *markdown) HTML(data []byte) string {
	d := blackfriday.MarkdownCommon(data)
	return string(d)
}

func New(conf config.Config) page.Reader {
	return &markdown{conf}

}

func init() {
	page.Register(".md", New)
}
