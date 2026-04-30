package content

import (
	"strings"

	"github.com/honmaple/snow/internal/site/content/types"
	"github.com/spf13/viper"
)

type (
	Taxonomy      = types.Taxonomy
	Taxonomies    = types.Taxonomies
	TaxonomyTerm  = types.TaxonomyTerm
	TaxonomyTerms = types.TaxonomyTerms
)

func (d *ContentParser) sortTaxonomyTermsPages(terms types.TaxonomyTerms, key string) {
	for _, term := range terms {
		term.Pages.SortBy(key)

		d.sortTaxonomyTermsPages(term.Children, key)
	}
}

func (d *ContentParser) ParseTaxonomies(pages types.Pages, lang string) types.Taxonomies {
	lctx := d.ctx.For(lang)

	taxonomies := make(types.Taxonomies, 0)
	for taxonomyName := range lctx.Config.GetStringMap("taxonomies") {
		if taxonomyName == "_default" {
			continue
		}
		taxonomy := &types.Taxonomy{
			Lang: lang,
			Name: taxonomyName,
		}
		customPath := lctx.GetTaxonomyConfig(taxonomyName, "path").String()
		taxonomy.Path = lctx.GetRelURL(d.resolveTaxonomyPath(taxonomy, customPath))
		taxonomy.Permalink = lctx.GetURL(taxonomy.Path)
		taxonomy.Terms = d.ParseTaxonomyTerms(taxonomy, pages, lang)

		taxonomies = append(taxonomies, taxonomy)
	}
	taxonomies.SortBy("weight desc")
	return taxonomies
}

func (d *ContentParser) ParseTaxonomyTerms(taxonomy *types.Taxonomy, pages types.Pages, lang string) types.TaxonomyTerms {
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

	filter := types.FilterExpr(lctx.GetTaxonomyConfig(taxonomy.Name, "term.filter_by").String())

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

func (d *ContentParser) ParseTaxonomyTermFormats(term *types.TaxonomyTerm, lang string) types.Formats {
	lctx := d.ctx.For(lang)

	customFormats := lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.formats").StringMap()

	v := viper.New()
	v.MergeConfigMap(customFormats)

	formats := make(types.Formats, 0)
	for name := range customFormats {
		customPath := v.GetString(name + ".path")
		customTemplate := v.GetString(name + ".template")
		if customTemplate == "" {
			customTemplate = lctx.Config.GetString("formats." + name + ".template")
		}
		if customPath == "" || customTemplate == "" {
			continue
		}

		format := &types.Format{
			Name:     name,
			Template: customTemplate,
		}

		format.Path = lctx.GetRelURL(d.resolveTaxonomyTermPath(term, customPath))
		format.Permalink = lctx.GetURL(format.Path)
		formats = append(formats, format)
	}
	return formats
}

func (d *ContentParser) resolveTaxonomyPath(taxonomy *types.Taxonomy, customPath string) string {
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

func (d *ContentParser) resolveTaxonomyTermPath(term *types.TaxonomyTerm, customPath string) string {
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
