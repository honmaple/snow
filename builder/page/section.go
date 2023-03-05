package page

import (
	"fmt"
	"path/filepath"
	"regexp"
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
		File         string
		Meta         Meta
		Path         string
		Permalink    string
		Slug         string
		Title        string
		Content      string
		Pages        Pages
		HiddenPages  Pages
		SectionPages Pages
		Assets       []string
		Parent       *Section
		Children     Sections
		Lang         string
	}
	Sections []*Section
)

func (sec *Section) vars() map[string]string {
	return map[string]string{"{section}": sec.Name(), "{section:slug}": sec.Slug}
}

func (sec *Section) allPages() Pages {
	pages := make(Pages, 0)

	pages = append(pages, sec.Pages...)
	for _, child := range sec.Children {
		pages = append(pages, child.allPages()...)
	}
	return pages
}

func (sec *Section) allHiddenPages() Pages {
	pages := make(Pages, 0)

	pages = append(pages, sec.HiddenPages...)
	for _, child := range sec.Children {
		pages = append(pages, child.allHiddenPages()...)
	}
	return pages
}

func (sec *Section) allSectionPages() Pages {
	pages := make(Pages, 0)

	pages = append(pages, sec.SectionPages...)
	for _, child := range sec.Children {
		pages = append(pages, child.allSectionPages()...)
	}
	return pages
}

func (sec *Section) allSections() Sections {
	sections := make(Sections, 0)
	for _, child := range sec.Children {
		sections = append(sections, child)
		sections = append(sections, child.allSections()...)
	}
	return sections
}

func (sec *Section) isRoot() bool {
	return sec.Parent == nil
}

func (sec *Section) isEmpty() bool {
	return len(sec.Children) == 0 && len(sec.Pages) == 0 && len(sec.HiddenPages) == 0 && len(sec.SectionPages) == 0
}

func (sec *Section) isPaginate() bool {
	return sec.Meta.GetInt("paginate") > 0
}

func (sec *Section) Paginator() []*paginator {
	return sec.Pages.Filter(sec.Meta.GetString("paginate_filter")).Paginator(
		sec.Meta.GetInt("paginate"),
		sec.Path,
		sec.Meta.GetString("paginate_path"),
	)
}

func (sec *Section) Root() *Section {
	if sec.Parent == nil {
		return sec
	}
	return sec.Parent.Root()
}

func (sec *Section) Name() string {
	if sec.Parent == nil || sec.Parent.Parent == nil {
		return sec.Title
	}
	return fmt.Sprintf("%s/%s", sec.Parent.Name(), sec.Title)
}

func (sec *Section) FirstName() string {
	if sec.Parent == nil || sec.Parent.Title == "" {
		return sec.Title
	}
	return sec.Parent.FirstName()
}

func (secs Sections) Sort() {
	sort.SliceStable(secs, func(i, j int) bool {
		ti := secs[i]
		tj := secs[j]
		if wi, wj := ti.Meta.GetInt("weight"), tj.Meta.GetInt("weight"); wi == wj {
			return ti.Title > tj.Title
		} else {
			return wi > wj
		}
	})
}

func (b *Builder) findSection(file string, langs ...string) *Section {
	lang := b.getLang(langs...)
	b.mu.RLock()
	defer b.mu.RUnlock()
	m, ok := b.sections[lang]
	if !ok {
		return nil
	}
	return m[file]
}

func (b *Builder) findSectionIndex(prefix string, files map[string]bool) string {
	for ext := range b.readers {
		file := prefix + ext
		if files[file] {
			return file
		}
	}
	return ""
}

func (b *Builder) insertSection(path string) *Section {
	names, _ := utils.FileList(path)
	namem := make(map[string]bool)
	for _, name := range names {
		namem[name] = true
	}

	b.mu.Lock()
	b.ignoreFiles = b.ignoreFiles[:0]
	b.mu.Unlock()

	b.languageRange(func(lang string, isdefault bool) {
		prefix := "_index"
		if !isdefault {
			prefix = prefix + "." + lang
		}
		filemeta := make(Meta)
		if index := b.findSectionIndex(prefix, namem); index != "" {
			filemeta, _ = b.readFile(filepath.Join(path, index))
		}

		section := &Section{
			File: path,
			Lang: lang,
		}
		section.Parent = b.findSection(filepath.Dir(section.File), lang)
		// 根目录
		if section.isRoot() {
			section.Meta = make(Meta)
			section.Meta.load(b.conf.GetStringMap("sections._default"))
		} else {
			section.Meta = section.Parent.Meta.clone()
			section.Title = filepath.Base(section.File)
		}
		section.Meta.load(filemeta)

		name := section.Name()
		if !section.isRoot() {
			section.Meta.load(b.conf.GetStringMap("sections." + name))
			if !isdefault {
				section.Meta.load(b.conf.GetStringMap("languages." + lang + ".sections." + name))
			}
		}

		for k, v := range section.Meta {
			switch strings.ToLower(k) {
			case "title":
				section.Title = v.(string)
			case "content":
				section.Content = v.(string)
			}
		}

		slug := section.Meta.GetString("slug")
		if slug == "" {
			names := strings.Split(name, "/")
			slugs := make([]string, len(names))
			for i, name := range names {
				slugs[i] = b.conf.GetSlug(name)
			}
			slug = strings.Join(slugs, "/")
		}
		section.Slug = slug
		section.Path = b.conf.GetRelURL(utils.StringReplace(section.Meta.GetString("path"), section.vars()), lang)
		section.Permalink = b.conf.GetURL(section.Path)

		b.mu.Lock()
		defer b.mu.Unlock()

		if section.Parent != nil {
			section.Parent.Children = append(section.Parent.Children, section)
		}
		ignoreFiles := filemeta.GetSlice("ignore_files")
		for _, file := range ignoreFiles {
			re, err := regexp.Compile(filepath.Join(path, file))
			if err == nil {
				b.ignoreFiles = append(b.ignoreFiles, re)
			}
		}
		if _, ok := b.sections[lang]; !ok {
			b.sections[lang] = make(map[string]*Section)
		}
		b.sections[lang][section.File] = section
	})
	return nil
}

func (b *Builder) writeSection(section *Section) {
	var (
		vars = section.vars()
	)
	if section.Meta.GetString("path") != "" {
		lookups := []string{
			utils.StringReplace(section.Meta.GetString("template"), vars),
			"section.html",
			"_default/section.html",
		}
		if tpl, ok := b.theme.LookupTemplate(lookups...); ok {
			for _, por := range section.Paginator() {
				b.write(tpl, por.URL, map[string]interface{}{
					"section":       section,
					"paginator":     por,
					"pages":         section.Pages,
					"current_lang":  section.Lang,
					"current_path":  por.URL,
					"current_index": por.PageNum,
				})
			}
		}
	}
	b.writeFormats(section.Meta, vars, map[string]interface{}{
		"section":      section,
		"pages":        section.Pages,
		"current_lang": section.Lang,
	})
}
