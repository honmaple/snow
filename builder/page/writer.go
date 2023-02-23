package page

import (
	"strings"
	"sync"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/utils"
	"github.com/panjf2000/ants/v2"
)

type taskPool struct {
	*ants.PoolWithFunc
	wg *sync.WaitGroup
}

func (p *taskPool) Invoke(i interface{}) {
	p.wg.Add(1)
	p.PoolWithFunc.Invoke(i)
}

func (p *taskPool) Wait() {
	p.wg.Wait()
}

func newTaskPool(wg *sync.WaitGroup, size int, f func(interface{})) *taskPool {
	p, _ := ants.NewPoolWithFunc(size, f)
	return &taskPool{
		PoolWithFunc: p,
		wg:           wg,
	}
}

func (b *Builder) getSection(name string) *Section {
	for _, section := range b.sections {
		if name == section.Name() {
			return section
		}
	}
	return nil
}

func (b *Builder) getSectionURL(name string) string {
	if sec := b.getSection(name); sec != nil {
		return sec.Permalink
	}
	return ""
}

func (b *Builder) getTaxonomy(name string) *Taxonomy {
	for _, taxonomy := range b.taxonomies {
		if name == taxonomy.Name {
			return taxonomy
		}
	}
	return nil
}

func (b *Builder) getTaxonomyTerm(kind, name string) *TaxonomyTerm {
	for _, taxonomy := range b.taxonomies {
		if kind != taxonomy.Name {
			continue
		}
		for _, term := range taxonomy.Terms {
			if strings.ToLower(name) == term.Name {
				return term
			}
		}
	}
	return nil
}

func (b *Builder) getTaxonomyURL(kind string, names ...string) string {
	if len(names) >= 1 {
		if term := b.getTaxonomyTerm(kind, names[0]); term != nil {
			return term.Permalink
		}
		return ""
	}
	taxonomy := b.getTaxonomy(kind)
	if taxonomy != nil {
		return taxonomy.Permalink
	}
	return ""
}

func (b *Builder) write(tpl template.Writer, path string, vars map[string]interface{}) {
	if path == "" {
		return
	}
	rvars := map[string]interface{}{
		"pages":            b.pages,
		"taxonomies":       b.taxonomies,
		"get_section":      b.getSection,
		"get_section_url":  b.getSectionURL,
		"get_taxonomy":     b.getTaxonomy,
		"get_taxonomy_url": b.getTaxonomyURL,
		"current_url":      b.conf.GetURL(path),
		"current_path":     b.conf.GetRelURL(path),
		"current_template": tpl.Name(),
	}
	for k, v := range rvars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}
	if err := tpl.Write(path, vars); err != nil {
		b.conf.Log.Error(err.Error())
	}
}

func (b *Builder) writePage(page *Page) {
	if !page.isSection() {
		if tpl, ok := b.theme.LookupTemplate(page.Meta.GetString("page_template")); ok {
			aliases := append(page.Aliases, page.Path)
			for _, aliase := range aliases {
				b.write(tpl, aliase, map[string]interface{}{
					"page": page,
				})
			}
		}
		return
	}

	var (
		path     = page.Meta.GetString("path")
		template = page.Meta.GetString("template")
	)
	if path == "" {
		return
	}
	section := &Section{
		Meta:    page.Meta,
		Title:   page.Title,
		Content: page.Content,
		Pages:   page.Section.allPages(),
	}
	section.Path = b.conf.GetRelURL(path)
	section.Permalink = b.conf.GetURL(path)

	pages := page.Section.allPages()
	section.Pages = pages.Filter(page.Meta.GetString("filter")).OrderBy(page.Meta.GetString("orderby"))
	if tpl, ok := b.theme.LookupTemplate(template); ok {
		for _, por := range section.Paginator() {
			b.write(tpl, por.URL, map[string]interface{}{
				"section":       section,
				"paginator":     por,
				"pages":         section.Pages,
				"current_index": por.PageNum,
			})
		}
	}
}

