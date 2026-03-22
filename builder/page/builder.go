package page

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/honmaple/snow/builder/parser"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type (
	Builder struct {
		ctx    *Context
		conf   config.Config
		theme  theme.Theme
		hooks  Hooks
		parser parser.Parser
	}
)

func (b *Builder) findLang(path string, filemeta Meta) string {
	if filemeta != nil {
		if lang := filemeta.GetString("lang"); lang != "" && b.conf.IsValidLanguage(lang) {
			return lang
		}
	}
	ext := filepath.Ext(path)
	if ext != "" {
		lang := filepath.Ext(path[:len(path)-len(ext)])
		if lang != "" && b.conf.IsValidLanguage(lang[1:]) {
			return lang[1:]
		}
	}
	return b.conf.DefaultLanguage
}

func (b *Builder) findFiles(path string, pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(path, pattern))
	if len(matches) == 0 {
		return nil
	}

	files := make([]string, 0)
	for _, m := range matches {
		if !b.parser.IsSupport(filepath.Ext(m)) {
			continue
		}
		files = append(files, m)
	}
	return files
}

func (b *Builder) Build(ctx context.Context) error {
	rootDir := b.conf.ContentDir
	if rootDir == "" {
		return fmt.Errorf("The content dir of %s is null", b.conf.Site.Language)
	}
	b.conf.Watch(rootDir)

	now := time.Now()
	defer func() {
		ps := make([]string, 0)
		ls := make([]string, 0)
		ts := make([]string, 0)

		lang := ""
		if !b.conf.IsDefaultLanguage(b.conf.Site.Language) {
			lang = "[" + b.conf.Site.Language + "]"
		}
		if count := len(b.ctx.Pages()); count > 0 {
			ps = append(ps, fmt.Sprintf("%d normal pages", count))
		}
		if count := len(b.ctx.HiddenPages()); count > 0 {
			ps = append(ps, fmt.Sprintf("%d hidden pages", count))
		}
		if count := len(b.ctx.SectionPages()); count > 0 {
			ps = append(ps, fmt.Sprintf("%d section pages", count))
		}

		for _, section := range b.ctx.Sections() {
			if section.isRoot() {
				continue
			}
			if count := len(section.Pages) + len(section.HiddenPages) + len(section.SectionPages); count > 0 {
				ls = append(ls, fmt.Sprintf("%d %s", count, section.RealName()))
			}
		}

		for _, taxonomy := range b.ctx.Taxonomies() {
			if count := len(taxonomy.Terms); count > 0 {
				ts = append(ts, fmt.Sprintf("%d %s", count, taxonomy.Name))
			}
		}

		duration := time.Since(now)
		if len(ps) > 0 {
			b.conf.Log.Infof("Done: %sPage Processed %s in %v", lang, strings.Join(ps, ", "), duration)
		}
		if len(ls) > 0 {
			b.conf.Log.Infof("Done: %sSection Processed %s in %v", lang, strings.Join(ls, ", "), duration)
		}
		if len(ts) > 0 {
			b.conf.Log.Infof("Done: %sTaxonomy Processed %s in %v", lang, strings.Join(ts, ", "), duration)
		}
	}()

	var wg sync.WaitGroup

	tasks := utils.NewTaskPool(&wg, 100, func(i any) {
		defer wg.Done()
		b.insertPage(i.(string))
	})
	defer tasks.Release()

	ignoreFiles := make(map[string]bool)
	ignoreRegex := make([]*regexp.Regexp, 0)

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		for _, re := range ignoreRegex {
			if re.MatchString(path) {
				return nil
			}
		}
		if info.IsDir() {
			// is not root
			if path != rootDir {
				files := b.findFiles(path, "index.*")
				for _, file := range files {
					tasks.Invoke(file)
				}
				if len(files) > 0 {
					return fs.SkipDir
				}
			}
			ignoreFiles = make(map[string]bool)
			ignoreRegex = ignoreRegex[:0]

			section := b.insertSection(path)
			if section == nil {
				return nil
			}
			for _, file := range section.Meta.GetSlice("ignore_files") {
				if ignoreFiles[file] {
					continue
				}
				ignoreFiles[file] = true

				re, err := regexp.Compile(filepath.Join(path, file))
				if err == nil {
					ignoreRegex = append(ignoreRegex, re)
				}
			}
			return nil
		}
		if !b.parser.IsSupport(filepath.Ext(path)) {
			b.insertAsset(path)
			return nil
		}
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_index.") || strings.HasPrefix(basename, ".") {
			return nil
		}
		tasks.Invoke(path)
		return nil
	}
	if err := filepath.WalkDir(rootDir, walkDir); err != nil {
		return err
	}
	tasks.Wait()
	return b.Write()
}

func (b *Builder) write(tpl template.Writer, path string, vars map[string]any) {
	if path == "" {
		return
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	rvars := map[string]any{
		"pages":                 b.ctx.Pages(),
		"hidden_pages":          b.ctx.HiddenPages(),
		"taxonomies":            b.ctx.Taxonomies(),
		"get_section":           b.ctx.findSection,
		"get_section_url":       b.ctx.findSectionURL,
		"get_taxonomy":          b.ctx.findTaxonomy,
		"get_taxonomy_url":      b.ctx.findTaxonomyURL,
		"get_taxonomy_term":     b.ctx.findTaxonomyTerm,
		"get_taxonomy_term_url": b.ctx.findTaxonomyTermURL,
		"current_url":           b.conf.GetURL(path),
		"current_path":          path,
		"current_template":      tpl.Name(),
	}
	for k, v := range rvars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	if err := tpl.Write(path, vars); err != nil {
		b.conf.Log.Error(err.Error())
	}
}

func (b *Builder) Write() error {
	var wg sync.WaitGroup

	tasks := utils.NewTaskPool(&wg, 10, func(i any) {
		defer wg.Done()

		switch v := i.(type) {
		case *Page:
			b.writePage(v)
		case *Section:
			b.writeSection(v)
		case *Taxonomy:
			b.writeTaxonomy(v)
		case *TaxonomyTerm:
			b.writeTaxonomyTerm(v)
		}
	})
	defer tasks.Release()

	b.ctx.ensure()
	for _, page := range b.hooks.Pages(b.ctx.Pages()) {
		tasks.Invoke(page)
	}
	for _, page := range b.hooks.Pages(b.ctx.HiddenPages()) {
		tasks.Invoke(page)
	}
	for _, page := range b.hooks.Pages(b.ctx.SectionPages()) {
		tasks.Invoke(page)
	}
	for _, section := range b.hooks.Sections(b.ctx.Sections()) {
		if section.isRoot() || section.isEmpty() {
			continue
		}
		tasks.Invoke(section)
	}
	for _, taxonomy := range b.hooks.Taxonomies(b.ctx.Taxonomies()) {
		tasks.Invoke(taxonomy)

		for _, term := range b.hooks.TaxonomyTerms(taxonomy.Terms) {
			tasks.Invoke(term)
		}
	}
	tasks.Wait()
	return nil
}

func NewBuilder(conf config.Config, parser parser.Parser, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{
		ctx:    newContext(conf),
		conf:   conf,
		theme:  theme,
		parser: parser,
		hooks:  hooks,
	}
}
