package markup

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/russross/blackfriday"
)

var (
	MAKRDOWN_MORE = "<!--more-->"
	MAKRDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

func markdownHTML(data []byte) string {
	d := blackfriday.MarkdownCommon(data)
	return string(d)
}

func (m *Markup) markdown(file string) (map[string]string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var (
		summary bytes.Buffer
		meta    = make(map[string]string)
		index   = 0
		scanner = bufio.NewScanner(bytes.NewReader(data))
	)

	limit := m.conf.GetInt("params.markdown.meta_limit")
	if limit == 0 {
		limit = 10
	}
	for scanner.Scan() {
		line := scanner.Text()
		if index <= limit {
			if match := MAKRDOWN_META.FindStringSubmatch(line); match != nil {
				meta[strings.ToLower(match[1])] = strings.TrimSpace(match[3])
			}
		}
		if line == MAKRDOWN_MORE {
			break
		}
		if _, err := summary.WriteString(line); err != nil {
			return nil, err
		}
		index++
	}
	meta["summary"] = markdownHTML(summary.Bytes())
	meta["content"] = markdownHTML(data)

	if title, ok := meta["titie"]; !ok || title == "" {
		meta["title"] = utils.FileBaseName(file)
	}
	return meta, nil
}
