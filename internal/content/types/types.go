package types

import (
	"github.com/spf13/viper"
)

type (
	Store interface {
		Pages() Pages
		GetPage(string) *Page
		GetPageURL(string) string

		Sections() Sections
		GetSection(string) *Section
		GetSectionURL(string) string

		Taxonomies() Taxonomies
		GetTaxonomy(string) *Taxonomy
		GetTaxonomyURL(string) string
		GetTaxonomyTerm(string, string) *TaxonomyTerm
		GetTaxonomyTermURL(string, string) string
	}
	Loader interface {
		Load() (Store, error)
	}
)

type (
	Asset struct {
		File      string
		Path      string
		Permalink string
	}
	Assets []*Asset
)

type (
	Static struct {
		File      string
		Path      string
		Permalink string
	}
	Statics []*Static
)

type (
	Format struct {
		Name      string
		Path      string
		Template  string
		Permalink string
	}
	Formats []*Format
)

type FrontMatter struct {
	*viper.Viper
}

func NewFrontMatter(m map[string]any) *FrontMatter {
	c := viper.New()
	for k, v := range m {
		c.Set(k, v)
	}
	return &FrontMatter{c}
}
