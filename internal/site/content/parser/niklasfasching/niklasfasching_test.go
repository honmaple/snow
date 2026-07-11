package niklasfasching

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	_ "github.com/honmaple/snow/internal/site/content/parser/orgmode"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errOrgReader = errors.New("niklasfasching org reader failed")

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

func TestPropertiesDrawerMeta(t *testing.T) {
	result := parseOrg(t, `:PROPERTIES:
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
`, nil)

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

func TestKeywordMeta(t *testing.T) {
	result := parseOrg(t, `#+TITLE: aaa
#+DATE: 2023-02-24 20:35:51
#+PROPERTY: TAGS [snow,"hello, world"]
#+PROPERTY: AUTHORS [snow]
#+PROPERTY: AUTHORS [snow1]
#+PROPERTY: formats.atom.path     index.html
#+PROPERTY: formats.atom.template index.json

summary
#+snow: more
content
`, nil)

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

func TestReaderErrorIsReturned(t *testing.T) {
	r := New(&Option{})
	_, err := r.Parse(orgErrReader{})

	require.Error(t, err)
	assert.ErrorIs(t, err, errOrgReader)
	assert.Contains(t, err.Error(), "niklasfasching org parser scan")
}

func TestNewWithContextUsesNiklasfaschingConfig(t *testing.T) {
	conf := core.DefaultConfig()
	conf.Set("markups.niklasfasching.style", "none")
	conf.Set("markups.niklasfasching.show_line_numbers", false)
	conf.Set("markups.niklasfasching.prevent_pre_code", false)
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	r := NewWithContext(ctx)
	assert.Equal(t, "none", r.opt.Style)
	assert.False(t, r.opt.ShowLineNumbers)
	assert.False(t, r.opt.PreventPreCode)
}

func TestParserRegistryFallsBackUnlessEnabled(t *testing.T) {
	fsys := fstest.MapFS{
		"index.org": &fstest.MapFile{Data: []byte("* One\n")},
	}

	conf := core.DefaultConfig()
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)
	result, err := parser.New(ctx).Parse(fsys, "index.org")
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)
	assert.NotContains(t, result.Content, "outline-container-headline-1")

	conf.Set("markups.niklasfasching.enabled", true)
	ctx, err = core.NewContext(conf)
	require.NoError(t, err)
	result, err = parser.New(ctx).Parse(fsys, "index.org")
	require.NoError(t, err)
	assert.Contains(t, result.Content, "outline-container-headline-1")
}

func TestTemplateFilterRequiresEnabledParser(t *testing.T) {
	conf := core.DefaultConfig()
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	tplset, err := template.NewSet(ctx)
	require.NoError(t, err)
	tpl, err := tplset.FromString(`{{ "* One" | parser:"niklasfasching" }}`)
	require.NoError(t, err)
	_, err = tpl.Execute(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "niklasfasching")
	assert.Contains(t, err.Error(), "no niklasfasching parser")
}

func TestTemplateFilter(t *testing.T) {
	conf := core.DefaultConfig()
	conf.Set("markups.niklasfasching.enabled", true)
	conf.Set("markups.niklasfasching.style", "none")
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	tplset, err := template.NewSet(ctx)
	require.NoError(t, err)
	tpl, err := tplset.FromString(`{{ "* One" | parser:"niklasfasching" }}`)
	require.NoError(t, err)

	out, err := tpl.Execute(nil)
	require.NoError(t, err)
	assert.Contains(t, out, "One")
	assert.Contains(t, out, "outline-container-headline-1")
}
