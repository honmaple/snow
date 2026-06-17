package markdown

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkParser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	// 兼容hugo
	MARKDOWN_LINE = regexp.MustCompile(`^[-|\+]{3}\s*$`)
	MARKDOWN_MORE = regexp.MustCompile(`^\s*(?i:<!--more-->)\s*$`)
	MARKDOWN_META = regexp.MustCompile(`^([^:]+):(\s+(.*)|$)`)
)

const scannerMaxTokenSize = 1024 * 1024

type (
	Option struct {
		parser.MarkupOption
	}
	Heading = parser.Heading
)

type mdParser struct {
	md  goldmark.Markdown
	opt *Option
}

func (m *mdParser) parse(data []byte) ([]*parser.Heading, string, error) {
	var buf bytes.Buffer

	ctx := goldmarkParser.NewContext()
	doc := m.md.Parser().Parse(text.NewReader(data), goldmarkParser.WithContext(ctx))
	if err := m.md.Renderer().Render(&buf, data, doc); err != nil {
		return nil, "", err
	}
	if toc, ok := ctx.Get(tocKey).([]*parser.Heading); ok {
		return toc, buf.String(), nil
	}
	return nil, buf.String(), nil
}

func (m *mdParser) Parse(r io.Reader) (*parser.Result, error) {
	var (
		summary   bytes.Buffer
		content   bytes.Buffer
		isMeta    = true
		isFormat  = true
		isSummery = true
	)

	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), scannerMaxTokenSize)
	for scanner.Scan() {
		line := scanner.Text()
		if isFormat && MARKDOWN_LINE.MatchString(line) {
			var b bytes.Buffer
			closed := false
			fence := strings.TrimSpace(line)
			for scanner.Scan() {
				l := scanner.Text()
				if strings.TrimSpace(l) == fence {
					closed = true
					break
				}
				b.WriteString(l)
				b.WriteString("\n")
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("markdown parser scan: %w", err)
			}
			if !closed {
				return nil, fmt.Errorf("markdown front matter is not closed: %s", fence)
			}

			cf := viper.New()
			if fence == "---" {
				cf.SetConfigType("yaml")
			} else {
				cf.SetConfigType("toml")
			}

			if err := cf.ReadConfig(&b); err != nil {
				return nil, err
			}
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
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("markdown parser scan: %w", err)
	}

	toc, res, err := m.parse(content.Bytes())
	if err != nil {
		return nil, err
	}
	result.Toc = toc
	result.Content = res
	result.RawSummary = summary.String()
	result.RawContent = content.String()

	if summary.Len() > 0 {
		_, res, err := m.parse(summary.Bytes())
		if err != nil {
			return nil, err
		}
		result.Summary = res
	}
	return result, nil
}

func (m *mdParser) SupportedExtensions() []string {
	return []string{".md"}
}

func New(opt *Option) *mdParser {
	exts := []goldmark.Extender{
		extension.GFM,
	}
	if opt.Style != "" && opt.Style != "none" {
		exts = append(exts, NewHighlightExtension(opt))
	}
	if opt.ShowToc {
		exts = append(exts, NewTocExtension(opt))
	}
	md := goldmark.New(goldmark.WithExtensions(exts...))
	return &mdParser{md: md, opt: opt}
}

func NewWithContext(ctx *core.Context) *mdParser {
	opt := &Option{
		MarkupOption: parser.NewMarkupOption(ctx, "markdown"),
	}
	return New(opt)
}

func markdownFilter(ctx *core.Context) pongo2.FilterFunction {
	r := NewWithContext(ctx)
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:markdown",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
		_, res, err := r.parse([]byte(v))
		if err != nil {
			return nil, err
		}
		return pongo2.AsValue(res), nil
	}
}

func init() {
	parser.Register("markdown", func(ctx *core.Context) parser.MarkupParser {
		return NewWithContext(ctx)
	})

	template.Register("markdown", func(ctx *core.Context, set template.TemplateSet) error {
		set.RegisterFilter("markdown", markdownFilter(ctx))
		return nil
	})
}
