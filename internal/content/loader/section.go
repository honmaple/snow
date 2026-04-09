package loader

import (
	stdpath "path"
	"path/filepath"

	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
	"strings"
)

func (d *DiskLoader) Sections() types.Sections {
	return d.sections.List()
}

func (d *DiskLoader) GetSection(name string) *types.Section {
	result, _ := d.sections.Find(name)
	return result
}

func (d *DiskLoader) GetSectionURL(name string) string {
	result := d.GetSection(name)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) getSectionPath(section *types.Section, customPath string) string {
	return utils.StringReplace(customPath, map[string]string{
		"{path}":         section.File.Dir,
		"{path:slug}":    d.ctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
}

func (d *DiskLoader) insertIndexSection(fullpath string) error {
	var result *parser.Result

	sectionFiles := d.findFiles(fullpath, "_index.*")
	if len(sectionFiles) > 0 {
		r, err := d.parser.Parse(sectionFiles[0])
		if err != nil {
			return err
		}
		result = r
	} else {
		result = &parser.Result{
			FrontMatter: make(map[string]any),
		}
	}

	file, err := d.loadSectionFile(fullpath)
	if err != nil {
		return err
	}

	meta := types.NewFrontMatter(result.FrontMatter)

	section := &types.Section{
		IsHome:      true,
		FrontMatter: meta,
		File:        file,
		Title:       meta.GetString("title"),
		Content:     result.Content,
		RawContent:  result.RawContent,
		Summary:     meta.GetString("summary"),
		Slug:        meta.GetString("slug"),
		Lang:        meta.GetString("lang"),
		Draft:       meta.GetBool("draft"),
		Pages:       make(types.Pages, 0),
		Assets:      make([]*types.Asset, 0),
	}
	if section.Summary == "" {
		section.Summary = d.ctx.GetSummary(result.Content)
	}
	if section.Title == "" {
		section.Title = "index"
	}
	if section.Slug == "" {
		section.Slug = d.ctx.GetSlug(section.Title)
	}

	section.Path = d.ctx.GetRelURL("/index.html")
	section.Permalink = d.ctx.GetURL(section.Path)
	section.RelPermalink = section.Path

	d.loadSectionFormats(section)

	d.sections.Add("@home", section)
	return nil
}

func (d *DiskLoader) insertSection(fullpath string) error {
	result, err := d.parser.Parse(fullpath)
	if err != nil {
		return err
	}

	file, err := d.loadSectionFile(fullpath)
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
		section.Title = stdpath.Base(section.File.Dir)
	}
	if section.Slug == "" {
		section.Slug = lctx.GetSlug(section.Title)
	}

	customPath := meta.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		customPath = lctx.GetSectionConfig(fullpath, "path")
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

	d.sections.Add(section.File.Path, section)
	return nil
}

func (d *DiskLoader) insertSectionAsset(fullpath string) error {
	section := d.findSection(filepath.Dir(fullpath))

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

func (d *DiskLoader) loadSectionFile(fullpath string) (*types.File, error) {
	relPath, err := filepath.Rel(d.ctx.GetContentDir(), fullpath)
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	ext := stdpath.Ext(relPath)
	nameWithExt := stdpath.Base(relPath)
	nameWithoutExt := strings.TrimSuffix(nameWithExt, ext)

	dir := stdpath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	return &types.File{
		Path:     relPath,
		Dir:      dir,
		Ext:      ext,
		Name:     nameWithExt,
		BaseName: nameWithoutExt,
	}, nil
}

func (d *DiskLoader) loadSectionFormats(section *types.Section) types.Formats {
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

		format.Path = d.ctx.GetRelURL(outputPath)
		format.Permalink = d.ctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}
