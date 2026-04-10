package loader

import (
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/viper"
)

func (d *DiskLoader) Taxonomies() types.Taxonomies {
	return d.taxonomies.List()
}

func (d *DiskLoader) GetTaxonomy(path string) *types.Taxonomy {
	result, _ := d.taxonomies.Find(path)
	return result
}

func (d *DiskLoader) GetTaxonomyURL(name string) string {
	result, ok := d.taxonomies.Find(name)
	if !ok {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) GetTaxonomyTerms(taxonomyName string) types.TaxonomyTerms {
	result, ok := d.taxonomies.Find(taxonomyName)
	if !ok {
		return nil
	}
	return result.Terms
}

func (d *DiskLoader) GetTaxonomyTerm(taxonomyName string, name string) *types.TaxonomyTerm {
	result, _ := d.taxonomyTermMap[fmt.Sprintf("%s:%s", taxonomyName, name)]
	return result
}

func (d *DiskLoader) GetTaxonomyTermURL(taxonomyName string, name string) string {
	result := d.GetTaxonomyTerm(taxonomyName, name)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) insertTaxonomies(page *types.Page) error {
	lctx := d.ctx.For(page.Lang)
	for taxonomyName := range d.ctx.Config.GetStringMap("taxonomies") {
		if taxonomyName == "_default" {
			continue
		}
		taxonomy := d.GetTaxonomy(taxonomyName)
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
			d.taxonomies.Add(taxonomyName, taxonomy)
		}
		d.insertTaxonomyTerm(page, taxonomy)
	}
	return nil
}

func (d *DiskLoader) insertTaxonomyTerm(page *types.Page, taxonomy *types.Taxonomy) error {
	lctx := d.ctx.For(page.Lang)
	for _, name := range page.FrontMatter.GetStringSlice(taxonomy.Name) {
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

			// get_taxonomy("categories", "tech")
			// get_taxonomy("categories", "tech/develop")
			// get_taxonomy("categories", "tech/develop/go")
			term := d.GetTaxonomyTerm(taxonomy.Name, currentName)
			if term == nil {
				term = &types.TaxonomyTerm{
					Name:     name,
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
					"{term}":      term.Name,
					"{term:slug}": term.Slug,
				})
				term.Slug = lctx.GetSlug(name)
				term.Path = lctx.GetRelURL(outputPath)
				term.Permalink = lctx.GetURL(term.Path)
				term.Formats = d.loadTaxonomyTermFormats(term, page.Lang)

				if currentTerm != nil {
					currentTerm.Children = append(currentTerm.Children, term)
					// 只插入第一个分类值
					taxonomy.Terms = append(taxonomy.Terms, term)
				}
				d.taxonomyTermMap[fmt.Sprintf("%s:%s", taxonomy.Name, term.FullName)] = term
			}
			term.Pages = append(term.Pages, page)
			currentTerm = term
		}
	}
	return nil
}

func (d *DiskLoader) loadTaxonomyTermFormats(term *types.TaxonomyTerm, lang string) types.Formats {
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
