package loader

import (
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/core"
)

type (
	Loader struct {
		ctx        *core.Context
		hook       types.Hook
		parser     parser.Parser
		parserExts map[string]bool

		pages         map[string]*Set[*types.Page]
		sections      map[string]*Set[*types.Section]
		taxonomies    map[string]*Set[*types.Taxonomy]
		taxonomyTerms map[string]*Set[*types.TaxonomyTerm]
	}
	LoaderOption func(*Loader)
)

func (d *Loader) isIgnored(path string, isDir bool) bool {
	// 忽略以.或者_开头的文件或目录，不要忽略_index.md
	if basename := filepath.Base(path); !strings.HasPrefix(basename, "_index.") && (strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".")) {
		return true
	}

	matchPath := strings.TrimPrefix(path, d.ctx.GetContentDir()+"/")
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range d.ctx.Config.GetStringSlice("ignored_content") {
		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			d.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (d *Loader) findIndexFiles(fullpath string, prefix string) []string {
	// 如果有多个扩展: index.md, index.org只返回第一个
	allowedFiles := make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		allowedFiles[prefix+ext] = true
		for lang := range d.ctx.Locales {
			allowedFiles[prefix+lang+ext] = true
		}
	}

	files, err := os.ReadDir(fullpath)
	if err != nil {
		return nil
	}

	results := make([]string, 0)
	resultMap := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
		if allowedFiles[name] && !resultMap[nameWithoutExt] {
			results = append(results, name)
			resultMap[nameWithoutExt] = true
		}
	}
	return results
}

func (d *Loader) parseFile(fullpath string) (*types.File, error) {
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

func (d *Loader) parseNode(fullpath string) (*types.Node, error) {
	file, err := d.parseFile(fullpath)
	if err != nil {
		return nil, err
	}

	result, err := d.parser.Parse(fullpath)
	if err != nil {
		return nil, err
	}

	meta := types.NewFrontMatter(result.FrontMatter)

	lang := meta.GetString("lang")
	if lang == "" {
		langExt := stdpath.Ext(file.BaseName)
		if langExt != "" {
			lang = strings.TrimPrefix(langExt, ".")
		}
	}
	if lang == "" || (lang != d.ctx.GetDefaultLanguage() && !d.ctx.Config.IsSet("languages."+lang)) {
		lang = d.ctx.GetDefaultLanguage()
	}

	if ext := "." + lang; strings.HasSuffix(file.BaseName, ext) {
		file.BaseName = strings.TrimSuffix(file.BaseName, ext)
		file.LanguageName = lang
	}
	lctx := d.ctx.For(lang)

	node := &types.Node{
		FrontMatter: meta,
		File:        file,
		Lang:        lang,
		Slug:        meta.GetString("slug"),
		Title:       meta.GetString("title"),
		Description: meta.GetString("desc"),
		Content:     result.Content,
		Summary:     meta.GetString("summary"),
	}
	if node.Summary == "" {
		node.Summary = lctx.GetSummary(result.Content)
	}
	return node, nil
}

func (d *Loader) Load() (types.Store, error) {
	contentDir := d.ctx.GetContentDir()
	if contentDir == "" {
		return nil, fmt.Errorf("The content dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == contentDir {
			sectionFiles := d.findIndexFiles(path, "_index")
			for _, file := range sectionFiles {
				d.insertSectionByPath(filepath.Join(path, file))
			}
			d.insertRootSection()
			return nil
		}
		// 忽略指定的文件
		if d.isIgnored(path, info.IsDir()) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			indexFiles := d.findIndexFiles(path, "index")
			if len(indexFiles) > 0 {
				for _, file := range indexFiles {
					d.insertPageByPath(filepath.Join(path, file), true)
				}
				return fs.SkipDir
			}
			sectionFiles := d.findIndexFiles(path, "_index")
			if len(sectionFiles) > 0 {
				for _, file := range sectionFiles {
					d.insertSectionByPath(filepath.Join(path, file))
				}
				return nil
			}
			return nil
		}
		// 忽略_index.md文件
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_") {
			return nil
		}

		if !d.parserExts[filepath.Ext(path)] {
			return nil
		}
		d.insertPageByPath(path, false)
		return nil
	}

	if err := filepath.WalkDir(contentDir, walkDir); err != nil {
		return nil, err
	}
	return d, nil
}

func WithHook(h types.Hook) LoaderOption {
	return func(d *Loader) {
		d.hook = h
	}
}

func WithParser(p parser.Parser) LoaderOption {
	return func(d *Loader) {
		d.parser = p
	}
}

func New(ctx *core.Context, opts ...LoaderOption) *Loader {
	d := &Loader{
		ctx:           ctx,
		pages:         make(map[string]*Set[*types.Page]),
		sections:      make(map[string]*Set[*types.Section]),
		taxonomies:    make(map[string]*Set[*types.Taxonomy]),
		taxonomyTerms: make(map[string]*Set[*types.TaxonomyTerm]),
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.parser == nil {
		d.parser = parser.New(ctx)
	}

	d.parserExts = make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		d.parserExts[ext] = true
	}
	return d
}
