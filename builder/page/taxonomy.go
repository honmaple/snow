package page

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/honmaple/snow/utils"
)

type (
	Taxonomy struct {
		// slug:
		// weight:
		// path:
		// template:
		// orderby:
		Meta      Meta
		Name      string
		Lang      string
		Path      string
		Permalink string
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
	return term.List.Filter(term.Meta.GetString("term_paginate_filter")).Paginator(
		term.Meta.GetInt("term_paginate"),
		term.Path,
		term.Meta.GetString("term_paginate_path"),
	)
}

func (terms TaxonomyTerms) Has(name string) bool {
	for _, term := range terms {
		if term.FullName() == name {
			return true
		}
	}
	return false
}

func (terms TaxonomyTerms) Find(name string) *TaxonomyTerm {
	for _, term := range terms {
		if term.FullName() == name {
			return term
		}
		if result := term.Children.Find(name); result != nil {
			return result
		}
	}
	return nil
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
	for _, term := range terms {
		term.Children.OrderBy(key)
	}
	return terms
}

func (terms TaxonomyTerms) Paginator(number int, path string, paginatePath string) []*paginator {
	list := make([]interface{}, len(terms))
	for i, term := range terms {
		list[i] = term
	}
	return Paginator(list, number, path, paginatePath)
}

func (b *Builder) findTaxonomy(kind string, langs ...string) *Taxonomy {
	lang := b.getLang(langs...)

	b.mu.RLock()
	defer b.mu.RUnlock()
	m, ok := b.taxonomies[lang]
	if !ok {
		return nil
	}
	return m[kind]
}

func (b *Builder) findTaxonomyTerm(kind, name string, langs ...string) *TaxonomyTerm {
	lang := b.getLang(langs...)

	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.taxonomyTerms[lang]
	if !ok {
		return nil
	}
	m, ok := b.taxonomyTerms[lang][kind]
	if !ok {
		return nil
	}
	return m[name]
}

func (b *Builder) insertTaxonomies(pages Pages, lang string) {
	for name := range b.conf.GetStringMap("taxonomies") {
		if name == "_default" {
			continue
		}
		taxonomy := &Taxonomy{
			Name: name,
			Lang: lang,
		}
		taxonomy.Meta = make(Meta)
		taxonomy.Meta.load(b.conf.GetStringMap("taxonomies._default"))
		taxonomy.Meta.load(b.conf.GetStringMap("taxonomies." + name))
		if lang != b.conf.Site.Language {
			taxonomy.Meta.load(b.conf.GetStringMap("languages." + lang + ".taxonomies." + name))
		}
		taxonomy.Path = b.conf.GetRelURL(utils.StringReplace(taxonomy.Meta.GetString("path"), taxonomy.vars()), lang)
		taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)
		taxonomy.Terms = pages.GroupBy(name).OrderBy(taxonomy.Meta.GetString("orderby"))

		b.insertTaxonomyTerms(taxonomy, taxonomy.Terms)

		b.mu.Lock()
		if _, ok := b.taxonomies[lang]; !ok {
			b.taxonomies[lang] = make(map[string]*Taxonomy)
		}
		b.taxonomies[lang][name] = taxonomy
		b.mu.Unlock()
	}
}

func (b *Builder) insertTaxonomyTerms(taxonomy *Taxonomy, terms TaxonomyTerms) {
	lang := taxonomy.Lang
	for _, term := range terms {
		term.Meta = taxonomy.Meta.clone()
		term.Taxonomy = taxonomy

		name := term.FullName()
		names := strings.Split(name, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = b.conf.GetSlug(name)
		}
		term.Slug = strings.Join(slugs, "/")
		term.Path = b.conf.GetRelURL(utils.StringReplace(term.Meta.GetString("term_path"), term.vars()), taxonomy.Lang)
		term.Permalink = b.conf.GetURL(term.Path)
		term.List = term.List.Filter(term.Meta.GetString("term_filter")).OrderBy(term.Meta.GetString("term_orderby"))

		b.mu.Lock()
		if _, ok := b.taxonomyTerms[lang]; !ok {
			b.taxonomyTerms[lang] = make(map[string]map[string]*TaxonomyTerm)
		}
		if _, ok := b.taxonomyTerms[lang][taxonomy.Name]; !ok {
			b.taxonomyTerms[lang][taxonomy.Name] = make(map[string]*TaxonomyTerm)
		}
		b.taxonomyTerms[lang][taxonomy.Name][name] = term
		b.mu.Unlock()

		b.insertTaxonomyTerms(taxonomy, term.Children)
	}
}

func (b *Builder) writeTaxonomy(taxonomy *Taxonomy) {
	if taxonomy.Meta.GetString("path") != "" {
		lookups := []string{
			utils.StringReplace(taxonomy.Meta.GetString("template"), taxonomy.vars()),
			fmt.Sprintf("%s/taxonomy.html", taxonomy.Name),
			"taxonomy.html",
			"_default/taxonomy.html",
		}
		if tpl, ok := b.theme.LookupTemplate(lookups...); ok {
			// example.com/tags/index.html
			b.write(tpl, taxonomy.Path, map[string]interface{}{
				"taxonomy":     taxonomy,
				"terms":        taxonomy.Terms,
				"current_lang": taxonomy.Lang,
			})
		}
	}
}

func (b *Builder) writeTaxonomyTerm(term *TaxonomyTerm) {
	var (
		vars = term.vars()
	)
	if term.Meta.GetString("term_path") != "" {
		lookups := []string{
			utils.StringReplace(term.Meta.GetString("term_template"), vars),
			fmt.Sprintf("%s/taxonomy.terms.html", term.Taxonomy.Name),
			"taxonomy.terms.html",
			"_default/taxonomy.terms.html",
		}
		if tpl, ok := b.theme.LookupTemplate(lookups...); ok {
			for _, por := range term.Paginator() {
				b.write(tpl, por.URL, map[string]interface{}{
					"term":          term,
					"pages":         term.List,
					"taxonomy":      term.Taxonomy,
					"paginator":     por,
					"current_lang":  term.Taxonomy.Lang,
					"current_path":  por.URL,
					"current_index": por.PageNum,
				})
			}
		}
	}
	for _, child := range term.Children {
		b.writeTaxonomyTerm(child)
	}
	b.writeFormats(term.Meta, vars, map[string]interface{}{
		"term":         term,
		"pages":        term.List,
		"taxonomy":     term.Taxonomy,
		"current_lang": term.Taxonomy.Lang,
	})
}
