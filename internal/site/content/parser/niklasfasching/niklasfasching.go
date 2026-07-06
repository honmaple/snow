package niklasfasching

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	org "github.com/niklasfasching/go-org/org"

	_ "github.com/honmaple/snow/internal/site/content/parser/orgmode"
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

func (p *orgParser) parse(data []byte) ([]*parser.Heading, string, error) {
	conf := org.New().Silent()
	conf.DefaultSettings["OPTIONS"] = "toc:nil title:nil <:t e:t f:t pri:t todo:t tags:t ealb:nil"

	doc := conf.Parse(bytes.NewReader(data), ".")
	if doc.Error != nil {
		return nil, "", doc.Error
	}

	writer := org.NewHTMLWriter()
	writer.HighlightCodeBlock = p.renderer.highlightCodeBlock
	toc := p.toc(doc, writer)

	out, err := doc.Write(writer)
	if err != nil {
		return nil, "", err
	}
	return toc, out, nil
}

func (p *orgParser) toc(doc *org.Document, writer *org.HTMLWriter) []*parser.Heading {
	var walk func([]*org.Section) []*parser.Heading
	walk = func(sections []*org.Section) []*parser.Heading {
		headings := make([]*parser.Heading, 0, len(sections))
		for _, section := range sections {
			if section.Headline == nil || section.Headline.IsExcluded(doc) {
				continue
			}
			heading := &parser.Heading{
				Id:       section.Headline.ID(),
				Level:    section.Headline.Lvl,
				Title:    writer.WriteNodesAsString(section.Headline.Title...),
				Children: walk(section.Children),
			}
			headings = append(headings, heading)
		}
		return headings
	}
	return walk(doc.Outline.Children)
}

func (p *orgParser) Parse(r io.Reader) (*parser.Result, error) {
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
						drawerErr = fmt.Errorf("niklasfasching org parser invalid properties line: %s", l)
					}
					continue
				}
				result.SetFrontMatter(match[1], match[2])
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("niklasfasching org parser scan: %w", err)
			}
			if !closed {
				return nil, errors.New("niklasfasching org properties drawer is not closed")
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
		return nil, fmt.Errorf("niklasfasching org parser scan: %w", err)
	}

	toc, res, err := p.parse(content.Bytes())
	if err != nil {
		return nil, err
	}
	result.Toc = toc
	result.Content = res
	result.RawSummary = summary.String()
	result.RawContent = content.String()

	if summary.Len() > 0 {
		_, res, err := p.parse(summary.Bytes())
		if err != nil {
			return nil, err
		}
		result.Summary = res
	}
	return result, nil
}

func (p *orgParser) SupportedExtensions() []string {
	return []string{".org"}
}

func New(opt *Option) *orgParser {
	return &orgParser{opt: opt, renderer: NewRenderer(opt)}
}

func NewWithContext(ctx *core.Context) *orgParser {
	opt := &Option{
		MarkupOption: parser.NewMarkupOption(ctx, "niklasfasching"),
	}
	return New(opt)
}

func init() {
	parser.Register("niklasfasching", func(ctx *core.Context) parser.MarkupParser {
		return NewWithContext(ctx)
	})
}
