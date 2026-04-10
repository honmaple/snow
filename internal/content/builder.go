package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/honmaple/snow/internal/content/loader"
	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/theme/template"
)

type (
	Builder struct {
		ctx    *core.Context
		once   sync.Once
		hook   types.Hook
		store  types.Store
		loader types.Loader
		tplset template.TemplateSet
		writer core.Writer
	}
	BuilderOption func(*Builder)
)

func (b *Builder) write(ctx context.Context, path string, tpl template.Template, vars map[string]any) error {
	if path == "" {
		return nil
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	commonVars := map[string]any{
		"pages":                 b.store.Pages(),
		"sections":              b.store.Sections(),
		"taxonomies":            b.store.Taxonomies(),
		"get_page":              b.store.GetPage,
		"get_page_url":          b.store.GetPageURL,
		"get_section":           b.store.GetSection,
		"get_section_url":       b.store.GetSectionURL,
		"get_taxonomy":          b.store.GetTaxonomy,
		"get_taxonomy_url":      b.store.GetTaxonomyURL,
		"get_taxonomy_term":     b.store.GetTaxonomyTerm,
		"get_taxonomy_term_url": b.store.GetTaxonomyTermURL,
		"current_url":           b.ctx.GetURL(path),
		"current_path":          path,
		"current_lang":          b.ctx.GetDefaultLanguage(),
		"current_template":      tpl.Name(),
	}
	for k, v := range commonVars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	result, err := tpl.Execute(b.ctx, vars)
	if err != nil {
		b.ctx.Logger.Errorf("execute tpl %s err: %s", tpl.Name(), err.Error())
		return err
	}
	if err := b.writer.Write(ctx, path, strings.NewReader(result)); err != nil {
		b.ctx.Logger.Errorf("write to %s err: %s", path, err.Error())
	}
	return nil
}

func (b *Builder) writeAsset(ctx context.Context, asset *types.Asset) error {
	dst := asset.Path
	if strings.HasSuffix(dst, "/") {
		dst = filepath.Join(dst, filepath.Base(asset.File))
	}
	srcFile, err := os.Open(asset.File)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	return b.writer.Write(ctx, dst, srcFile)
}

func (b *Builder) writeSection(ctx context.Context, section *types.Section) error {
	customTemplate := section.FrontMatter.GetString("template")
	if customTemplate == "" {
		customTemplate = b.ctx.Config.GetString(fmt.Sprintf("sections.%s.template", section.File))
	}

	if tpl := b.tplset.Lookup(customTemplate, "section.html", "internal/section.html"); tpl != nil {
		paginators := section.Pages.Paginate(
			section.FrontMatter.GetInt("paginate"),
			section.Path,
			section.FrontMatter.GetString("paginate_path"),
		)
		for _, por := range paginators {
			b.write(ctx, por.Path, tpl, map[string]any{
				"section":       section,
				"paginator":     por,
				"pages":         section.Pages,
				"current_lang":  section.Lang,
				"current_path":  por.Path,
				"current_index": por.PageNum,
			})
		}
	}
	for _, format := range section.Formats {
		if tpl := b.tplset.Lookup(format.Template); tpl != nil {
			b.write(ctx, format.Path, tpl, map[string]any{
				"section":      section,
				"pages":        section.Pages,
				"current_lang": section.Lang,
			})
		}
	}
	for _, asset := range section.Assets {
		if err := b.writeAsset(ctx, asset); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writePage(ctx context.Context, page *types.Page) error {
	vars := map[string]any{
		"page":         page,
		"current_url":  page.Permalink,
		"current_path": page.Path,
		"current_lang": page.Lang,
	}

	customTemplate := page.FrontMatter.GetString("template")
	if customTemplate == "" {
		if page.File.BaseName == "index" {
			customTemplate = b.ctx.GetSectionConfig(filepath.Dir(page.File.Dir), "page_template")
		} else {
			customTemplate = b.ctx.GetSectionConfig(page.File.Dir, "page_template")
		}
	}

	if tpl := b.tplset.Lookup(customTemplate, "page.html", "internal/page.html"); tpl != nil {
		b.write(ctx, page.Path, tpl, vars)
	}
	if tpl := b.tplset.Lookup("alias.html", "internal/partials/alias.html"); tpl != nil {
		for _, alias := range page.FrontMatter.GetStringSlice("aliases") {
			if !strings.HasPrefix(alias, "/") {
				if strings.HasSuffix(page.Path, "/") {
					alias = filepath.Join(page.Path, alias)
				} else {
					alias = filepath.Join(filepath.Dir(page.Path), alias)
				}
			}
			b.write(ctx, alias, tpl, vars)
		}
	}
	for _, format := range page.Formats {
		if tpl := b.tplset.Lookup(format.Template); tpl != nil {
			b.write(ctx, format.Path, tpl, map[string]any{
				"page":         page,
				"current_lang": page.Lang,
				"current_url":  format.Permalink,
				"current_path": format.Path,
			})
		}
	}
	for _, asset := range page.Assets {
		if err := b.writeAsset(ctx, asset); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writeTaxonomy(ctx context.Context, taxonomy *types.Taxonomy) error {
	customTemplate := b.ctx.Config.GetString(fmt.Sprintf("taxonomies.%s.template", taxonomy.Name))

	lookups := []string{
		customTemplate,
		fmt.Sprintf("%s/taxonomy.html", taxonomy.Name),
		"taxonomy.html",
		"internal/taxonomy.html",
	}
	if tpl := b.tplset.Lookup(lookups...); tpl != nil {
		// example.com/tags/index.html
		b.write(ctx, taxonomy.Path, tpl, map[string]any{
			"taxonomy": taxonomy,
		})
	}
	for _, term := range taxonomy.Terms {
		b.writeTaxonomyTerm(ctx, term)
	}
	return nil
}
func (b *Builder) writeTaxonomyTerm(ctx context.Context, term *types.TaxonomyTerm) error {
	customTemplate := b.ctx.Config.GetString(fmt.Sprintf("taxonomies.%s.term_template", term.Taxonomy.Name))

	lookups := []string{
		customTemplate,
		fmt.Sprintf("%s/taxonomy.terms.html", term.Taxonomy.Name),
		"taxonomy.terms.html",
		"internal/taxonomy.terms.html",
	}
	if tpl := b.tplset.Lookup(lookups...); tpl != nil {
		for _, por := range term.Pages.Paginate(
			b.ctx.Config.GetInt(fmt.Sprintf("taxonomies.%s.term_paginate", term.Taxonomy.Name)),
			term.Path,
			b.ctx.Config.GetString(fmt.Sprintf("taxonomies.%s.term_paginate_path", term.Taxonomy.Name)),
		) {
			b.write(ctx, por.Path, tpl, map[string]any{
				"term":          term,
				"pages":         term.Pages,
				"taxonomy":      term.Taxonomy,
				"paginator":     por,
				"current_path":  por.Path,
				"current_index": por.PageNum,
			})
		}
	}
	for _, format := range term.Formats {
		if tpl := b.tplset.Lookup(format.Template); tpl != nil {
			b.write(ctx, format.Path, tpl, map[string]any{
				"term":     term,
				"pages":    term.Pages,
				"taxonomy": term.Taxonomy,
			})
		}
	}
	return nil
}

func (b *Builder) Build(ctx context.Context) error {
	store, err := b.loader.Load()
	if err != nil {
		return err
	}

	b.once.Do(func() {
		b.store = store
	})

	pages := b.hook.HandlePages(store.Pages())
	sections := b.hook.HandleSections(store.Sections())
	taxonomies := b.hook.HandleTaxonomies(store.Taxonomies())

	b.ctx.Logger.Infof("%d pages", len(pages))
	b.ctx.Logger.Infof("%d sections", len(sections))
	b.ctx.Logger.Infof("%d taxonomies", len(taxonomies))

	for _, page := range pages {
		b.writePage(ctx, page)
		break
	}
	for _, section := range sections {
		b.writeSection(ctx, section)
	}
	for _, taxonomy := range taxonomies {
		b.writeTaxonomy(ctx, taxonomy)
	}
	return nil
}

func WithLoader(l types.Loader) BuilderOption {
	return func(b *Builder) {
		b.loader = l
	}
}

func WithWriter(w core.Writer) BuilderOption {
	return func(b *Builder) {
		b.writer = w
	}
}

func WithTemplateSet(set template.TemplateSet) BuilderOption {
	return func(b *Builder) {
		b.tplset = set
	}
}

func New(ctx *core.Context, opts ...BuilderOption) (*Builder, error) {
	b := &Builder{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(b)
	}
	if b.hook == nil {
		b.hook = &types.EmptyHook{}
	}
	if b.writer == nil {
		return nil, errors.New("static writer is required")
	}
	if b.loader == nil {
		b.loader = loader.New(ctx)
	}
	if b.tplset == nil {
		b.tplset = template.NewSet(ctx, ctx.Theme)
	}
	return b, nil
}

func Build(ctx *core.Context, opts ...BuilderOption) error {
	b, err := New(ctx, opts...)
	if err != nil {
		return err
	}
	return b.Build(context.TODO())
}
