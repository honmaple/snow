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
	orgmodeParser "github.com/honmaple/org-golang/parser"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/template"
)

var (
	ORGMODE_MORE       = regexp.MustCompile(`^(?i:#\+more)\s*$`)
	ORGMODE_KEYWORD    = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
	ORGMODE_PROPERTIES = regexp.MustCompile(`^(?i::PROPERTIES:)$`)
	ORGMODE_META       = regexp.MustCompile(`^:([^:]+):(\s+(.*)|$)`)
)

type Option struct {
	parser.MarkupOption
}

type orgParser struct {
	opt      *Option
	renderer *Renderer
}

func (m *orgParser) parse(data []byte) ([]*parser.Heading, string, error) {
	rd := render.HTML{
		Toc:            false,
		Document:       org.New(bytes.NewBuffer(data)),
		RenderNodeFunc: m.renderer.RenderNode,
	}

	var ch func([]*orgmodeParser.Section) []*parser.Heading

	ch = func(children []*orgmodeParser.Section) []*parser.Heading {
		headings := make([]*parser.Heading, 0)
		for _, child := range children {
			h := &parser.Heading{
				Id:       child.Id(),
				Title:    rd.RenderNodes(child.Title, ""),
				Level:    child.Stars,
				Children: ch(child.Children),
			}
			headings = append(headings, h)
		}
		return headings
	}
	toc := ch(rd.Document.Sections.Children)
	return toc, rd.String(), nil
}

func (m *orgParser) Parse(r io.Reader) (*parser.Result, error) {
	var (
		content   bytes.Buffer
		summary   bytes.Buffer
		isMeta    = true
		isFormat  = true
		isSummary = true
	)

	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}

	scanner := bufio.NewScanner(r)
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

func New(opt *Option) *orgParser {
	return &orgParser{opt: opt, renderer: NewRenderer(opt)}
}

func NewWithContext(ctx *core.Context) *orgParser {
	opt := &Option{
		MarkupOption: parser.NewMarkupOption(ctx, "org"),
	}
	return New(opt)
}

func orgFilter(ctx *core.Context) pongo2.FilterFunction {
	r := NewWithContext(ctx)
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:org",
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
	parser.Register(".org", func(ctx *core.Context) parser.MarkupParser {
		return NewWithContext(ctx)
	})
	template.Register("orgParser", func(ctx *core.Context, set template.TemplateSet) error {
		set.RegisterFilter("org", orgFilter(ctx))
		return nil
	})
}
