package markdown

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
	"github.com/russross/blackfriday/v2"
	"github.com/spf13/viper"
)

var (
	// 兼容hugo
	MARKDOWN_LINE = regexp.MustCompile(`^[-|\+]{3}\s*$`)
	MARKDOWN_MORE = regexp.MustCompile(`^\s*(?i:<!--more-->)\s*$`)
	MARKDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

type markdown struct {
	conf config.Config
	opts []blackfriday.Option
}

func readMeta(r io.Reader, content *bytes.Buffer, summary *bytes.Buffer) (page.Meta, error) {
	var (
		isMeta    = true
		isFormat  = true
		isSummery = true
		meta      = make(page.Meta)
		scanner   = bufio.NewScanner(r)
	)

	for scanner.Scan() {
		line := scanner.Text()
		if isFormat && MARKDOWN_LINE.MatchString(line) {
			var b bytes.Buffer
			for scanner.Scan() {
				l := scanner.Text()
				if strings.TrimSpace(l) == "" || MARKDOWN_LINE.MatchString(l) {
					break
				}
				b.WriteString(l)
				b.WriteString("\n")
			}
			cf := viper.New()
			if line == "---" {
				cf.SetConfigType("yaml")
			} else {
				cf.SetConfigType("toml")
			}

			if err := cf.ReadConfig(&b); err != nil {
				return nil, err
			}
			// 不要直接使用meta反序列化数据, 否则子元素map类型也会是page.Meta
			meta = page.Meta(cf.AllSettings())
			isFormat = false
			continue
		}
		isFormat = false

		if isMeta {
			if match := MARKDOWN_META.FindStringSubmatch(line); match != nil {
				meta.Set(strings.ToLower(match[1]), strings.TrimSpace(match[3]))
				continue
			}
		}
		isMeta = false
		if isSummery && MARKDOWN_MORE.MatchString(line) {
			summary.WriteString(content.String())
			isSummery = false
		}
		content.WriteString(line)
		content.WriteString("\n")
	}
	return meta, nil
}

func (m *markdown) Read(r io.Reader) (page.Meta, error) {
	var (
		summary bytes.Buffer
		content bytes.Buffer
	)
	meta, err := readMeta(r, &content, &summary)
	if err != nil {
		return nil, err
	}
	buf := content.Bytes()
	if summary.Len() == 0 {
		meta["summary"] = m.HTML(buf, true)
	} else {
		meta["summary"] = m.HTML(summary.Bytes(), false)
	}
	meta["content"] = m.HTML(buf, false)
	return meta, nil
}

func (m *markdown) HTML(data []byte, summary bool) string {
	d := blackfriday.Run(data, m.opts...)
	if summary {
		return m.conf.GetSummary(string(d))
	}
	return string(d)
}

func New(conf config.Config) page.Reader {
	return &markdown{conf, []blackfriday.Option{
		blackfriday.WithRenderer(NewChromaRenderer(conf.GetHighlightStyle())),
	}}
}

func NewPongo2Filter(conf config.Config) pongo2.FilterFunction {
	r := &markdown{conf, []blackfriday.Option{
		blackfriday.WithRenderer(NewChromaRenderer(conf.GetHighlightStyle())),
	}}
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:markdown",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
		return pongo2.AsValue(r.HTML([]byte(v), false)), nil
	}
}

func init() {
	page.Register(".md", New)
	template.RegisterConfigFilter("markdown", NewPongo2Filter)
}
