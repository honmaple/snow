package page

import (
	"sync"

	"github.com/honmaple/snow/config"
)

type (
	listContext struct {
		pages        Pages
		hiddenPages  Pages
		sectionPages Pages
		sections     Sections
		taxonomies   Taxonomies
	}
	Context struct {
		mu     sync.RWMutex
		filter func(*Page) bool

		list          map[string]*listContext
		pages         map[string]map[string]*Page
		sections      map[string]map[string]*Section
		taxonomies    map[string]map[string]*Taxonomy
		taxonomyTerms map[string]map[string]map[string]*TaxonomyTerm
	}
)

func (ctx *Context) Pages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.list[lang]
	if !ok {
		return nil
	}
	return m.pages
}

func (ctx *Context) HiddenPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.list[lang]
	if !ok {
		return nil
	}
	return m.hiddenPages
}

func (ctx *Context) SectionPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.list[lang]
	if !ok {
		return nil
	}
	return m.sectionPages
}

func (ctx *Context) Sections(lang string) Sections {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.list[lang]
	if !ok {
		return nil
	}
	return m.sections
}

func (ctx *Context) Taxonomies(lang string) Taxonomies {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.list[lang]
	if !ok {
		return nil
	}
	return m.taxonomies
}

func (ctx *Context) withLock(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	f()
}

func (ctx *Context) findPage(file string, lang string) *Page {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.pages[lang]
	if !ok {
		return nil
	}
	return m[file]
}

func (ctx *Context) findSection(file string, lang string) *Section {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.sections[lang]
	if !ok {
		return nil
	}
	return m[file]
}

func (ctx *Context) findTaxonomy(kind string, lang string) *Taxonomy {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	m, ok := ctx.taxonomies[lang]
	if !ok {
		return nil
	}
	return m[kind]
}

func (ctx *Context) findTaxonomyTerm(kind, name string, lang string) *TaxonomyTerm {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	_, ok := ctx.taxonomyTerms[lang]
	if !ok {
		return nil
	}
	m, ok := ctx.taxonomyTerms[lang][kind]
	if !ok {
		return nil
	}
	return m[name]
}

func (ctx *Context) ensure() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for lang, sections := range ctx.sections {
		for _, section := range sections {
			key := section.Meta.GetString("page_orderby")

			section.Pages.setSort(key)
			section.HiddenPages.setSort(key)
			section.SectionPages.setSort(key)

			section.Pages.setRelation(true)
			section.HiddenPages.setRelation(true)
			section.SectionPages.setRelation(true)

			section.Children.setSort(section.Meta.GetString("orderby"))
		}
		for _, taxonomy := range ctx.taxonomies[lang] {
			taxonomy.Terms.setSort(taxonomy.Meta.GetString("orderby"))
		}

		for _, terms := range ctx.taxonomyTerms[lang] {
			for _, term := range terms {
				term.List.setSort(term.Meta.GetString("term_orderby"))
			}
		}
		list := ctx.list[lang]
		list.pages.setSort("date desc")
		list.hiddenPages.setSort("date desc")
		list.sectionPages.setSort("date desc")

		list.pages.setRelation(false)
		list.hiddenPages.setRelation(false)
		list.sectionPages.setRelation(false)

		list.taxonomies.setSort("weight")
	}
}

func newContext(conf config.Config) *Context {
	ctx := &Context{
		filter:        filterExpr(conf.GetString("build_filter")),
		list:          make(map[string]*listContext),
		pages:         make(map[string]map[string]*Page),
		sections:      make(map[string]map[string]*Section),
		taxonomies:    make(map[string]map[string]*Taxonomy),
		taxonomyTerms: make(map[string]map[string]map[string]*TaxonomyTerm),
	}
	for lang := range conf.Languages {
		ctx.list[lang] = &listContext{}
	}
	ctx.list[conf.Site.Language] = &listContext{}
	return ctx
}
