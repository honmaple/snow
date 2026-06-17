package html

import (
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseHTML(t *testing.T, text string) *parser.Result {
	t.Helper()

	r := &htmlParser{}

	result, err := r.Parse(strings.NewReader(text))
	require.NoError(t, err)
	return result
}

func assertFunc(t *testing.T, text string) {
	result := parseHTML(t, text)

	assert.Equal(t, map[string]any{
		"title":   "aaa",
		"date":    "2023-02-24 20:35:51",
		"tags":    []string{"snow", "hello", "world"},
		"authors": []string{"snow", "snow1"},
		"formats": map[string]any{
			"atom": map[string]any{
				"path":     "index.html",
				"template": "index.json",
			},
		},
		"links":   []any{"./main.css", "./body.css"},
		"scripts": []string{"./main.js"},
	}, result.FrontMatter)
	assert.Equal(t, "content", result.RawContent)
	assert.Equal(t, result.RawContent, result.Content)
}

func TestMeta(t *testing.T) {
	text := `<html>
  <head>
	<title>aaa</title>
	<meta name="date" content="2023-02-24 20:35:51" />
	<meta name="tags" content="[snow, hello, world]" />
	<meta name="authors" content="[snow, snow1]" />
	<meta name="formats.atom.path" content="index.html" />
	<meta name="formats.atom.template" content="index.json" />
	<link href="./main.css" rel="stylesheet">
	<link href="./body.css" rel="stylesheet">
	<script src="./main.js"></script>
	<script type="text/javascript">sss</script>
  </head>
  <body>
	content
  </body>
</html>
`
	assertFunc(t, text)
}

func TestBodyKeepsRenderedChildMarkup(t *testing.T) {
	result := parseHTML(t, `<html><head><title>Title</title></head><body>
<main><h1>Hello</h1><p>World</p></main>
</body></html>`)

	assert.Equal(t, "Title", result.FrontMatter["title"])
	assert.Equal(t, "<main><h1>Hello</h1><p>World</p></main>", result.Content)
}

func TestMetaTypeConversionAndMerge(t *testing.T) {
	result := parseHTML(t, `<html><head>
<meta name="draft" content="true">
<meta name="weight" content="7">
<meta name="tags" content="[go, snow]">
<meta name="tags" content="[site]">
<meta name="params.author.name" content="Ada">
</head><body></body></html>`)

	assert.Equal(t, true, result.FrontMatter["draft"])
	assert.Equal(t, 7, result.FrontMatter["weight"])
	assert.Equal(t, []any{"go", "snow", "site"}, result.FrontMatter["tags"])
	assert.Equal(t, map[string]any{
		"author": map[string]any{
			"name": "Ada",
		},
	}, result.FrontMatter["params"])
}

func TestIgnoresInlineScriptsAndLinksWithoutHref(t *testing.T) {
	result := parseHTML(t, `<html><head>
<link rel="stylesheet">
<script>alert("inline")</script>
</head><body>ok</body></html>`)

	assert.NotContains(t, result.FrontMatter, "links")
	assert.NotContains(t, result.FrontMatter, "scripts")
	assert.Equal(t, "ok", result.Content)
}

func TestFragmentBodyCanBeParsed(t *testing.T) {
	result := parseHTML(t, `<h1>Hello</h1><p>World</p>`)

	assert.Equal(t, "<h1>Hello</h1><p>World</p>", result.Content)
	assert.Equal(t, result.RawContent, result.Content)
}

func TestTitleIsTrimmedAndDecoded(t *testing.T) {
	result := parseHTML(t, `<html><head><title> Hello &amp; World </title></head><body>ok</body></html>`)

	assert.Equal(t, "Hello & World", result.FrontMatter["title"])
	assert.Equal(t, "ok", result.Content)
}

func TestMetaPropertyAndItemprop(t *testing.T) {
	result := parseHTML(t, `<html><head>
<meta property="og:title" content="Open Graph Title">
<meta itemprop="description" content="Item description">
</head><body>ok</body></html>`)

	assert.Equal(t, map[string]any{
		"og": map[string]any{
			"title": "Open Graph Title",
		},
		"description": "Item description",
	}, result.FrontMatter)
}

func TestBodyMetadataIsNotCollected(t *testing.T) {
	result := parseHTML(t, `<html><head></head><body>
<meta name="title" content="Wrong">
<link href="./body.css" rel="stylesheet">
<script src="./body.js"></script>
<p>ok</p>
</body></html>`)

	assert.Empty(t, result.FrontMatter)
	assert.Contains(t, result.Content, `<meta name="title" content="Wrong"/>`)
	assert.Contains(t, result.Content, `<p>ok</p>`)
}

func TestSummaryMarker(t *testing.T) {
	result := parseHTML(t, `<html><body><p>summary</p><!--more--><p>content</p></body></html>`)

	assert.Equal(t, "<p>summary</p>", result.RawSummary)
	assert.Equal(t, "<p>summary</p>", result.Summary)
	assert.Equal(t, "<p>summary</p><p>content</p>", result.Content)
	assert.Equal(t, result.Content, result.RawContent)
}

func TestToc(t *testing.T) {
	r := &htmlParser{
		opt: parser.MarkupOption{
			ShowToc: true,
		},
	}
	result, err := r.Parse(strings.NewReader(`<html><body>
<h1>One</h1>
<h2 id="two">Two</h2>
<h3>Three</h3>
</body></html>`))
	require.NoError(t, err)

	require.Len(t, result.Toc, 1)
	assert.Equal(t, "heading-1", result.Toc[0].Id)
	assert.Equal(t, "One", result.Toc[0].Title)
	assert.Equal(t, 1, result.Toc[0].Level)
	require.Len(t, result.Toc[0].Children, 1)
	assert.Equal(t, "two", result.Toc[0].Children[0].Id)
	assert.Equal(t, "Two", result.Toc[0].Children[0].Title)
	assert.Equal(t, 2, result.Toc[0].Children[0].Level)
	require.Len(t, result.Toc[0].Children[0].Children, 1)
	assert.Equal(t, "heading-1.1.1", result.Toc[0].Children[0].Children[0].Id)
	assert.Contains(t, result.Content, `<h1 id="heading-1">One</h1>`)
	assert.Contains(t, result.Content, `<h2 id="two">Two</h2>`)
}
