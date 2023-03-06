package page

import (
	"sync"

	"github.com/honmaple/snow/config"
)

type (
	listContext struct {
		pages        map[string]Pages
		hiddenPages  map[string]Pages
		sectionPages map[string]Pages
		sections     map[string]Sections
		taxonomies   map[string]Taxonomies
	}
	Context struct {
		mu        sync.RWMutex
		list      *listContext
		filter    func(*Page) bool
		languages []string

		pages         map[string]map[string]*Page
		hiddenPages   map[string]map[string]*Page
		sectionPages  map[string]map[string]*Page
		sections      map[string]map[string]*Section
		taxonomies    map[string]map[string]*Taxonomy
		taxonomyTerms map[string]map[string]map[string]*TaxonomyTerm
	}
)

func (ctx *Context) Pages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.list.pages[lang]
}

func (ctx *Context) HiddenPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.list.hiddenPages[lang]
}

func (ctx *Context) SectionPages(lang string) Pages {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.list.sectionPages[lang]
}

func (ctx *Context) Sections(lang string) Sections {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.list.sections[lang]
}

func (ctx *Context) Taxonomies(lang string) Taxonomies {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.list.taxonomies[lang]
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

func (ctx *Context) insertPage(page *Page) {
	if ctx.filter != nil && !ctx.filter(page) {
		return
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	lang := page.Lang
	if page.isHidden() {
		if _, ok := ctx.hiddenPages[lang]; !ok {
			ctx.hiddenPages[lang] = make(map[string]*Page)
		}
		ctx.hiddenPages[lang][page.File] = page
		ctx.list.hiddenPages[lang] = append(ctx.list.hiddenPages[lang], page)
		page.Section.HiddenPages = append(page.Section.HiddenPages, page)
	} else if page.isSection() {
		if _, ok := ctx.sectionPages[lang]; !ok {
			ctx.sectionPages[lang] = make(map[string]*Page)
		}
		ctx.sectionPages[lang][page.File] = page
		ctx.list.sectionPages[lang] = append(ctx.list.sectionPages[lang], page)
		page.Section.SectionPages = append(page.Section.SectionPages, page)
	} else {
		if _, ok := ctx.pages[lang]; !ok {
			ctx.pages[lang] = make(map[string]*Page)
		}
		ctx.pages[lang][page.File] = page
		ctx.list.pages[lang] = append(ctx.list.pages[lang], page)
		page.Section.Pages = append(page.Section.Pages, page)
	}
}

func (ctx *Context) insertSection(section *Section) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if section.Parent != nil {
		section.Parent.Children = append(section.Parent.Children, section)
	}
	lang := section.Lang
	if _, ok := ctx.sections[lang]; !ok {
		ctx.sections[lang] = make(map[string]*Section)
	}
	ctx.sections[lang][section.File] = section
	ctx.list.sections[lang] = append(ctx.list.sections[lang], section)
}

func (ctx *Context) insertTaxonomy(taxonomy *Taxonomy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	lang := taxonomy.Lang
	if _, ok := ctx.taxonomies[lang]; !ok {
		ctx.taxonomies[lang] = make(map[string]*Taxonomy)
	}
	ctx.taxonomies[lang][taxonomy.Name] = taxonomy
	ctx.list.taxonomies[lang] = append(ctx.list.taxonomies[lang], taxonomy)
}

func (ctx *Context) insertTaxonomyTerm(term *TaxonomyTerm) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	lang := term.Taxonomy.Lang
	if _, ok := ctx.taxonomyTerms[lang]; !ok {
		ctx.taxonomyTerms[lang] = make(map[string]map[string]*TaxonomyTerm)
	}
	if _, ok := ctx.taxonomyTerms[lang][term.Taxonomy.Name]; !ok {
		ctx.taxonomyTerms[lang][term.Taxonomy.Name] = make(map[string]*TaxonomyTerm)
	}
	ctx.taxonomyTerms[lang][term.Taxonomy.Name][term.FullName()] = term
}

func newContext(conf config.Config) *Context {
	return &Context{
		list: &listContext{
			pages:        make(map[string]Pages),
			hiddenPages:  make(map[string]Pages),
			sectionPages: make(map[string]Pages),
			sections:     make(map[string]Sections),
			taxonomies:   make(map[string]Taxonomies),
		},
		filter:        filterExpr(conf.GetString("build_filter")),
		pages:         make(map[string]map[string]*Page),
		hiddenPages:   make(map[string]map[string]*Page),
		sectionPages:  make(map[string]map[string]*Page),
		sections:      make(map[string]map[string]*Section),
		taxonomies:    make(map[string]map[string]*Taxonomy),
		taxonomyTerms: make(map[string]map[string]map[string]*TaxonomyTerm),
	}
}
