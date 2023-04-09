package page

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/honmaple/snow/utils"
)

type (
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

		Taxonomy *Taxonomy
	}
	TaxonomyTerms []*TaxonomyTerm
)

func (term *TaxonomyTerm) canWrite() bool {
	return term.Meta.GetString("term_path") != ""
}

func (term *TaxonomyTerm) realPath(pathstr string) string {
	return utils.StringReplace(pathstr,
		map[string]string{
			"{taxonomy}":  term.Taxonomy.Name,
			"{term}":      term.RealName(),
			"{term:slug}": term.Slug,
		})
}

func (term *TaxonomyTerm) RealName() string {
	if term.Parent == nil {
		return term.Name
	}
	return filepath.Join(term.Parent.RealName(), term.Name)
}

func (term *TaxonomyTerm) Paginator() []*paginator {
	return term.List.Filter(term.Meta.GetString("term_paginate_filter")).Paginator(
		term.Meta.GetInt("term_paginate"),
		term.Path,
		term.Meta.GetString("term_paginate_path"),
	)
}

func (terms TaxonomyTerms) setSort(key string) {
	sort.SliceStable(terms, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(terms[i].Name, terms[j].Name)
		case "name":
			return strings.Compare(terms[i].Name, terms[j].Name)
		case "count":
			return utils.Compare(len(terms[i].List), len(terms[j].List))
		default:
			return 0
		}
	}))
	for _, term := range terms {
		term.Children.setSort(key)
	}
}

func (terms TaxonomyTerms) Has(name string) bool {
	return terms.Find(name) != nil
}

func (terms TaxonomyTerms) Find(name string) *TaxonomyTerm {
	for _, term := range terms {
		if term.RealName() == name {
			return term
		}
	}
	return nil
}

func (terms TaxonomyTerms) OrderBy(key string) TaxonomyTerms {
	newTerms := make(TaxonomyTerms, len(terms))
	copy(newTerms, terms)

	newTerms.setSort(key)
	return newTerms
}

func (terms TaxonomyTerms) Paginator(number int, path string, paginatePath string) []*paginator {
	list := make([]interface{}, len(terms))
	for i, term := range terms {
		list[i] = term
	}
	return Paginator(list, number, path, paginatePath)
}

func (b *Builder) insertTaxonomyTerms(taxonomy *Taxonomy, page *Page) {
	lang := page.Lang
	kind := taxonomy.Name

	var values []string

	if strings.HasPrefix(kind, "date:") {
		values = []string{page.Date.Format(kind[5:])}
	} else if v, ok := page.Meta[kind]; ok {
		switch value := v.(type) {
		case string:
			values = []string{value}
		case []string:
			values = value
		}
	}

	for _, value := range values {
		var parent *TaxonomyTerm
		for _, subname := range utils.SplitPrefix(value, "/") {
			term := b.ctx.findTaxonomyTerm(kind, subname, lang)
			if term == nil {
				term = &TaxonomyTerm{
					Meta:     taxonomy.Meta.clone(),
					Name:     subname[strings.LastIndex(subname, "/")+1:],
					Parent:   parent,
					Taxonomy: taxonomy,
				}
				names := strings.Split(subname, "/")
				slugs := make([]string, len(names))
				for i, name := range names {
					slugs[i] = b.conf.GetSlug(name)
				}
				term.Slug = strings.Join(slugs, "/")
				term.Path = b.conf.GetRelURL(term.realPath(term.Meta.GetString("term_path")), taxonomy.Lang)
				term.Permalink = b.conf.GetURL(term.Path)

				b.ctx.withLock(func() {
					if parent == nil {
						taxonomy.Terms = append(taxonomy.Terms, term)
					} else {
						parent.Children = append(parent.Children, term)
					}
					if _, ok := b.ctx.taxonomyTerms[lang]; !ok {
						b.ctx.taxonomyTerms[lang] = make(map[string]map[string]*TaxonomyTerm)
					}
					if _, ok := b.ctx.taxonomyTerms[lang][term.Taxonomy.Name]; !ok {
						b.ctx.taxonomyTerms[lang][term.Taxonomy.Name] = make(map[string]*TaxonomyTerm)
					}
					b.ctx.taxonomyTerms[lang][term.Taxonomy.Name][term.RealName()] = term
				})
			}
			b.ctx.withLock(func() {
				term.List = append(term.List, page)
			})
			parent = term
		}
	}
}

func (b *Builder) writeTaxonomyTerm(term *TaxonomyTerm) {
	if term.canWrite() {
		lookups := []string{
			term.realPath(term.Meta.GetString("term_template")),
			term.realPath("{taxonomy}/taxonomy.terms.html"),
			"taxonomy.terms.html",
			"_default/taxonomy.terms.html",
		}
		if tpl := b.theme.LookupTemplate(lookups...); tpl != nil {
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
	b.writeFormats(term.Meta, term.realPath, map[string]interface{}{
		"term":         term,
		"pages":        term.List,
		"taxonomy":     term.Taxonomy,
		"current_lang": term.Taxonomy.Lang,
	})
}
