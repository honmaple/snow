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

		Slug      string
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

func (terms TaxonomyTerms) SortBy(key string) {
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
	for _, term := range terms {
		term.Children.SortBy(key)
	}
}

func (terms TaxonomyTerms) OrderBy(key string) TaxonomyTerms {
	newTerms := make(TaxonomyTerms, len(terms))
	copy(newTerms, terms)

	newTerms.SortBy(key)
	return newTerms
}

func (d *Processor) sortTaxonomyTermsPages(terms TaxonomyTerms, key string) {
	for _, term := range terms {
		term.Pages.SortBy(key)

		d.sortTaxonomyTermsPages(term.Children, key)
	}
}

func (d *Processor) resolveTaxonomyTermPath(term *TaxonomyTerm, customPath string) string {
	lctx := d.ctx.For(term.Taxonomy.Lang)

	vars := map[string]string{
		"{lang}":      term.Taxonomy.Lang,
		"{taxonomy}":  term.Taxonomy.Name,
		"{term}":      term.GetFullName(),
		"{term:slug}": lctx.GetPathSlug(term.GetFullName()),
	}
	if term.Taxonomy.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = term.Taxonomy.Lang
	}
	return d.resolvePath(customPath, vars)
}

func (d *Processor) ParseTaxonomyTerms(taxonomy *Taxonomy, pages Pages, lang string) TaxonomyTerms {
	var groupf func(*Page) []string

	if strings.HasPrefix(taxonomy.Name, "date:") {
		format := taxonomy.Name[5:]
		groupf = func(page *Page) []string {
			return []string{page.Date.Format(format)}
		}
	} else {
		groupf = func(page *Page) []string {
			value := page.FrontMatter.Get(taxonomy.Name)
			switch v := value.(type) {
			case string:
				return []string{v}
			case []string:
				return v
			default:
				return nil
			}
		}
	}

	lctx := d.ctx.For(lang)

	filter := FilterExpr(lctx.GetTaxonomyConfig(taxonomy.Name, "term.filter_by").String())

	results := make(TaxonomyTerms, 0)
	resultMap := make(map[string]*TaxonomyTerm)
	for _, page := range pages {
		if filter != nil && !filter(page) {
			continue
		}
		for _, name := range groupf(page) {
			var (
				currentTerm *TaxonomyTerm
				currentName string
			)
			for part := range strings.SplitSeq(strings.Trim(name, "/"), "/") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if currentName == "" {
					currentName = part
				} else {
					currentName += "/" + part
				}

				term, ok := resultMap[currentName]
				if !ok {
					term = &TaxonomyTerm{
						Name:     part,
						Pages:    make(Pages, 0),
						Parent:   currentTerm,
						Children: make(TaxonomyTerms, 0),
						Taxonomy: taxonomy,
					}
					term.Slug = lctx.GetSlug(part)

					customPath := lctx.GetTaxonomyConfig(taxonomy.Name, "term.path").String()

					term.Path = lctx.GetRelURL(d.resolveTaxonomyTermPath(term, customPath))
					term.Permalink = lctx.GetURL(term.Path)
					term.Formats = d.ParseTaxonomyTermFormats(term, page.Lang)

					resultMap[currentName] = term

					if currentTerm == nil {
						results = append(results, term)
					} else {
						currentTerm.Children = append(currentTerm.Children, term)
					}
				}
				term.Pages = append(term.Pages, page)

				currentTerm = term
			}
		}
	}
	d.sortTaxonomyTermsPages(results, lctx.GetTaxonomyConfig(taxonomy.Name, "term.sort_by").String())

	results.SortBy(lctx.GetTaxonomyConfig(taxonomy.Name, "sort_by").String())
	return results
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

		format.Path = lctx.GetRelURL(d.resolveTaxonomyTermPath(term, customPath))
		format.Permalink = lctx.GetURL(format.Path)
		formats = append(formats, format)
	}
	return formats
}

func (d *Processor) RenderTaxonomyTerm(term *TaxonomyTerm, tplset template.TemplateSet, writer core.Writer) error {
	d.ctx.Logger.Debugf("write taxonomy term [%s:%s] -> %s", term.Taxonomy.Name, term.GetFullName(), term.Path)

	lctx := d.ctx.For(term.Taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.template").String(),
		fmt.Sprintf("%s/taxonomy.terms.html", term.Taxonomy.Name),
		"taxonomy.terms.html",
	}
	if tpl := tplset.Lookup(lookups...); tpl != nil {
		for _, por := range term.Pages.
			FilterBy(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_filter_by").String(),
			).
			Paginate(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate").Int(),
				term.Path,
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_path").String(),
			) {
			if err := d.RenderTemplate(por.Path, tpl, map[string]any{
				"term":          term,
				"pages":         term.Pages,
				"taxonomy":      term.Taxonomy,
				"paginator":     por,
				"current_path":  por.Path,
				"current_index": por.PageNum,
				"current_lang":  term.Taxonomy.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}
	for _, format := range term.Formats {
		if tpl := tplset.Lookup(format.Template); tpl != nil {
			if err := d.RenderTemplate(format.Path, tpl, map[string]any{
				"term":         term,
				"pages":        term.Pages,
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
