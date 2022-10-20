package markdown

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/honmaple/snow/config"
	"github.com/russross/blackfriday"
)

var (
	MAKRDOWN_LINE = "---" // 兼容hugo
	MAKRDOWN_MORE = "<!--more-->"
	MAKRDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

type markdown struct {
	conf config.Config
}

func (s *markdown) Exts() []string {
	return []string{".md"}
}

func (m *markdown) Read(r io.Reader) (map[string]string, error) {
	var (
		summary    bytes.Buffer
		content    bytes.Buffer
		summeryEnd = false
		metaEnd    = false
		meta       = make(map[string]string)
		scanner    = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		line := scanner.Text()
		if !metaEnd {
			if match := MAKRDOWN_META.FindStringSubmatch(line); match != nil {
				meta[strings.ToLower(match[1])] = strings.TrimSpace(match[3])
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

func New(conf config.Config) *markdown {
	return &markdown{conf}

}
