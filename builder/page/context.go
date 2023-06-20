package page

import (
	"sync"

	"github.com/honmaple/snow/config"
)

type (
	Context struct {
		pages        Pages
		hiddenPages  Pages
		sectionPages Pages
		sections     Sections
		taxonomies   Taxonomies

		pageMap         map[string]*Page
		sectionMap      map[string]*Section
		taxonomyMap     map[string]*Taxonomy
		taxonomyTermMap map[string]map[string]*TaxonomyTerm
	}
	LanguagesContext struct {
		mu     sync.RWMutex
		langs  map[string]*Context
		filter func(*Page) bool
	}
)

func (ctx *LanguagesContext) Pages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.pages
}

func (ctx *LanguagesContext) HiddenPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.hiddenPages
}

func (ctx *LanguagesContext) SectionPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.sectionPages
}

func (ctx *LanguagesContext) Sections(lang string) Sections {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.sections
}

func (ctx *LanguagesContext) Taxonomies(lang string) Taxonomies {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.taxonomies
}

func (ctx *LanguagesContext) withLock(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	f()
}

func (ctx *LanguagesContext) insertPage(page *Page) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	langx, ok := ctx.langs[page.Lang]
	if !ok {
		return
	}
	section := page.Section

	if page.isHidden() {
		section.HiddenPages = append(section.HiddenPages, page)
		langx.hiddenPages = append(langx.hiddenPages, page)
	} else if page.isSection() {
		section.SectionPages = append(section.SectionPages, page)
		langx.sectionPages = append(langx.sectionPages, page)
	} else {
		section.Pages = append(section.Pages, page)
		langx.pages = append(langx.pages, page)
	}
	langx.pageMap[page.File] = page
}

func (ctx *LanguagesContext) insertSection(section *Section) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	langx, ok := ctx.langs[section.Lang]
	if !ok {
		return
	}

	if section.Parent != nil {
		section.Parent.Children = append(section.Parent.Children, section)
	}

	langx.sections = append(langx.sections, section)
	langx.sectionMap[section.File] = section
}

func (ctx *LanguagesContext) insertTaxonomy(taxonomy *Taxonomy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	langx, ok := ctx.langs[taxonomy.Lang]
	if !ok {
		return
	}

	if _, ok := langx.taxonomyMap[taxonomy.Name]; !ok {
		langx.taxonomies = append(langx.taxonomies, taxonomy)
		langx.taxonomyMap[taxonomy.Name] = taxonomy
	}
}

func (ctx *LanguagesContext) insertTaxonomyTerm(term *TaxonomyTerm) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	langx, ok := ctx.langs[term.Taxonomy.Lang]
	if !ok {
		return
	}
	if _, ok := langx.taxonomyTermMap[term.Taxonomy.Name]; !ok {
		langx.taxonomyTermMap[term.Taxonomy.Name] = make(map[string]*TaxonomyTerm)
	}
	termName := term.RealName()
	if _, ok := langx.taxonomyTermMap[term.Taxonomy.Name][termName]; !ok {
		if term.Parent == nil {
			term.Taxonomy.Terms = append(term.Taxonomy.Terms, term)
		} else {
			term.Parent.Children = append(term.Parent.Children, term)
		}
		langx.taxonomyTermMap[term.Taxonomy.Name][termName] = term
	}
}

func (ctx *LanguagesContext) findPage(file string, lang string) *Page {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.pageMap[file]
}

func (ctx *LanguagesContext) findSection(file string, lang string) *Section {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.sectionMap[file]
}

func (ctx *LanguagesContext) findTaxonomy(kind string, lang string) *Taxonomy {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	return langx.taxonomyMap[kind]
}

func (ctx *LanguagesContext) findTaxonomyTerm(kind, name string, lang string) *TaxonomyTerm {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	langx, ok := ctx.langs[lang]
	if !ok {
		return nil
	}
	terms, ok := langx.taxonomyTermMap[kind]
	if !ok {
		return nil
	}
	return terms[name]
}

func (ctx *LanguagesContext) ensure() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for _, langx := range ctx.langs {
		for _, section := range langx.sectionMap {
			key := section.Meta.GetString("page_orderby")

			section.Pages.setSort(key)
			section.HiddenPages.setSort(key)
			section.SectionPages.setSort(key)

			section.Pages.setRelation(true)
			section.HiddenPages.setRelation(true)
			section.SectionPages.setRelation(true)

			section.Children.setSort(section.Meta.GetString("orderby"))
		}

		for _, taxonomy := range langx.taxonomyMap {
			taxonomy.Terms.setSort(taxonomy.Meta.GetString("orderby"))
		}

		for _, terms := range langx.taxonomyTermMap {
			for _, term := range terms {
				term.List.setSort(term.Meta.GetString("term_orderby"))
			}
		}
		langx.pages.setSort("date desc")
		langx.hiddenPages.setSort("date desc")
		langx.sectionPages.setSort("date desc")

		langx.pages.setRelation(false)
		langx.hiddenPages.setRelation(false)
		langx.sectionPages.setRelation(false)

		langx.taxonomies.setSort("weight")
	}
}

func newContext(conf config.Config) *LanguagesContext {
	ctx := &LanguagesContext{
		langs:  make(map[string]*Context),
		filter: filterExpr(conf.GetString("build_filter")),
	}
	for lang := range conf.Languages {
		ctx.langs[lang] = &Context{
			pageMap:         make(map[string]*Page),
			sectionMap:      make(map[string]*Section),
			taxonomyMap:     make(map[string]*Taxonomy),
			taxonomyTermMap: make(map[string]map[string]*TaxonomyTerm),
		}
	}
	ctx.langs[conf.Site.Language] = &Context{
		pageMap:         make(map[string]*Page),
		sectionMap:      make(map[string]*Section),
		taxonomyMap:     make(map[string]*Taxonomy),
		taxonomyTermMap: make(map[string]map[string]*TaxonomyTerm),
	}
	return ctx
}
