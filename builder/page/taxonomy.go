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
	sortfs := make([]func(int, int) int, 0)
	for _, k := range strings.Split(key, ",") {
		var (
			sortf   func(int, int) int
			reverse = false
		)
		if strings.HasSuffix(strings.ToUpper(key), " DESC") {
			k = k[:len(k)-5]
			reverse = true
		}
		switch k {
		case "name":
			sortf = func(i, j int) int {
				return strings.Compare(terms[i].Name, terms[j].Name)
			}
		case "count":
			sortf = func(i, j int) int {
				return utils.Compare(len(terms[i].List), len(terms[j].List))
			}
		}
		if sortf != nil {
			if reverse {
				sortfs = append(sortfs, func(i, j int) int {
					return 0 - sortf(i, j)
				})
			} else {
				sortfs = append(sortfs, sortf)
			}
		}
	}
	sort.SliceStable(terms, func(i, j int) bool {
		for _, f := range sortfs {
			result := f(i, j)
			if result != 0 {
				return result < 0
			}
		}
		// 增加一个默认排序, 避免时间相同时排序混乱
		return strings.Compare(terms[i].Name, terms[j].Name) <= 0
	})
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
		if taxonomy.Meta.GetBool("disabled") {
			continue
		}
		taxonomy.Path = b.conf.GetRelURL(utils.StringReplace(taxonomy.Meta.GetString("path"), taxonomy.vars()), lang)
		taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)
		taxonomy.Terms = pages.GroupBy(name).OrderBy(taxonomy.Meta.GetString("orderby"))

		b.insertTaxonomyTerms(taxonomy, taxonomy.Terms)

		b.ctx.insertTaxonomy(taxonomy)
	}
	taxonomies := b.ctx.Taxonomies(lang)

	sort.SliceStable(taxonomies, func(i, j int) bool {
		ti := taxonomies[i]
		tj := taxonomies[j]
		if wi, wj := ti.Meta.GetInt("weight"), tj.Meta.GetInt("weight"); wi == wj {
			return ti.Name > tj.Name
		} else {
			return wi > wj
		}
	})
}

func (b *Builder) insertTaxonomyTerms(taxonomy *Taxonomy, terms TaxonomyTerms) {
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

		b.insertTaxonomyTerms(taxonomy, term.Children)

		b.ctx.insertTaxonomyTerm(term)
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
