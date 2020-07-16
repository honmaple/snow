package markup

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/niklasfasching/go-org/org"
)

var (
	orgProperty    = []byte("#+")
	orgMore        = []byte("#+more")
	KEYWORD_REGEXP = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
)

func (m *Markup) html(file string, content []byte) string {
	d := org.New().Parse(bytes.NewReader(content), file)
	writer := org.NewHTMLWriter()
	writer.HighlightCodeBlock = highlightCodeBlock
	out, err := d.Write(writer)
	if err != nil {
		return ""
	}
	return out
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
	)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		if match := KEYWORD_REGEXP.FindStringSubmatch(string(b)); match != nil {
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
		if bytes.Equal(b, orgMore) {
			break
		}
		if _, err := summary.Write(b); err != nil {
			return nil, err
		}
	}
	meta["summary"] = summary.String()
	return meta, nil

	meta["summary"] = m.html(file, []byte(meta["summary"]))
	meta["content"] = m.html(file, data)
	if title, ok := meta["titie"]; !ok || title == "" {
		meta["title"] = utils.FileBaseName(file)
	}
	return meta, nil
}
