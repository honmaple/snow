package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func taxonomyNames(taxonomies Taxonomies) []string {
	names := make([]string, len(taxonomies))
	for i, taxonomy := range taxonomies {
		names[i] = taxonomy.Name
	}
	return names
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
