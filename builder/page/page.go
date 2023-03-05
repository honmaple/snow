package page

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/utils"
	"github.com/spf13/cast"
)

type Meta map[string]interface{}

func (m Meta) load(other map[string]interface{}) {
	for k, v := range other {
		m[k] = v
	}
}

func (m Meta) clone() Meta {
	return utils.DeepCopy(m)
}

func (m Meta) Done() {
	for k, v := range m {
		switch value := v.(type) {
		case []interface{}:
			newvalue := make([]string, len(value))
			for i, v := range value {
				newvalue[i] = v.(string)
			}
			m[k] = newvalue
		case string:
			switch k {
			case "date", "modified":
				if t, err := utils.ParseTime(value); err == nil {
					m[k] = t
				}
			}
		}
	}
}

func (m Meta) Get(k string) interface{} {
	return m[k]
}

func (m Meta) GetInt(k string) int {
	return cast.ToInt(m[k])
}

func (m Meta) GetBool(k string) bool {
	return cast.ToBool(m[k])
}

func (m Meta) GetString(k string) string {
	return cast.ToString(m[k])
}

func (m Meta) GetSlice(k string) []string {
	return cast.ToStringSlice(m[k])
}

func (m Meta) GetStringMap(k string) map[string]interface{} {
	return cast.ToStringMap(m[k])
}

func (m Meta) Set(k, v string) {
	var realVal interface{}

	k = strings.ToLower(k)
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '[' && v[len(v)-1] == ']' {
		realVal = utils.SplitTrim(v[1:len(v)-1], ",")
	} else if b, err := strconv.Atoi(v); err == nil {
		realVal = b
	} else if b, err := strconv.ParseBool(v); err == nil {
		realVal = b
	} else {
		realVal = v
	}

	ss := utils.SplitTrim(k, ".")
	if len(ss) == 1 {
		oldv, ok := m[k]
		if ok {
			m[k] = utils.Merge(oldv, realVal)
		} else {
			m[k] = realVal
		}
		return
	}
	var result map[string]interface{}
	for i := len(ss) - 1; i >= 0; i-- {
		if i == len(ss)-1 {
			result = map[string]interface{}{
				ss[i]: realVal,
			}
		} else {
			result = map[string]interface{}{
				ss[i]: result,
			}
		}
	}
	for key, val := range result {
		if oldv, ok := m[key]; ok {
			m[key] = utils.Merge(oldv, val)
		} else {
			m[key] = val
		}
	}
}

type (
	Page struct {
		File     string
		Meta     Meta
		Type     string
		Lang     string
		Date     time.Time
		Modified time.Time

		Path      string
		Permalink string
		Aliases   []string
		Assets    []string

		Title   string
		Summary string
		Content string

		Prev       *Page
		Next       *Page
		PrevInType *Page
		NextInType *Page

		Section *Section
	}
	Pages []*Page
)

func filterExpr(filter string) func(*Page) bool {
	if filter == "" {
		return func(*Page) bool {
			return true
		}
	}
	newstr := make([]byte, 0, len(filter))
	for i := 0; i < len(filter); i++ {
		newstr = append(newstr, filter[i])
		if filter[i] == '=' {
			if i > 0 && filter[i-1] != '!' {
				newstr = append(newstr, '=')
			}
			if i < len(filter)-1 && filter[i+1] == '=' {
				i++
			}
		}
	}
	tpl, err := template.Expr(string(newstr))
	if err != nil {
		panic(err)
	}
	return func(page *Page) bool {
		args := page.Meta.clone()
		args["page"] = page
		args["type"] = page.Type

		result, err := tpl.Execute(map[string]interface{}(args))
		if err == nil {
			return result == "True"
		}
		return false
	}
}

func (page *Page) isHidden() bool {
	return page.Meta.GetBool("hidden")
}

func (page *Page) isSection() bool {
	return page.Meta.GetBool("section")
}

func (page *Page) Get(k string) interface{} {
	return page.Meta.Get(k)
}

func (page *Page) HasPrev() bool {
	return page.Prev != nil
}

func (page *Page) HasNext() bool {
	return page.Next != nil
}

func (page *Page) HasPrevInType() bool {
	return page.PrevInType != nil
}

func (page *Page) HasNextInType() bool {
	return page.NextInType != nil
}

func (pages Pages) First() *Page {
	if len(pages) > 0 {
		return pages[0]
	}
	return nil
}

