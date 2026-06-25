package content

import (
	"fmt"
	"sort"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Taxonomy struct {
		Lang string
		Name string

		Path      string
		Permalink string
		Terms     TaxonomyTerms
	}
	Taxonomies []*Taxonomy
)

func (d *Processor) resolveTaxonomyPath(taxonomy *Taxonomy, customPath string) string {
	vars := map[string]string{
		"{lang}":     taxonomy.Lang,
		"{taxonomy}": taxonomy.Name,
	}
	if taxonomy.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = taxonomy.Lang
	}
	return d.resolvePath(customPath, vars)
}

func (d *Processor) ParseTaxonomies(pages Pages, lang string) Taxonomies {
	lctx := d.ctx.For(lang)

	taxonomies := make(Taxonomies, 0)
	for taxonomyName := range lctx.Config.GetStringMap("taxonomies") {
		if taxonomyName == "_default" {
			continue
		}
		taxonomy := &Taxonomy{
			Lang: lang,
			Name: taxonomyName,
		}
		customPath := lctx.GetTaxonomyConfig(taxonomyName, "path").String()
		if customPath == "" {
			customPath = "/{lang:optional}/{taxonomy}/"
		}
		taxonomy.Path = lctx.GetRelURL(d.resolveTaxonomyPath(taxonomy, customPath))
		taxonomy.Permalink = lctx.GetURL(taxonomy.Path)
		taxonomy.Terms = d.ParseTaxonomyTerms(taxonomy, pages, lang)

		taxonomies = append(taxonomies, taxonomy)
	}

	sort.SliceStable(taxonomies, func(i, j int) bool {
		wi := lctx.GetTaxonomyConfig(taxonomies[i].Name, "weight").Int64()
		wj := lctx.GetTaxonomyConfig(taxonomies[j].Name, "weight").Int64()
		if wi == wj {
			return taxonomies[i].Name < taxonomies[j].Name
		}
		return wi < wj
	})
	return taxonomies
}

func (d *Processor) RenderTaxonomy(taxonomy *Taxonomy, tplset template.TemplateSet, writer core.Writer) error {
	lctx := d.ctx.For(taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(taxonomy.Name, "template").String(),
		fmt.Sprintf("%s/list.html", taxonomy.Name),
		"taxonomy_list.html",
	}
	if tpl := tplset.Lookup(lookups...); tpl != nil {
		d.ctx.Logger.Debugf("write taxonomy [%s] -> %s", taxonomy.Name, taxonomy.Path)
		// example.com/tags/index.html
		if err := d.RenderTemplate(taxonomy.Path, tpl, map[string]any{
			"terms":        taxonomy.Terms,
			"taxonomy":     taxonomy,
			"current_lang": taxonomy.Lang,
		}, writer); err != nil {
			return err
		}
	}

	for _, term := range taxonomy.Terms {
		if err := d.RenderTaxonomyTerm(term, tplset, writer); err != nil {
			return err
		}
	}
	return nil
}
