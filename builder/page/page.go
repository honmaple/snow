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
		Lang     string
		Date     time.Time
		Modified time.Time

		Slug      string
		Path      string
		Permalink string
		Aliases   []string
		Assets    []string

		Title   string
		Summary string
		Content string

		Prev          *Page
		Next          *Page
		PrevInSection *Page
		NextInSection *Page

		Formats Formats
		Section *Section
	}
	Pages []*Page
)

func FilterExpr(filter string) func(*Page) bool {
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
		args["type"] = page.Section.FirstName()
		args["section"] = page.Section.RealName()

		result, err := tpl.Execute(map[string]interface{}(args))
		return err == nil && result == "True"
	}
}

func (page *Page) realPath(pathstr string) string {
	filename := utils.FileBaseName(page.File)
	if filename == "index" && !page.Section.isRoot() {
		filename = filepath.Base(filepath.Dir(page.File))
	}
	vars := map[string]string{
		"{date:%Y}":      page.Date.Format("2006"),
		"{date:%m}":      page.Date.Format("01"),
		"{date:%d}":      page.Date.Format("02"),
		"{date:%H}":      page.Date.Format("15"),
		"{slug}":         page.Slug,
		"{filename}":     filename,
		"{section}":      page.Section.RealName(),
		"{section:slug}": page.Section.Slug,
	}
	return utils.StringReplace(pathstr, vars)
}

func (page *Page) isNormal() bool {
	return !page.isHidden() && !page.isSection()
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

func (pages Pages) setRelation(section bool) {
	var prev *Page

	for _, page := range pages {
		if section {
			page.PrevInSection = prev
		} else {
			page.Prev = prev
		}
		if prev != nil {
			if section {
				prev.NextInSection = page
			} else {
				prev.Next = page
			}
		}
		prev = page
	}
}

func (pages Pages) setSort(key string) {
	sort.SliceStable(pages, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			// "-"表示默认排序, 避免时间相同时排序混乱
			return 0 - strings.Compare(pages[i].Title, pages[j].Title)
		case "title":
			return strings.Compare(pages[i].Title, pages[j].Title)
		case "date":
			return utils.Compare(pages[i].Date, pages[j].Date)
		case "modified":
			return utils.Compare(pages[i].Modified, pages[j].Modified)
		default:
			return utils.Compare(pages[i].Meta[k], pages[j].Meta[k])
		}
	}))
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
	npages := make(Pages, 0, len(pages))

	expr := FilterExpr(filter)
	for _, page := range pages {
		if expr(page) {
			npages = append(npages, page)
		}
	}
	return npages
}

func (pages Pages) OrderBy(key string) Pages {
	newPs := make(Pages, len(pages))
	copy(newPs, pages)

	newPs.setSort(key)
	return newPs
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
			default:
				return nil
			}
		}
	}
	for _, page := range pages {
		for _, name := range groupf(page) {
			// 不要忽略大小写
			var parent *TaxonomyTerm

			for _, subname := range utils.SplitPrefix(name, "/") {
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

func (b *Builder) insertPage(file string) *Page {
	section := b.ctx.findSection(filepath.Dir(file))
	if section == nil {
		return nil
	}

	filemeta, err := b.readFile(file)
	if err != nil {
		return nil
	}

	meta := section.Meta.clone()
	meta["path"] = meta["page_path"]
	meta["template"] = meta["page_template"]
	meta["formats"] = meta["page_formats"]
	delete(meta, "slug")
	delete(meta, "title")
	delete(meta, "content")
	delete(meta, "summary")
	meta.load(filemeta)

	lang := b.findLang(file, meta)
	if lang != b.conf.Site.Language {
		return nil
	}

	page := &Page{
		Meta:    meta,
		Lang:    lang,
		File:    file,
		Date:    time.Now(),
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
		case "slug":
			page.Slug = v.(string)
		case "title":
			page.Title = v.(string)
		case "date":
			if t, ok := v.(time.Time); ok {
				page.Date = t
			} else if t, err := utils.ParseTime(v.(string)); err == nil {
				page.Date = t
			}
		case "modified":
			if t, ok := v.(time.Time); ok {
				page.Modified = t
			} else if t, err := utils.ParseTime(v.(string)); err == nil {
				page.Modified = t
			}
		case "url", "save_as":
			page.Path = v.(string)
		case "aliases":
			page.Aliases = v.([]string)
		case "summary":
			page.Summary = v.(string)
		case "content":
			page.Content = v.(string)
		}
		meta[k] = v
	}
	if page.Title == "" {
		filename := utils.FileBaseName(file)
		if filename == "index" && !section.isRoot() {
			filename = filepath.Base(filepath.Dir(file))
		}
		page.Title = filename
	}
	if page.Modified.IsZero() {
		page.Modified = page.Date
	}
	if page.Slug == "" {
		page.Slug = b.conf.GetSlug(page.Title)
	}
	if page.Path == "" {
		page.Path = page.realPath(meta.GetString("path"))
	}
	page.Path = b.conf.GetRelURL(page.Path)
	page.Permalink = b.conf.GetURL(page.Path)
	page.Formats = b.formats(page.Meta, nil)

	page = b.hooks.Page(page)
	if page == nil {
		return nil
	}

	b.ctx.insertPage(page)

	b.insertTaxonomies(page)
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
		if tpl := b.theme.LookupTemplate(page.Meta.GetString("template")); tpl != nil {
			b.write(tpl, page.Path, ctx)
		}
		if tpl := b.theme.LookupTemplate("alias.html", "_internal/partials/alias.html"); tpl != nil {
			for _, aliase := range page.Aliases {
				if !strings.HasPrefix(aliase, "/") {
					aliase = filepath.Join(filepath.Dir(page.Path), aliase)
				}
				b.write(tpl, aliase, ctx)
			}
		}
		for _, format := range page.Formats {
			if tpl := b.theme.LookupTemplate(format.Template); tpl != nil {
				b.write(tpl, format.Path, map[string]interface{}{
					"page":         page,
					"current_lang": page.Lang,
					"current_url":  format.Permalink,
					"current_path": format.Path,
				})
			}
		}
		return
	}
	section := &Section{
		Meta:      page.Meta,
		Lang:      page.Lang,
		File:      page.File,
		Slug:      page.Slug,
		Title:     page.Title,
		Content:   page.Content,
		Path:      page.Path,
		Permalink: page.Permalink,
		Parent:    page.Section,
		Formats:   page.Formats,
	}
	section.Pages = b.ctx.Pages().Filter(page.Meta.GetString("filter")).OrderBy(page.Meta.GetString("orderby"))

	b.writeSection(section)
}
