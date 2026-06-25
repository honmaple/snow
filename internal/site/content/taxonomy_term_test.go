package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func taxonomyTermNames(terms TaxonomyTerms) []string {
	names := make([]string, len(terms))
	for i, term := range terms {
		names[i] = term.Name
	}
	return names
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

			SortTaxonomyTerms(terms, tt.key, true)

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

	SortTaxonomyTerms(terms, "name", true)

	assert.Equal(t, []string{"first", "second"}, taxonomyTermNames(terms[0].Children))
}
