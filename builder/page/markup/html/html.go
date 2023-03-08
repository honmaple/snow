package html

import (
	"bytes"
	"io"
	"strings"

	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/config"
	"golang.org/x/net/html"
)

type htmlReader struct {
	conf config.Config
}

func parseMeta(meta page.Meta, n *html.Node) error {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if n.FirstChild != nil {
				meta["title"] = n.FirstChild.Data
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
				meta.Set(strings.ToLower(key), strings.TrimSpace(val))
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
				meta.Set("custom_css", "["+href+"]")
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
				meta.Set("custom_js", "["+src+"]")
			}
		case "body":
			var buf bytes.Buffer

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := html.Render(&buf, c); err != nil {
					return err
				}
			}
			meta["content"] = strings.TrimSpace(buf.String())
			return nil
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if err := parseMeta(meta, child); err != nil {
			return err
		}
	}
	return nil
}

func readMeta(r io.Reader) (page.Meta, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	meta := make(page.Meta)
	if err := parseMeta(meta, doc); err != nil {
		return nil, err
	}
	return meta, nil
}

func (s *htmlReader) Read(r io.Reader) (page.Meta, error) {
	return readMeta(r)
}

func New(conf config.Config) page.Reader {
	return &htmlReader{conf}
}

func init() {
	page.Register(".html", New)
}
