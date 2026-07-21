package markdown

import (
	"errors"
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errMarkdownReader = errors.New("markdown reader failed")

type markdownErrReader struct{}

func (markdownErrReader) Read([]byte) (int, error) {
	return 0, errMarkdownReader
}

func parseMarkdown(t *testing.T, text string, opt *Option) *parser.Result {
	t.Helper()
	if opt == nil {
		opt = &Option{}
	}
	r := New(opt)

	result, err := r.Parse(strings.NewReader(text))
	require.NoError(t, err)
	return result
}

func assertFunc(t *testing.T, text string) {
	r := New(&Option{})

	result, err := r.Parse(strings.NewReader(text))
	assert.Nil(t, err)

	date, _ := utils.ParseTime("2023-02-24 20:35:51")

	expected := map[string]any{
		"title":   "aaa",
		"date":    date,
		"tags":    []any{"snow", "hello, world"},
		"authors": []any{"snow", "snow1"},
		"formats": map[string]any{
			"atom": map[string]any{
				"path":     "index.html",
				"template": "index.json",
			},
			"js": map[string]any{
				"atom": "index.xml",
			},
		},
	}
	for k, v := range result.FrontMatter {
		assert.Equal(t, expected[k], v)
	}
	assert.Equal(t, "\nsummary\n", result.RawSummary)
	assert.Equal(t, "\nsummary\n<!--more-->\ncontent\n", result.RawContent)
}

func TestMeta(t *testing.T) {
	text := `---
title: aaa
date: 2023-02-24 20:35:51
tags:
  - snow
  - hello, world
authors:
  - snow
  - snow1
formats:
  atom:
   path: index.html
   template: index.json
"formats.js":
  atom: index.xml
---

summary
<!--more-->
content
`
	//	text1 := `Title:   aaa
	// Date:     2023-02-24 20:35:51
	// Tags:     [snow,"hello, world"]
	// Authors:  [snow]
	// Authors:  [snow1]
	// Formats.atom.path: index.html
	// Formats.atom.template: index.json
	// "formats.js".atom: index.xml

	// summary
	// <!--more-->
	// content
	// `

	//	text2 := `+++
	// title =   "aaa"
	// date =    "2023-02-24 20:35:51"
	// tags =     ["snow","hello, world"]
	// authors =  ["snow", "snow1"]
	// formats.atom.path = "index.html"
	// formats.atom.template = "index.json"
	// ["formats.js"]
	//   atom = "index.xml"
	// +++

	// summary
	// <!--more-->
	// content
	// `
	assertFunc(t, text)
	// assertFunc(t, text1)
	// assertFunc(t, text2)
}

func TestInlineMeta(t *testing.T) {
	result := parseMarkdown(t, `Title:   Inline
Draft:   true
Weight:  3
Tags:    [snow, parser]
Authors: [Ada]
Authors: [Grace]
Formats.atom.path: index.xml

body
`, nil)

	assert.Equal(t, map[string]any{
		"title":   "Inline",
		"draft":   true,
		"weight":  3,
		"tags":    []string{"snow", "parser"},
		"authors": []any{"Ada", "Grace"},
		"formats": map[string]any{
			"atom": map[string]any{
				"path": "index.xml",
			},
		},
	}, result.FrontMatter)
	assert.Equal(t, "\nbody\n", result.RawContent)
	assert.Contains(t, result.Content, "<p>body</p>")
}

func TestTOMLFrontMatter(t *testing.T) {
	result := parseMarkdown(t, `+++
title = "TOML"
draft = true
weight = 2
tags = ["snow", "parser"]
[formats.atom]
path = "index.xml"
+++

content
`, nil)

	assert.Equal(t, "TOML", result.FrontMatter["title"])
	assert.Equal(t, true, result.FrontMatter["draft"])
	assert.Equal(t, int64(2), result.FrontMatter["weight"])
	assert.Equal(t, []any{"snow", "parser"}, result.FrontMatter["tags"])
	assert.Equal(t, map[string]any{
		"atom": map[string]any{
			"path": "index.xml",
		},
	}, result.FrontMatter["formats"])
	assert.Equal(t, "\ncontent\n", result.RawContent)
}

func TestSummaryWithoutMoreMarker(t *testing.T) {
	result := parseMarkdown(t, `# Title

content
`, nil)

	assert.Empty(t, result.RawSummary)
	assert.Empty(t, result.Summary)
	assert.Equal(t, "# Title\n\ncontent\n", result.RawContent)
	assert.Contains(t, result.Content, "<h1")
}

func TestTOC(t *testing.T) {
	result := parseMarkdown(t, `# One

## Two

### Three
`, &Option{
		MarkupOption: parser.MarkupOption{
			ShowToc: true,
			Style:   "none",
		},
	})

	require.Len(t, result.Toc, 1)
	assert.Equal(t, "One", result.Toc[0].Title)
	assert.Equal(t, 1, result.Toc[0].Level)
	require.Len(t, result.Toc[0].Children, 1)
	assert.Equal(t, "Two", result.Toc[0].Children[0].Title)
	assert.Equal(t, 2, result.Toc[0].Children[0].Level)
	require.Len(t, result.Toc[0].Children[0].Children, 1)
	assert.Equal(t, "Three", result.Toc[0].Children[0].Children[0].Title)
	assert.Equal(t, 3, result.Toc[0].Children[0].Children[0].Level)
}

func TestTOCHeadingTitleWithInlineMarkup(t *testing.T) {
	result := parseMarkdown(t, "# One **Two** [Three](https://example.com) `Four`\n", &Option{
		MarkupOption: parser.MarkupOption{
			ShowToc: true,
			Style:   "none",
		},
	})

	require.Len(t, result.Toc, 1)
	assert.Equal(t, "One Two Three Four", result.Toc[0].Title)
}

func TestExportHTMLBlock(t *testing.T) {
	result := parseMarkdown(t, `before

:::export html
<div class="raw">
  <span>HTML</span>
</div>
:::

after
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<p>before</p>`)
	assert.Contains(t, result.Content, `<div class="raw">`+"\n"+`  <span>HTML</span>`)
	assert.Contains(t, result.Content, `<p>after</p>`)
	assert.NotContains(t, result.Content, ":::export")
	assert.NotContains(t, result.Content, ":::")
}

func TestCenterBlock(t *testing.T) {
	result := parseMarkdown(t, `:::center
**centered**
:::
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<div style="text-align: center;">`)
	assert.Contains(t, result.Content, `<strong>centered</strong>`)
	assert.Contains(t, result.Content, `</div>`)
	assert.NotContains(t, result.Content, ":::")
}

func TestQuoteBlock(t *testing.T) {
	result := parseMarkdown(t, `:::quote
quoted **text**
:::
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<blockquote>`)
	assert.Contains(t, result.Content, `<strong>text</strong>`)
	assert.Contains(t, result.Content, `</blockquote>`)
	assert.NotContains(t, result.Content, ":::")
}

func TestShortcodeBlock(t *testing.T) {
	result := parseMarkdown(t, `:::shortcode notice type=info
**Snow**

- static-site
- go
:::
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<shortcode notice type="info">`)
	assert.Contains(t, result.Content, `<strong>Snow</strong>`)
	assert.Contains(t, result.Content, `<li>static-site</li>`)
	assert.Contains(t, result.Content, `</shortcode>`)
	assert.NotContains(t, result.Content, ":::shortcode")
	assert.NotContains(t, result.Content, ":::")
}

func TestNestedDirectiveBlocks(t *testing.T) {
	result := parseMarkdown(t, `:::center
outer

:::quote
inner **text**
:::

after
:::
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<div style="text-align: center;">`)
	assert.Contains(t, result.Content, `<blockquote>`)
	assert.Contains(t, result.Content, `<strong>text</strong>`)
	assert.Contains(t, result.Content, `after`)
	assert.Contains(t, result.Content, `</blockquote>`)
	assert.Contains(t, result.Content, `</div>`)
	assert.NotContains(t, result.Content, ":::")
}

func TestSiblingDirectiveBlocks(t *testing.T) {
	result := parseMarkdown(t, `:::quote
first
:::

:::center
second
:::
`, &Option{DirectiveBlocks: true})

	assert.Contains(t, result.Content, `<blockquote>`)
	assert.Contains(t, result.Content, `first`)
	assert.Contains(t, result.Content, `<div style="text-align: center;">`)
	assert.Contains(t, result.Content, `second`)
	assert.NotContains(t, result.Content, ":::")
}

func TestDirectiveBlocksDisabledByDefault(t *testing.T) {
	result := parseMarkdown(t, `:::center
**centered**
:::
`, nil)

	assert.NotContains(t, result.Content, `<div style="text-align: center;">`)
	assert.Contains(t, result.Content, `:::center`)
}

func TestLongLineCanBeParsed(t *testing.T) {
	longLine := strings.Repeat("a", 70*1024)
	result := parseMarkdown(t, longLine+"\n", nil)

	assert.Equal(t, longLine+"\n", result.RawContent)
	assert.Contains(t, result.Content, longLine)
}

func TestUnclosedFrontMatterReturnsError(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "yaml",
			text: "---\ntitle: missing close\ncontent\n",
		},
		{
			name: "toml",
			text: "+++\ntitle = \"missing close\"\ncontent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&Option{})
			_, err := r.Parse(strings.NewReader(tt.text))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "front matter is not closed")
		})
	}
}

func TestInvalidFrontMatterReturnsError(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "yaml",
			text: "---\nformats: [\n---\ncontent\n",
		},
		{
			name: "toml",
			text: "+++\nformats = [\n+++\ncontent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&Option{})
			_, err := r.Parse(strings.NewReader(tt.text))
			require.Error(t, err)
		})
	}
}

func TestReaderErrorIsReturned(t *testing.T) {
	r := New(&Option{})
	_, err := r.Parse(markdownErrReader{})

	require.Error(t, err)
	assert.ErrorIs(t, err, errMarkdownReader)
	assert.Contains(t, err.Error(), "markdown parser scan")
}
