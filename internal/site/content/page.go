package content

import (
	"fmt"
	"io/fs"
	stdpath "path"
	"sort"
	"strings"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Page struct {
		*Node

		Draft    bool
		Hidden   bool
		IsBundle bool

		Date     time.Time
		Modified time.Time

		Path      string
		Permalink string

		Section *Section
		Assets  Assets

		Formats Formats
	}
	Pages []*Page
)

func (page *Page) Ancestors() Sections {
	if page == nil || page.Section == nil {
		return nil
	}
	return append(Sections{page.Section}, page.Section.Ancestors()...)
}

func SortPages(pages Pages, key string) {
	if key == "" {
		key = "date DESC"
	}
	sort.SliceStable(pages, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			// "-"表示默认排序, 避免时间相同时排序混乱
			return 0 - strings.Compare(pages[i].Title, pages[j].Title)
		case "title":
			return strings.Compare(pages[i].Title, pages[j].Title)
		case "date":
			return utils.Compare(pages[i].Date, pages[j].Date)
		case "modified":
			return utils.Compare(pages[i].Modified, pages[j].Modified)
		case "weight":
			return utils.Compare(pages[i].FrontMatter.Get(k), pages[j].FrontMatter.Get(k))
		default:
			return utils.Compare(pages[i].FrontMatter.Get(k), pages[j].FrontMatter.Get(k))
		}
	}))
}

func FilterExpr(filter string) func(*Page) bool {
	tpl, err := pongo2.FromString("{{" + filter + "}}")
	if err != nil {
		return func(page *Page) bool {
			return true
		}
	}
	return func(page *Page) bool {
		args := page.FrontMatter.AllSettings()

		result, err := tpl.Execute(args)
		return err == nil && result == "True"
	}
}

func (pages Pages) First() *Page {
	if len(pages) > 0 {
		return pages[0]
	}
	return nil
}

func (pages Pages) Last() *Page {
	if len(pages) > 0 {
		return pages[len(pages)-1]
	}
	return nil
}

func (pages Pages) Related(page *Page) *Related[*Page] {
	return &Related[*Page]{list: pages, cur: page}
}

func (pages Pages) Limit(n int) Pages {
	if n >= len(pages) {
		return pages
	}
	return pages[:n]
}

func (pages Pages) Reverse() Pages {
	ns := make(Pages, len(pages))
	for i, j := 0, len(pages)-1; j >= 0; i, j = i+1, j-1 {
		ns[i] = pages[j]
	}
	return ns
}

func (pages Pages) FilterBy(filter string) Pages {
	if filter == "" {
		return pages
	}
	npages := make(Pages, 0, len(pages))

	expr := FilterExpr(filter)
	for _, page := range pages {
		if expr(page) {
			npages = append(npages, page)
		}
	}
	return npages
}

func (pages Pages) OrderBy(key string) Pages {
	newPs := make(Pages, len(pages))
	copy(newPs, pages)

	SortPages(newPs, key)
	return newPs
}

func (pages Pages) GroupBy(key string) PageGroups {
	var groupf func(*Page) []string

	if strings.HasPrefix(key, "date:") {
		format := key[5:]
		groupf = func(page *Page) []string {
			return []string{page.Date.Format(format)}
		}
	} else {
		groupf = func(page *Page) []string {
			if v := page.FrontMatter.GetStringSlice(key); len(v) > 0 {
				return v
			}
			if v := page.FrontMatter.GetString(key); v != "" {
				return []string{v}
			}
			return nil
		}
	}

	results := make(PageGroups, 0)
	resultMap := make(map[string]*PageGroup)
	for _, page := range pages {
		for _, name := range groupf(page) {
			var (
				currentTerm *PageGroup
				currentName string
			)
			for part := range strings.SplitSeq(strings.Trim(name, "/"), "/") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if currentName == "" {
					currentName = part
				} else {
					currentName += "/" + part
				}

				term, ok := resultMap[currentName]
				if !ok {
					term = &PageGroup{
						Name:     part,
						Pages:    make(Pages, 0),
						Parent:   currentTerm,
						Children: make(PageGroups, 0),
					}
					resultMap[currentName] = term

					if currentTerm == nil {
						results = append(results, term)
					} else {
						currentTerm.Children = append(currentTerm.Children, term)
					}
				}
				term.Pages = append(term.Pages, page)

				currentTerm = term
			}
		}
	}
	return results
}

func (d *Processor) resolvePagePath(page *Page, customPath string) string {
	lctx := d.ctx.For(page.Lang)

	vars := map[string]string{
		"{lang}":       page.Lang,
		"{date:%Y}":    page.Date.Format("2006"),
		"{date:%m}":    page.Date.Format("01"),
		"{date:%d}":    page.Date.Format("02"),
		"{date:%H}":    page.Date.Format("15"),
		"{path}":       page.File.Dir,
		"{path:slug}":  lctx.GetPathSlug(page.File.Dir),
		"{slug}":       page.Slug,
		"{title}":      page.Title,
		"{title:slug}": lctx.GetSlug(page.Title),
	}
	if page.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = page.Lang
	}

	customPath = d.resolvePath(customPath, vars)
	return lctx.ApplyPathStyle(customPath, page.FrontMatter.GetString("path_style"))
}

func (d *Processor) IsPage(fullpath string) bool {
	return d.parserExts[stdpath.Ext(fullpath)]
}

