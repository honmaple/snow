package loader

import (
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/content/parser"
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
	result := d.GetSection(name)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) getSectionPath(section *types.Section, customPath string) string {
	return utils.StringReplace(customPath, map[string]string{
		"{path}":         filepath.Dir(section.File),
		"{path:slug}":    d.ctx.GetPathSlug(filepath.Dir(section.File)),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	})
}

func (d *DiskLoader) insertSection(file string, isHome bool) error {
	if isHome {
		var result *parser.Result

		sectionFiles := d.findFiles(file, "_index.*")
		if len(sectionFiles) > 0 {
			r, err := d.parser.Parse(file)
			if err != nil {
				return err
			}
			result = r
		} else {
			result = &parser.Result{
				FrontMatter: make(map[string]any),
			}
		}

		meta := types.NewFrontMatter(result.FrontMatter)

		section := &types.Section{
			IsHome:      isHome,
			FrontMatter: meta,
			File:        strings.TrimPrefix(file, d.ctx.GetContentDir()+"/"),
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

		d.insertSectionFormats(section)

		d.sections.Add("@home", section)
		return nil
	}

	result, err := d.parser.Parse(file)
	if err != nil {
		return err
	}

	meta := types.NewFrontMatter(result.FrontMatter)
	section := &types.Section{
		FrontMatter: meta,
		File:        strings.TrimPrefix(file, d.ctx.GetContentDir()+"/"),
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
		section.Summary = d.ctx.GetSummary(result.Content)
	}
	if section.Title == "" {
		section.Title = filepath.Base(filepath.Dir(file))
	}
	if section.Slug == "" {
		section.Slug = d.ctx.GetSlug(section.Title)
	}

	customPath := meta.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		customPath = d.ctx.GetSectionConfig(file, "path")
	}

	outputPath := d.getSectionPath(section, customPath)
	section.Path = d.ctx.GetRelURL(outputPath)
	section.Permalink = d.ctx.GetURL(section.Path)
	section.RelPermalink = section.Path

	d.insertSectionFormats(section)

	if d.hook != nil {
		section = d.hook.HandleSection(section)
		if section == nil {
			return nil
		}
	}

	d.sections.Add(section.File, section)
	return nil
}

func (d *DiskLoader) insertSectionAsset(file string) error {
	section := d.findSection(filepath.Dir(file))

	asset := &types.Asset{
		File: file,
	}
	customPath := section.FrontMatter.GetString("asset_path")
	if customPath == "" {
		customPath = d.ctx.GetSectionConfig(section.File, "asset_path")
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

func (d *DiskLoader) insertSectionFormats(section *types.Section) error {
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
	section.Formats = formats
	return nil
}
