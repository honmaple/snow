package site

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/site/content"
)

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
		matched, err := doublestar.Match(pattern, matchPath)
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

	pages := make(content.Pages, 0)
	sections := make(content.Sections, 0)

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == contentDir {
			sections, err := site.contentProcessor.ParseRootSection(path)
			if err != nil {
				return err
			}
			for _, section := range sections {
				sections = append(sections, section)
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
			indexFiles, ok := site.contentProcessor.IsPageBundle(path)
			if ok {
				for _, file := range indexFiles {
					page, err := site.contentProcessor.ParsePage(filepath.Join(path, file), true)
					if err != nil {
						return err
					}
					pages = append(pages, page)
				}
				return fs.SkipDir
			}
			sectionFiles, ok := site.contentProcessor.IsSection(path)
			if len(sectionFiles) > 0 {
				for _, file := range sectionFiles {
					section, err := site.contentProcessor.ParseSection(filepath.Join(path, file))
					if err != nil {
						return err
					}
					sections = append(sections, section)
				}
				return nil
			}
			return nil
		}
		// 忽略_index.md文件
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_") {
			return nil
		}

		if !site.contentProcessor.IsPage(path) {
			return nil
		}

		page, err := site.contentProcessor.ParsePage(path, false)
		if err != nil {
			return err
		}
		pages = append(pages, page)
		return nil
	}

	if err := filepath.WalkDir(contentDir, walkDir); err != nil {
		return err
	}

	for _, section := range sections {
		section = site.hook.HandleSection(section)
		if section == nil {
			continue
		}
		site.store.insertSection(section)
	}
	for _, page := range pages {
		page = site.hook.HandlePage(page)
		if page == nil || (!site.option.IncludeDrafts && page.Draft) {
			continue
		}
		site.store.insertPage(page)
	}

	for _, sections := range site.store.AllSections() {
		sort.SliceStable(sections, func(i, j int) bool {
			return sections[i].File.Path > sections[j].File.Path
		})

		for _, section := range sections {
			section.Pages.SortBy(section.FrontMatter.GetString("sort_by"))
		}
	}

	for lang, pages := range site.store.AllPages() {
		pages.SortBy("date desc")

		taxonomies := site.contentProcessor.ParseTaxonomies(pages, lang)
		for _, taxonomy := range taxonomies {
			for _, term := range taxonomy.Terms {
				site.store.insertTaxonomyTerm(term)
			}
			site.store.insertTaxonomy(taxonomy)
		}
	}
	return nil
}

func (site *Site) buildContent(ctx context.Context) error {
	for _, lang := range site.ctx.GetAllLanguages() {
		site.ctx.Logger.Debugf("write %s site", lang)

		tplset := &ContentTemplateSet{
			lang:        lang,
			store:       site.store,
			TemplateSet: site.tplset,
		}

		for _, section := range site.store.Sections(lang) {
			if err := site.contentProcessor.RenderSection(section, tplset, site.writer); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}

		for _, page := range site.store.Pages(lang) {
			if err := site.contentProcessor.RenderPage(page, tplset, site.writer); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}

		for _, page := range site.store.HiddenPages(lang) {
			if err := site.contentProcessor.RenderPage(page, tplset, site.writer); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}

		for _, taxonomy := range site.store.Taxonomies(lang) {
			if err := site.contentProcessor.RenderTaxonomy(taxonomy, tplset, site.writer); err != nil {
				site.ctx.Logger.Error(err.Error())
			}
		}
	}
	return nil
}
