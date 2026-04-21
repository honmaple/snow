package markdown

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/russross/blackfriday/v2"
	"github.com/spf13/viper"
)

var (
	// 兼容hugo
	MARKDOWN_LINE = regexp.MustCompile(`^[-|\+]{3}\s*$`)
	MARKDOWN_MORE = regexp.MustCompile(`^\s*(?i:<!--more-->)\s*$`)
	MARKDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

type mdParser struct {
	opts []blackfriday.Option
}

func readMeta(r io.Reader, content *bytes.Buffer, summary *bytes.Buffer, result *parser.Result) error {
	var (
		isMeta    = true
		isFormat  = true
		isSummery = true
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
				return err
			}
			// 不要直接使用meta反序列化数据, 否则子元素map类型也会是page.Meta
			result.FrontMatter = cf.AllSettings()
			isFormat = false
			continue
		}
		isFormat = false

		if isMeta {
			if match := MARKDOWN_META.FindStringSubmatch(line); match != nil {
				result.SetFrontMatter(strings.ToLower(match[1]), strings.TrimSpace(match[3]))
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
	return nil
}

func (m *mdParser) Parse(r io.Reader) (*parser.Result, error) {
	var (
		summary bytes.Buffer
		content bytes.Buffer
	)
	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}
	if err := readMeta(r, &content, &summary, result); err != nil {
		return nil, err
	}
	if summary.Len() > 0 {
		result.Summary = m.HTML(summary.Bytes())
	}
	result.Content = m.HTML(content.Bytes())
	result.RawContent = content.String()
	return result, nil
}

func (m *mdParser) HTML(data []byte) string {
	d := blackfriday.Run(data, m.opts...)
	return string(d)
}

func New(ctx *core.Context) *mdParser {
	return &mdParser{
		opts: []blackfriday.Option{
			blackfriday.WithRenderer(NewChromaRenderer(ctx.GetHighlightStyle())),
		},
	}
}

func markdownFilter(ctx *core.Context) pongo2.FilterFunction {
	r := New(ctx)
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:markdown",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
		return pongo2.AsValue(r.HTML([]byte(v))), nil
	}
}

func init() {
	parser.Register(".md", func(ctx *core.Context) parser.MarkupParser {
		return New(ctx)
	})
	template.RegisterContextFilter("markdown", markdownFilter)
}
