package page

import (
	"strings"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/utils"
)

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
	if err := tpl.Write(path, vars); err != nil {
		b.conf.Log.Error(err.Error())
	}
}

func (b *Builder) writePages(pages Pages) {
	for _, page := range pages {
		if !page.Meta.GetBool("section") {
			if tpl, ok := b.theme.LookupTemplate(page.Meta.GetString("page_template")); ok {
				aliases := append(page.Aliases, page.Path)
				for _, aliase := range aliases {
					b.write(tpl, aliase, map[string]interface{}{
						"page": page,
					})
				}
			}
			continue
		}

		var (
			path     = page.Meta.GetString("path")
			template = page.Meta.GetString("template")
		)
		if path == "" {
			continue
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
		section.Pages = pages.Filter(page.Meta.Get("filter")).OrderBy(page.Meta.GetString("orderby"))
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
}

func (b *Builder) writeSections(sections Sections) {
	for _, section := range sections {
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
}

// 写入分类系统, @templates/{name}/list.html, single.html
func (b *Builder) writeTaxonomies(taxonomies Taxonomies) {
	for _, taxonomy := range taxonomies {
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
		b.writeTaxonomyTerms(taxonomy, taxonomy.Terms)
	}
}

func (b *Builder) writeTaxonomyTerms(taxonomy *Taxonomy, terms TaxonomyTerms) {
	for _, term := range terms {
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
						"taxonomy":  taxonomy,
						"paginator": por,
					})
				}
			}
		}
		if feedPath != "" {
			b.writeFeeds(feedPath, feedTemplate, map[string]interface{}{
				"term":     term,
				"pages":    term.List,
				"taxonomy": taxonomy,
			})
		}
		b.writeTaxonomyTerms(taxonomy, term.Children)
	}
}

func (b *Builder) writeFeeds(path string, template string, vars map[string]interface{}) {
	if tpl, ok := b.theme.LookupTemplate(template, "_internal/feed.xml"); ok {
		b.write(tpl, path, vars)
	}
}

func (b *Builder) Write() error {
	b.writePages(b.hooks.BeforePagesWrite(b.pages))
	b.writePages(b.hooks.BeforePagesWrite(b.hiddenPages))
	b.writePages(b.hooks.BeforePagesWrite(b.sectionPages))

	b.writeSections(b.hooks.BeforeSectionsWrite(b.sections))
	b.writeTaxonomies(b.hooks.BeforeTaxonomiesWrite(b.taxonomies))
	return nil
}
