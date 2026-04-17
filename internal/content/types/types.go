package types

import (
	"github.com/spf13/viper"
)

type (
	Store interface {
		Pages(string) Pages
		GetPage(string, string) *Page
		GetPageURL(string, string) string

		Sections(string) Sections
		GetSection(string, string) *Section
		GetSectionURL(string, string) string

		Taxonomies(string) Taxonomies
		GetTaxonomy(string, string) *Taxonomy
		GetTaxonomyURL(string, string) string
		GetTaxonomyTerm(string, string, string) *TaxonomyTerm
		GetTaxonomyTermURL(string, string, string) string
	}
	Loader interface {
		Load() (Store, error)
	}
)

type (
	Node struct {
		File        *File
		FrontMatter *FrontMatter

		Lang        string
		Slug        string
		Title       string
		Description string
		Summary     string
		Content     string
		RawContent  string
	}
	Asset struct {
		File      string
		Path      string
		Permalink string
	}
	Assets []*Asset
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
