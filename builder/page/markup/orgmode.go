package markup

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/honmaple/org-golang"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/utils"
)

var (
	ORGMODE_MORE = "#+more"
	ORGMODE_META = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
)

func orgmodeHTML(data []byte) string {
	r := render.HTML{
		Toc:       false,
		Document:  org.New(bytes.NewReader(data)),
		Highlight: highlightCodeBlock,
	}
	return r.String()
}

func (m *Markup) orgmode(file string) (map[string]string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)

	var (
		summary bytes.Buffer
		meta    = make(map[string]string)
		index   = 0
		scanner = bufio.NewScanner(r)
	)
	limit := m.conf.GetInt("params.orgmode.meta_limit")
	if limit == 0 {
		limit = 10
	}
	for scanner.Scan() {
		line := scanner.Text()
		if index <= limit {
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
			}
		}
		if line == ORGMODE_MORE {
			break
		}
		if _, err := summary.WriteString(line); err != nil {
			return nil, err
		}
		index++
	}
	meta["summary"] = orgmodeHTML(summary.Bytes())
	meta["content"] = orgmodeHTML(data)
	if title, ok := meta["titie"]; !ok || title == "" {
		meta["title"] = utils.FileBaseName(file)
	}
	return meta, nil
}
