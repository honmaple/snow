package content

import (
	stdpath "path"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Section struct {
		*Node

		Path      string
		Permalink string

		Pages    Pages
		Assets   Assets
		Formats  Formats
		Children Sections
	}
	Sections []*Section
)

func (secs Sections) Len() int           { return len(secs) }
func (secs Sections) Swap(i, j int)      { secs[i], secs[j] = secs[j], secs[i] }
func (secs Sections) Less(i, j int) bool { return secs[i].Title < secs[j].Title }

func (d *Processor) resolveSectionPath(section *Section, customPath string) string {
	lctx := d.ctx.For(section.Lang)

	vars := map[string]string{
		"{lang}":         section.Lang,
		"{path}":         section.File.Dir,
		"{path:slug}":    lctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	}
	if section.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = section.Lang
	}
	return d.resolvePath(customPath, vars)
}

func (d *Processor) IsSection(fullpath string) ([]string, bool) {
	sectionFiles := d.findIndexFiles(fullpath, "_index")
	if len(sectionFiles) > 0 {
		return sectionFiles, true
	}
	return nil, false
}

func (d *Processor) ParseRootSection(fullpath string) (Sections, error) {
	sections := make(Sections, 0)
	sectionFiles := d.findIndexFiles(fullpath, "_index")
	for _, file := range sectionFiles {
		section, err := d.ParseSection(filepath.Join(fullpath, file))
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, nil
}

func (d *Processor) ParseSection(fullpath string) (*Section, error) {
	node, err := d.parseNode(fullpath, false)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	section := &Section{
		Node:     node,
		Pages:    make(Pages, 0),
		Assets:   make(Assets, 0),
		Children: make(Sections, 0),
	}
	if section.Title == "" {
		if section.File.Dir == "" {
			section.Title = "index"
		} else {
			section.Title = stdpath.Base(section.File.Dir)
		}
	}
	if section.Slug == "" {
		section.Slug = lctx.GetSlug(section.Title)
	}

	section.Assets = d.ParseSectionAssets(fullpath, section)
	section.Formats = d.ParseSectionFormats(section)

	customPath := d.resolveSectionPath(section, section.FrontMatter.GetString("path"))
	section.Path = lctx.GetRelURL(customPath)
	section.Permalink = lctx.GetURL(section.Path)
	return section, nil
}

func (d *Processor) ParseSectionAssets(fullpath string, section *Section) Assets {
	lctx := d.ctx.For(section.Lang)

	assets := make(Assets, 0)
	for _, file := range section.FrontMatter.GetStringSlice("assets") {
		asset := &Asset{
			File: file,
		}
		customPath := section.FrontMatter.GetString("asset_path")
		outputPath := d.resolveSectionPath(section, customPath)
		asset.Path = lctx.GetRelURL(outputPath)
		asset.Permalink = lctx.GetURL(asset.Path)

		assets = append(assets, asset)
	}
	return assets
}

func (d *Processor) ParseSectionFormats(section *Section) Formats {
	lctx := d.ctx.For(section.Lang)

	formats := make(Formats, 0)
	for name := range section.FrontMatter.GetStringMap("formats") {
		customPath := section.FrontMatter.GetString(name + ".path")
		customTemplate := section.FrontMatter.GetString(name + ".template")
		if customTemplate == "" {
			customTemplate = d.ctx.Config.GetString("formats." + name + ".template")
		}
		if customPath == "" || customTemplate == "" {
			continue
		}

		format := &Format{
			Name:     name,
			Template: customTemplate,
		}
		outputPath := d.resolveSectionPath(section, customPath)
		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *Processor) RenderSection(section *Section, tplset template.TemplateSet, writer core.Writer) error {
	d.ctx.Logger.Debugf("write section [%s] -> %s", section.File.Path, section.Path)

	customTemplate := section.FrontMatter.GetString("template")

	lookups := []string{
		customTemplate,
		"section.html",
	}
	// 首页content/_index.md
	if section.File.Dir == "" {
		lookups = []string{
			customTemplate,
			"index.html",
			"section.html",
		}
	}

	if tpl := tplset.Lookup(lookups...); tpl != nil {
		for _, por := range section.Pages.
			FilterBy(
				section.FrontMatter.GetString("paginate_filter_by"),
			).
			Paginate(
				section.FrontMatter.GetInt("paginate"),
				section.Path,
				section.FrontMatter.GetString("paginate_path"),
			) {
			if err := d.RenderTemplate(por.Path, tpl, map[string]any{
				"section":       section,
				"paginator":     por,
				"pages":         section.Pages,
				"current_lang":  section.Lang,
				"current_index": por.PageNum,
			}, writer); err != nil {
				return err
			}
		}
	}

	for _, format := range section.Formats {
		if tpl := tplset.Lookup(format.Template); tpl != nil {
			if err := d.RenderTemplate(format.Path, tpl, map[string]any{
				"section":      section,
				"pages":        section.Pages,
				"current_lang": section.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}
	// for _, asset := range section.Assets {
	//	if err := r.renderAsset(asset); err != nil {
	//		return err
	//	}
	// }
	return nil
}
