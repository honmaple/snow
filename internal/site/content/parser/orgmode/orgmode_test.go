package orgmode

import (
	"errors"
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errOrgReader = errors.New("org reader failed")

type orgErrReader struct{}

func (orgErrReader) Read([]byte) (int, error) {
	return 0, errOrgReader
}

func parseOrg(t *testing.T, text string, opt *Option) *parser.Result {
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
	result := parseOrg(t, text, nil)

	assert.Equal(t, "aaa", result.FrontMatter["title"])
	assert.Equal(t, "2023-02-24 20:35:51", result.FrontMatter["date"])
	assert.Equal(t, []string{"snow", "hello, world"}, result.FrontMatter["tags"])
	assert.Equal(t, []any{"snow", "snow1"}, result.FrontMatter["authors"])
	assert.Equal(t, map[string]any{
		"atom": map[string]any{
			"path":     "index.html",
			"template": "index.json",
		},
	}, result.FrontMatter["formats"])

	assert.Equal(t, "\nsummary\n", result.RawSummary)
	assert.Equal(t, "\nsummary\n#+snow: more\ncontent\n", result.RawContent)
	assert.Contains(t, result.Content, "summary")
	assert.Contains(t, result.Content, "content")
}

func TestPropertiesDrawerMeta(t *testing.T) {
	text := `:PROPERTIES:
:title:   aaa
:date:     2023-02-24 20:35:51
:tags:     [snow,"hello, world"]
:authors:  [snow]
:authors:  [snow1]
:formats.atom.path: index.html
:formats.atom.template: index.json
:END:

summary
#+snow: more
content
`
	assertFunc(t, text)
}

func TestKeywordMeta(t *testing.T) {
	text := `#+TITLE: aaa
#+DATE: 2023-02-24 20:35:51
#+PROPERTY: TAGS [snow,"hello, world"]
#+PROPERTY: AUTHORS [snow]
#+PROPERTY: AUTHORS [snow1]
#+PROPERTY: formats.atom.path     index.html
#+PROPERTY: formats.atom.template index.json

summary
#+snow: more
content
`
	assertFunc(t, text)
}

func TestMeta(t *testing.T) {
	TestPropertiesDrawerMeta(t)
	TestKeywordMeta(t)
}

func TestPropertyKeywordWithoutValue(t *testing.T) {
	result := parseOrg(t, `#+PROPERTY: draft

body
`, nil)

	assert.Equal(t, "", result.FrontMatter["draft"])
	assert.Equal(t, "\nbody\n", result.RawContent)
}

func TestSummaryWithoutMoreMarker(t *testing.T) {
	result := parseOrg(t, `* Title

content
`, nil)

	assert.Empty(t, result.RawSummary)
	assert.Empty(t, result.Summary)
	assert.Equal(t, "* Title\n\ncontent\n", result.RawContent)
	assert.Contains(t, result.Content, "Title")
}

func TestSummaryWithSnowMoreKeyword(t *testing.T) {
	result := parseOrg(t, `#+TITLE: aaa

summary
#+snow: more
content
`, nil)

	assert.Equal(t, "aaa", result.FrontMatter["title"])
	assert.NotContains(t, result.FrontMatter, "snow")
	assert.Equal(t, "\nsummary\n", result.RawSummary)
	assert.Equal(t, "\nsummary\n#+snow: more\ncontent\n", result.RawContent)
	assert.Contains(t, result.Summary, "summary")
}

func TestSummaryWithHTMLMoreKeyword(t *testing.T) {
	result := parseOrg(t, `#+TITLE: aaa

summary
#+html: <!--more-->
content
`, nil)

	assert.Equal(t, "aaa", result.FrontMatter["title"])
	assert.NotContains(t, result.FrontMatter, "html")
	assert.Equal(t, "\nsummary\n", result.RawSummary)
	assert.Equal(t, "\nsummary\n#+html: <!--more-->\ncontent\n", result.RawContent)
	assert.Contains(t, result.Summary, "summary")
}

func TestTOC(t *testing.T) {
	result := parseOrg(t, `* One
** Two
*** Three
`, &Option{
		MarkupOption: parser.MarkupOption{
			Style: "none",
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

func TestDottedPropertyMerge(t *testing.T) {
	result := parseOrg(t, `:PROPERTIES:
:formats.atom.path: index.html
:formats.atom.template: index.json
:END:
`, nil)

	assert.Equal(t, map[string]any{
		"atom": map[string]any{
			"path":     "index.html",
			"template": "index.json",
		},
	}, result.FrontMatter["formats"])
}

func TestQuotedDottedPropertyIsNotNormalized(t *testing.T) {
	result := parseOrg(t, `:PROPERTIES:
:"formats.js".atom: index.xml
:END:
`, nil)

	assert.NotEqual(t, map[string]any{
		"js": map[string]any{
			"atom": "index.xml",
		},
	}, result.FrontMatter["formats"])
	assert.Equal(t, map[string]any{
		"atom": "index.xml",
	}, result.FrontMatter["formats.js"])
}

func TestLongBodyLineCanBeParsed(t *testing.T) {
	longLine := strings.Repeat("a", 70*1024)
	result := parseOrg(t, longLine+"\n", nil)

	assert.Equal(t, longLine+"\n", result.RawContent)
}

func TestLongPropertyLineCanBeParsed(t *testing.T) {
	longValue := strings.Repeat("a", 70*1024)
	result := parseOrg(t, ":PROPERTIES:\n:title: "+longValue+"\n:END:\n", nil)

	assert.Equal(t, longValue, result.FrontMatter["title"])
}

func TestUnclosedPropertiesDrawerReturnsError(t *testing.T) {
	r := New(&Option{})
	_, err := r.Parse(strings.NewReader(":PROPERTIES:\n:title: missing close\ncontent\n"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "properties drawer is not closed")
}

func TestInvalidPropertiesLineReturnsError(t *testing.T) {
	r := New(&Option{})
	_, err := r.Parse(strings.NewReader(":PROPERTIES:\ntitle: missing colon prefix\n:END:\n"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid properties line")
}

func TestReaderErrorIsReturned(t *testing.T) {
	r := New(&Option{})
	_, err := r.Parse(orgErrReader{})

	require.Error(t, err)
	assert.ErrorIs(t, err, errOrgReader)
	assert.Contains(t, err.Error(), "org parser scan")
}