func (pages Pages) Last() *Page {
	if len(pages) > 0 {
		return pages[len(pages)-1]
	}
	return nil
}

func (pages Pages) Filter(filter string) Pages {
	if filter == "" {
		return pages
	}
	npages := make(Pages, 0)

	expr := filterExpr(filter)
	for _, page := range pages {
		if expr(page) {
			npages = append(npages, page)
		}
	}
	return npages
}

func (pages Pages) OrderBy(key string) Pages {
	sortfs := make([]func(int, int) int, 0)
	for _, k := range strings.Split(key, ",") {
		var (
			sortf   func(int, int) int
			reverse bool
		)
		newk := k
		if strings.HasSuffix(strings.ToUpper(k), " DESC") {
			newk = newk[:len(k)-5]
			reverse = true
		}
		switch newk {
		case "title":
			sortf = func(i, j int) int {
				return strings.Compare(pages[i].Title, pages[j].Title)
			}
		case "date":
			sortf = func(i, j int) int {
				return utils.Compare(pages[i].Date, pages[j].Date)
			}
		case "modified":
			sortf = func(i, j int) int {
				return utils.Compare(pages[i].Modified, pages[j].Modified)
			}
		case "type":
			sortf = func(i, j int) int {
				return strings.Compare(pages[i].Type, pages[j].Type)
			}
		default:
			sortf = func(i, j int) int {
				return utils.Compare(pages[i].Meta[newk], pages[j].Meta[newk])
			}
		}
		if reverse {
			sortfs = append(sortfs, func(i, j int) int {
				return 0 - sortf(i, j)
			})
		} else {
			sortfs = append(sortfs, sortf)
		}
	}
	sort.SliceStable(pages, func(i, j int) bool {
		for _, f := range sortfs {
			result := f(i, j)
			if result != 0 {
				return result > 0
			}
		}
		// 增加一个默认排序, 避免时间相同时排序混乱
		return strings.Compare(pages[i].Title, pages[j].Title) >= 0
	})
	return pages
}

func (pages Pages) GroupBy(key string) TaxonomyTerms {
	var groupf func(*Page) []string

	terms := make(TaxonomyTerms, 0)
	termm := make(map[string]*TaxonomyTerm)

	if strings.HasPrefix(key, "date:") {
		format := key[5:]
		groupf = func(page *Page) []string {
			return []string{page.Date.Format(format)}
		}
	} else {
		groupf = func(page *Page) []string {
			value := page.Meta[key]
			switch v := value.(type) {
			case string:
				return []string{v}
			case []string:
				return v
			case []interface{}:
				as := make([]string, len(v))
				for i, vv := range v {
					as[i] = vv.(string)
				}
				return as
			default:
				return nil
			}
		}
	}
	for _, page := range pages {
		for _, name := range groupf(page) {
			// 不要忽略大小写
			// name = strings.ToLower(name)
			var parent *TaxonomyTerm

			for _, subname := range utils.SplitPrefix(name, "/") {
				// for _, subname := range names {
				term, ok := termm[subname]
				if !ok {
					term = &TaxonomyTerm{Name: subname[strings.LastIndex(subname, "/")+1:], Parent: parent}
					if parent == nil {
						terms = append(terms, term)
					}
				}
				term.List = append(term.List, page)
				termm[subname] = term
				if parent != nil {
					if !parent.Children.Has(subname) {
						parent.Children = append(parent.Children, term)
					}
				}
				parent = term
			}
		}
	}
	return terms
}

func (pages Pages) Paginator(number int, path string, paginatePath string) []*paginator {
	list := make([]interface{}, len(pages))
	for i, page := range pages {
		list[i] = page
	}
	return Paginator(list, number, path, paginatePath)
}

func (b *Builder) findPage(file string, langs ...string) *Page {
	lang := b.getLang(langs...)

	b.mu.RLock()
	defer b.mu.RUnlock()
	m, ok := b.pages[lang]
	if !ok {
		return nil
	}
	return m[file]
}

