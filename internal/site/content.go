package site

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/utils/taskutil"
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

func (site *Site) loadContent() (*ContentStore, error) {
	contentDir := site.ctx.GetContentDir()
	if contentDir == "" {
		return nil, fmt.Errorf("The content dir is null")
	}

	pages := make(content.Pages, 0)
	sections := make(content.Sections, 0)

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == contentDir {
			roots, err := site.contentProcessor.ParseRootSection(path)
			if err != nil {
				return err
			}
			for _, section := range roots {
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
		return nil, err
	}

	store := NewContentStore()

	for _, section := range sections {
		if section.FrontMatter.IsSet("render") && !section.FrontMatter.GetBool("render") {
			continue
		}
		section = site.hook.HandleSection(section)
		if section == nil {
			continue
		}
		store.insertSection(section)
	}
	for _, page := range pages {
		if page.Draft && !site.option.IncludeDrafts {
			continue
		}
		if page.FrontMatter.IsSet("render") && !page.FrontMatter.GetBool("render") {
			continue
		}
		page = site.hook.HandlePage(page)
		if page == nil {
			continue
		}
		store.insertPage(page)
	}

	for _, sections := range store.AllSections() {
		sort.SliceStable(sections, func(i, j int) bool {
			return sections[i].File.Path > sections[j].File.Path
		})

		for _, section := range sections {
			section.Pages.SortBy(section.FrontMatter.GetString("sort_by"))
		}
	}

	for lang, pages := range store.AllPages() {
		pages.SortBy("date desc")

		taxonomies := site.contentProcessor.ParseTaxonomies(pages, lang)
		for _, taxonomy := range taxonomies {
			for _, term := range taxonomy.Terms {
				store.insertTaxonomyTerm(term)
			}
			store.insertTaxonomy(taxonomy)
		}
	}
	return store, nil
}

func (site *Site) buildContent(ctx context.Context) error {
	store, err := site.loadContent()
	if err != nil {
		return err
	}

	for _, lang := range site.ctx.GetAllLanguages() {
		site.ctx.Logger.Infof("Building %s site...", lang)

		tplset := &ContentTemplateSet{
			lang:        lang,
			store:       store,
			TemplateSet: site.tplset,
		}

		tasks := taskutil.NewPool[any](20, func(arg any) (err error) {
			switch v := arg.(type) {
			case *content.Section:
				err = site.contentProcessor.RenderSection(v, tplset, site.writer)
			case *content.Page:
				err = site.contentProcessor.RenderPage(v, tplset, site.writer)
			case *content.Taxonomy:
				err = site.contentProcessor.RenderTaxonomy(v, tplset, site.writer)
			}
			if err != nil {
				site.ctx.Logger.Error(err.Error())
			}
			return nil
		})

		now := time.Now()
		for _, section := range store.Sections(lang) {
			tasks.Invoke(section)
		}

		for _, page := range store.Pages(lang) {
			tasks.Invoke(page)
		}

		for _, page := range store.HiddenPages(lang) {
			tasks.Invoke(page)
		}

		ts := make([]string, 0)
		for _, taxonomy := range store.Taxonomies(lang) {
			if count := len(taxonomy.Terms); count > 0 {
				ts = append(ts, fmt.Sprintf("%d %s", count, taxonomy.Name))
			}
			tasks.Invoke(taxonomy)
		}
		tasks.StopAndWait()

		site.ctx.Logger.Infof("Done: %d sections, %d pages, %d hidden pages and %d taxonomies (%s) in %v",
			len(store.Sections(lang)),
			len(store.Pages(lang)),
			len(store.HiddenPages(lang)),
			len(store.Taxonomies(lang)),
			strings.Join(ts, ", "),
			time.Since(now),
		)
	}
	return nil
}
