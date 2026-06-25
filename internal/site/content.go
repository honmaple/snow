package site

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/utils/taskutil"
)

func (site *Site) IsIgnoredContent(path string, isDir bool) bool {
	// 忽略以.或者_开头的文件或目录，不要忽略_index.md
	if basename := stdpath.Base(path); !strings.HasPrefix(basename, "_index.") && (strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".")) {
		return true
	}

	matchPath := path
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

func (site *Site) loadContent(processor *content.Processor) (*ContentStore, error) {
	store := NewContentStore()

	insertPageByFile := func(file string, isBundle bool) error {
		page, err := processor.ParsePage(file, isBundle)
		if err != nil {
			return err
		}

		if page.Draft && !site.includeDrafts {
			return nil
		}
		if !page.FrontMatter.GetBool("render", true) {
			return nil
		}
		page = site.hook.HandlePage(page)
		if page == nil {
			return nil
		}
		store.insertPage(page)
		return nil
	}

	insertSection := func(section *content.Section) {
		if !section.FrontMatter.GetBool("render", true) {
			return
		}
		section = site.hook.HandleSection(section)
		if section == nil {
			return
		}
		store.insertSection(section)
	}

	insertSectionByFile := func(file string) error {
		section, err := processor.ParseSection(file)
		if err != nil {
			return err
		}
		insertSection(section)
		return nil
	}

	type node struct {
		File     string
		IsBundle bool
	}

	tasks := taskutil.NewPool[node](100, func(arg node) (err error) {
		return insertPageByFile(arg.File, arg.IsBundle)
	})

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			sections, err := processor.ParseHomeSections(".")
			if err != nil {
				return err
			}
			for _, section := range sections {
				insertSection(section)
			}
			return nil
		}
		// 忽略指定的文件
		if site.IsIgnoredContent(path, info.IsDir()) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			indexFiles, ok := processor.IsPageBundle(path)
			if ok {
				for _, indexFile := range indexFiles {
					tasks.Invoke(node{
						File:     stdpath.Join(path, indexFile),
						IsBundle: true,
					})
				}
				return fs.SkipDir
			}
			sectionFiles, ok := processor.IsSection(path)
			if len(sectionFiles) > 0 {
				for _, sectionFile := range sectionFiles {
					if err := insertSectionByFile(stdpath.Join(path, sectionFile)); err != nil {
						return err
					}
				}
				return nil
			}
			return nil
		}
		// 忽略_index.md文件
		if basename := stdpath.Base(path); strings.HasPrefix(basename, "_") {
			return nil
		}

		if !processor.IsPage(path) {
			return nil
		}

		tasks.Invoke(node{
			File:     path,
			IsBundle: false,
		})
		return nil
	}

	if err := processor.WalkDir(walkDir); err != nil {
		return nil, err
	}
	tasks.StopAndWait()

	for _, lang := range site.ctx.GetAllLanguages() {
		sections := store.Sections(lang)
		if len(sections) > 0 {
			content.SortSections(sections, "weight", true)

			for _, section := range sections {
				content.SortPages(section.Pages, section.FrontMatter.GetString("sort_by"))
			}
		}

		pages := store.Pages(lang)
		if len(pages) > 0 {
			// 使用首页配置的顺序
			root := store.GetSection("/", lang)
			if root == nil {
				content.SortPages(pages, "date desc")
			} else {
				content.SortPages(pages, root.FrontMatter.GetString("sort_by"))
			}

			taxonomies := processor.ParseTaxonomies(pages, lang)
			for _, taxonomy := range taxonomies {
				for _, term := range taxonomy.Terms {
					store.insertTaxonomyTerm(term)
				}
				store.insertTaxonomy(taxonomy)
			}
		}
	}
	return store, nil
}

func (site *Site) BuildContent(ctx context.Context, writer core.Writer) error {
	contentDir := site.ctx.GetContentDir()
	if contentDir == "" {
		return fmt.Errorf("The content dir is null")
	}
	processor := content.NewProcessor(site.ctx, os.DirFS(contentDir))

	now := time.Now()

	store, err := site.loadContent(processor)
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

		tasks := taskutil.NewPool[any](100, func(arg any) (err error) {
			switch v := arg.(type) {
			case *content.Section:
				err = processor.RenderSection(v, tplset, writer)
			case *content.Page:
				err = processor.RenderPage(v, tplset, writer)
			case *content.Taxonomy:
				err = processor.RenderTaxonomy(v, tplset, writer)
			}
			if err != nil {
				site.ctx.Logger.Error(err.Error())
			}
			return nil
		})

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

		tstat := ""
		if len(ts) > 0 {
			tstat = fmt.Sprintf(" (%s)", strings.Join(ts, ", "))
		}
		site.ctx.Logger.Infof("Done: %d sections, %d pages, %d hidden pages and %d taxonomies%s in %v",
			len(store.Sections(lang)),
			len(store.Pages(lang)),
			len(store.HiddenPages(lang)),
			len(store.Taxonomies(lang)),
			tstat,
			time.Since(now),
		)
	}
	return nil
}
