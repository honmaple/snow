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

	file, err := d.loadPageFile(fullpath)
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
		assets := make([]string, 0)

		root := filepath.Dir(fullpath)
		if err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == root || path == fullpath || info.IsDir() {
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

	sectionPath := page.File.Dir
	if isBundle {
		sectionPath = stdpath.Dir(sectionPath)
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

func (d *DiskLoader) loadPageFile(fullpath string) (*types.File, error) {
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

func (d *DiskLoader) loadPageFormats(page *types.Page) types.Formats {
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
	return formats
}