func (d *Processor) IsPageBundle(fullpath string) ([]string, bool) {
	indexFiles := d.findIndexFiles(fullpath, "index")
	if len(indexFiles) > 0 {
		return indexFiles, true
	}
	return nil, false
}

func (d *Processor) ParsePage(fullpath string, isBundle bool) (*Page, error) {
	node, err := d.parseNode(fullpath)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	page := &Page{
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
		page.Slug = lctx.GetSlug(page.File.BaseName)
	}
	if page.Date.IsZero() {
		stat, err := fs.Stat(d.contentFS, fullpath)
		if err != nil {
			page.Date = time.Now()
		} else {
			page.Date = stat.ModTime()
		}
	}
	if page.Modified.IsZero() {
		page.Modified = page.Date
	}

	page.Path = d.resolvePagePath(page, page.FrontMatter.GetString("path"))
	page.Permalink = lctx.GetURL(page.Path)

	// 添加附属资源
	if isBundle {
		assets, err := d.ParsePageAssets(page)
		if err != nil {
			return nil, err
		}
		page.Assets = assets
	}

	page.Formats = d.ParsePageFormats(page)
	return page, nil
}

func (d *Processor) ParsePageAssets(page *Page) (Assets, error) {
	root := stdpath.Dir(page.File.Path)

	assetPaths := make([]string, 0)
	if files := page.FrontMatter.GetStringSlice("assets"); len(files) > 0 {
		paths, err := d.parseAssetPaths(root, files)
		if err != nil {
			return nil, err
		}
		assetPaths = paths
	} else {
		rootFS, err := fs.Sub(d.contentFS, root)
		if err != nil {
			return nil, err
		}
		if err := fs.WalkDir(rootFS, ".", func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == "." || info.IsDir() {
				return nil
			}
			if stdpath.Join(root, path) == page.File.Path {
				return nil
			}

			assetPaths = append(assetPaths, path)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	lctx := d.ctx.For(page.Lang)

	assets := make(Assets, 0)
	for _, assetPath := range assetPaths {
		file, err := d.parseFile(stdpath.Join(root, assetPath))
		if err != nil {
			return nil, err
		}
		asset := &Asset{
			File: file,
		}
		asset.Path = d.resolveAssetPath(page.Path, assetPath)
		asset.Permalink = lctx.GetURL(asset.Path)
		assets = append(assets, asset)
	}
	return assets, nil
}

func (d *Processor) ParsePageFormats(page *Page) Formats {
	lctx := d.ctx.For(page.Lang)

	formats := make(Formats, 0)
	for name := range page.FrontMatter.GetStringMap("formats") {
		customPath := page.FrontMatter.GetString(fmt.Sprintf("formats.%s.path", name))
		customTemplate := page.FrontMatter.GetString(fmt.Sprintf("formats.%s.template", name))
		// 从全局配置获取
		if customTemplate == "" {
			customTemplate = lctx.Config.GetString("formats." + name + ".template")
		}
		if customPath == "" || customTemplate == "" {
			continue
		}

		format := &Format{
			Name:     name,
			Template: customTemplate,
		}

		format.Path = d.resolvePagePath(page, customPath)
		format.Permalink = lctx.GetURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *Processor) RenderPage(page *Page, tplset template.TemplateSet, writer core.Writer) error {
	if tpl := tplset.Lookup(page.FrontMatter.GetString("template"), "page.html"); tpl != nil {
		d.ctx.Logger.Debugf("write page [%s] -> %s", page.File.Path, page.Path)
		if err := d.RenderTemplate(page.Path, tpl, map[string]any{
			"page":         page,
			"current_lang": page.Lang,
			"current_url":  page.Permalink,
			"current_path": page.Path,
		}, writer); err != nil {
			return err
		}
	}
	if tpl := tplset.Lookup("alias.html", "partials/alias.html"); tpl != nil {
		for _, alias := range page.FrontMatter.GetStringSlice("aliases") {
			if alias == "" || alias == "." || stdpath.Clean(alias) != alias {
				d.ctx.Logger.Warnf("invalid alias '%s' for %s", alias, page.File.Path)
				continue
			}
			// aliases: ["alias.html", "/alias.html"]
			if !strings.HasPrefix(alias, "/") {
				if strings.HasSuffix(page.Path, "/") {
					alias = stdpath.Join(page.Path, alias)
				} else {
					alias = stdpath.Join(stdpath.Dir(page.Path), alias)
				}
			}
			d.ctx.Logger.Debugf("write page alias [%s] -> %s", page.File.Path, alias)
			if err := d.RenderTemplate(alias, tpl, map[string]any{
				"page":         page,
				"current_lang": page.Lang,
				"current_url":  d.ctx.GetURL(alias),
				"current_path": alias,
			}, writer); err != nil {
				return err
			}
		}
	}
	for _, format := range page.Formats {
		if tpl := tplset.Lookup(format.Template); tpl != nil {
			d.ctx.Logger.Debugf("write page format [%s] -> %s", page.File.Path, format.Path)
			if err := d.RenderTemplate(format.Path, tpl, map[string]any{
				"page":         page,
				"current_lang": page.Lang,
				"current_url":  format.Permalink,
				"current_path": format.Path,
			}, writer); err != nil {
				return err
			}
		}
	}
	for _, asset := range page.Assets {
		if err := d.RenderAsset(asset, writer); err != nil {
			return err
		}
	}
	return nil
}
