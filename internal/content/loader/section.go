package loader

import (
	stdpath "path"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
)

func (d *Loader) Sections(lang string) types.Sections {
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Loader) GetSection(name string, lang string) *types.Section {
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(name)
	return result
}

func (d *Loader) GetSectionURL(name string, lang string) string {
	result := d.GetSection(name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Loader) findSection(dir string, lang string) *types.Section {
	currentDir := dir
	for {
		if currentDir == "" || currentDir == "." {
			break
		}
		result := d.GetSection(currentDir, lang)
		if result != nil {
			return result
		}
		currentDir = stdpath.Dir(currentDir)
	}
	return d.GetSection("/", lang)
}

func (d *Loader) parseSectionPath(section *types.Section, customPath string) string {
	lctx := d.ctx.For(section.Lang)
	return utils.StringReplace(customPath, map[string]string{
		"{path}":         section.File.Dir,
		"{path:slug}":    lctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
}

func (d *Loader) parseSection(fullpath string) (*types.Section, error) {
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

	customPath := section.FrontMatter.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		customPath = lctx.GetSectionConfig(section.File.Dir, "path")
	}
	section.Assets = d.parseSectionAssets(fullpath, section)
	section.Formats = d.parseSectionFormats(section)

	section.Path = lctx.GetRelURL(d.parseSectionPath(section, customPath))
	section.Permalink = lctx.GetURL(section.Path)
	return section, nil
}

func (d *Loader) parseSectionAssets(fullpath string, section *types.Section) types.Assets {
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

func (d *Loader) parseSectionFormats(section *types.Section) types.Formats {
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

func (d *Loader) insertSection(section *types.Section) {
	set, ok := d.sections[section.Lang]
	if !ok {
		set = newSet[*types.Section]()

		d.sections[section.Lang] = set
	}
	if section.File.Dir == "" {
		set.Add("/", section)
	} else {
		set.Add(section.File.Dir, section)
	}
}

func (d *Loader) insertRootSection() {
	for _, lang := range d.ctx.GetAllLanguages() {
		section := d.GetSection("/", lang)
		if section != nil {
			continue
		}
		section = &types.Section{
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

		d.insertSection(section)
	}
}

func (d *Loader) insertSectionByPath(fullpath string) {
	section, err := d.parseSection(fullpath)
	if err != nil {
		d.ctx.Logger.Warnf("parse root section %s err: %s", fullpath, err.Error())
		return
	}
	d.insertSection(section)
}
