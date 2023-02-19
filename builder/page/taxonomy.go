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

func (term TaxonomyTerm) RealName() string {
	if term.Parent == nil {
		return term.Name
	}
	return filepath.Join(term.Parent.Name, term.Name)
}

func (term TaxonomyTerm) Paginator() []*paginator {
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
	keys := []string{"term_path", "term_template", "feed_path", "feed_template"}
	for _, term := range terms {
		term.Meta = taxonomy.Meta.Copy()

		name := term.RealName()
		names := strings.Split(name, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = b.conf.GetSlug(name)
		}
		slug := strings.Join(slugs, "/")

		vars := map[string]string{"{term}": name, "{term:slug}": slug}
		for _, k := range keys {
			term.Meta[k] = utils.StringReplace(term.Meta.GetString(k), vars)
		}
		term.List = term.List.Filter(term.Meta.Get("term_filter")).OrderBy(term.Meta.GetString("term_orderby"))
		term.Path = b.conf.GetRelURL(term.Meta.GetString("term_path"))
		term.Permalink = b.conf.GetURL(term.Path)
		term.Taxonomy = taxonomy
		b.loadTaxonomyTerms(taxonomy, term.Children)
	}
}

func (b *Builder) loadTaxonomies() error {
	for name := range b.conf.GetStringMap("taxonomies") {
		if name == "_default" {
			continue
		}
		taxonomy := &Taxonomy{Name: name}
		taxonomy.Meta = b.newTaxonomyConfig(name)
		taxonomy.Path = b.conf.GetRelURL(taxonomy.Meta.GetString("path"))
		taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)
		taxonomy.Terms = b.pages.GroupBy(name).OrderBy(taxonomy.Meta.GetString("orderby"))

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

func (b *Builder) newTaxonomyConfig(name string) Meta {
	meta := make(Meta)
	for k, v := range b.conf.GetStringMap("taxonomies._default") {
		meta[k] = v
	}
	for k, v := range b.conf.GetStringMap("taxonomies." + name) {
		meta[k] = v
	}
	vars := map[string]string{"{taxonomy}": name}
	keys := []string{"path", "template", "term_path", "term_template", "feed_path", "feed_template"}
	for _, k := range keys {
		meta[k] = utils.StringReplace(meta.GetString(k), vars)
	}
	return meta
}
