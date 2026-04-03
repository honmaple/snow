package html

import (
	"bytes"
	"io"
	"strings"

	"github.com/honmaple/snow/builder/content/parser"
	"github.com/honmaple/snow/config"
	"golang.org/x/net/html"
)

type htmlParser struct {
	conf config.Config
}

func (d *htmlParser) parse(result *parser.Result, n *html.Node) error {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if n.FirstChild != nil {
				result.FrontMatter["title"] = n.FirstChild.Data
			}
		case "meta":
			key, val := "", ""
			for _, a := range n.Attr {
				if a.Key == "name" {
					key = a.Val
				} else if a.Key == "content" {
					val = a.Val
				}
			}
			if key != "" {
				result.SetFrontMatter(strings.ToLower(key), strings.TrimSpace(val))
			}
		case "link":
			href := ""
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
					break
				}
			}
			if href != "" {
				result.SetFrontMatter("custom_css", "["+href+"]")
			}
		case "script":
			src := ""
			for _, a := range n.Attr {
				if a.Key == "src" {
					src = a.Val
					break
				}
			}
			if src != "" {
				result.SetFrontMatter("custom_js", "["+src+"]")
			}
		case "body":
			var buf bytes.Buffer

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := html.Render(&buf, c); err != nil {
					return err
				}
			}
			result.RawContent = strings.TrimSpace(buf.String())
			return nil
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if err := d.parse(result, child); err != nil {
			return err
		}
	}
	return nil
}

func (d *htmlParser) Parse(r io.Reader) (*parser.Result, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}
	if err := d.parse(result, doc); err != nil {
		return nil, err
	}
	result.Content = result.RawContent
	return result, nil
}

func New(conf config.Config) parser.MarkupParser {
	return &htmlParser{conf}
}

func init() {
	parser.Register(".html", New)
}
