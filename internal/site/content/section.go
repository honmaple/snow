package content

import (
	"fmt"
	stdpath "path"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Section struct {
		*Node

		Path      string
		Permalink string

		Pages       Pages
		HiddenPages Pages
		Assets      Assets
		Formats     Formats

		Parent   *Section
		Children Sections
	}
	Sections []*Section
)

func (sections Sections) Reverse() Sections {
	ns := make(Sections, len(sections))
	for i, j := 0, len(sections)-1; j >= 0; i, j = i+1, j-1 {
		ns[i] = sections[j]
	}
	return ns
}

func (sections Sections) SortBy(key string) {
	sort.SliceStable(sections, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return strings.Compare(sections[i].File.Path, sections[j].File.Path)
		case "title":
			return strings.Compare(sections[i].Title, sections[j].Title)
		default:
			return utils.Compare(sections[i].FrontMatter.Get(k), sections[j].FrontMatter.Get(k))
		}
	}))

	for _, section := range sections {
		section.Children.SortBy(key)
	}
}

func (sections Sections) OrderBy(key string) Sections {
	ns := make(Sections, len(sections))
	copy(ns, sections)

	ns.SortBy(key)
	return ns
}

func (d *Processor) resolveSectionPath(section *Section, customPath string) string {
	lctx := d.ctx.For(section.Lang)

	vars := map[string]string{
		"{lang}":         section.Lang,
		"{path}":         section.File.Dir,
		"{path:slug}":    lctx.GetPathSlug(section.File.Dir),
		"{section}":      section.Title,
		"{section:slug}": section.Slug,
	}
	if section.Lang == d.ctx.GetDefaultLanguage() {
		vars["{lang:optional}"] = ""
	} else {
		vars["{lang:optional}"] = section.Lang
	}

	return d.resolvePath(customPath, vars)
}

func (d *Processor) IsSection(fullpath string) ([]string, bool) {
	sectionFiles := d.findIndexFiles(fullpath, "_index")
	if len(sectionFiles) > 0 {
		return sectionFiles, true
	}
	return nil, false
}

func (d *Processor) ParseSection(fullpath string) (*Section, error) {
	node, err := d.parseNode(fullpath, false)
	if err != nil {
		return nil, err
	}
	lctx := d.ctx.For(node.Lang)

	section := &Section{
		Node:        node,
		Pages:       make(Pages, 0),
		HiddenPages: make(Pages, 0),
		Assets:      make(Assets, 0),
		Children:    make(Sections, 0),
	}
	if section.Title == "" {
		if section.File.Dir == "" {
			section.Title = "index"
		} else {
			section.Title = stdpath.Base(section.File.Dir)
		}
	}
	if section.Slug == "" {
		if section.File.Dir == "" {
			section.Slug = "index"
		} else {
			section.Slug = lctx.GetSlug(stdpath.Base(section.File.Dir))
		}
	}

	section.Assets = d.ParseSectionAssets(fullpath, section)
	section.Formats = d.ParseSectionFormats(section)

	customPath := d.resolveSectionPath(section, section.FrontMatter.GetString("path"))
	section.Path = lctx.GetRelURL(customPath)
	section.Permalink = lctx.GetURL(section.Path)
	return section, nil
}

func (d *Processor) ParseSectionAssets(fullpath string, section *Section) Assets {
	lctx := d.ctx.For(section.Lang)

	assets := make(Assets, 0)
	for _, file := range section.FrontMatter.GetStringSlice("assets") {
		asset := &Asset{
			File: file,
		}
		customPath := section.FrontMatter.GetString("asset_path")
		outputPath := d.resolveSectionPath(section, customPath)
		asset.Path = lctx.GetRelURL(outputPath)
		asset.Permalink = lctx.GetURL(asset.Path)

		assets = append(assets, asset)
	}
	return assets
}

func (d *Processor) ParseSectionFormats(section *Section) Formats {
	lctx := d.ctx.For(section.Lang)

	formats := make(Formats, 0)
	for name := range section.FrontMatter.GetStringMap("formats") {
		customPath := section.FrontMatter.GetString(fmt.Sprintf("formats.%s.path", name))
		customTemplate := section.FrontMatter.GetString(fmt.Sprintf("formats.%s.template", name))
		if customTemplate == "" {
			customTemplate = d.ctx.Config.GetString("formats." + name + ".template")
		}
		if customPath == "" || customTemplate == "" {
			continue
		}

		format := &Format{
			Name:     name,
			Template: customTemplate,
		}
		format.Path = lctx.GetRelURL(d.resolveSectionPath(section, customPath))
		format.Permalink = lctx.GetRelURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}

func (d *Processor) RenderSection(section *Section, tplset template.TemplateSet, writer core.Writer) error {
	d.ctx.Logger.Debugf("write section [%s] -> %s", section.File.Path, section.Path)

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

	if tpl := tplset.Lookup(lookups...); tpl != nil {
		for _, por := range section.Pages.
			FilterBy(
				section.FrontMatter.GetString("paginate_filter_by"),
			).
			PaginateBy(
				section.FrontMatter.GetInt("paginate"),
				section.Path,
				section.FrontMatter.GetString("paginate_path"),
			) {
			if err := d.RenderTemplate(por.Path, tpl, map[string]any{
				"paginator":     por,
				"section":       section,
				"pages":         section.Pages,
				"current_lang":  section.Lang,
				"current_index": por.PageNum,
			}, writer); err != nil {
				return err
			}
		}
	}

	for _, format := range section.Formats {
		if tpl := tplset.Lookup(format.Template); tpl != nil {
			d.ctx.Logger.Debugf("write section format [%s] -> %s", section.File.Path, format.Path)
			if err := d.RenderTemplate(format.Path, tpl, map[string]any{
				"section":      section,
				"pages":        section.Pages,
				"current_lang": section.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}
	// for _, asset := range section.Assets {
	//	if err := r.renderAsset(asset); err != nil {
	//		return err
	//	}
	// }
	return nil
}
