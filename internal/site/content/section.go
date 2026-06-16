package content

import (
	"fmt"
	stdpath "path"
	"slices"
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

func (sec *Section) IsHome() bool {
	return sec.File.Dir == ""
}

func (sec *Section) Ancestors() Sections {
	if sec == nil {
		return nil
	}
	sections := make(Sections, 0)
	for current := sec.Parent; current != nil; current = current.Parent {
		sections = append(sections, current)
	}
	return sections
}

func SortSections(sections Sections, key string) {
	sort.SliceStable(sections, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return strings.Compare(sections[i].File.Path, sections[j].File.Path)
		case "title":
			return strings.Compare(sections[i].Title, sections[j].Title)
		case "weight":
			// 默认weight越小越在前
			return utils.Compare(sections[i].FrontMatter.Get(k), sections[j].FrontMatter.Get(k))
		default:
			return utils.Compare(sections[i].FrontMatter.Get(k), sections[j].FrontMatter.Get(k))
		}
	}))

	for _, section := range sections {
		SortSections(section.Children, key)
	}
}

func (sec *Section) RecursivePages() Pages {
	ns := slices.Clone(sec.Pages)
	for _, child := range sec.Children {
		ns = append(ns, child.RecursivePages()...)
	}
	SortPages(ns, sec.FrontMatter.GetString("sort_by"))
	return ns
}

func (sec *Section) RecursiveHiddenPages() Pages {
	ns := slices.Clone(sec.HiddenPages)
	for _, child := range sec.Children {
		ns = append(ns, child.RecursiveHiddenPages()...)
	}
	SortPages(ns, sec.FrontMatter.GetString("sort_by"))
	return ns
}

func (sections Sections) Reverse() Sections {
	ns := make(Sections, len(sections))
	for i, j := 0, len(sections)-1; j >= 0; i, j = i+1, j-1 {
		ns[i] = sections[j]
	}
	return ns
}

func (sections Sections) OrderBy(key string) Sections {
	ns := make(Sections, len(sections))
	copy(ns, sections)

	SortSections(ns, key)
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

func (d *Processor) ParseHomeSections(fullpath string) (Sections, error) {
	langs := make(map[string]bool)

	sections := make(Sections, 0)
	sectionFiles, _ := d.IsSection(fullpath)
	for _, file := range sectionFiles {
		section, err := d.ParseSection(stdpath.Join(fullpath, file))
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)

		langs[section.Lang] = true
	}

	for _, lang := range d.ctx.GetAllLanguages() {
		if langs[lang] {
			continue
		}
		indexFile := stdpath.Join(fullpath, "_index.md")
		if lang != d.ctx.GetDefaultLanguage() {
			indexFile = stdpath.Join(fullpath, "_index."+lang+".md")
		}
		file, err := d.parseFile(indexFile)
		if err != nil {
			return nil, err
		}
		section := &Section{
			Node: &Node{
				File:        file,
				Lang:        lang,
				Slug:        "index",
				Title:       "index",
				FrontMatter: NewFrontMatter(nil),
			},
			Pages:  make(Pages, 0),
			Assets: make([]*Asset, 0),
		}
		if lang == d.ctx.GetDefaultLanguage() {
			section.Path = "/index.html"
		} else {
			section.Path = "/" + lang + "/index.html"
		}
		section.Permalink = d.ctx.GetURL(section.Path)

		sections = append(sections, section)
	}
	return sections, nil
}

func (d *Processor) ParseSection(fullpath string) (*Section, error) {
	node, err := d.parseNode(fullpath)
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
		if section.IsHome() {
			section.Title = "index"
		} else {
			section.Title = stdpath.Base(section.File.Dir)
		}
	}
	if section.Slug == "" {
		if section.IsHome() {
			section.Slug = "index"
		} else {
			section.Slug = lctx.GetSlug(stdpath.Base(section.File.Dir))
		}
	}

	section.Path = lctx.GetRelURL(d.resolveSectionPath(section, section.FrontMatter.GetString("path")))
	section.Permalink = lctx.GetURL(section.Path)

	assets, err := d.ParseSectionAssets(section)
	if err != nil {
		return nil, err
	}
	section.Assets = assets
	section.Formats = d.ParseSectionFormats(section)
	return section, nil
}

func (d *Processor) ParseSectionAssets(section *Section) (Assets, error) {
	root := stdpath.Dir(section.File.Path)

	assetPaths, err := d.parseAssetPaths(root, section.FrontMatter.GetStringSlice("assets"))
	if err != nil {
		return nil, err
	}

	lctx := d.ctx.For(section.Lang)
	assets := make(Assets, 0)
	for _, assetPath := range assetPaths {
		file, err := d.parseFile(stdpath.Join(root, assetPath))
		if err != nil {
			return nil, err
		}
		asset := &Asset{
			File: file,
		}
		asset.Path = lctx.GetRelURL(d.resolveAssetPath(section.Path, assetPath))
		asset.Permalink = lctx.GetURL(asset.Path)
		assets = append(assets, asset)
	}
	return assets, nil
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
		fmt.Sprintf("%s/section.html", section.File.Dir),
		"section.html",
	}
	// 首页content/_index.md
	if section.IsHome() {
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
				d.ctx.For(section.Lang).GetURL,
			) {
			if err := d.RenderTemplate(por.Path, tpl, map[string]any{
				"paginator":     por,
				"pages":         section.Pages,
				"section":       section,
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
				"pages":        section.Pages,
				"section":      section,
				"current_lang": section.Lang,
			}, writer); err != nil {
				return err
			}
		}
	}
	for _, asset := range section.Assets {
		if err := d.RenderAsset(asset, writer); err != nil {
			return err
		}
	}
	return nil
}