func (b *Builder) writeSection(section *Section) {
	var (
		vars         = section.vars()
		path         = utils.StringReplace(section.Meta.GetString("path"), vars)
		template     = utils.StringReplace(section.Meta.GetString("template"), vars)
		feedPath     = utils.StringReplace(section.Meta.GetString("feed_path"), vars)
		feedTemplate = utils.StringReplace(section.Meta.GetString("feed_template"), vars)
	)

	if path != "" {
		if tpl, ok := b.theme.LookupTemplate(template, "section.html", "_internal/section.html"); ok {
			for _, por := range section.Paginator() {
				b.write(tpl, por.URL, map[string]interface{}{
					"section":       section,
					"paginator":     por,
					"pages":         section.Pages,
					"current_index": por.PageNum,
				})
			}
		}
	}
	if feedPath != "" {
		b.writeFeeds(feedPath, feedTemplate, map[string]interface{}{
			"section": section,
			"pages":   section.Pages,
		})
	}
}

func (b *Builder) writeTaxonomy(taxonomy *Taxonomy) {
	var (
		vars     = taxonomy.vars()
		path     = utils.StringReplace(taxonomy.Meta.GetString("path"), vars)
		template = utils.StringReplace(taxonomy.Meta.GetString("template"), vars)
	)
	if path != "" {
		if tpl, ok := b.theme.LookupTemplate(template, "taxonomy.html", "_default/list.html", "_internal/taxonomy.html"); ok {
			// example.com/tags/index.html
			b.write(tpl, path, map[string]interface{}{
				"taxonomy": taxonomy,
				"terms":    taxonomy.Terms,
			})
		}
	}
	b.writeTaxonomyTerms(taxonomy.Terms)
}

func (b *Builder) writeTaxonomyTerm(term *TaxonomyTerm) {
	var (
		vars         = term.vars()
		termPath     = utils.StringReplace(term.Meta.GetString("term_path"), vars)
		termTemplate = utils.StringReplace(term.Meta.GetString("term_template"), vars)
		feedPath     = utils.StringReplace(term.Meta.GetString("feed_path"), vars)
		feedTemplate = utils.StringReplace(term.Meta.GetString("feed_template"), vars)
	)
	if termPath != "" {
		if tpl, ok := b.theme.LookupTemplate(termTemplate, "taxonomy.terms.html", "_default/single.html", "_internal/taxonomy.terms.html"); ok {
			for _, por := range term.Paginator() {
				b.write(tpl, por.URL, map[string]interface{}{
					"term":      term,
					"pages":     term.List,
					"taxonomy":  term.Taxonomy,
					"paginator": por,
				})
			}
		}
	}
	if feedPath != "" {
		b.writeFeeds(feedPath, feedTemplate, map[string]interface{}{
			"term":     term,
			"pages":    term.List,
			"taxonomy": term.Taxonomy,
		})
	}
	b.writeTaxonomyTerms(term.Children)
}

func (b *Builder) writeFeeds(path string, template string, vars map[string]interface{}) {
	if tpl, ok := b.theme.LookupTemplate(template, "_internal/feed.xml"); ok {
		b.write(tpl, path, vars)
	}
}

func (b *Builder) writePages(pages Pages) {
	for _, page := range pages {
		b.tasks.Invoke(page)
	}
}

func (b *Builder) writeSections(sections Sections) {
	for _, section := range sections {
		b.tasks.Invoke(section)
	}
}

func (b *Builder) writeTaxonomies(taxonomies Taxonomies) {
	for _, taxonomy := range taxonomies {
		b.tasks.Invoke(taxonomy)
	}
}

func (b *Builder) writeTaxonomyTerms(terms TaxonomyTerms) {
	for _, term := range terms {
		b.tasks.Invoke(term)
	}
}

func (b *Builder) Write() error {
	var wg sync.WaitGroup

	tasks := newTaskPool(&wg, 100, func(i interface{}) {
		switch v := i.(type) {
		case *Page:
			b.writePage(v)
		case *Section:
			b.writeSection(v)
		case *Taxonomy:
			b.writeTaxonomy(v)
		case *TaxonomyTerm:
			b.writeTaxonomyTerm(v)
		}
		wg.Done()
	})
	defer tasks.Release()

	b.tasks = tasks

	b.writePages(b.hooks.BeforePagesWrite(b.pages))
	b.writePages(b.hooks.BeforePagesWrite(b.hiddenPages))
	b.writePages(b.hooks.BeforePagesWrite(b.sectionPages))

	b.writeSections(b.hooks.BeforeSectionsWrite(b.sections))
	b.writeTaxonomies(b.hooks.BeforeTaxonomiesWrite(b.taxonomies))

	tasks.Wait()
	return nil
}
