package content

import (
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"time"

	"github.com/honmaple/snow/internal/site/content/types"
)

type (
	Page  = types.Page
	Pages = types.Pages
)

func (d *ContentParser) IsPage(fullpath string) bool {
	return d.parserExts[filepath.Ext(fullpath)]
}

func (d *ContentParser) IsPageBundle(fullpath string) ([]string, bool) {
	indexFiles := d.findIndexFiles(fullpath, "index")
	if len(indexFiles) > 0 {
		return indexFiles, true
	}
	return nil, false
}

func (d *ContentParser) ParsePage(fullpath string, isBundle bool) (*types.Page, error) {
	node, err := d.parseNode(fullpath, true)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	page := &types.Page{
		Node:     node,
		Draft:    node.FrontMatter.GetBool("draft"),
		Hidden:   node.FrontMatter.GetBool("hidden"),
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
		assets, err := d.ParsePageAssets(fullpath, page)
		if err != nil {
			return nil, err
		}
		page.Assets = assets
	}

	sectionPath := page.File.Dir
	if isBundle {
		sectionPath = stdpath.Dir(sectionPath)
	}

	customPath := d.resolvePagePath(page, page.FrontMatter.GetString("path"))

	page.Path = lctx.GetRelURL(customPath)
	page.Permalink = lctx.GetURL(page.Path)
	page.Formats = d.ParsePageFormats(page)
	return page, nil
}

func (d *ContentParser) ParsePageAssets(fullpath string, page *types.Page) (types.Assets, error) {
	lctx := d.ctx.For(page.Lang)

	customPath := page.FrontMatter.GetString("asset_path")
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
		outputPath := d.resolvePagePath(page, customPath)

		asset.Path = lctx.GetRelURL(outputPath)
		asset.Permalink = lctx.GetURL(asset.Path)

		assets = append(assets, asset)
		return nil
	}); err != nil {
		return nil, err
	}
	return assets, nil
}

func (d *ContentParser) ParsePageFormats(page *types.Page) types.Formats {
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
		outputPath := d.resolvePagePath(page, customPath)

		format.Path = lctx.GetRelURL(outputPath)
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *ContentParser) resolvePagePath(page *types.Page, customPath string) string {
	lctx := d.ctx.For(page.Lang)

	vars := map[string]string{
		"{lang}":      page.Lang,
		"{date:%Y}":   page.Date.Format("2006"),
		"{date:%m}":   page.Date.Format("01"),
		"{date:%d}":   page.Date.Format("02"),
		"{date:%H}":   page.Date.Format("15"),
		"{path}":      page.File.Dir,
		"{path:slug}": lctx.GetPathSlug(page.File.Dir),
		"{slug}":      page.Slug,
		"{title}":     page.Title,
	}
	if page.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = page.Lang
	}
	return d.resolvePath(customPath, vars)
}
