package html

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func firstAttr(n *html.Node, keys ...string) string {
	for _, key := range keys {
		if val := attr(n, key); val != "" {
			return val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func hasRel(rel, name string) bool {
	for _, part := range strings.Fields(rel) {
		if strings.EqualFold(part, name) {
			return true
		}
	}
	return false
}

func isHeading(tag string) bool {
	return len(tag) == 2 && tag[0] == 'h' && tag[1] >= '1' && tag[1] <= '6'
}

func isMoreMarker(n *html.Node) bool {
	return n.Type == html.CommentNode && strings.EqualFold(strings.TrimSpace(n.Data), "more")
}

func nodeText(n *html.Node) string {
	var buf strings.Builder
	walk(n, func(child *html.Node) bool {
		if child.Type == html.TextNode {
			buf.WriteString(child.Data)
		}
		return true
	})
	return buf.String()
}

func findElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func walk(n *html.Node, fn func(*html.Node) bool) {
	if n == nil {
		return
	}
	if !fn(n) {
		return
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		walk(child, fn)
	}
}

func renderContent(n *html.Node) (string, string, bool, error) {
	var content, summary bytes.Buffer
	hasSummary := false

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if isMoreMarker(child) {
			hasSummary = true
			continue
		}

		var buf bytes.Buffer
		if err := html.Render(&buf, child); err != nil {
			return "", "", false, err
		}
		if !hasSummary {
			summary.Write(buf.Bytes())
		}
		content.Write(buf.Bytes())
	}
	return strings.TrimSpace(content.String()), strings.TrimSpace(summary.String()), hasSummary, nil
}

func renderChildren(n *html.Node) (string, error) {
	var buf bytes.Buffer
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if isMoreMarker(child) {
			continue
		}
		if err := html.Render(&buf, child); err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(buf.String()), nil
}
