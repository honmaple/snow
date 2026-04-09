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
		FullName string
		Name     string

		Slug      string
		Path      string
		Permalink string

		Parent   *TaxonomyTerm
		Children TaxonomyTerms

		Pages    Pages
		Formats  Formats
		Taxonomy *Taxonomy
	}
	TaxonomyTerms []*TaxonomyTerm
)

func (term *TaxonomyTerm) FindChild(name string) *TaxonomyTerm {
	for _, child := range term.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func (terms TaxonomyTerms) OrderBy() TaxonomyTerms {
	return nil
}
