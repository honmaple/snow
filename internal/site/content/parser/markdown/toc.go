package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var tocKey = parser.NewContextKey()

type tocExtension struct{}

func (e *tocExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		// parser.WithAutoHeadingID(),
		parser.WithASTTransformers(
			util.Prioritized(e, 100),
		),
	)
}

func (e *tocExtension) headingTitle(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	e.writeNodeText(&buf, h, source)
	return strings.TrimSpace(buf.String())
}

func (e *tocExtension) writeNodeText(buf *bytes.Buffer, node ast.Node, source []byte) {
	if text, ok := node.(*ast.Text); ok {
		buf.Write(text.Value(source))
		if text.SoftLineBreak() || text.HardLineBreak() {
			buf.WriteByte(' ')
		}
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		e.writeNodeText(buf, child, source)
	}
}

func (e *tocExtension) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	var (
		toc      []*Heading
		stack    []*Heading
		counters = make([]int, 0)
	)
	source := reader.Source()
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindHeading {
			heading := n.(*ast.Heading)

			for len(stack) > 0 && stack[len(stack)-1].Level >= heading.Level {
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

			parts := make([]string, 0)
			for _, c := range counters {
				parts = append(parts, fmt.Sprintf("%d", c))
			}
			id := "heading-" + strings.Join(parts, ".")
			heading.SetAttributeString("id", []byte(id))

			child := &Heading{
				Id:       id,
				Title:    e.headingTitle(heading, source),
				Level:    heading.Level,
				Children: make([]*Heading, 0),
			}

			if len(stack) == 0 {
				toc = append(toc, child)
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, child)
			}
			stack = append(stack, child)
		}
		return ast.WalkContinue, nil
	})

	if err == nil {
		pc.Set(tocKey, toc)
	}
}

func NewTocExtension(opt *Option) goldmark.Extender {
	r := &tocExtension{}
	return r
}
