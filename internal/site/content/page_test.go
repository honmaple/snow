package content

import (
	"testing"
	"time"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPage(title string, frontMatter map[string]any) *Page {
	return &Page{
		Node: &Node{
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

func pageGroupNames(groups PageGroups) []string {
	names := make([]string, len(groups))
	for i, group := range groups {
		names[i] = group.Name
	}
	return names
}

func TestParseNodeSetsWordCountAndReadingTime(t *testing.T) {
	processor := newAssetTestProcessor(t, t.TempDir(), &parser.Result{
		Content: `<p>Hello <strong>world</strong> 你好</p><script>ignored words</script><style>.ignored { color: red; }</style>`,
	})

	page, err := processor.ParsePage("hello.md", false)
	require.NoError(t, err)
	assert.Equal(t, int64(4), page.WordCount)
	assert.Equal(t, int64(1), page.ReadingTime)

	section, err := processor.ParseSection("blog/_index.md")
	require.NoError(t, err)
	assert.Equal(t, int64(4), section.WordCount)
	assert.Equal(t, int64(1), section.ReadingTime)

	assert.Equal(t, int64(0), processor.countReadingTime(0))
	assert.Equal(t, int64(1), processor.countReadingTime(1))
	assert.Equal(t, int64(1), processor.countReadingTime(200))
	assert.Equal(t, int64(2), processor.countReadingTime(201))
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

func TestPageGroupsOrderByFields(t *testing.T) {
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
			groups := PageGroups{
				{Name: "gamma", Pages: Pages{testPage("one", nil), testPage("two", nil)}},
				{Name: "alpha", Pages: Pages{testPage("one", nil), testPage("two", nil), testPage("three", nil)}},
				{Name: "beta", Pages: Pages{testPage("one", nil)}},
			}

			got := groups.OrderBy(tt.key)

			assert.Equal(t, tt.want, pageGroupNames(got))
			assert.Equal(t, []string{"gamma", "alpha", "beta"}, pageGroupNames(groups))
		})
	}
}

func TestPageGroupsOrderBySortsChildren(t *testing.T) {
	groups := PageGroups{
		{
			Name: "root",
			Children: PageGroups{
				{Name: "second"},
				{Name: "first"},
			},
		},
	}

	got := groups.OrderBy("name")

	assert.Equal(t, []string{"first", "second"}, pageGroupNames(got[0].Children))
}
