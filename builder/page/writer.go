package page

import (
	"strings"

	"fmt"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/utils"
)

func (b *Builder) getTaxonomyURL(kind string, name string) string {
	conf := b.newTaxonomyConfig(kind)
	return b.conf.GetRelURL(utils.StringReplace(conf.TermPath, map[string]string{"{taxonomy}": kind, "{slug}": name}))
}

func (b *Builder) getSectionURL(name string) string {
	conf := b.newSectionConfig(name, false)
	return b.conf.GetRelURL(utils.StringReplace(conf.Path, map[string]string{"{section}": name}))
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
			tpl, ok := b.theme.LookupTemplate(conf.Template)
			if ok {
				pors := section.Pages.Paginator(conf.Paginate, conf.Path, map[string]string{"{slug}": strings.ToLower(name)})
				for _, por := range pors {
					b.write(tpl, por.URL, map[string]interface{}{
						"section":   section,
						"paginator": por,
					})
				}
			}
		}
		if conf.PagePath == "" {
			continue
		}
		tpl, ok := b.theme.LookupTemplate(conf.PageTemplate)
		if !ok {
			continue
		}
		for _, page := range section.Pages {
			// posts/first-page.html
			aliases := append(page.Aliases, page.Path)
			for _, aliase := range aliases {
				b.write(tpl, aliase, map[string]interface{}{
					"page": page,
				})
			}
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
		b.writeTaxonomyTerms(taxonomy.Terms, conf)
	}
}

func (b *Builder) writeTaxonomyTerms(terms TaxonomyTerms, conf TaxonomyConfig) {
	tpl, ok := b.theme.LookupTemplate(conf.TermTemplate, "taxonomy.terms.html", "_default/single.html")
	if !ok {
		return
	}
	for _, term := range terms {
		pors := term.List.
			Filter(conf.TermFilter).
			OrderBy(conf.TermOrderby).
			Paginator(conf.TermPaginate, conf.TermPath, map[string]string{"{slug}": strings.ToLower(term.Name)})
		for _, por := range pors {
			b.write(tpl, por.URL, map[string]interface{}{
				"term":      term,
				"slug":      term.Name,
				"paginator": por,
			})
		}
		b.writeTaxonomyTerms(term.Children, conf)
	}
}

func (b *Builder) writeCustoms() {
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
		pages := b.pages.Filter(conf.Filter).OrderBy(conf.Orderby)
		for _, por := range pages.Paginator(conf.Paginate, conf.Path, nil) {
			b.write(tpl, por.URL, map[string]interface{}{
				"pages":     pages,
				"paginator": por,
			})
		}
	}
}

func (b *Builder) Write() error {
	b.writeSections(b.sections)
	b.writeTaxonomies(b.taxonomies)
	b.writeCustoms()
	return nil
}
