package page

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
	"github.com/spf13/cast"
)

type Meta map[string]interface{}

func (m Meta) load(other map[string]interface{}) {
	for k, v := range other {
		m[k] = v
	}
}

func (m Meta) copy() Meta {
	nm := make(Meta)
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

func (m Meta) Fix() {
	for k, v := range m {
		switch value := v.(type) {
		case []interface{}:
			newvalue := make([]string, len(value))
			for i, v := range value {
				newvalue[i] = v.(string)
			}
			m[k] = newvalue
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

func (m Meta) Set(conf config.Config, k, v string) {
	k = strings.ToLower(k)
	switch k {
	case "date", "modified":
		if t, err := utils.ParseTime(v); err == nil {
			m[k] = t
		}
	default:
		if a, ok := m[k]; ok {
			switch b := a.(type) {
			case string:
				m[k] = b + "," + strings.TrimSpace(v)
			case []string:
				m[k] = append(b, strings.TrimSpace(v))
			}
		} else if conf.IsSet(fmt.Sprintf("taxonomies.%s", k)) {
			m[k] = utils.SplitTrim(v, ",")
		} else {
			m[k] = strings.TrimSpace(v)
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
		args := page.Meta.copy()
		args["page"] = page
		args["type"] = page.Type

		result, err := tpl.Execute(map[string]interface{}(args))
		if err == nil {
			return result == "True"
		}
		return false
	}
}

func (page *Page) isDraft() bool {
	return page.Meta.GetBool("draft")
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

func (b *Builder) readFile(file string) (Meta, error) {
	reader, ok := b.readers[filepath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no reader for %s", file)
	}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return reader.Read(bytes.NewBuffer(buf))
}

func (b *Builder) newPage(section *Section, file string, filemeta Meta) *Page {
	meta := section.Meta.copy()
	meta["path"] = meta["page_path"]
	meta["template"] = meta["page_template"]
	meta.load(filemeta)

	now := time.Now()
	page := &Page{
		File:    file,
		Type:    section.FirstName(),
		Meta:    meta,
		Date:    now,
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
		case "lang":
			page.Lang = v.(string)
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
	if page.Lang == "" {
		lang := filepath.Ext(filename)
		if lang == "" {
			page.Lang = b.conf.GetString("site.language")
		} else {
			langs := b.conf.GetStringMap("languages")
			lang = lang[1:]
			if _, ok := langs[lang]; ok {
				page.Lang = lang
			}
		}
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
			"{lang}":         page.Lang,
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
	page.Path = b.conf.GetRelURL(page.Path)
	page.Permalink = b.conf.GetURL(page.Path)
	return b.hooks.AfterPageParse(page)
}
