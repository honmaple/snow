package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginateSetsPermalinkWithURLFunc(t *testing.T) {
	pages := Pages{
		testPage("one", nil),
		testPage("two", nil),
		testPage("three", nil),
	}

	paginators := pages.PaginateBy(1, "/posts/", "{name}{number:optional}{extension}", func(path string) string {
		return "https://example.com" + path
	})

	require.Len(t, paginators, 3)
	assert.Equal(t, "/posts/index.html", paginators[0].Path)
	assert.Equal(t, "https://example.com/posts/index.html", paginators[0].Permalink)
	assert.Equal(t, "/posts/index2.html", paginators[1].Path)
	assert.Equal(t, "https://example.com/posts/index2.html", paginators[1].Permalink)
	assert.Equal(t, "/posts/index3.html", paginators[2].Path)
	assert.Equal(t, "https://example.com/posts/index3.html", paginators[2].Permalink)

	assert.Equal(t, paginators[1].Permalink, paginators[0].Next.Permalink)
	assert.Equal(t, paginators[0].Permalink, paginators[1].Prev.Permalink)
}

func TestPaginateKeepsPermalinkEmptyWithoutURLFunc(t *testing.T) {
	paginators := Paginate([]string{"one", "two"}, 1, "/posts/", "")

	require.Len(t, paginators, 2)
	assert.Equal(t, "/posts/index.html", paginators[0].Path)
	assert.Empty(t, paginators[0].Permalink)
}
