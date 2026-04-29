package renderer

import (
	"context"
	"fmt"
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/types"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Store interface {
		Sections(string) types.Sections
		GetSection(string, string) *types.Section
		GetSectionURL(string, string) string

		Pages(string) types.Pages
		GetPage(string, string) *types.Page
		GetPageURL(string, string) string

		Taxonomies(string) types.Taxonomies
		GetTaxonomy(string, string) *types.Taxonomy
		GetTaxonomyURL(string, string) string

		GetTaxonomyTerms(string, string) types.TaxonomyTerms
		GetTaxonomyTerm(string, string, string) *types.TaxonomyTerm
		GetTaxonomyTermURL(string, string, string) string
	}
	Renderer struct {
		ctx    *core.Context
		store  Store
		tplset template.TemplateSet
		writer core.Writer
	}
)

func (r *Renderer) RenderTemplate(path string, tpl template.Template, vars map[string]any) error {
	if path == "" {
		return nil
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	lang := r.ctx.GetDefaultLanguage()
	if l, ok := vars["current_lang"]; ok {
		lang = l.(string)
	}
	lctx := r.ctx.For(lang)

	commonVars := map[string]any{
		"pages":                 r.store.Pages(lang),
		"sections":              r.store.Sections(lang),
		"taxonomies":            r.store.Taxonomies(lang),
		"get_page":              r.store.GetPage,
		"get_page_url":          r.store.GetPageURL,
		"get_section":           r.store.GetSection,
		"get_section_url":       r.store.GetSectionURL,
		"get_taxonomy":          r.store.GetTaxonomy,
		"get_taxonomy_url":      r.store.GetTaxonomyURL,
		"get_taxonomy_term":     r.store.GetTaxonomyTerm,
		"get_taxonomy_term_url": r.store.GetTaxonomyTermURL,
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
	if err := r.writer.Write(context.TODO(), path, strings.NewReader(result)); err != nil {
		return &core.Error{
			Op:   "write tpl",
			Err:  err,
			Path: path,
		}
	}
	return nil
}

func (r *Renderer) renderAsset(asset *types.Asset) error {
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

	// return r.writer.Write(nil, dst, srcFile)
}

func (r *Renderer) RenderPage(page *types.Page) error {
	r.ctx.Logger.Debugf("write page [%s] -> %s", page.File.Path, page.Path)

	vars := map[string]any{
		"page":         page,
		"current_url":  page.Permalink,
		"current_path": page.Path,
		"current_lang": page.Lang,
	}
	if tpl := r.tplset.Lookup(page.FrontMatter.GetString("template"), "page.html"); tpl != nil {
		if err := r.RenderTemplate(page.Path, tpl, vars); err != nil {
			return err
		}
	}
	if tpl := r.tplset.Lookup("alias.html", "partials/alias.html"); tpl != nil {
		for _, alias := range page.FrontMatter.GetStringSlice("aliases") {
			if !strings.HasPrefix(alias, "/") {
				if strings.HasSuffix(page.Path, "/") {
					alias = stdpath.Join(page.Path, alias)
				} else {
					alias = stdpath.Join(stdpath.Dir(page.Path), alias)
				}
			}
			if err := r.RenderTemplate(alias, tpl, vars); err != nil {
				return err
			}
		}
	}
	for _, format := range page.Formats {
		if tpl := r.tplset.Lookup(format.Template); tpl != nil {
			if err := r.RenderTemplate(format.Path, tpl, map[string]any{
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
		if err := r.renderAsset(asset); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) RenderSection(section *types.Section) error {
	r.ctx.Logger.Debugf("write section [%s] -> %s", section.File.Path, section.Path)

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

	if tpl := r.tplset.Lookup(lookups...); tpl != nil {
		for _, por := range section.Pages.
			Filter(
				section.FrontMatter.GetString("paginate_filter"),
			).
			Paginate(
				section.FrontMatter.GetInt("paginate"),
				section.Path,
				section.FrontMatter.GetString("paginate_path"),
			) {
			if err := r.RenderTemplate(por.Path, tpl, map[string]any{
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
		if tpl := r.tplset.Lookup(format.Template); tpl != nil {
			if err := r.RenderTemplate(format.Path, tpl, map[string]any{
				"section":      section,
				"pages":        section.Pages,
				"current_lang": section.Lang,
			}); err != nil {
				return err
			}
		}
	}
	for _, asset := range section.Assets {
		if err := r.renderAsset(asset); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) RenderTaxonomy(taxonomy *types.Taxonomy) error {
	r.ctx.Logger.Debugf("write taxonomy [%s] -> %s", taxonomy.Name, taxonomy.Path)

	lctx := r.ctx.For(taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(taxonomy.Name, "template").String(),
		fmt.Sprintf("%s/taxonomy.html", taxonomy.Name),
		"taxonomy.html",
	}
	if tpl := r.tplset.Lookup(lookups...); tpl != nil {
		// example.com/tags/index.html
		if err := r.RenderTemplate(taxonomy.Path, tpl, map[string]any{
			"taxonomy":     taxonomy,
			"current_lang": taxonomy.Lang,
		}); err != nil {
			return err
		}
	}

	for _, term := range taxonomy.Terms {
		if err := r.RenderTaxonomyTerm(term); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) RenderTaxonomyTerm(term *types.TaxonomyTerm) error {
	r.ctx.Logger.Debugf("write taxonomy term [%s:%s] -> %s", term.Taxonomy.Name, term.GetFullName(), term.Path)

	lctx := r.ctx.For(term.Taxonomy.Lang)

	lookups := []string{
		lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.template").String(),
		fmt.Sprintf("%s/taxonomy.terms.html", term.Taxonomy.Name),
		"taxonomy.terms.html",
	}
	if tpl := r.tplset.Lookup(lookups...); tpl != nil {
		for _, por := range term.Pages.
			Filter(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_filter").String(),
			).
			Paginate(
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate").Int(),
				term.Path,
				lctx.GetTaxonomyConfig(term.Taxonomy.Name, "term.paginate_path").String(),
			) {
			if err := r.RenderTemplate(por.Path, tpl, map[string]any{
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
		if tpl := r.tplset.Lookup(format.Template); tpl != nil {
			if err := r.RenderTemplate(format.Path, tpl, map[string]any{
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
		if err := r.RenderTaxonomyTerm(child); err != nil {
			return err
		}
	}
	return nil
}

func New(ctx *core.Context, store Store, tplset template.TemplateSet, writer core.Writer) *Renderer {
	r := &Renderer{
		ctx:    ctx,
		store:  store,
		tplset: tplset,
		writer: writer,
	}
	return r
}
