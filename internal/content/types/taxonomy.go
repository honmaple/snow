package types

type (
	Taxonomy struct {
		Lang      string
		Name      string
		Path      string
		Permalink string
		Terms     TaxonomyTerms
	}
	Taxonomies []*Taxonomy
)

type (
	TaxonomyTerm struct {
		Name      string
		Slug      string
		Path      string
		Permalink string

		Pages    Pages
		Formats  Formats
		Taxonomy *Taxonomy

		// Parent   *TaxonomyTerm
		// Children TaxonomyTerms
	}
	TaxonomyTerms []*TaxonomyTerm
)
