package content

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testPage(title string, frontMatter map[string]any) *Page {
	return &Page{
		Node: &Node{
			Title:       title,
			FrontMatter: NewFrontMatter(frontMatter),
		},
	}
}

func testSection(title string, path string, frontMatter map[string]any) *Section {
	return &Section{
		Node: &Node{
			File:        &File{Path: path},
			Title:       title,
			FrontMatter: NewFrontMatter(frontMatter),
		},
	}
}

func pageTitles(pages Pages) []string {
	titles := make([]string, len(pages))
	for i, page := range pages {
		titles[i] = page.Title
	}
	return titles
}

func sectionTitles(sections Sections) []string {
	titles := make([]string, len(sections))
	for i, section := range sections {
		titles[i] = section.Title
	}
	return titles
}

func taxonomyNames(taxonomies Taxonomies) []string {
	names := make([]string, len(taxonomies))
	for i, taxonomy := range taxonomies {
		names[i] = taxonomy.Name
	}
	return names
}

func taxonomyTermNames(terms TaxonomyTerms) []string {
	names := make([]string, len(terms))
	for i, term := range terms {
		names[i] = term.Name
	}
	return names
}

func TestSortPagesWeightAscByDefault(t *testing.T) {
	pages := Pages{
		testPage("third", map[string]any{"weight": 30}),
		testPage("first", map[string]any{"weight": 10}),
		testPage("second", map[string]any{"weight": 20}),
	}

	SortPages(pages, "weight")

	assert.Equal(t, []string{"first", "second", "third"}, pageTitles(pages))
}

func TestSortPagesWeightDesc(t *testing.T) {
	pages := Pages{
		testPage("first", map[string]any{"weight": 10}),
		testPage("third", map[string]any{"weight": 30}),
		testPage("second", map[string]any{"weight": 20}),
	}

	SortPages(pages, "weight desc")

	assert.Equal(t, []string{"third", "second", "first"}, pageTitles(pages))
}

func TestSortPagesMultipleFieldsUseEachFieldDirection(t *testing.T) {
	day := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pages := Pages{
		{Node: &Node{Title: "beta", FrontMatter: NewFrontMatter(nil)}, Date: day},
		{Node: &Node{Title: "alpha", FrontMatter: NewFrontMatter(nil)}, Date: day},
		{Node: &Node{Title: "gamma", FrontMatter: NewFrontMatter(nil)}, Date: day.AddDate(0, 0, -1)},
	}

	SortPages(pages, "date desc, title asc")

	assert.Equal(t, []string{"alpha", "beta", "gamma"}, pageTitles(pages))
}

func TestSortPagesByBuiltInFields(t *testing.T) {
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

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
			name: "date asc",
			key:  "date",
			want: []string{"gamma", "beta", "alpha"},
		},
		{
			name: "date desc",
			key:  "date desc",
			want: []string{"alpha", "gamma", "beta"},
		},
		{
			name: "modified asc",
			key:  "modified",
			want: []string{"gamma", "alpha", "beta"},
		},
		{
			name: "modified desc",
			key:  "modified desc",
			want: []string{"beta", "alpha", "gamma"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := Pages{
				{Node: &Node{Title: "gamma", FrontMatter: NewFrontMatter(nil)}, Date: older, Modified: older},
				{Node: &Node{Title: "alpha", FrontMatter: NewFrontMatter(nil)}, Date: newer, Modified: older.Add(1 * time.Hour)},
				{Node: &Node{Title: "beta", FrontMatter: NewFrontMatter(nil)}, Date: older, Modified: newer},
			}

			SortPages(pages, tt.key)

			assert.Equal(t, tt.want, pageTitles(pages))
		})
	}
}

func TestSortPagesByFrontMatterFields(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "custom int asc",
			key:  "priority",
			want: []string{"beta", "gamma", "alpha"},
		},
		{
			name: "custom int desc",
			key:  "priority desc",
			want: []string{"alpha", "gamma", "beta"},
		},
		{
			name: "custom string asc",
			key:  "category",
			want: []string{"alpha", "gamma", "beta"},
		},
		{
			name: "custom string desc",
			key:  "category desc",
			want: []string{"beta", "gamma", "alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := Pages{
				testPage("alpha", map[string]any{"priority": 30, "category": "a"}),
				testPage("beta", map[string]any{"priority": 10, "category": "c"}),
				testPage("gamma", map[string]any{"priority": 20, "category": "b"}),
			}

			SortPages(pages, tt.key)

			assert.Equal(t, tt.want, pageTitles(pages))
		})
	}
}

func TestSectionRecursivePagesUsesWeightAscending(t *testing.T) {
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

	pages := section.RecursivePages()

	assert.Equal(t, []string{"first", "second", "third"}, pageTitles(pages))
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

			SortSections(sections, tt.key)

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

	SortSections(sections, "weight")

	assert.Equal(t, []string{"first", "second"}, sectionTitles(sections[0].Children))
}

func TestSortTaxonomiesByFields(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "name asc",
			key:  "name",
			want: []string{"authors", "categories", "tags"},
		},
		{
			name: "name desc",
			key:  "name desc",
			want: []string{"tags", "categories", "authors"},
		},
		{
			name: "weight asc",
			key:  "weight",
			want: []string{"categories", "tags", "authors"},
		},
		{
			name: "weight desc",
			key:  "weight desc",
			want: []string{"authors", "tags", "categories"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxonomies := Taxonomies{
				{Name: "tags", Weight: 20},
				{Name: "authors", Weight: 30},
				{Name: "categories", Weight: 10},
			}

			SortTaxonomies(taxonomies, tt.key)

			assert.Equal(t, tt.want, taxonomyNames(taxonomies))
		})
	}
}

func TestSortTaxonomyTermsByFields(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "name asc",
			key:  "name",
			want: []string{"alpha", "beta", "gamma"},
		},
		{
			name: "name desc",
			key:  "name desc",
			want: []string{"gamma", "beta", "alpha"},
		},
		{
			name: "count asc",
			key:  "count",
			want: []string{"beta", "gamma", "alpha"},
		},
		{
			name: "count desc",
			key:  "count desc",
			want: []string{"alpha", "gamma", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms := TaxonomyTerms{
				{Name: "gamma", Pages: Pages{testPage("one", nil), testPage("two", nil)}},
				{Name: "alpha", Pages: Pages{testPage("one", nil), testPage("two", nil), testPage("three", nil)}},
				{Name: "beta", Pages: Pages{testPage("one", nil)}},
			}

			SortTaxonomyTerms(terms, tt.key)

			assert.Equal(t, tt.want, taxonomyTermNames(terms))
		})
	}
}

func TestSortTaxonomyTermsSortsChildren(t *testing.T) {
	terms := TaxonomyTerms{
		{
			Name: "root",
			Children: TaxonomyTerms{
				{Name: "second"},
				{Name: "first"},
			},
		},
	}

	SortTaxonomyTerms(terms, "name")

	assert.Equal(t, []string{"first", "second"}, taxonomyTermNames(terms[0].Children))
}
