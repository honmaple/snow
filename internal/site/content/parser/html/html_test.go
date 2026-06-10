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
		"custom_css": []any{"./main.css", "./body.css"},
		"custom_js":  []string{"./main.js"},
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

	assert.NotContains(t, result.FrontMatter, "custom_css")
	assert.NotContains(t, result.FrontMatter, "custom_js")
	assert.Equal(t, "ok", result.Content)
}
