package orgmode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/org-golang"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
)

var (
	ORGMODE_MORE       = regexp.MustCompile(`^(?i:#\+more)\s*$`)
	ORGMODE_KEYWORD    = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
	ORGMODE_PROPERTIES = regexp.MustCompile(`^(?i::PROPERTIES:)$`)
	ORGMODE_META       = regexp.MustCompile(`^:([^:]+):(\s+(.*)|$)`)
)

type orgmode struct {
	conf config.Config
}

func readMeta(r io.Reader, content *bytes.Buffer, summary *bytes.Buffer) (page.Meta, error) {
	var (
		meta      = make(page.Meta)
		scanner   = bufio.NewScanner(r)
		isMeta    = true
		isFormat  = true
		isSummary = true
	)
	for scanner.Scan() {
		line := scanner.Text()
		if isFormat && ORGMODE_PROPERTIES.MatchString(line) {
			for scanner.Scan() {
				l := scanner.Text()
				if strings.TrimSpace(l) == "" {
					break
				}
				match := ORGMODE_META.FindStringSubmatch(l)
				if match == nil || match[1] == "END" {
					break
				}
				meta.Set(match[1], match[2])
			}
			isFormat = false
			continue
		}
		if isMeta {
			if match := ORGMODE_KEYWORD.FindStringSubmatch(line); match != nil {
				if match[1] == "PROPERTY" {
					s := strings.SplitN(match[3], " ", 2)
					k := strings.ToLower(s[0])
					v := ""
					if len(s) > 1 {
						v = strings.TrimSpace(s[1])
					}
					meta.Set(k, v)
				} else {
					meta.Set(strings.ToLower(match[1]), strings.TrimSpace(match[3]))
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
	return meta, nil
}

func (m *orgmode) Read(r io.Reader) (page.Meta, error) {
	var (
		content bytes.Buffer
		summary bytes.Buffer
	)
	meta, err := readMeta(r, &content, &summary)
	if err != nil {
		return nil, err
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

func NewPongo2Filter(conf config.Config) pongo2.FilterFunction {
	r := &orgmode{conf}
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:org",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
		return pongo2.AsValue(r.HTML([]byte(v), false, false)), nil
	}
}

func init() {
	page.Register(".org", New)
	template.RegisterConfigFilter("org", NewPongo2Filter)
}
