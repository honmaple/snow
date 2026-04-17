package loader

import (
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"time"

	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/utils"
)

func (d *Loader) Pages(lang string) types.Pages {
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Loader) GetPage(path string, lang string) *types.Page {
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(path)
	return result
}

func (d *Loader) GetPageURL(path string, lang string) string {
	result := d.GetPage(path, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Loader) parsePagePath(page *types.Page, customPath string) string {
	lctx := d.ctx.For(page.Lang)
	return utils.StringReplace(customPath, map[string]string{
		"{date:%Y}":   page.Date.Format("2006"),
		"{date:%m}":   page.Date.Format("01"),
		"{date:%d}":   page.Date.Format("02"),
		"{date:%H}":   page.Date.Format("15"),
		"{slug}":      page.Slug,
		"{filename}":  page.File.BaseName,
		"{path}":      page.File.Dir,
		"{path:slug}": lctx.GetPathSlug(page.File.Dir),
	})
}

func (d *Loader) parsePage(fullpath string, isBundle bool) (*types.Page, error) {
	node, err := d.parseNode(fullpath)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	page := &types.Page{
		Node:     node,
		Draft:    node.FrontMatter.GetBool("draft"),
		Date:     node.FrontMatter.GetTime("date"),
		Modified: node.FrontMatter.GetTime("modified"),
		IsBundle: isBundle,
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
		assets, err := d.parsePageAssets(fullpath, page)
		if err != nil {
			return nil, err
		}
		page.Assets = assets
	}

	sectionPath := page.File.Dir
	if isBundle {
		sectionPath = stdpath.Dir(sectionPath)
	}

	customPath := page.FrontMatter.GetString("path")
	if customPath == "" {
		// 查找顺序: posts/linux/emacs -> posts/linux -> posts -> _default
		customPath = lctx.GetSectionConfig(sectionPath, "page_path")
	}

	page.Path = lctx.GetRelURL(d.parsePagePath(page, customPath))
	page.Permalink = lctx.GetURL(page.Path)
	page.Formats = d.parsePageFormats(page)

	section := d.findSection(sectionPath, page.Lang)
	section.Pages = append(section.Pages, page)
	return page, nil
}

func (d *Loader) parsePageAssets(fullpath string, page *types.Page) (types.Assets, error) {
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
			"{path}":      page.File.Dir,
			"{path:slug}": lctx.GetPathSlug(page.File.Dir),
			"{slug}":      page.Slug,
			"{filename}":  page.File.BaseName,
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

func (d *Loader) parsePageFormats(page *types.Page) types.Formats {
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

		outputPath := d.parsePagePath(page, customPath)

		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *Loader) insertPage(page *types.Page) {
	set, ok := d.pages[page.Lang]
	if !ok {
		set = newSet[*types.Page]()
		d.pages[page.Lang] = set
	}
	set.Add(page.File.Path, page)

	d.insertTaxonomies(page)
}

func (d *Loader) insertPageByPath(fullpath string, isBundle bool) {
	page, err := d.parsePage(fullpath, isBundle)
	if err != nil {
		d.ctx.Logger.Warnf("parse page %s err: %s", fullpath, err.Error())
		return
	}
	d.insertPage(page)
}
