package loader

import (
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/viper"
)

func (d *Loader) Taxonomies(lang string) types.Taxonomies {
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Loader) GetTaxonomy(name string, lang string) *types.Taxonomy {
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(name)
	return result
}

func (d *Loader) GetTaxonomyURL(name string, lang string) string {
	result := d.GetTaxonomy(name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Loader) GetTaxonomyTerms(name string, lang string) types.TaxonomyTerms {
	taxonomy := d.GetTaxonomy(name, lang)
	if taxonomy == nil {
		return nil
	}
	return taxonomy.Terms
}

func (d *Loader) GetTaxonomyTerm(taxonomyName string, name string, lang string) *types.TaxonomyTerm {
	set, ok := d.taxonomyTerms[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(fmt.Sprintf("%s:%s", taxonomyName, name))
	return result
}

func (d *Loader) GetTaxonomyTermURL(taxonomyName string, name string, lang string) string {
	result := d.GetTaxonomyTerm(taxonomyName, name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Loader) insertTaxonomies(page *types.Page) error {
	lctx := d.ctx.For(page.Lang)

	for taxonomyName := range lctx.Config.GetStringMap("taxonomies") {
		if taxonomyName == "_default" {
			continue
		}
		taxonomy := d.GetTaxonomy(taxonomyName, page.Lang)
		if taxonomy == nil {
			taxonomy = &types.Taxonomy{
				Name:  taxonomyName,
				Terms: make(types.TaxonomyTerms, 0),
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

			d.insertTaxonomy(taxonomy, page.Lang)
		}
		d.insertTaxonomyTerms(taxonomy, page)
	}
	return nil
}

func (d *Loader) insertTaxonomyTerms(taxonomy *types.Taxonomy, page *types.Page) error {
	lctx := d.ctx.For(page.Lang)

	var names []string
	if strings.HasPrefix(taxonomy.Name, "date:") {
		names = []string{page.Date.Format(taxonomy.Name[5:])}
	} else if result := page.FrontMatter.Get(taxonomy.Name); result != nil {
		switch value := result.(type) {
		case string:
			names = []string{value}
		case []string:
			names = value
		}
	}

	for _, name := range names {
		var (
			currentName string
			currentTerm *types.TaxonomyTerm
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

			// get_taxonomy_term("categories", "tech")
			// get_taxonomy_term("categories", "tech/develop")
			// get_taxonomy_term("categories", "tech/develop/go")
			term := d.GetTaxonomyTerm(taxonomy.Name, currentName, page.Lang)
			if term == nil {
				term = &types.TaxonomyTerm{
					Name:     part,
					Slug:     lctx.GetSlug(part),
					Pages:    make(types.Pages, 0),
					Parent:   currentTerm,
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
				term.Formats = d.parseTaxonomyTermFormats(term, page.Lang)

				if currentTerm == nil {
					// 只插入第一个分类值
					taxonomy.Terms = append(taxonomy.Terms, term)
				} else {
					currentTerm.Children = append(currentTerm.Children, term)
				}
				d.insertTaxonomyTerm(term, page.Lang)
			}
			term.Pages = append(term.Pages, page)
			currentTerm = term
		}
	}
	return nil
}

func (d *Loader) parseTaxonomyTermFormats(term *types.TaxonomyTerm, lang string) types.Formats {
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

func (d *Loader) insertTaxonomy(taxonomy *types.Taxonomy, lang string) {
	set, ok := d.taxonomies[lang]
	if !ok {
		set = newSet[*types.Taxonomy]()
		d.taxonomies[lang] = set
	}
	set.Add(taxonomy.Name, taxonomy)
}

func (d *Loader) insertTaxonomyTerm(term *types.TaxonomyTerm, lang string) {
	set, ok := d.taxonomyTerms[lang]
	if !ok {
		set = newSet[*types.TaxonomyTerm]()
		d.taxonomyTerms[lang] = set
	}
	set.Add(fmt.Sprintf("%s:%s", term.Taxonomy.Name, term.GetFullName()), term)
}
