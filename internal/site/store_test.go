package site

import (
	"testing"

	"github.com/honmaple/snow/internal/site/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStoreSection(path string) *content.Section {
	title := "home"
	if path != "" {
		title = path
	}
	return &content.Section{
		Node: &content.Node{
			File:  &content.File{Dir: path},
			Lang:  "zh",
			Title: title,
		},
		Pages:       make(content.Pages, 0),
		HiddenPages: make(content.Pages, 0),
		Children:    make(content.Sections, 0),
	}
}

func testStorePage(path string, dir string, isBundle bool, hidden bool) *content.Page {
	return &content.Page{
		Node: &content.Node{
			File: &content.File{Path: path, Dir: dir},
			Lang: "zh",
		},
		IsBundle: isBundle,
		Hidden:   hidden,
	}
}

func TestContentStoreLinksSectionParent(t *testing.T) {
	store := NewContentStore()
	root := testStoreSection("")
	blog := testStoreSection("blog")
	post := testStoreSection("blog/posts")

	store.insertSection(root)
	store.insertSection(blog)
	store.insertSection(post)

	require.Same(t, root, blog.Parent)
	require.Same(t, blog, post.Parent)
	assert.Equal(t, content.Sections{blog}, root.Children)
	assert.Equal(t, content.Sections{post}, blog.Children)
}

func TestContentStoreLinksPageSection(t *testing.T) {
	store := NewContentStore()
	blog := testStoreSection("blog")
	page := testStorePage("blog/hello.md", "blog", false, false)
	bundle := testStorePage("blog/bundle/index.md", "blog/bundle", true, false)
	hidden := testStorePage("blog/hidden.md", "blog", false, true)

	store.insertSection(testStoreSection(""))
	store.insertSection(blog)
	store.insertPage(page)
	store.insertPage(bundle)
	store.insertPage(hidden)

	require.Same(t, blog, page.Section)
	require.Same(t, blog, bundle.Section)
	require.Same(t, blog, hidden.Section)
	assert.Equal(t, content.Pages{page, bundle}, blog.Pages)
	assert.Equal(t, content.Pages{hidden}, blog.HiddenPages)
}

func TestContentAncestorsReturnSectionList(t *testing.T) {
	store := NewContentStore()
	root := testStoreSection("")
	blog := testStoreSection("blog")
	post := testStoreSection("blog/posts")
	page := testStorePage("blog/posts/hello.md", "blog/posts", false, false)

	store.insertSection(root)
	store.insertSection(blog)
	store.insertSection(post)
	store.insertPage(page)

	assert.Equal(t, content.Sections{blog, root}, post.Ancestors())
	assert.Equal(t, content.Sections{post, blog, root}, page.Ancestors())
}
