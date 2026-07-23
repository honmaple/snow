package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPaginatorTestProcessor(t *testing.T, setup ...func(*core.Config)) *Processor {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "content"), 0755))

	conf := core.DefaultConfig()
	conf.Set("base_url", "https://example.com")
	for _, fn := range setup {
		fn(conf)
	}

	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	return NewProcessor(ctx, os.DirFS(filepath.Join(root, "content")), WithParser(&assetTestParser{
		result: &parser.Result{Content: "<p>content</p>"},
	}))
}

func TestPaginateBySetsPathPermalinkAndLinks(t *testing.T) {
	processor := newPaginatorTestProcessor(t)
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
		testPage("three", nil),
	}

	pagers := processor.PaginateBy(pages, 1, "/posts/", "{name}{number}{extension}", "en")
	paginator := NewPaginator(pagers[0], pagers)

	require.Len(t, pagers, 3)
	assert.Equal(t, "/posts/", pagers[0].Path)
	assert.Equal(t, "https://example.com/posts/", pagers[0].Permalink)
	assert.Equal(t, "/posts/index2.html", pagers[1].Path)
	assert.Equal(t, "https://example.com/posts/index2.html", pagers[1].Permalink)
	assert.Equal(t, "/posts/index3.html", pagers[2].Path)
	assert.Equal(t, "https://example.com/posts/index3.html", pagers[2].Permalink)

	assert.False(t, paginator.HasPrev())
	assert.True(t, paginator.HasNext())
	assert.Equal(t, pagers[0].Path, paginator.Path)
	assert.Equal(t, pagers[0].Permalink, paginator.Permalink)
	assert.Equal(t, pagers[0].Pages, paginator.Pages)
	assert.Same(t, pagers[1], pagers[0].Next)
	assert.Same(t, pagers[0], pagers[1].Prev)
	assert.Same(t, pagers[0], paginator.First())
	assert.Same(t, pagers[2], paginator.Last())
	assert.Same(t, pagers[1], paginator.Page(2))
	assert.Same(t, pagers[0], paginator.Page(0))
	assert.Same(t, pagers[2], paginator.Page(99))
	assert.Equal(t, pagers, paginator.Pagers)
	assert.Equal(t, 3, paginator.Total)
	assert.Equal(t, Pages{pages[1]}, pagers[1].Pages)
}

func TestPaginatorPageHandlesEmptyPaginator(t *testing.T) {
	var nilPaginator *Paginator
	assert.Nil(t, nilPaginator.Page(1))
	assert.Nil(t, (&Paginator{}).Page(1))
}

func TestPaginateByUsesDefaultPaginatePathWhenEmpty(t *testing.T) {
	processor := newPaginatorTestProcessor(t)
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
	}

	pagers := processor.PaginateBy(pages, 1, "/posts/", "", "en")

	require.Len(t, pagers, 2)
	assert.Equal(t, "/posts/", pagers[0].Path)
	assert.Equal(t, "/posts/page/2/", pagers[1].Path)
	assert.Equal(t, "https://example.com/posts/", pagers[0].Permalink)
}

func TestPaginateByUsesLanguageContextForPermalink(t *testing.T) {
	processor := newPaginatorTestProcessor(t, func(conf *core.Config) {
		conf.Set("languages.zh.base_url", "https://zh.example.com")
	})
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
	}

	pagers := processor.PaginateBy(pages, 1, "/posts/", "", "zh")

	require.Len(t, pagers, 2)
	assert.Equal(t, "https://zh.example.com/posts/", pagers[0].Permalink)
	assert.Equal(t, "https://zh.example.com/posts/page/2/", pagers[1].Permalink)
}

func TestPaginateByUsesUglyDefaultPaginatePathWhenEmpty(t *testing.T) {
	processor := newPaginatorTestProcessor(t)
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
	}

	pagers := processor.PaginateBy(pages, 1, "/posts.html", "", "en")

	require.Len(t, pagers, 2)
	assert.Equal(t, "/posts.html", pagers[0].Path)
	assert.Equal(t, "/posts2.html", pagers[1].Path)
}

func TestPaginateByUsesBaseFileNameForFilePaths(t *testing.T) {
	processor := newPaginatorTestProcessor(t)
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
	}

	pagers := processor.PaginateBy(pages, 1, "/posts.html", "{name}-{number}{extension}", "en")

	require.Len(t, pagers, 2)
	assert.Equal(t, "/posts.html", pagers[0].Path)
	assert.Equal(t, "/posts-2.html", pagers[1].Path)
}

func TestPaginateByDisablesPaginationWhenSizeIsZero(t *testing.T) {
	processor := newPaginatorTestProcessor(t)
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
	}

	pagers := processor.PaginateBy(pages, 0, "/posts/", "{name}{number}{extension}", "en")
	paginator := NewPaginator(pagers[0], pagers)

	require.Len(t, pagers, 1)
	assert.Equal(t, "/posts/", pagers[0].Path)
	assert.Equal(t, pages, pagers[0].Pages)
	assert.Equal(t, 1, paginator.Total)
	assert.Equal(t, 1, pagers[0].PageNum)
}

func TestPaginateByRendersEmptyPageWhenPaginationEnabled(t *testing.T) {
	processor := newPaginatorTestProcessor(t)

	pagers := processor.PaginateBy(nil, 10, "/posts/", "{name}{number}{extension}", "en")
	paginator := NewPaginator(pagers[0], pagers)

	require.Len(t, pagers, 1)
	assert.Equal(t, "/posts/", pagers[0].Path)
	assert.Empty(t, pagers[0].Pages)
	assert.Equal(t, 1, paginator.Total)
	assert.Equal(t, 1, pagers[0].PageNum)
	assert.False(t, paginator.HasPrev())
	assert.False(t, paginator.HasNext())
}
