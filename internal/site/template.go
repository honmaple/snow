package site

import (
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
)

type ContentTemplate struct {
	lang  string
	store *ContentStore
	template.Template
}

func (t *ContentTemplate) Execute(vars map[string]any) (string, error) {
	commonVars := map[string]any{
		"pages":                 t.Pages(),
		"hidden_pages":          t.Pages(),
		"sections":              t.Sections(),
		"taxonomies":            t.Taxonomies(),
		"get_page":              t.GetPage,
		"get_page_url":          t.GetPageURL,
		"get_section":           t.GetSection,
		"get_section_url":       t.GetSectionURL,
		"get_taxonomy":          t.GetTaxonomy,
		"get_taxonomy_url":      t.GetTaxonomyURL,
		"get_taxonomy_term":     t.GetTaxonomyTerm,
		"get_taxonomy_term_url": t.GetTaxonomyTermURL,
	}
	for k, v := range commonVars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}
	return t.Template.Execute(vars)
}

func (t *ContentTemplate) Sections(args ...string) content.Sections {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.Sections(lang)
}

func (t *ContentTemplate) GetSection(path string, args ...string) *content.Section {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetSection(path, lang)
}

func (t *ContentTemplate) GetSectionURL(path string, args ...string) string {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetSectionURL(path, lang)
}

func (t *ContentTemplate) Pages(args ...string) content.Pages {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.Pages(lang)
}

func (t *ContentTemplate) HiddenPages(args ...string) content.Pages {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.HiddenPages(lang)
}

func (t *ContentTemplate) GetPage(path string, args ...string) *content.Page {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetPage(path, lang)
}

func (t *ContentTemplate) GetPageURL(path string, args ...string) string {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetPageURL(path, lang)
}

func (t *ContentTemplate) Taxonomies(args ...string) content.Taxonomies {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.Taxonomies(lang)
}

func (t *ContentTemplate) GetTaxonomy(name string, args ...string) *content.Taxonomy {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetTaxonomy(name, lang)
}

func (t *ContentTemplate) GetTaxonomyURL(name string, args ...string) string {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetTaxonomyURL(name, lang)
}

func (t *ContentTemplate) GetTaxonomyTerm(taxonomyName string, name string, args ...string) *content.TaxonomyTerm {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetTaxonomyTerm(taxonomyName, name, lang)
}

func (t *ContentTemplate) GetTaxonomyTermURL(taxonomyName string, name string, args ...string) string {
	lang := t.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return t.store.GetTaxonomyTermURL(taxonomyName, name, lang)
}

type ContentTemplateSet struct {
	lang  string
	store *ContentStore
	template.TemplateSet
}

func (set *ContentTemplateSet) newTemplate(tpl template.Template) template.Template {
	return &ContentTemplate{Template: tpl, lang: set.lang, store: set.store}
}

func (set *ContentTemplateSet) Lookup(names ...string) template.Template {
	tpl := set.TemplateSet.Lookup(names...)
	if tpl == nil {
		return nil
	}
	return set.newTemplate(tpl)
}

func (set *ContentTemplateSet) FromFile(name string) (template.Template, error) {
	tpl, err := set.TemplateSet.FromFile(name)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}

func (set *ContentTemplateSet) FromBytes(b []byte) (template.Template, error) {
	tpl, err := set.TemplateSet.FromBytes(b)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}

func (set *ContentTemplateSet) FromString(b string) (template.Template, error) {
	tpl, err := set.TemplateSet.FromString(b)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}
