package loader

import (
	"io/fs"
	"os"
	stdpath "path"
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
	result, ok := d.pages.Find(path)
	if !ok {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) getPagePath(page *types.Page, customPath string) string {
	return utils.StringReplace(customPath, map[string]string{
		"{date:%Y}":   page.Date.Format("2006"),
		"{date:%m}":   page.Date.Format("01"),
		"{date:%d}":   page.Date.Format("02"),
		"{date:%H}":   page.Date.Format("15"),
		"{slug}":      page.Slug,
		"{filename}":  page.File.BaseName,
		"{path}":      page.File.Dir,
		"{path:slug}": d.ctx.GetPathSlug(page.File.Dir),
	})
}

func (d *DiskLoader) insertPage(fullpath string, isBundle bool) error {
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

	sectionPath := file.Dir
	if isBundle {
		sectionPath = stdpath.Dir(sectionPath)
	}

	customPath := meta.GetString("path")
	if customPath == "" {
		// 查找顺序: posts/linux/emacs -> posts/linux -> posts -> _default
		customPath = lctx.GetSectionConfig(sectionPath, "page_path")
	}
	if customPath == "" || customPath == "none" {
		return nil
	}

	page := &types.Page{
		FrontMatter: meta,
		File:        file,
		Lang:        lang,
		Title:       meta.GetString("title"),
		Summary:     meta.GetString("summary"),
		Content:     result.Content,
		RawContent:  result.RawContent,
		Slug:        meta.GetString("slug"),
		Draft:       meta.GetBool("draft"),
		Date:        meta.GetTime("date"),
		Modified:    meta.GetTime("modified"),
	}
	if page.Summary == "" {
		page.Summary = lctx.GetSummary(result.Content)
	}
	if page.Title == "" {
		if isBundle && page.File.Dir != "" {
			page.Title = stdpath.Base(page.File.Dir)
		} else {
			page.Title = page.File.BaseName
		}
	}
	if page.Slug == "" {
		page.Slug = lctx.GetSlug(page.Title)
	}
	if page.Date.IsZero() {
		stat, err := os.Stat(fullpath)
		if err != nil {
			page.Date = time.Now()
		} else {
			page.Date = stat.ModTime()
		}
	}
	if page.Modified.IsZero() {
		page.Modified = page.Date
	}

	// 添加附属资源
	if isBundle {
		assets, err := d.loadPageAssets(fullpath, page)
		if err != nil {
			return err
		}
		page.Assets = assets
	}
	page.Path = lctx.GetRelURL(d.getPagePath(page, customPath))
	page.Permalink = lctx.GetURL(page.Path)
	page.RelPermalink = page.Path
	page.Formats = d.loadPageFormats(page)

	section := d.findSection(sectionPath)
	section.Pages = append(section.Pages, page)

	if d.hook != nil {
		page = d.hook.HandlePage(page)
		if page == nil {
			return nil
		}
	}
	d.insertTaxonomies(page)

	d.pages.Add(page.File.Path, page)
	return nil
}

func (d *DiskLoader) loadPageAssets(fullpath string, page *types.Page) (types.Assets, error) {
	lctx := d.ctx.For(page.Lang)

	customPath := page.FrontMatter.GetString("asset_path")
	if customPath == "" {
		customPath = lctx.GetSectionConfig(stdpath.Dir(page.File.Dir), "page_asset_path")
	}
	if customPath == "" || customPath == "none" {
		return nil, nil
	}
	assets := make(types.Assets, 0)

	root := filepath.Dir(fullpath)
	if err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root || path == fullpath || info.IsDir() {
			return nil
		}

		asset := &types.Asset{
			File: path,
		}
		outputPath := utils.StringReplace(customPath, map[string]string{
			"{date:%Y}":   page.Date.Format("2006"),
			"{date:%m}":   page.Date.Format("01"),
			"{date:%d}":   page.Date.Format("02"),
			"{date:%H}":   page.Date.Format("15"),
			"{slug}":      page.Slug,
			"{filename}":  page.File.BaseName,
			"{path}":      page.File.Dir,
			"{path:slug}": lctx.GetPathSlug(page.File.Dir),
		})

		asset.Path = lctx.GetRelURL(outputPath)
		asset.Permalink = lctx.GetURL(asset.Path)

		assets = append(assets, asset)
		return nil
	}); err != nil {
		return nil, err
	}
	return assets, nil
}

func (d *DiskLoader) loadPageFormats(page *types.Page) types.Formats {
	lctx := d.ctx.For(page.Lang)

	formats := make(types.Formats, 0)
	for name := range page.FrontMatter.GetStringMap("formats") {
		customPath := page.FrontMatter.GetString(name + ".path")
		customTemplate := page.FrontMatter.GetString(name + ".template")
		// 从全局配置获取
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

		outputPath := d.getPagePath(page, customPath)

		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}
