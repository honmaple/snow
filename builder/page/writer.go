package page

import (
	"strings"

	"fmt"

	"github.com/honmaple/snow/builder/theme/template"
)

func (b *Builder) getSectionURL(name string) string {
	conf := b.newSectionConfig(name, false)
	return b.conf.GetURL(conf.Path)
}

func (b *Builder) getTaxonomyURL(kind string, names ...string) string {
	if len(names) >= 1 {
		conf := b.newTaxonomyTermConfig(kind, names[0])
		return b.conf.GetURL(conf.Path)
	}
	conf := b.newTaxonomyConfig(kind)
	return b.conf.GetURL(conf.Path)
}

func (b *Builder) write(tpl template.Writer, path string, vars map[string]interface{}) {
	vars["sections"] = b.sections
	vars["taxonomies"] = b.taxonomies
	vars["get_section_url"] = b.getSectionURL
	vars["get_taxonomy_url"] = b.getTaxonomyURL
	vars["current_url"] = b.conf.GetURL(path)
	if err := tpl.Write(path, vars); err != nil {
		b.conf.Log.Error(err.Error())
	}
}

func (b *Builder) writeSections(sections Sections) {
	for _, section := range sections {
		name := section.Name()
		conf := section.Config
		if conf.Path != "" {
			if tpl, ok := b.theme.LookupTemplate(conf.Template); ok {
				pors := section.Pages.Paginator(conf.Paginate, conf.Path, map[string]string{"{slug}": strings.ToLower(name)})
				for _, por := range pors {
					b.write(tpl, por.URL, map[string]interface{}{
						"section":   section,
						"paginator": por,
					})
				}
			}
		}
		if conf.PagePath != "" {
			if tpl, ok := b.theme.LookupTemplate(conf.PageTemplate); ok {
				for _, page := range section.Pages {
					// posts/first-page.html
					aliases := append(page.Aliases, page.Path)
					for _, aliase := range aliases {
						b.write(tpl, aliase, map[string]interface{}{
							"page": page,
						})
					}
				}
			}
		}
		if conf.FeedPath != "" {
			b.writeFeeds(conf.FeedPath, conf.Template, map[string]interface{}{
				"section": section,
				"pages":   section.Pages,
			})
		}
		b.writeSections(section.Children)
	}
}

// 写入分类系统, @templates/{name}/list.html, single.html
func (b *Builder) writeTaxonomies(taxonomies Taxonomies) {
	for _, taxonomy := range taxonomies {
		conf := taxonomy.Config
		if tpl, ok := b.theme.LookupTemplate(conf.Template, "taxonomy.html", "_default/list.html"); ok {
			// example.com/tags/index.html
			b.write(tpl, conf.Path, map[string]interface{}{
				"taxonomy": taxonomy,
				"terms":    taxonomy.Terms,
			})
		}
		b.writeTaxonomyTerms(taxonomy.Terms)
	}
}

func (b *Builder) writeTaxonomyTerms(terms TaxonomyTerms) {
	for _, term := range terms {
		conf := term.Config

		if conf.Path != "" {
			if tpl, ok := b.theme.LookupTemplate(conf.Template, "taxonomy.terms.html", "_default/single.html"); ok {
				pors := term.List.Paginator(conf.Paginate, conf.Path, map[string]string{"{slug}": strings.ToLower(term.Name)})
				for _, por := range pors {
					b.write(tpl, por.URL, map[string]interface{}{
						"term":      term,
						"slug":      term.Name,
						"paginator": por,
					})
				}
			}
		}
		if conf.FeedPath != "" {
			b.writeFeeds(conf.FeedPath, conf.FeedTemplate, map[string]interface{}{
				"term":  term,
				"pages": term.List,
			})
		}
		b.writeTaxonomyTerms(term.Children)
	}
}

func (b *Builder) writeCustoms(pages Pages) {
	meta := b.conf.GetStringMap("sections")
	for name := range meta {
		if !b.conf.GetBool(fmt.Sprintf("sections.%s.custom", name)) {
			continue
		}
		conf := b.newSectionConfig(name, true)
		if conf.Path == "" {
			continue
		}
		tpl, ok := b.theme.LookupTemplate(conf.Template)
		if !ok {
			return
		}
		for _, por := range pages.Filter(conf.Filter).OrderBy(conf.Orderby).Paginator(conf.Paginate, conf.Path, nil) {
			b.write(tpl, por.URL, map[string]interface{}{
				"pages":     pages,
				"paginator": por,
			})
		}
	}
}

func (b *Builder) writeFeeds(path string, template string, vars map[string]interface{}) {
	if path == "" {
		return
	}
	tpl, ok := b.theme.LookupTemplate(template, "_internal/feed.xml")
	if !ok {
		return
	}
	b.write(tpl, path, vars)
}

func (b *Builder) Write(pages Pages) error {
	b.writeSections(b.sections)
	b.writeTaxonomies(b.taxonomies)
	b.writeCustoms(pages)
	return nil
}
