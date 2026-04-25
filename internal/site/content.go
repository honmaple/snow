package site

import (
	"context"
	"fmt"
	"io/fs"
	stdpath "path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/content/types"
	"github.com/honmaple/snow/internal/site/template"
)

func (site *Site) writeTemplate(path string, tpl template.Template, vars map[string]any) error {
	if path == "" {
		return nil
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	lang := site.ctx.GetDefaultLanguage()
	if l, ok := vars["current_lang"]; ok {
		lang = l.(string)
	}
	lctx := site.ctx.For(lang)

	commonVars := map[string]any{
		"pages":                 site.store.Pages(lang),
		"sections":              site.store.Sections(lang),
		"taxonomies":            site.store.Taxonomies(lang),
		"get_page":              site.store.GetPage,
		"get_page_url":          site.store.GetPageURL,
		"get_section":           site.store.GetSection,
		"get_section_url":       site.store.GetSectionURL,
		"get_taxonomy":          site.store.GetTaxonomy,
		"get_taxonomy_url":      site.store.GetTaxonomyURL,
		"get_taxonomy_term":     site.store.GetTaxonomyTerm,
		"get_taxonomy_term_url": site.store.GetTaxonomyTermURL,
		"current_url":           lctx.GetURL(path),
		"current_path":          path,
		"current_lang":          lang,
		"current_template":      tpl.Name(),
	}
	for k, v := range commonVars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	result, err := tpl.Execute(vars)
	if err != nil {
		return &core.Error{
			Op:   "execute tpl",
			Err:  err,
			Path: tpl.Name(),
		}
	}
	if err := site.writer.Write(context.TODO(), path, strings.NewReader(result)); err != nil {
		return &core.Error{
			Op:   "write tpl",
			Err:  err,
			Path: path,
		}
	}
	return nil
}

func (site *Site) writeAsset(asset *types.Asset) error {
	return nil
	// dst := asset.Path
	// if strings.HasSuffix(dst, "/") {
	//	dst = stdpath.Join(dst, stdpath.Base(asset.File))
	// }
	// srcFile, err := os.Open(asset.File)
	// if err != nil {
	//	return err
	// }
	// defer srcFile.Close()

	// return site.writer.Write(nil, dst, srcFile)
}

func (site *Site) writePage(page *types.Page) error {
	site.ctx.Logger.Debugf("write page [%s] -> %s", page.File.Path, page.Path)

	vars := map[string]any{
		"page":         page,
		"current_url":  page.Permalink,
		"current_path": page.Path,
		"current_lang": page.Lang,
	}
	if tpl := site.tplset.Lookup(page.FrontMatter.GetString("template"), "page.html"); tpl != nil {
		if err := site.writeTemplate(page.Path, tpl, vars); err != nil {
			return err
		}
	}
	if tpl := site.tplset.Lookup("alias.html", "partials/alias.html"); tpl != nil {
		for _, alias := range page.FrontMatter.GetStringSlice("aliases") {
			if !strings.HasPrefix(alias, "/") {
				if strings.HasSuffix(page.Path, "/") {
					alias = stdpath.Join(page.Path, alias)
				} else {
					alias = stdpath.Join(stdpath.Dir(page.Path), alias)
				}
			}
			if err := site.writeTemplate(alias, tpl, vars); err != nil {
				return err
			}
		}
	}
	for _, format := range page.Formats {
		if tpl := site.tplset.Lookup(format.Template); tpl != nil {
			if err := site.writeTemplate(format.Path, tpl, map[string]any{
				"page":         page,
				"current_lang": page.Lang,
				"current_url":  format.Permalink,
				"current_path": format.Path,
			}); err != nil {
				return err
			}
		}
	}
	for _, asset := range page.Assets {
		if err := site.writeAsset(asset); err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) writeSection(section *types.Section) error {
	site.ctx.Logger.Debugf("write section [%s] -> %s", section.File.Path, section.Path)

	customTemplate := section.FrontMatter.GetString("template")

	lookups := []string{
		customTemplate,
		"section.html",
	}
	// 首页content/_index.md
	if section.File.Dir == "" {
		lookups = []string{
			customTemplate,
			"index.html",
			"section.html",
		}
	}

	if tpl := site.tplset.Lookup(lookups...); tpl != nil {
		for _, por := range section.Pages.
			Filter(
				section.FrontMatter.GetString("paginate_filter"),
			).
			Paginate(
				section.FrontMatter.GetInt("paginate"),
				section.Path,
				section.FrontMatter.GetString("paginate_path"),
			) {
			if err := site.writeTemplate(por.Path, tpl, map[string]any{
				"section":       section,
				"paginator":     por,
				"pages":         section.Pages,
				"current_lang":  section.Lang,
				"current_index": por.PageNum,
			}); err != nil {
				return err
			}
		}
	}

	for _, format := range section.Formats {
		if tpl := site.tplset.Lookup(format.Template); tpl != nil {
			if err := site.writeTemplate(format.Path, tpl, map[string]any{
				"section":      section,
				"pages":        section.Pages,
				"current_lang": section.Lang,
			}); err != nil {
				return err
			}
		}
	}
	for _, asset := range section.Assets {
		if err := site.writeAsset(asset); err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) writeTaxonomy(taxonomy *types.Taxonomy) error {
	site.ctx.Logger.Debugf("write taxonomy [%s] -> %s", taxonomy.Name, taxonomy.Path)

	lctx := site.ctx.For(taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(taxonomy.Name, "template").String(),
		fmt.Sprintf("%s/taxonomy.html", taxonomy.Name),
		"taxonomy.html",
	}
	if tpl := site.tplset.Lookup(lookups...); tpl != nil {
		// example.com/tags/index.html
		if err := site.writeTemplate(taxonomy.Path, tpl, map[string]any{
			"taxonomy":     taxonomy,
			"current_lang": taxonomy.Lang,
		}); err != nil {
			return err
		}
	}

	for _, term := range taxonomy.Terms {
		if err := site.writeTaxonomyTerm(term); err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) writeTaxonomyTerm(term *types.TaxonomyTerm) error {
	site.ctx.Logger.Debugf("write taxonomy term [%s:%s] -> %s", term.Taxonomy.Name, term.GetFullName(), term.Path)

	lctx := site.ctx.For(term.Taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.template").String(),
		fmt.Sprintf("%s/taxonomy.terms.html", term.Taxonomy.Name),
		"taxonomy.terms.html",
	}
	if tpl := site.tplset.Lookup(lookups...); tpl != nil {
		for _, por := range term.Pages.
			Filter(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_filter").String(),
			).
			Paginate(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate").Int(),
				term.Path,
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_path").String(),
			) {
			if err := site.writeTemplate(por.Path, tpl, map[string]any{
				"term":          term,
				"pages":         term.Pages,
				"taxonomy":      term.Taxonomy,
				"paginator":     por,
				"current_path":  por.Path,
				"current_index": por.PageNum,
				"current_lang":  term.Taxonomy.Lang,
			}); err != nil {
				return err
			}
		}
	}
	for _, format := range term.Formats {
		if tpl := site.tplset.Lookup(format.Template); tpl != nil {
			if err := site.writeTemplate(format.Path, tpl, map[string]any{
				"term":         term,
				"pages":        term.Pages,
				"taxonomy":     term.Taxonomy,
				"current_lang": term.Taxonomy.Lang,
			}); err != nil {
				return err
			}
		}
	}

	for _, child := range term.Children {
		if err := site.writeTaxonomyTerm(child); err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) buildContent(ctx context.Context) error {
	for _, lang := range site.ctx.GetAllLanguages() {
		site.ctx.Logger.Debugf("write %s site", lang)

		for _, section := range site.store.Sections(lang) {
			if err := site.writeSection(section); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}

		for _, page := range site.store.Pages(lang) {
			if err := site.writePage(page); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}

		for _, taxonomy := range site.store.Taxonomies(lang) {
			if err := site.writeTaxonomy(taxonomy); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}
	}
	return nil
}

func (site *Site) isIgnoredContent(path string, isDir bool) bool {
	// 忽略以.或者_开头的文件或目录，不要忽略_index.md
	if basename := filepath.Base(path); !strings.HasPrefix(basename, "_index.") && (strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".")) {
		return true
	}

	matchPath := strings.TrimPrefix(path, site.ctx.GetContentDir()+"/")
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range site.ctx.Config.GetStringSlice("ignored_content") {
		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			site.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (site *Site) loadContent() error {
	contentDir := site.ctx.GetContentDir()
	if contentDir == "" {
		return fmt.Errorf("The content dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == contentDir {
			sections, err := site.contentParser.ParseRootSection(path)
			if err != nil {
				return err
			}
			for _, section := range sections {
				site.store.insertSection(section)
			}
			return nil
		}
		// 忽略指定的文件
		if site.isIgnoredContent(path, info.IsDir()) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			indexFiles, ok := site.contentParser.IsPageBundle(path)
			if ok {
				for _, file := range indexFiles {
					page, err := site.contentParser.ParsePage(filepath.Join(path, file), true)
					if err != nil {
						return err
					}
					site.store.insertPage(page)
				}
				return fs.SkipDir
			}
			sectionFiles, ok := site.contentParser.IsSection(path)
			if len(sectionFiles) > 0 {
				for _, file := range sectionFiles {
					section, err := site.contentParser.ParseSection(filepath.Join(path, file))
					if err != nil {
						return err
					}
					site.store.insertSection(section)
				}
				return nil
			}
			return nil
		}
		// 忽略_index.md文件
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_") {
			return nil
		}

		if !site.contentParser.IsPage(path) {
			return nil
		}

		page, err := site.contentParser.ParsePage(path, false)
		if err != nil {
			return err
		}
		site.store.insertPage(page)
		return nil
	}

	if err := filepath.WalkDir(contentDir, walkDir); err != nil {
		return err
	}

	for _, sections := range site.store.AllSections() {
		sort.SliceStable(sections, func(i, j int) bool {
			return sections[i].File.Path > sections[j].File.Path
		})

		for _, section := range sections {
			content.SortPages(section.Pages, section.FrontMatter.GetString("sort_by"))
		}
	}

	for lang, pages := range site.store.AllPages() {
		sort.SliceStable(pages, func(i, j int) bool {
			return pages[i].Date.After(pages[j].Date)
		})

		taxonomies := site.contentParser.ParseTaxonomies(pages, lang)

		for _, taxonomy := range taxonomies {
			for _, term := range taxonomy.Terms {
				site.store.insertTaxonomyTerm(term)
			}
			site.store.insertTaxonomy(taxonomy)
		}
	}
	return nil
}
