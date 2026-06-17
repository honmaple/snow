package html

import (
	"io"
	"strconv"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"golang.org/x/net/html"
)

type htmlParser struct {
	opt parser.MarkupOption
}

func (d *htmlParser) parseHead(result *parser.Result, head *html.Node) {
	walk(head, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return true
		}

		switch n.Data {
		case "title":
			if title := strings.TrimSpace(nodeText(n)); title != "" {
				result.FrontMatter["title"] = title
			}
		case "meta":
			key := firstAttr(n, "name", "property", "itemprop")
			if key != "" {
				result.SetFrontMatter(strings.ReplaceAll(key, ":", "."), attr(n, "content"))
			}
		case "link":
			href := attr(n, "href")
			rel := attr(n, "rel")
			if href != "" && (rel == "" || hasRel(rel, "stylesheet")) {
				result.SetFrontMatter("links", "["+href+"]")
			}
		case "script":
			if src := attr(n, "src"); src != "" {
				result.SetFrontMatter("scripts", "["+src+"]")
			}
		}
		return true
	})
}

func (d *htmlParser) parseBody(result *parser.Result, body *html.Node) error {
	content, summary, hasSummary, err := renderContent(body)
	if err != nil {
		return err
	}

	result.RawContent = content
	result.Content = content
	if hasSummary {
		result.RawSummary = summary
		result.Summary = summary
	}
	if d.opt.ShowToc {
		result.Toc = d.toc(body)
		result.RawContent, err = renderChildren(body)
		if err != nil {
			return err
		}
		result.Content = result.RawContent
	}
	return nil
}

func (d *htmlParser) toc(body *html.Node) []*parser.Heading {
	var (
		toc      []*parser.Heading
		stack    []*parser.Heading
		counters []int
	)
	walk(body, func(n *html.Node) bool {
		if n.Type != html.ElementNode || !isHeading(n.Data) {
			return true
		}

		level, _ := strconv.Atoi(strings.TrimPrefix(n.Data, "h"))
		for len(stack) > 0 && stack[len(stack)-1].Level >= level {
			stack = stack[:len(stack)-1]
		}
		if len(counters) > len(stack)+1 {
			counters = counters[:len(stack)+1]
		}
		if len(counters) < len(stack)+1 {
			counters = append(counters, 1)
		} else {
			counters[len(stack)]++
		}

		id := attr(n, "id")
		if id == "" {
			parts := make([]string, 0, len(counters))
			for _, c := range counters {
				parts = append(parts, strconv.Itoa(c))
			}
			id = "heading-" + strings.Join(parts, ".")
			setAttr(n, "id", id)
		}

		heading := &parser.Heading{
			Id:       id,
			Title:    strings.TrimSpace(nodeText(n)),
			Level:    level,
			Children: make([]*parser.Heading, 0),
		}
		if len(stack) == 0 {
			toc = append(toc, heading)
		} else {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, heading)
		}
		stack = append(stack, heading)
		return false
	})
	return toc
}

func (d *htmlParser) Parse(r io.Reader) (*parser.Result, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}
	if head := findElement(doc, "head"); head != nil {
		d.parseHead(result, head)
	}
	body := findElement(doc, "body")
	if body == nil {
		body = doc
	}
	if err := d.parseBody(result, body); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *htmlParser) SupportedExtensions() []string {
	return []string{".html"}
}

func New(ctx *core.Context) parser.MarkupParser {
	return &htmlParser{
		opt: parser.NewMarkupOption(ctx, "html"),
	}
}

func init() {
	parser.Register("html", New)
}
