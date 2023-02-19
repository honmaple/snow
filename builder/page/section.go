package page

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/honmaple/snow/utils"
)

type (
	Section struct {
		// slug:
		// weight:
		// aliases:
		// transparent:
		// filter:
		// orderby:
		// paginate:
		// paginate_path: {name}{number}{extension}
		// path:
		// template:
		// page_path:
		// page_template:
		// feed_path:
		// feed_template:
		Meta      Meta
		Path      string
		Permalink string
		Title     string
		Content   string
		Hidden    bool
		Pages     Pages
		Assets    []string
		Parent    *Section
		Children  Sections
	}
	Sections []*Section
)

func (sec Section) allPages() Pages {
	pages := make(Pages, 0)
	if !sec.Meta.GetBool("transparent") {
		pages = append(pages, sec.Pages...)
	}
	for _, child := range sec.Children {
		pages = append(pages, child.allPages()...)
	}
	return pages
}

func (sec Section) allSections() Sections {
	sections := make(Sections, 0)
	for _, child := range sec.Children {
		sections = append(sections, child)
		sections = append(sections, child.allSections()...)
	}
	return sections
}

func (sec Section) Paginator() []*paginator {
	return sec.Pages.Paginator(
		sec.Meta.GetInt("paginate"),
		sec.Path,
		sec.Meta.GetString("paginate_path"),
	)
}

func (sec Section) Name() string {
	if sec.Parent == nil || sec.Parent.Title == "" {
		return sec.Title
	}
	return fmt.Sprintf("%s/%s", sec.Parent.Name(), sec.Title)
}

func (b *Builder) loadSectionDone(section *Section) {
	pages := section.Pages
	if section.Hidden {
		pages = section.Parent.allPages()
	}
	section.Path = b.conf.GetRelURL(section.Meta.GetString("path"))
	section.Permalink = b.conf.GetURL(section.Path)
	section.Pages = pages.Filter(section.Meta.Get("filter")).OrderBy(section.Meta.GetString("orderby"))

	sort.SliceStable(section.Children, func(i, j int) bool {
		ti := section.Children[i]
		tj := section.Children[j]
		if wi, wj := ti.Meta.GetInt("weight"), tj.Meta.GetInt("weight"); wi == wj {
			return ti.Title > tj.Title
		} else {
			return wi > wj
		}
	})
}

func (b *Builder) loadSection(parent *Section, path string) (*Section, error) {
	section := &Section{Parent: parent}
	if parent != nil {
		section.Title = utils.FileBaseName(path)
	}
	section.Meta = b.newSectionConfig(section.Name())
	// _index.md包括配置信息
	matches, err := filepath.Glob(filepath.Join(path, "_index.*"))
	if err == nil && len(matches) > 0 {
		meta, err := b.readFile(matches[0])
		if err != nil {
			return nil, err
		}
		for k, v := range meta {
			switch strings.ToLower(k) {
			case "title":
				section.Title = v.(string)
			case "content":
				section.Content = v.(string)
			default:
				section.Meta[k] = v
			}
		}
	}
	name := section.Name()
	slug := section.Meta.GetString("slug")
	if slug == "" {
		names := strings.Split(name, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = b.conf.GetSlug(name)
		}
		slug = strings.Join(slugs, "/")
	}
	vars := map[string]string{"{section}": name, "{section:slug}": slug}
	keys := []string{"path", "template", "page_path", "page_template", "feed_path", "feed_template"}
	for _, k := range keys {
		section.Meta[k] = utils.StringReplace(section.Meta.GetString(k), vars)
	}
	defer b.loadSectionDone(section)

	names, err := utils.FileList(path)
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if strings.HasPrefix(name, "_index.") {
			continue
		}
		file := filepath.Join(path, name)
		info, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(file, "index.*"))
			if err == nil && len(matches) > 0 {
				if page, err := b.loadPage(section, matches[0]); err == nil {
					section.Pages = append(section.Pages, page)
				}
				continue
			}
			sec, err := b.loadSection(section, file)
			if err != nil {
				return nil, err
			}
			if sec.Meta.GetBool("transparent") {
				section.Pages = append(section.Pages, sec.Pages...)
			}
			section.Children = append(section.Children, sec)
			continue
		}
		if _, ok := b.readers[filepath.Ext(file)]; !ok {
			section.Assets = append(section.Assets, file)
			continue
		}

		filemeta, err := b.readFile(file)
		if err != nil {
			return nil, err
		}
		// 自定义页面, 该页面的page列表与父级section一致
		if template, ok := filemeta["template"]; ok && template != "" {
			meta := section.Meta.Copy()
			for k, v := range filemeta {
				meta[k] = v
			}
			child := &Section{
				Meta:    meta,
				Title:   meta.GetString("title"),
				Content: meta.GetString("content"),
				Hidden:  true,
				Parent:  section,
			}
			section.Children = append(section.Children, child)
			defer b.loadSectionDone(child)
			continue
		}
		if page, err := b.loadPage(section, file); err == nil {
			section.Pages = append(section.Pages, page)
		}
	}
	return section, nil
}

func (b *Builder) loadSections() error {
	dir := b.conf.GetString("content_dir")
	root, err := b.loadSection(nil, dir)
	if err != nil {
		return err
	}
	b.pages = root.allPages()
	b.sections = root.allSections()
	return nil
}

func (b *Builder) newSectionConfig(name string) Meta {
	meta := make(Meta)
	for k, v := range b.conf.GetStringMap("sections._default") {
		meta[k] = v
	}
	if name == "" {
		return meta
	}
	list := make([]string, 0)
	for m := range b.conf.GetStringMap("sections") {
		if m == "_default" {
			continue
		}
		list = append(list, m)
	}
	// 从最短路径开始匹配, 子目录的配置可以继承父目录
	sort.SliceStable(list, func(i, j int) bool {
		return len(list[i]) < len(list[j])
	})

	for _, m := range list {
		if m != name && !strings.HasPrefix(name, m+"/") {
			continue
		}
		for k, v := range b.conf.GetStringMap("sections." + m) {
			meta[k] = v
		}
		if m == name {
			break
		}
	}
	return meta
}
