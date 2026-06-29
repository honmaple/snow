package content

import (
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/viper"
)

type (
	TaxonomyTerm struct {
		Name string
		Slug string

		Path      string
		Permalink string

		Parent   *TaxonomyTerm
		Children TaxonomyTerms

		Pages    Pages
		Formats  Formats
		Taxonomy *Taxonomy
	}
	TaxonomyTerms []*TaxonomyTerm
)

func SortTaxonomyTerms(terms TaxonomyTerms, key string, recursive bool) {
	sort.SliceStable(terms, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(terms[i].Name, terms[j].Name)
		case "name":
			return strings.Compare(terms[i].Name, terms[j].Name)
		case "count":
			return utils.Compare(len(terms[i].Pages), len(terms[j].Pages))
		default:
			return 0
		}
	}))
	if recursive {
		for _, term := range terms {
			SortTaxonomyTerms(term.Children, key, true)
		}
	}
}

func (term *TaxonomyTerm) GetFullName() string {
	currentTerm := term
	currentName := ""
	for {
		if currentTerm == nil {
			break
		}
		if currentName == "" {
			currentName = currentTerm.Name
		} else {
			currentName = currentTerm.Name + "/" + currentName
		}
		currentTerm = currentTerm.Parent
	}
	return currentName
}

func (term *TaxonomyTerm) FindChild(name string) *TaxonomyTerm {
	for _, child := range term.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func (terms TaxonomyTerms) Reverse() TaxonomyTerms {
	ns := make(TaxonomyTerms, len(terms))
	for i, j := 0, len(terms)-1; j >= 0; i, j = i+1, j-1 {
		ns[i] = terms[j]
	}
	return ns
}

func (terms TaxonomyTerms) OrderBy(key string) TaxonomyTerms {
	newTerms := make(TaxonomyTerms, len(terms))
	copy(newTerms, terms)

	SortTaxonomyTerms(newTerms, key, true)
	return newTerms
}

func (d *Processor) resolveTaxonomyTermPath(term *TaxonomyTerm, customPath string) string {
	lctx := d.ctx.For(term.Taxonomy.Lang)

	name := term.GetFullName()
	vars := map[string]string{
		"{lang}":      term.Taxonomy.Lang,
		"{taxonomy}":  term.Taxonomy.Name,
		"{term}":      name,
		"{term:slug}": lctx.GetPathSlug(name),
	}
	if term.Taxonomy.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = term.Taxonomy.Lang
	}

	customPath = d.resolvePath(customPath, vars)
	return lctx.ApplyPathStyle(customPath, lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.path_style").String())
}

func (d *Processor) parseTaxonomyTermsFromGroups(taxonomy *Taxonomy, groups PageGroups, parent *TaxonomyTerm) TaxonomyTerms {
	lctx := d.ctx.For(taxonomy.Lang)

	results := make(TaxonomyTerms, 0, len(groups))
	for _, group := range groups {
		term := &TaxonomyTerm{
			Name:     group.Name,
			Slug:     lctx.GetSlug(group.Name),
			Pages:    group.Pages,
			Parent:   parent,
			Children: make(TaxonomyTerms, 0),
			Taxonomy: taxonomy,
		}

		customPath := lctx.GetTaxonomyConfig(taxonomy.Name, "term.path").String()
		if customPath == "" {
			customPath = "/{lang:optional}/{taxonomy}/{term:slug}/"
		}
		term.Path = d.resolveTaxonomyTermPath(term, customPath)
		term.Permalink = lctx.GetURL(term.Path)
		term.Children = d.parseTaxonomyTermsFromGroups(taxonomy, group.Children, term)
		term.Formats = d.ParseTaxonomyTermFormats(term, taxonomy.Lang)

		results = append(results, term)
	}
	SortTaxonomyTerms(results, lctx.GetTaxonomyConfig(taxonomy.Name, "sort_by").String(), false)
	return results
}

func (d *Processor) ParseTaxonomyTerms(taxonomy *Taxonomy, pages Pages, lang string) TaxonomyTerms {
	lctx := d.ctx.For(lang)

	groups := pages.FilterBy(lctx.GetTaxonomyConfig(taxonomy.Name, "term.filter_by").String()).
		OrderBy(lctx.GetTaxonomyConfig(taxonomy.Name, "term.sort_by").String()).
		GroupBy(taxonomy.Name)
	return d.parseTaxonomyTermsFromGroups(taxonomy, groups, nil)
}

func (d *Processor) ParseTaxonomyTermFormats(term *TaxonomyTerm, lang string) Formats {
	lctx := d.ctx.For(lang)

	customFormats := lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.formats").StringMap()

	v := viper.New()
	v.MergeConfigMap(customFormats)

	formats := make(Formats, 0)
	for name := range customFormats {
		customPath := v.GetString(name + ".path")
		customTemplate := v.GetString(name + ".template")
		if customTemplate == "" {
			customTemplate = lctx.Config.GetString("formats." + name + ".template")
		}
		if customPath == "" || customTemplate == "" {
			continue
		}

		format := &Format{
			Name:     name,
			Template: customTemplate,
		}

		format.Path = d.resolveTaxonomyTermPath(term, customPath)
		format.Permalink = lctx.GetURL(format.Path)
		formats = append(formats, format)
	}
	return formats
}

func (d *Processor) RenderTaxonomyTerm(term *TaxonomyTerm, tplset template.TemplateSet, writer core.Writer) error {
	lctx := d.ctx.For(term.Taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.template").String(),
		fmt.Sprintf("%s/single.html", term.Taxonomy.Name),
		"taxonomy_single.html",
	}
	if tpl := tplset.Lookup(lookups...); tpl != nil {
		d.ctx.Logger.Debugf("write taxonomy term [%s:%s] -> %s", term.Taxonomy.Name, term.GetFullName(), term.Path)

		pagers := d.PaginateBy(term.Pages.FilterBy(lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_filter_by").String()),
			lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate").Int(),
			term.Path,
			lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_path").String(),
			term.Taxonomy.Lang,
		)
		for _, pager := range pagers {
			if err := d.RenderTemplate(pager.Path, tpl, map[string]any{
				"paginator":     NewPaginator(pager, pagers),
				"term":          term,
				"taxonomy":      term.Taxonomy,
				"current_index": pager.PageNum,
				"current_lang":  term.Taxonomy.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}
	for _, format := range term.Formats {
		if tpl := tplset.Lookup(format.Template); tpl != nil {
			d.ctx.Logger.Debugf("write taxonomy term format [%s:%s] -> %s", term.Taxonomy.Name, term.GetFullName(), format.Path)
			if err := d.RenderTemplate(format.Path, tpl, map[string]any{
				"term":         term,
				"taxonomy":     term.Taxonomy,
				"current_lang": term.Taxonomy.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}

	for _, child := range term.Children {
		if err := d.RenderTaxonomyTerm(child, tplset, writer); err != nil {
			return err
		}
	}
	return nil
}
