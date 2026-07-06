package orgmode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/honmaple/org-golang"
	orgmodeParser "github.com/honmaple/org-golang/parser"
	"github.com/honmaple/org-golang/render"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
)

var (
	ORGMODE_MORE       = regexp.MustCompile(`^(?i:#\+more)\s*$`)
	ORGMODE_KEYWORD    = regexp.MustCompile(`^#\+([^:]+):(\s+(.*)|$)`)
	ORGMODE_PROPERTIES = regexp.MustCompile(`^(?i::PROPERTIES:)$`)
	ORGMODE_META       = regexp.MustCompile(`^:([^:]+):(\s+(.*)|$)`)
)

const scannerMaxTokenSize = 1024 * 1024

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
	scanner.Buffer(make([]byte, 1024), scannerMaxTokenSize)
	for scanner.Scan() {
		line := scanner.Text()
		if isFormat && ORGMODE_PROPERTIES.MatchString(line) {
			closed := false
			var drawerErr error
			for scanner.Scan() {
				l := scanner.Text()
				if strings.EqualFold(strings.TrimSpace(l), ":END:") {
					closed = true
					break
				}
				match := ORGMODE_META.FindStringSubmatch(l)
				if match == nil {
					if drawerErr == nil {
						drawerErr = fmt.Errorf("org parser invalid properties line: %s", l)
					}
					continue
				}
				result.SetFrontMatter(match[1], match[2])
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("org parser scan: %w", err)
			}
			if !closed {
				return nil, errors.New("org properties drawer is not closed")
			}
			if drawerErr != nil {
				return nil, drawerErr
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
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("org parser scan: %w", err)
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

func (m *orgParser) SupportedExtensions() []string {
	return []string{".org"}
}

func New(opt *Option) *orgParser {
	return &orgParser{opt: opt, renderer: NewRenderer(opt)}
}

func NewWithContext(ctx *core.Context) *orgParser {
	opt := &Option{
		MarkupOption: parser.NewMarkupOption(ctx, "orgmode"),
	}
	return New(opt)
}

func init() {
	parser.Register("orgmode", func(ctx *core.Context) parser.MarkupParser {
		return NewWithContext(ctx)
	})
}
