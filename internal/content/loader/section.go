package loader

import (
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
)

func (d *DiskLoader) Sections() types.Sections {
	return d.sections.List()
}

func (d *DiskLoader) GetSection(name string) *types.Section {
	result, _ := d.sections.Find(name)
	return result
}

func (d *DiskLoader) GetSectionURL(name string) string {
	result, ok := d.sections.Find(name)
	if !ok {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) findSection(dir string) *types.Section {
	currentDir := dir
	for {
		if currentDir == "" || currentDir == "." {
			break
		}
		result, ok := d.sections.Find(currentDir)
		if ok {
			return result
		}
		currentDir = stdpath.Dir(currentDir)
	}
	result, _ := d.sections.Find("/")
	return result
}

func (d *DiskLoader) getSectionPath(section *types.Section, customPath string) string {
	lctx := d.ctx.For(section.Lang)
	return utils.StringReplace(customPath, map[string]string{
		"{path}":         section.File.Dir,
		"{path:slug}":    lctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
}

func (d *DiskLoader) insertSection(fullpath string, isRoot bool) error {
	result, err := d.parser.Parse(fullpath)
	if err != nil {
		return err
	}

	file, err := d.loadFile(fullpath)
	if err != nil {
		return err
	}
	meta := types.NewFrontMatter(result.FrontMatter)

	lang := meta.GetString("lang")
	if lang == "" {
		langExt := stdpath.Ext(file.BaseName)
		if langExt != "" {
			lang = strings.TrimPrefix(langExt, ".")
		}
	}
	if lang == "" || !d.ctx.IsValidLanguage(lang) {
		lang = d.ctx.GetDefaultLanguage()
	}
	if ext := "." + lang; strings.HasSuffix(file.BaseName, ext) {
		file.BaseName = strings.TrimSuffix(file.BaseName, ext)
		file.LanguageName = lang
	}
	lctx := d.ctx.For(lang)

	section := &types.Section{
		FrontMatter: meta,
		File:        file,
		Title:       meta.GetString("title"),
		Content:     result.Content,
		Summary:     meta.GetString("summary"),
		Slug:        meta.GetString("slug"),
		Lang:        meta.GetString("lang"),
		Draft:       meta.GetBool("draft"),
		Pages:       make(types.Pages, 0),
		Assets:      make([]*types.Asset, 0),
	}
	if section.Summary == "" {
		section.Summary = lctx.GetSummary(result.Content)
	}
	if section.Title == "" {
		if isRoot {
			section.Title = "index"
		} else {
			section.Title = stdpath.Base(section.File.Dir)
		}
	}
	if section.Slug == "" {
		section.Slug = lctx.GetSlug(section.Title)
	}

	customPath := meta.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		customPath = lctx.GetSectionConfig(section.File.Dir, "path")
	}

	outputPath := d.getSectionPath(section, customPath)
	section.Path = lctx.GetRelURL(outputPath)
	section.Permalink = lctx.GetURL(section.Path)
	section.RelPermalink = section.Path
	section.Formats = d.loadSectionFormats(section)

	if d.hook != nil {
		section = d.hook.HandleSection(section)
		if section == nil {
			return nil
		}
	}

	if isRoot {
		d.sections.Add("/", section)
	} else {
		d.sections.Add(section.File.Dir, section)
	}
	return nil
}

func (d *DiskLoader) insertSectionAsset(fullpath string) error {
	file, err := d.loadFile(fullpath)
	if err != nil {
		return err
	}

	section := d.findSection(file.Dir)

	asset := &types.Asset{
		File: fullpath,
	}
	customPath := section.FrontMatter.GetString("asset_path")
	if customPath == "" {
		customPath = d.ctx.GetSectionConfig(section.File.Path, "asset_path")
	}
	outputPath := utils.StringReplace(customPath, map[string]string{
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
	asset.Path = d.ctx.GetRelURL(outputPath)
	asset.Permalink = d.ctx.GetURL(asset.Path)

	section.Assets = append(section.Assets, asset)
	return nil
}

func (d *DiskLoader) loadSectionFormats(section *types.Section) types.Formats {
	lctx := d.ctx.For(section.Lang)

	customFormats := section.FrontMatter.GetStringMap("formats")

	formats := make(types.Formats, 0)
	for name := range customFormats {
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
