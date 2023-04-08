package page

import (
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/utils"
)

type (
	Taxonomy struct {
		// slug:
		// weight:
		// path:
		// template:
		// orderby:
		Meta      Meta
		Name      string
		Lang      string
		Path      string
		Permalink string
		Terms     TaxonomyTerms
	}
	Taxonomies []*Taxonomy
)

func (t *Taxonomy) vars() map[string]string {
	return map[string]string{"{taxonomy}": t.Name}
}

func (t *Taxonomy) canWrite() bool {
	return t.Meta.GetString("path") != ""
}

func (ts Taxonomies) setSort(key string) {
	sort.SliceStable(ts, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(ts[i].Name, ts[j].Name)
		case "name":
			return strings.Compare(ts[i].Name, ts[j].Name)
		case "weight":
			return utils.Compare(ts[i].Meta.GetInt("weight"), ts[j].Meta.GetInt("weight"))
		default:
			return 0
		}
	}))
}

func (ts Taxonomies) OrderBy(key string) Taxonomies {
	newTs := make(Taxonomies, len(ts))
	copy(newTs, ts)

	newTs.setSort(key)
	return newTs
}

func (b *Builder) insertTaxonomies(page *Page) {
	lang := page.Lang

	for kind := range b.conf.GetStringMap("taxonomies") {
		if kind == "_default" {
			continue
		}

		taxonomy := b.ctx.findTaxonomy(kind, lang)
		if taxonomy == nil {
			taxonomy = &Taxonomy{
				Lang: lang,
				Name: kind,
			}
			taxonomy.Meta = make(Meta)
			taxonomy.Meta.load(b.conf.GetStringMap("taxonomies._default"))
			taxonomy.Meta.load(b.conf.GetStringMap("taxonomies." + kind))
			if lang != b.conf.Site.Language {
				taxonomy.Meta.load(b.conf.GetStringMap("languages." + lang + ".taxonomies." + kind))
			}
			if taxonomy.Meta.GetBool("disabled") {
				continue
			}
			taxonomy.Path = b.conf.GetRelURL(utils.StringReplace(taxonomy.Meta.GetString("path"), taxonomy.vars()), lang)
			taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)

			b.ctx.withLock(func() {
				if _, ok := b.ctx.taxonomies[lang]; !ok {
					b.ctx.taxonomies[lang] = make(map[string]*Taxonomy)
				}
				b.ctx.taxonomies[lang][taxonomy.Name] = taxonomy
				b.ctx.list[lang].taxonomies = append(b.ctx.list[lang].taxonomies, taxonomy)
			})
		}
		b.insertTaxonomyTerms(taxonomy, page)
	}
}

func (b *Builder) writeTaxonomy(taxonomy *Taxonomy) {
	if taxonomy.canWrite() {
		lookups := []string{
			utils.StringReplace(taxonomy.Meta.GetString("template"), taxonomy.vars()),
			fmt.Sprintf("%s/taxonomy.html", taxonomy.Name),
			"taxonomy.html",
			"_default/taxonomy.html",
		}
		if tpl := b.theme.LookupTemplate(lookups...); tpl != nil {
			// example.com/tags/index.html
			b.write(tpl, taxonomy.Path, map[string]interface{}{
				"taxonomy":     taxonomy,
				"terms":        taxonomy.Terms,
				"current_lang": taxonomy.Lang,
			})
		}
	}
}
