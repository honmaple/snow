package content

import (
	stdpath "path"
	"path/filepath"

	"github.com/honmaple/snow/internal/site/content/types"
)

type (
	Section  = types.Section
	Sections = types.Sections
)

func (d *ContentParser) IsSection(fullpath string) ([]string, bool) {
	sectionFiles := d.findIndexFiles(fullpath, "_index")
	if len(sectionFiles) > 0 {
		return sectionFiles, true
	}
	return nil, false
}

func (d *ContentParser) ParseRootSection(fullpath string) (types.Sections, error) {
	sections := make(types.Sections, 0)
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

func (d *ContentParser) ParseSection(fullpath string) (*types.Section, error) {
	node, err := d.parseNode(fullpath, false)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	section := &types.Section{
		Node:     node,
		Pages:    make(types.Pages, 0),
		Assets:   make(types.Assets, 0),
		Children: make(types.Sections, 0),
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

func (d *ContentParser) ParseSectionAssets(fullpath string, section *types.Section) types.Assets {
	lctx := d.ctx.For(section.Lang)

	assets := make(types.Assets, 0)
	for _, file := range section.FrontMatter.GetStringSlice("assets") {
		asset := &types.Asset{
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

func (d *ContentParser) ParseSectionFormats(section *types.Section) types.Formats {
	lctx := d.ctx.For(section.Lang)

	formats := make(types.Formats, 0)
	for name := range section.FrontMatter.GetStringMap("formats") {
		customPath := section.FrontMatter.GetString(name + ".path")
		customTemplate := section.FrontMatter.GetString(name + ".template")
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
		outputPath := d.resolveSectionPath(section, customPath)
		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *ContentParser) resolveSectionPath(section *types.Section, customPath string) string {
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
