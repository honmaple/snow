package orgmode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/org-golang"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/theme/template"
)

var (
	ORGMODE_MORE       = regexp.MustCompile(`^(?i:#\+more)\s*$`)
	ORGMODE_KEYWORD    = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
	ORGMODE_PROPERTIES = regexp.MustCompile(`^(?i::PROPERTIES:)$`)
	ORGMODE_META       = regexp.MustCompile(`^:([^:]+):(\s+(.*)|$)`)
)

type orgParser struct {
	ctx *core.Context
}

func readMeta(r io.Reader, content *bytes.Buffer, summary *bytes.Buffer, result *parser.Result) error {
	var (
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
				result.SetFrontMatter(match[1], match[2])
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
					result.SetFrontMatter(k, v)
				} else {
					result.SetFrontMatter(strings.ToLower(match[1]), strings.TrimSpace(match[3]))
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
	return nil
}

func (m *orgParser) Parse(r io.Reader) (*parser.Result, error) {
	var (
		content bytes.Buffer
		summary bytes.Buffer
	)
	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}
	if err := readMeta(r, &content, &summary, result); err != nil {
		return nil, err
	}
	result.RawContent = content.String()

	if summary.Len() > 0 {
		result.Summary = m.HTML(summary.Bytes(), false)
	}
	result.Content = m.HTML(content.Bytes(), true)
	return result, nil
}

func (m *orgParser) HTML(data []byte, showToc bool) string {
	rd := render.HTML{
		Toc:            showToc,
		Document:       org.New(bytes.NewBuffer(data)),
		RenderNodeFunc: m.renderNode,
	}
	return rd.String()
}

func New(ctx *core.Context) *orgParser {
	return &orgParser{ctx}
}

func orgFilter(ctx *core.Context) pongo2.FilterFunction {
	r := New(ctx)
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:org",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
		return pongo2.AsValue(r.HTML([]byte(v), false)), nil
	}
}

func init() {
	parser.Register(".org", func(ctx *core.Context) parser.MarkupParser {
		return New(ctx)
	})
	template.RegisterContextFilter("org", orgFilter)
}
