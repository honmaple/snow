package page

import (
	"strings"
	"sync"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
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

func (b *Builder) getSection(lang string) func(string, ...string) *Section {
	return func(name string, langs ...string) *Section {
		if len(langs) == 0 {
			langs = []string{lang}
		}
		return b.findSection(name, langs...)
	}
}

func (b *Builder) getSectionURL(lang string) func(string, ...string) string {
	return func(name string, langs ...string) string {
		if len(langs) == 0 {
			langs = []string{lang}
		}
		section := b.findSection(name, langs...)
		if section == nil {
			return ""
		}
		return section.Permalink
	}
}

func (b *Builder) getTaxonomy(lang string) func(string, ...string) *Taxonomy {
	return func(kind string, langs ...string) *Taxonomy {
		if len(langs) == 0 {
			langs = []string{lang}
		}
		return b.findTaxonomy(kind, langs...)
	}
}

func (b *Builder) getTaxonomyURL(lang string) func(string, ...string) string {
	return func(kind string, names ...string) string {
		if len(names) == 0 {
			taxonomy := b.findTaxonomy(kind, lang)
			if taxonomy == nil {
				return ""
			}
			return taxonomy.Permalink
		}
		if len(names) == 1 {
			term := b.findTaxonomyTerm(kind, names[0], lang)
			if term == nil {
				return ""
			}
			return term.Permalink
		}
		if names[1] == "" {
			taxonomy := b.findTaxonomy(kind, names[2])
			if taxonomy == nil {
				return ""
			}
			return taxonomy.Permalink
		}
		term := b.findTaxonomyTerm(kind, names[0], names[1])
		if term == nil {
			return ""
		}
		return term.Permalink
	}
}

func (b *Builder) getTaxonomyTerm(lang string) func(string, string, ...string) *TaxonomyTerm {
	return func(kind string, name string, langs ...string) *TaxonomyTerm {
		if len(langs) == 0 {
			langs = []string{lang}
		}
		return b.findTaxonomyTerm(kind, name, langs...)
	}
}

func (b *Builder) getTaxonomyTermURL(lang string) func(string, string, ...string) string {
	return func(kind string, name string, langs ...string) string {
		if len(langs) == 0 {
			langs = []string{lang}
		}
		term := b.findTaxonomyTerm(kind, name, langs...)
		if term == nil {
			return ""
		}
		return term.Permalink
	}
}

func (b *Builder) write(tpl template.Writer, path string, vars map[string]interface{}) {
	if path == "" {
		return
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	lang := b.conf.Site.Language
	if v, ok := vars["current_lang"]; ok && v != "" {
		lang = v.(string)
	}
	rvars := map[string]interface{}{
		"pages":                 b.list.pages[lang],
		"taxonomies":            b.list.taxonomies[lang],
		"get_section":           b.getSection(lang),
		"get_section_url":       b.getSectionURL(lang),
		"get_taxonomy":          b.getTaxonomy(lang),
		"get_taxonomy_url":      b.getTaxonomyURL(lang),
		"get_taxonomy_term":     b.getTaxonomyTerm(lang),
		"get_taxonomy_term_url": b.getTaxonomyTermURL(lang),
		"current_url":           b.conf.GetURL(path),
		"current_path":          path,
		"current_template":      tpl.Name(),
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

// write rss, atom, json
func (b *Builder) writeFormats(meta Meta, pathvars map[string]string, vars map[string]interface{}) {
	formats := meta.GetStringMap("formats")
	if len(formats) == 0 {
		return
	}

	conf := viper.New()
	conf.MergeConfigMap(formats)

	dconf := b.conf.Sub("formats")
	for _, k := range dconf.AllKeys() {
		if !conf.IsSet(k) {
			conf.Set(k, dconf.Get(k))
		}
	}

	lang := b.conf.Site.Language
	if v, ok := vars["current_lang"]; ok && v != "" {
		lang = v.(string)
	}
	for name := range formats {
		var (
			path     = utils.StringReplace(conf.GetString(name+".path"), pathvars)
			template = utils.StringReplace(conf.GetString(name+".template"), pathvars)
		)
		if path == "" || template == "" {
			continue
		}
		if tpl, ok := b.theme.LookupTemplate(template); ok {
			b.write(tpl, b.conf.GetRelURL(path, lang), vars)
		}
	}
}

func (b *Builder) Write() error {
	var wg sync.WaitGroup

	tasks := newTaskPool(&wg, 10, func(i interface{}) {
		defer wg.Done()

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
	})
	defer tasks.Release()

	b.list = &writeList{
		pages:      make(map[string]Pages),
		taxonomies: make(map[string]Taxonomies),
	}
	for lang, m := range b.sections {
		pages := make(Pages, 0)
		hiddenPages := make(Pages, 0)
		sectionPages := make(Pages, 0)
		for _, section := range m {
			if section.isEmpty() {
				continue
			}
			section.Pages = section.Pages.Filter(section.Meta.GetString("filter")).OrderBy(section.Meta.GetString("orderby"))
			section.Children.Sort()

			pages = append(pages, section.Pages...)
			hiddenPages = append(hiddenPages, section.HiddenPages...)
			sectionPages = append(sectionPages, section.SectionPages...)
		}
		b.insertTaxonomies(pages, lang)

		taxonomies := make(Taxonomies, 0)
		for _, taxonomy := range b.taxonomies[lang] {
			taxonomies = append(taxonomies, taxonomy)
		}
		b.list.pages[lang] = pages
		b.list.taxonomies[lang] = taxonomies

		for _, section := range m {
			if section.isEmpty() {
				continue
			}
			tasks.Invoke(section)
		}
		for _, page := range b.hooks.BeforePagesWrite(pages) {
			tasks.Invoke(page)
		}
		for _, page := range b.hooks.BeforePagesWrite(hiddenPages) {
			tasks.Invoke(page)
		}
		for _, page := range b.hooks.BeforePagesWrite(sectionPages) {
			tasks.Invoke(page)
		}
		for _, taxonomy := range b.taxonomies[lang] {
			tasks.Invoke(taxonomy)

			for _, term := range taxonomy.Terms {
				tasks.Invoke(term)
			}
		}
	}
	tasks.Wait()
	return nil
}
