package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSection(title string, path string, frontMatter map[string]any) *Section {
	fm := NewFrontMatter(frontMatter)
	return &Section{
		Node: &Node{
			File:        &File{Path: path},
			Title:       title,
			Weight:      fm.GetInt64("weight"),
			FrontMatter: fm,
		},
	}
}

func sectionTitles(sections Sections) []string {
	titles := make([]string, len(sections))
	for i, section := range sections {
		titles[i] = section.Title
	}
	return titles
}

func TestSectionAllPagesUsesWeightAscending(t *testing.T) {
	section := &Section{
		Node: &Node{
			FrontMatter: NewFrontMatter(map[string]any{"sort_by": "weight"}),
		},
		Pages: Pages{
			testPage("second", map[string]any{"weight": 20}),
			testPage("first", map[string]any{"weight": 10}),
		},
		Children: Sections{
			{
				Node: &Node{
					FrontMatter: NewFrontMatter(map[string]any{"sort_by": "weight"}),
				},
				Pages: Pages{
					testPage("third", map[string]any{"weight": 30}),
				},
			},
		},
	}

	pages := section.AllPages()

	assert.Equal(t, []string{"first", "second", "third"}, pageTitles(pages))
}

func TestParseSectionPathStyleSlug(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "content", "docs")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "_index.md"), []byte("# Section"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		FrontMatter: map[string]any{
			"path":       "/Docs.v1/Guide Page.html",
			"path_style": "slug",
		},
		Content: "<p>Section</p>",
	})

	section, err := processor.ParseSection("docs/_index.md")
	require.NoError(t, err)

	assert.Equal(t, "/docs-v1/guide-page.html", section.Path)
}

func TestSortSectionsByFields(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "title asc",
			key:  "title",
			want: []string{"alpha", "beta", "gamma"},
		},
		{
			name: "title desc",
			key:  "title desc",
			want: []string{"gamma", "beta", "alpha"},
		},
		{
			name: "weight asc",
			key:  "weight",
			want: []string{"beta", "gamma", "alpha"},
		},
		{
			name: "weight desc",
			key:  "weight desc",
			want: []string{"alpha", "gamma", "beta"},
		},
		{
			name: "custom field asc",
			key:  "priority",
			want: []string{"gamma", "alpha", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := Sections{
				testSection("gamma", "content/gamma/_index.md", map[string]any{"weight": 20, "priority": 30}),
				testSection("alpha", "content/alpha/_index.md", map[string]any{"weight": 30, "priority": 40}),
				testSection("beta", "content/beta/_index.md", map[string]any{"weight": 10, "priority": 50}),
			}

			SortSections(sections, tt.key, true)

			assert.Equal(t, tt.want, sectionTitles(sections))
		})
	}
}

func TestSortSectionsSortsChildren(t *testing.T) {
	sections := Sections{
		{
			Node: &Node{
				File:        &File{Path: "content/root/_index.md"},
				Title:       "root",
				FrontMatter: NewFrontMatter(map[string]any{"weight": 10}),
			},
			Children: Sections{
				testSection("second", "content/root/second/_index.md", map[string]any{"weight": 20}),
				testSection("first", "content/root/first/_index.md", map[string]any{"weight": 10}),
			},
		},
	}

	SortSections(sections, "weight", true)

	assert.Equal(t, []string{"first", "second"}, sectionTitles(sections[0].Children))
}
