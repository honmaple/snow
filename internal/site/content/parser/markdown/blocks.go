package markdown

import (
	"bytes"
	stdhtml "html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type blockKind string

const (
	blockKindExportHTML blockKind = "export-html"
	blockKindCenter     blockKind = "center"
	blockKindQuote      blockKind = "quote"
	blockKindShortcode  blockKind = "shortcode"
)

var kindDirectiveBlock = ast.NewNodeKind("DirectiveBlock")

type directiveBlock struct {
	ast.BaseBlock
	kind      blockKind
	name      string
	arguments []string
	nesting   int
}

func (b *directiveBlock) Kind() ast.NodeKind {
	return kindDirectiveBlock
}

func (b *directiveBlock) Dump(source []byte, level int) {
	ast.DumpHelper(b, source, level, map[string]string{
		"Kind": string(b.kind),
		"Name": b.name,
	}, nil)
}

type directiveBlockParser struct{}

func (p *directiveBlockParser) Trigger() []byte {
	return []byte{':'}
}

func (p *directiveBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if pc.BlockIndent() > 3 {
		return nil, parser.NoChildren
	}

	kind, name, args, ok := p.parseLine(line)
	if !ok || kind == "" {
		return nil, parser.NoChildren
	}

	reader.AdvanceToEOL()
	node := &directiveBlock{kind: kind, name: name, arguments: args}
	if kind == blockKindExportHTML {
		return node, parser.NoChildren
	}
	return node, parser.HasChildren
}

func (p *directiveBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	block, ok := node.(*directiveBlock)
	if !ok {
		return parser.Close
	}
	if p.isClose(line) {
		if block.nesting > 0 {
			block.nesting--
			return parser.Continue | parser.HasChildren
		}
		reader.AdvanceToEOL()
		return parser.Close
	}

	if block.kind == blockKindExportHTML {
		block.Lines().Append(text.NewSegment(segment.Start, segment.Stop))
		reader.AdvanceToEOL()
		return parser.Continue | parser.NoChildren
	}
	if p.isOpen(line) {
		block.nesting++
	}
	return parser.Continue | parser.HasChildren
}

func (p *directiveBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *directiveBlockParser) CanInterruptParagraph() bool {
	return true
}

func (p *directiveBlockParser) CanAcceptIndentedLine() bool {
	return false
}

func (p *directiveBlockParser) parseLine(line []byte) (blockKind, string, []string, bool) {
	trimmed := strings.TrimSpace(string(line))
	if !strings.HasPrefix(trimmed, ":::") {
		return "", "", nil, false
	}
	fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(trimmed, ":::")))
	if len(fields) == 0 {
		return "", "", nil, false
	}
	switch fields[0] {
	case "export":
		if len(fields) >= 2 && fields[1] == "html" {
			return blockKindExportHTML, "", nil, true
		}
	case "center":
		return blockKindCenter, "", nil, true
	case "quote":
		return blockKindQuote, "", nil, true
	case "shortcode":
		if len(fields) >= 2 {
			return blockKindShortcode, fields[1], fields[2:], true
		}
	}
	return "", "", nil, false
}

func (p *directiveBlockParser) isClose(line []byte) bool {
	return strings.TrimSpace(string(line)) == ":::"
}

func (p *directiveBlockParser) isOpen(line []byte) bool {
	kind, _, _, ok := p.parseLine(line)
	return ok && kind != ""
}

type directiveBlockRenderer struct{}

func (r *directiveBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindDirectiveBlock, r.renderDirectiveBlock)
}

func (r *directiveBlockRenderer) renderDirectiveBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	block, ok := node.(*directiveBlock)
	if !ok {
		return ast.WalkContinue, nil
	}
	switch block.kind {
	case blockKindExportHTML:
		if entering {
			var buf bytes.Buffer
			for i := 0; i < block.Lines().Len(); i++ {
				line := block.Lines().At(i)
				buf.Write(line.Value(source))
			}
			_, err := w.Write(buf.Bytes())
			return ast.WalkSkipChildren, err
		}
	case blockKindCenter:
		if entering {
			_, err := w.WriteString(`<div style="text-align: center;">` + "\n")
			return ast.WalkContinue, err
		}
		_, err := w.WriteString("</div>\n")
		return ast.WalkContinue, err
	case blockKindQuote:
		if entering {
			_, err := w.WriteString("<blockquote>\n")
			return ast.WalkContinue, err
		}
		_, err := w.WriteString("</blockquote>\n")
		return ast.WalkContinue, err
	case blockKindShortcode:
		if entering {
			_, err := w.WriteString(renderShortcodeOpenTag(block) + "\n")
			return ast.WalkContinue, err
		}
		_, err := w.WriteString("</shortcode>\n")
		return ast.WalkContinue, err
	}
	return ast.WalkContinue, nil
}

func renderShortcodeOpenTag(block *directiveBlock) string {
	var b strings.Builder
	b.WriteString("<shortcode ")
	b.WriteString(stdhtml.EscapeString(block.name))
	for _, arg := range block.arguments {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || key == "" {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(stdhtml.EscapeString(key))
		b.WriteString(`="`)
		b.WriteString(stdhtml.EscapeString(strings.Trim(value, `"'`)))
		b.WriteByte('"')
	}
	b.WriteByte('>')
	return b.String()
}

type directiveBlockExtension struct{}

func (e *directiveBlockExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&directiveBlockParser{}, 650),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&directiveBlockRenderer{}, 500),
	))
}

func NewDirectiveBlockExtension() goldmark.Extender {
	return &directiveBlockExtension{}
}
