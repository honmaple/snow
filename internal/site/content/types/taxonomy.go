package types

import (
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/utils"
)

type (
	Taxonomy struct {
		Lang      string
		Name      string
		Weight    int
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

func (terms TaxonomyTerms) OrderBy(key string) TaxonomyTerms {
	newTerms := make(TaxonomyTerms, len(terms))
	copy(newTerms, terms)

	SortTaxonomyTerms(newTerms, key)
	return newTerms
}

func SortTaxonomies(taxonomies Taxonomies, key string) {
	sort.SliceStable(taxonomies, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(taxonomies[i].Name, taxonomies[j].Name)
		case "name":
			return strings.Compare(taxonomies[i].Name, taxonomies[j].Name)
		case "weigt":
			return utils.Compare(taxonomies[i].Weight, taxonomies[j].Weight)
		default:
			return 0
		}
	}))
}

func SortTaxonomyTerms(terms TaxonomyTerms, key string) {
	sort.SliceStable(terms, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(terms[i].Name, terms[j].Name)
		case "name":
			return strings.Compare(terms[i].Name, terms[j].Name)
		case "count":
			return utils.Compare(len(terms[i].Pages), len(terms[j].Pages))
		default:
			return 0
		}
	}))
	for _, term := range terms {
		SortTaxonomyTerms(term.Children, key)
	}
}
