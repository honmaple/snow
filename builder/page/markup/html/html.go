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

func (m *htmlReader) parse(meta page.Meta, n *html.Node) error {
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
				meta.Set(m.conf, strings.ToLower(key), val)
			}
		case "body":
			var buf bytes.Buffer
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := html.Render(&buf, c); err == nil {
					return err
				}
			}
			meta["content"] = buf.String()
			return nil
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if err := m.parse(meta, child); err != nil {
			return err
		}
	}
	return nil
}

func (s *htmlReader) Read(r io.Reader) (page.Meta, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	meta := make(page.Meta)
	if err := s.parse(meta, doc); err != nil {
		return nil, err
	}
	return meta, nil
}

func New(conf config.Config) page.Reader {
	return &htmlReader{conf}
}

func init() {
	page.Register(".html", New)
}
