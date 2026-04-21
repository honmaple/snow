package content

import (
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/site/content/types"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/viper"
)

type (
	Taxonomy      = types.Taxonomy
	Taxonomies    = types.Taxonomies
	TaxonomyTerm  = types.TaxonomyTerm
	TaxonomyTerms = types.TaxonomyTerms
)

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
		customPath := lctx.Config.GetString(fmt.Sprintf("taxonomies.%s.path", taxonomyName))
		if customPath == "" {
			customPath = lctx.Config.GetString("taxonomies._default.path")
		}
		outputPath := utils.StringReplace(customPath, map[string]string{
			"{taxonomy}": taxonomy.Name,
		})

		taxonomy.Path = lctx.GetRelURL(outputPath)
		taxonomy.Permalink = lctx.GetURL(taxonomy.Path)
		taxonomy.Terms = d.ParseTaxonomyTerms(taxonomy, pages, lang)

		taxonomies = append(taxonomies, taxonomy)
	}
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

	results := make(TaxonomyTerms, 0)
	resultMap := make(map[string]*TaxonomyTerm)
	for _, page := range pages {
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

					customPath := lctx.Config.GetString(fmt.Sprintf("taxonomies.%s.term_path", taxonomy.Name))
					if customPath == "" {
						customPath = lctx.Config.GetString("taxonomies._default.term_path")
					}
					outputPath := utils.StringReplace(customPath, map[string]string{
						"{taxonomy}":  taxonomy.Name,
						"{term}":      currentName,
						"{term:slug}": lctx.GetPathSlug(currentName),
					})

					term.Slug = lctx.GetSlug(part)
					term.Path = lctx.GetRelURL(outputPath)
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
	return results
}

func (d *ContentParser) ParseTaxonomyTermFormats(term *types.TaxonomyTerm, lang string) types.Formats {
	lctx := d.ctx.For(lang)

	customFormats := lctx.Config.GetStringMap(fmt.Sprintf("taxonomies.%s.formats", term.Taxonomy.Name))
	if len(customFormats) == 0 {
		customFormats = lctx.Config.GetStringMap("taxonomies._default.formats")
	}

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
		outputPath := utils.StringReplace(customPath, map[string]string{
			"{taxonomy}":  term.Taxonomy.Name,
			"{term}":      term.Name,
			"{term:slug}": term.Slug,
		})

		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}
