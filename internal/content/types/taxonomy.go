package types

type (
	Taxonomy struct {
		Name      string
		Path      string
		Permalink string
		Terms     TaxonomyTerms
	}
	Taxonomies []*Taxonomy
)

type (
	TaxonomyTerm struct {
		Name string

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

func (term *TaxonomyTerm) GetFullName() string {
	currentTerm := term
	currentName := ""
	for {
		if currentTerm == nil {
			break
		}
		if currentName == "" {
			currentName = currentTerm.Name
		} else {
			currentName = currentTerm.Name + "/" + currentName
		}
		currentTerm = currentTerm.Parent
	}
	return currentName
}

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
