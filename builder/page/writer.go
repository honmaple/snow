package page

import (
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
	return b.getSection(name).Permalink
}

func (b *Builder) getTaxonomy(name string) *Taxonomy {
	for _, taxonomy := range b.taxonomies {
		if name == taxonomy.Name {
			return taxonomy
		}
	}
	return nil
}

func (b *Builder) getTaxonomyURL(kind string, names ...string) string {
	conf := b.newTaxonomyConfig(kind)
	if len(names) >= 1 {
		return b.conf.GetURL(utils.StringReplace(conf.GetString("term_path"), map[string]string{
			"{taxonomy}":  kind,
			"{term}":      names[0],
			"{term:slug}": b.conf.GetSlug(names[0]),
		}))
	}
	return b.conf.GetURL(conf.GetString("path"))
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
		if tpl, ok := b.theme.LookupTemplate(page.Meta.GetString("page_template")); ok {
			aliases := append(page.Aliases, page.Path)
			for _, aliase := range aliases {
				b.write(tpl, aliase, map[string]interface{}{
					"page": page,
				})
			}
		}
	}
}

func (b *Builder) writeSections(sections Sections) {
	for _, section := range sections {
		if path := section.Meta.GetString("path"); path != "" {
			lookup := []string{section.Meta.GetString("template")}
			if !section.Hidden {
				lookup = append(lookup, "section.html", "_internal/section.html")
			}
			if tpl, ok := b.theme.LookupTemplate(lookup...); ok {
				for _, por := range section.Paginator() {
					b.write(tpl, por.URL, map[string]interface{}{
						"section":       section,
						"paginator":     por,
						"pages":         section.Pages,
						"current_index": por.PageNum,
					})
				}
			}
			if section.Hidden {
				return
			}
		}
		if path := section.Meta.GetString("feed_path"); path != "" {
			b.writeFeeds(path, section.Meta.GetString("feed_template"), map[string]interface{}{
				"section": section,
				"pages":   section.Pages,
			})
		}
	}
}

// 写入分类系统, @templates/{name}/list.html, single.html
func (b *Builder) writeTaxonomies(taxonomies Taxonomies) {
	for _, taxonomy := range taxonomies {
		if path := taxonomy.Meta.GetString("path"); path != "" {
			if tpl, ok := b.theme.LookupTemplate(taxonomy.Meta.GetString("template"), "taxonomy.html", "_default/list.html", "_internal/taxonomy.html"); ok {
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
		if path := term.Meta.GetString("term_path"); path != "" {
			if tpl, ok := b.theme.LookupTemplate(term.Meta.GetString("term_template"), "taxonomy.terms.html", "_default/single.html", "_internal/taxonomy.terms.html"); ok {
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
		if path := term.Meta.GetString("feed_path"); path != "" {
			b.writeFeeds(path, term.Meta.GetString("feed_template"), map[string]interface{}{
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
	b.writeSections(b.hooks.BeforeSectionsWrite(b.sections))
	b.writeTaxonomies(b.hooks.BeforeTaxonomiesWrite(b.taxonomies))
	return nil
}
