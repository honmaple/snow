package content

import (
	stdpath "path"
	"path/filepath"

	"github.com/honmaple/snow/internal/site/content/types"
	"github.com/honmaple/snow/internal/utils"
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
	langs := make(map[string]bool)

	sections := make(types.Sections, 0)
	sectionFiles := d.findIndexFiles(fullpath, "_index")
	for _, file := range sectionFiles {
		section, err := d.ParseSection(filepath.Join(fullpath, file))
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)

		langs[section.Lang] = true
	}

	for _, lang := range d.ctx.GetAllLanguages() {
		if langs[lang] {
			continue
		}
		section := &types.Section{
			Node: &types.Node{
				File: &types.File{
					Path:     "_index.md",
					Dir:      "",
					Ext:      ".md",
					Name:     "_index.md",
					BaseName: "_index",
				},
				Lang:        lang,
				Slug:        "index",
				Title:       "index",
				FrontMatter: types.NewFrontMatter(nil),
			},
			Pages:  make(types.Pages, 0),
			Assets: make([]*types.Asset, 0),
		}
		if lang != d.ctx.GetDefaultLanguage() {
			section.File.LanguageName = lang
			section.Path = "/" + lang + "/index.html"
		} else {
			section.Path = "/index.html"
		}
		section.Permalink = d.ctx.GetURL(section.Path)

		sections = append(sections, section)
	}
	return sections, nil
}

func (d *ContentParser) ParseSection(fullpath string) (*types.Section, error) {
	node, err := d.parseNode(fullpath)
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

	section.Assets = d.parseSectionAssets(fullpath, section)
	section.Formats = d.ParseSectionFormats(section)

	customPath := section.FrontMatter.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		customPath = lctx.GetSectionConfig(section.File.Dir, "path")
	}
	section.Path = lctx.GetRelURL(d.parseSectionPath(section, customPath))
	section.Permalink = lctx.GetURL(section.Path)
	return section, nil
}

func (d *ContentParser) parseSectionAssets(fullpath string, section *types.Section) types.Assets {
	lctx := d.ctx.For(section.Lang)

	assets := make(types.Assets, 0)
	for _, file := range section.FrontMatter.GetStringSlice("assets") {
		asset := &types.Asset{
			File: file,
		}
		customPath := section.FrontMatter.GetString("asset_path")
		if customPath == "" {
			customPath = lctx.GetSectionConfig(section.File.Path, "asset_path")
		}
		outputPath := utils.StringReplace(customPath, map[string]string{
			"{section}":      section.Title,
			"{section:slug}": section.Slug,
		})
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
		outputPath := utils.StringReplace(customPath, map[string]string{
			"{section}":      section.Title,
			"{section:slug}": section.Slug,
		})

		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *ContentParser) parseSectionPath(section *types.Section, customPath string) string {
	lctx := d.ctx.For(section.Lang)
	return utils.StringReplace(customPath, map[string]string{
		"{path}":         section.File.Dir,
		"{path:slug}":    lctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
}
