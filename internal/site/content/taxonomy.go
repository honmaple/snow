package content

import (
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Taxonomy struct {
		Lang      string
		Name      string
		Weight    int
		Path      string
		Permalink string
		Terms     TaxonomyTerms
	}
	Taxonomies []*Taxonomy
)

func SortTaxonomies(ts Taxonomies, key string) {
	sort.SliceStable(ts, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(ts[i].Name, ts[j].Name)
		case "name":
			return strings.Compare(ts[i].Name, ts[j].Name)
		case "weight":
			// 默认weight越小越在前
			return utils.Compare(ts[i].Weight, ts[j].Weight)
		default:
			return 0
		}
	}))
}

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
		taxonomy.Path = lctx.GetRelURL(d.resolveTaxonomyPath(taxonomy, customPath))
		taxonomy.Permalink = lctx.GetURL(taxonomy.Path)
		taxonomy.Terms = d.ParseTaxonomyTerms(taxonomy, pages, lang)

		taxonomies = append(taxonomies, taxonomy)
	}
	SortTaxonomies(taxonomies, "weight")
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
