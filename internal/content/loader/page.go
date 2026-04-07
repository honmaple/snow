package loader

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
)

func (d *DiskLoader) Pages() types.Pages {
	return d.pages.List()
}

func (d *DiskLoader) GetPage(path string) *types.Page {
	result, _ := d.pages.Find(path)
	return result
}

func (d *DiskLoader) GetPageURL(path string) string {
	result, _ := d.pages.Find(path)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) getPagePath(page *types.Page, customPath string) string {
	filename := utils.FileBaseName(page.File)
	if filename == "index" {
		filename = filepath.Base(filepath.Dir(page.File))
	}

	path := strings.TrimPrefix(filepath.ToSlash(filepath.Dir(page.File)), d.ctx.GetContentDir()+"/")
	return utils.StringReplace(customPath, map[string]string{
		"{date:%Y}":   page.Date.Format("2006"),
		"{date:%m}":   page.Date.Format("01"),
		"{date:%d}":   page.Date.Format("02"),
		"{date:%H}":   page.Date.Format("15"),
		"{slug}":      page.Slug,
		"{filename}":  filename,
		"{path}":      path,
		"{path:slug}": d.ctx.GetPathSlug(path),
	})
}

func (d *DiskLoader) insertPage(file string, isDir bool) error {
	result, err := d.parser.Parse(file)
	if err != nil {
		return err
	}

	meta := types.NewFrontMatter(result.FrontMatter)

	page := &types.Page{
		FrontMatter: meta,
		File:        strings.TrimPrefix(file, d.ctx.GetContentDir()+"/"),
		Title:       meta.GetString("title"),
		Summary:     meta.GetString("summary"),
		Content:     result.Content,
		Slug:        meta.GetString("slug"),
		Draft:       meta.GetBool("draft"),
		Date:        meta.GetTime("date"),
		Modified:    meta.GetTime("modified"),
	}
	page.Lang = d.findLang(filepath.Base(file), meta.GetString("lang"))

	lctx := d.ctx.For(page.Lang)

	if page.Summary == "" {
		page.Summary = lctx.GetSummary(result.Content)
	}

	if page.Title == "" {
		filename := utils.FileBaseName(page.File)
		if filename == "index" {
			filename = filepath.Base(filepath.Dir(page.File))
		}
		page.Title = filename
	}
	if page.Slug == "" {
		page.Slug = lctx.GetSlug(page.Title)
	}
	if page.Date.IsZero() {
		page.Date = time.Now()
	}
	if page.Modified.IsZero() {
		page.Modified = page.Date
	}

	// 添加附属资源
	if isDir {
		assets := make([]string, 0)

		root := filepath.Dir(file)
		if err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == root || path == file || info.IsDir() {
				return nil
			}
			assets = append(assets, path)
			return nil
		}); err != nil {
			return err
		}

		page.Assets = make([]*types.Asset, len(assets))
		for i, assetFile := range assets {
			asset := &types.Asset{
				File: assetFile,
			}
			asset.Path = lctx.GetRelURL(meta.GetString("asset_path"))
			asset.Permalink = lctx.GetURL(asset.Path)

			page.Assets[i] = asset
		}
	}
	sectionPath := filepath.Dir(page.File)
	if isDir {
		sectionPath = filepath.Dir(sectionPath)
	}

	customPath := meta.GetString("path")
	// 如果自定义path为空，则从配置中获取
	if customPath == "" {
		// 获取配置 content/posts/linux/emacs/page-01.md
		// 查找顺序: posts/linux/emacs -> posts/linux -> posts -> _default
		customPath = lctx.GetSectionConfig(sectionPath, "page_path")
	}
	outputPath := d.getPagePath(page, customPath)

	page.Path = lctx.GetRelURL(outputPath)
	page.Permalink = lctx.GetURL(page.Path)
	page.RelPermalink = page.Path

	section := d.findSection(sectionPath)
	section.Pages = append(section.Pages, page)

	d.insertPageFormats(page)

	if d.hook != nil {
		page = d.hook.HandlePage(page)
		if page == nil {
			return nil
		}
	}
	d.insertTaxonomies(page)

	d.pages.Add(page.File, page)
	return nil
}

func (d *DiskLoader) insertPageFormats(page *types.Page) error {
	customFormats := page.FrontMatter.GetStringMap("formats")

	formats := make(types.Formats, 0)
	for name := range customFormats {
		customPath := page.FrontMatter.GetString(name + ".path")
		customTemplate := page.FrontMatter.GetString(name + ".template")
		// 从全局配置获取
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

		outputPath := d.getPagePath(page, customPath)

		format.Path = d.ctx.GetRelURL(outputPath)
		format.Permalink = d.ctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	page.Formats = formats
	return nil
}
