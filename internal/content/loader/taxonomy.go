package loader

import (
	"fmt"

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
	result := d.GetTaxonomy(name)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) GetTaxonomyTerm(taxonomyName string, name string) *types.TaxonomyTerm {
	result, _ := d.taxonomyTerms.Find(fmt.Sprintf("%s:%s", taxonomyName, name))
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
	for taxonomyName := range d.ctx.Config.GetStringMap("taxonomies") {
		if taxonomyName == "_default" {
			continue
		}
		taxonomy := d.GetTaxonomy(taxonomyName)
		if taxonomy == nil {
			taxonomy = &types.Taxonomy{
				Lang: page.Lang,
				Name: taxonomyName,
			}
			customPath := d.ctx.Config.GetString(fmt.Sprintf("taxonomies.%s.path", taxonomyName))
			if customPath == "" {
				customPath = d.ctx.Config.GetString("taxonomies._default.path")
			}
			outputPath := utils.StringReplace(customPath, map[string]string{
				"{taxonomy}": taxonomy.Name,
			})

			taxonomy.Path = d.ctx.GetRelURL(outputPath)
			taxonomy.Permalink = d.ctx.GetURL(taxonomy.Path)
			d.taxonomies.Add(taxonomyName, taxonomy)
		}
		d.insertTaxonomyTerm(page, taxonomy)
	}
	return nil
}

func (d *DiskLoader) insertTaxonomyTerm(page *types.Page, taxonomy *types.Taxonomy) error {
	values := page.FrontMatter.GetStringSlice(taxonomy.Name)
	for _, value := range values {
		term := d.GetTaxonomyTerm(taxonomy.Name, value)
		if term == nil {
			term = &types.TaxonomyTerm{
				Name:     value,
				Taxonomy: taxonomy,
				Pages:    make(types.Pages, 0),
			}
			customPath := d.ctx.Config.GetString(fmt.Sprintf("taxonomies.%s.term_path", taxonomy.Name))
			if customPath == "" {
				customPath = d.ctx.Config.GetString("taxonomies._default.term_path")
			}

			outputPath := utils.StringReplace(customPath, map[string]string{
				"{taxonomy}":  taxonomy.Name,
				"{term}":      term.Name,
				"{term:slug}": term.Slug,
			})
			term.Slug = d.ctx.GetSlug(value)
			term.Path = d.ctx.GetRelURL(outputPath)
			term.Permalink = d.ctx.GetURL(term.Path)
			term.Formats = d.loadTaxonomyTermFormats(term)

			d.taxonomyTerms.Add(fmt.Sprintf("%s:%s", taxonomy.Name, term.Name), term)
		}
		term.Pages = append(term.Pages, page)
	}
	return nil
}

func (d *DiskLoader) loadTaxonomyTermFormats(term *types.TaxonomyTerm) types.Formats {
	customFormats := d.ctx.Config.GetStringMap(fmt.Sprintf("taxonomies.%s.formats", term.Taxonomy.Name))
	if len(customFormats) == 0 {
		customFormats = d.ctx.Config.GetStringMap("taxonomies._default.formats")
	}

	v := viper.New()
	v.MergeConfigMap(customFormats)

	formats := make(types.Formats, 0)
	for name := range customFormats {
		customPath := v.GetString(name + ".path")
		customTemplate := v.GetString(name + ".template")
		if customTemplate == "" {
			customTemplate = d.ctx.Config.GetString("formats." + name + ".template")
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

		format.Path = d.ctx.GetRelURL(outputPath)
		format.Permalink = d.ctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}
