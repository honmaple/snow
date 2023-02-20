package page

import (
	"github.com/honmaple/snow/utils"
	"path/filepath"
	"sort"
	"strings"
)

type (
	Taxonomy struct {
		// slug:
		// weight:
		// path:
		// template:
		// orderby:
		Meta      Meta
		Path      string
		Permalink string
		Name      string
		Terms     TaxonomyTerms
	}
	Taxonomies   []*Taxonomy
	TaxonomyTerm struct {
		// term_path: /{taxonomy}/{slug}/index.html
		// term_template:
		// term_filter:
		// term_orderby:
		// term_paginate:
		// term_paginate_path: {name}{number}{extension}
		// feed_path:
		// feed_template:
		Meta      Meta
		Name      string
		Slug      string
		Path      string
		Permalink string

		List     Pages
		Parent   *TaxonomyTerm
		Children TaxonomyTerms

		Previous *TaxonomyTerm
		Next     *TaxonomyTerm

		Taxonomy *Taxonomy
	}
	TaxonomyTerms []*TaxonomyTerm
)

func (t *Taxonomy) vars() map[string]string {
	return map[string]string{"{taxonomy}": t.Name}
}

func (term *TaxonomyTerm) vars() map[string]string {
	return map[string]string{"{taxonomy}": term.Taxonomy.Name, "{term}": term.FullName(), "{term:slug}": term.Slug}
}

func (term *TaxonomyTerm) FullName() string {
	if term.Parent == nil {
		return term.Name
	}
	return filepath.Join(term.Parent.Name, term.Name)
}

func (term *TaxonomyTerm) Paginator() []*paginator {
	return term.List.Paginator(
		term.Meta.GetInt("term_paginate"),
		term.Path,
		term.Meta.GetString("term_paginate_path"),
	)
}

func (terms TaxonomyTerms) Has(name string) bool {
	for _, term := range terms {
		if term.Name == name {
			return true
		}
	}
	return false
}

func (terms TaxonomyTerms) OrderBy(key string) TaxonomyTerms {
	if key == "" {
		return terms
	}
	var (
		reverse = false
		sortf   func(int, int) bool
	)
	if strings.HasSuffix(strings.ToUpper(key), " DESC") {
		key = key[:len(key)-5]
		reverse = true
	}
	switch key {
	case "name":
		sortf = func(i, j int) bool {
			return terms[i].Name < terms[j].Name
		}
	case "count":
		sortf = func(i, j int) bool {
			return len(terms[i].List) < len(terms[j].List)
		}
	}
	if sortf == nil {
		return terms
	}
	if reverse {
		sort.SliceStable(terms, func(i, j int) bool {
			return !sortf(i, j)
		})
	} else {
		sort.SliceStable(terms, sortf)
	}
	return terms
}

func (b *Builder) loadTaxonomyTerms(taxonomy *Taxonomy, terms TaxonomyTerms) {
	for _, term := range terms {
		term.Meta = taxonomy.Meta.copy()
		term.Taxonomy = taxonomy

		name := term.FullName()
		names := strings.Split(name, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = b.conf.GetSlug(name)
		}
		term.Slug = strings.Join(slugs, "/")
		term.Path = b.conf.GetRelURL(utils.StringReplace(term.Meta.GetString("term_path"), term.vars()))
		term.Permalink = b.conf.GetURL(term.Path)
		term.List = term.List.Filter(term.Meta.Get("term_filter")).OrderBy(term.Meta.GetString("term_orderby"))
		b.loadTaxonomyTerms(taxonomy, term.Children)
	}
}

func (b *Builder) loadTaxonomies() error {
	for name := range b.conf.GetStringMap("taxonomies") {
		if name == "_default" {
			continue
		}
		taxonomy := b.newTaxonomy(name)

		b.loadTaxonomyTerms(taxonomy, taxonomy.Terms)
		b.taxonomies = append(b.taxonomies, taxonomy)
	}
	sort.SliceStable(b.taxonomies, func(i, j int) bool {
		ti := b.taxonomies[i]
		tj := b.taxonomies[j]
		if wi, wj := ti.Meta.GetInt("weight"), tj.Meta.GetInt("weight"); wi == wj {
			return ti.Name > tj.Name
		} else {
			return wi > wj
		}
	})
	return nil
}

func (b *Builder) newTaxonomy(name string) *Taxonomy {
	taxonomy := &Taxonomy{Name: name}
	taxonomy.Meta = make(Meta)
	taxonomy.Meta.load(b.conf.GetStringMap("taxonomies._default"))
	taxonomy.Meta.load(b.conf.GetStringMap("taxonomies." + name))

	taxonomy.Path = b.conf.GetRelURL(utils.StringReplace(taxonomy.Meta.GetString("path"), taxonomy.vars()))
	taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)
	taxonomy.Terms = b.pages.GroupBy(name).OrderBy(taxonomy.Meta.GetString("orderby"))
	return taxonomy
}