func (b *Builder) insertPage(file string) *Page {
	filemeta, err := b.readFile(file)
	if err != nil {
		return nil
	}
	lang := b.findLanguage(file, filemeta)
	section := b.findSection(filepath.Dir(file), lang)

	meta := section.Meta.clone()
	meta["path"] = meta["page_path"]
	meta["template"] = meta["page_template"]
	meta["formats"] = meta["page_formats"]
	delete(meta, "slug")
	delete(meta, "title")
	meta.load(filemeta)

	now := time.Now()
	page := &Page{
		File:    file,
		Type:    section.FirstName(),
		Meta:    meta,
		Date:    now,
		Lang:    lang,
		Section: section,
	}
	for k, v := range meta {
		if v == "" {
			continue
		}
		if vs, ok := v.([]interface{}); ok {
			res := make([]string, len(vs))
			for i, vv := range vs {
				res[i] = vv.(string)
			}
			v = res
		}
		switch strings.ToLower(k) {
		case "type":
			page.Type = v.(string)
		case "title":
			page.Title = v.(string)
		case "date":
			page.Date = v.(time.Time)
		case "modified":
			page.Modified = v.(time.Time)
		case "url", "save_as":
			page.Path = v.(string)
		case "aliases":
			page.Aliases = v.([]string)
		case "summary":
			page.Summary = v.(string)
		case "content":
			page.Content = v.(string)
		}
	}
	filename := utils.FileBaseName(file)
	if filename == "index" && section.Parent != nil {
		filename = filepath.Base(filepath.Dir(file))
	}
	if page.Title == "" {
		page.Title = filename
	}
	if page.Modified.IsZero() {
		page.Modified = page.Date
	}
	if page.Path == "" {
		vars := map[string]string{
			"{date:%Y}":      page.Date.Format("2006"),
			"{date:%m}":      page.Date.Format("01"),
			"{date:%d}":      page.Date.Format("02"),
			"{date:%H}":      page.Date.Format("15"),
			"{filename}":     filename,
			"{section}":      section.Name(),
			"{section:slug}": section.Slug,
		}
		if slug, ok := meta["slug"]; ok && slug != "" {
			vars["{slug}"] = slug.(string)
		} else {
			vars["{slug}"] = b.conf.GetSlug(page.Title)
		}
		page.Path = utils.StringReplace(meta.GetString("path"), vars)
	}
	page.Path = b.conf.GetRelURL(page.Path, page.Lang)
	page.Permalink = b.conf.GetURL(page.Path)

	if !b.buildFilter(page) {
		return nil
	}

	page = b.hooks.AfterPageParse(page)

	b.mu.Lock()
	defer b.mu.Unlock()
	if page.isHidden() {
		if _, ok := b.hiddenPages[lang]; !ok {
			b.hiddenPages[lang] = make(map[string]*Page)
		}
		b.hiddenPages[lang][file] = page
		section.HiddenPages = append(section.HiddenPages, page)
	} else if page.isSection() {
		if _, ok := b.sectionPages[lang]; !ok {
			b.sectionPages[lang] = make(map[string]*Page)
		}
		b.sectionPages[lang][file] = page
		section.SectionPages = append(section.SectionPages, page)
	} else {
		if _, ok := b.pages[lang]; !ok {
			b.pages[lang] = make(map[string]*Page)
		}
		b.pages[lang][file] = page
		section.Pages = append(section.Pages, page)
	}
	return page
}

func (b *Builder) writePage(page *Page) {
	if !page.isSection() {
		ctx := map[string]interface{}{
			"page":         page,
			"current_url":  page.Permalink,
			"current_path": page.Path,
			"current_lang": page.Lang,
		}
		if tpl, ok := b.theme.LookupTemplate(page.Meta.GetString("template")); ok {
			b.write(tpl, page.Path, ctx)
		}
		if tpl, ok := b.theme.LookupTemplate("aliase.html", "_internal/aliase.html"); ok {
			for _, aliase := range page.Aliases {
				b.write(tpl, aliase, ctx)
			}
		}
		b.writeFormats(page.Meta, nil, ctx)
		return
	}

	path := page.Meta.GetString("path")
	if path == "" {
		return
	}

	section := &Section{
		File:    page.File,
		Meta:    page.Meta,
		Title:   page.Title,
		Content: page.Content,
		Pages:   page.Section.allPages(),
		Lang:    page.Lang,
		Parent:  page.Section,
	}
	section.Slug = b.conf.GetSlug(section.Title)
	section.Path = b.conf.GetRelURL(path, page.Lang)
	section.Permalink = b.conf.GetURL(section.Path)
	section.Pages = section.Pages.Filter(page.Meta.GetString("filter")).OrderBy(page.Meta.GetString("orderby"))

	b.writeSection(section)
}
