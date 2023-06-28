package page

import (
	"strings"
	"sync"

	"github.com/honmaple/snow/config"
)

type (
	Context struct {
		mu   sync.RWMutex
		conf config.Config

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
)

func (ctx *Context) Pages() Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.pages
}

func (ctx *Context) HiddenPages() Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.hiddenPages
}

func (ctx *Context) SectionPages() Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.sectionPages
}

func (ctx *Context) Sections() Sections {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.sections
}

func (ctx *Context) Taxonomies() Taxonomies {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.taxonomies
}

func (ctx *Context) withLock(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	f()
}

func (ctx *Context) insertPage(page *Page) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	section := page.Section

	if page.isHidden() {
		section.HiddenPages = append(section.HiddenPages, page)
		ctx.hiddenPages = append(ctx.hiddenPages, page)
	} else if page.isSection() {
		section.SectionPages = append(section.SectionPages, page)
		ctx.sectionPages = append(ctx.sectionPages, page)
	} else {
		section.Pages = append(section.Pages, page)
		ctx.pages = append(ctx.pages, page)
	}
	ctx.pageMap[page.File] = page
}

func (ctx *Context) insertSection(section *Section) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if section.Parent != nil {
		section.Parent.Children = append(section.Parent.Children, section)
	}

	ctx.sections = append(ctx.sections, section)
	ctx.sectionMap[section.File] = section
}

func (ctx *Context) insertTaxonomy(taxonomy *Taxonomy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if _, ok := ctx.taxonomyMap[taxonomy.Name]; !ok {
		ctx.taxonomies = append(ctx.taxonomies, taxonomy)
		ctx.taxonomyMap[taxonomy.Name] = taxonomy
	}
}

func (ctx *Context) insertTaxonomyTerm(term *TaxonomyTerm) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if _, ok := ctx.taxonomyTermMap[term.Taxonomy.Name]; !ok {
		ctx.taxonomyTermMap[term.Taxonomy.Name] = make(map[string]*TaxonomyTerm)
	}
	termName := term.RealName()
	if _, ok := ctx.taxonomyTermMap[term.Taxonomy.Name][termName]; !ok {
		if term.Parent == nil {
			term.Taxonomy.Terms = append(term.Taxonomy.Terms, term)
		} else {
			term.Parent.Children = append(term.Parent.Children, term)
		}
		ctx.taxonomyTermMap[term.Taxonomy.Name][termName] = term
	}
}

func (ctx *Context) findPage(file string) *Page {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if strings.HasPrefix(file, "@") {
		file = ctx.conf.ContentDir + file[1:]
	}
	return ctx.pageMap[file]
}

func (ctx *Context) findSection(file string) *Section {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if strings.HasPrefix(file, "@") {
		file = ctx.conf.ContentDir + file[1:]
	}
	return ctx.sectionMap[file]
}

func (ctx *Context) findTaxonomy(kind string) *Taxonomy {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.taxonomyMap[kind]
}

func (ctx *Context) findTaxonomyTerm(kind, name string) *TaxonomyTerm {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	terms, ok := ctx.taxonomyTermMap[kind]
	if !ok {
		return nil
	}
	return terms[name]
}

func (ctx *Context) findSectionURL(name string) string {
	section := ctx.findSection(name)
	if section == nil {
		return ""
	}
	return section.Permalink
}

func (ctx *Context) findTaxonomyURL(kind string, names ...string) string {
	if len(names) > 0 {
		return ctx.findTaxonomyTermURL(kind, names[0])
	}
	taxonomy := ctx.findTaxonomy(kind)
	if taxonomy == nil {
		return ""
	}
	return taxonomy.Permalink
}

func (ctx *Context) findTaxonomyTermURL(kind, name string) string {
	term := ctx.findTaxonomyTerm(kind, name)
	if term == nil {
		return ""
	}
	return term.Permalink
}

func (ctx *Context) ensure() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for _, section := range ctx.sectionMap {
		key := section.Meta.GetString("page_orderby")

		section.Pages.setSort(key)
		section.HiddenPages.setSort(key)
		section.SectionPages.setSort(key)

		section.Pages.setRelation(true)
		section.HiddenPages.setRelation(true)
		section.SectionPages.setRelation(true)

		section.Children.setSort(section.Meta.GetString("orderby"))
	}

	for _, taxonomy := range ctx.taxonomyMap {
		taxonomy.Terms.setSort(taxonomy.Meta.GetString("orderby"))
	}

	for _, terms := range ctx.taxonomyTermMap {
		for _, term := range terms {
			term.List.setSort(term.Meta.GetString("term_orderby"))
		}
	}
	ctx.pages.setSort("date desc")
	ctx.hiddenPages.setSort("date desc")
	ctx.sectionPages.setSort("date desc")

	ctx.pages.setRelation(false)
	ctx.hiddenPages.setRelation(false)
	ctx.sectionPages.setRelation(false)

	ctx.taxonomies.setSort("weight")
}

func newContext(conf config.Config) *Context {
	ctx := &Context{
		conf:            conf,
		pageMap:         make(map[string]*Page),
		sectionMap:      make(map[string]*Section),
		taxonomyMap:     make(map[string]*Taxonomy),
		taxonomyTermMap: make(map[string]map[string]*TaxonomyTerm),
	}
	return ctx
}
