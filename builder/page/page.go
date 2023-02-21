package page

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
				m[k] = []string{b, strings.TrimSpace(v)}
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

func (page Page) Get(k string) interface{} {
	return page.Meta.Get(k)
}

func (page Page) HasPrev() bool {
	return page.Prev != nil
}

func (page Page) HasNext() bool {
	return page.Next != nil
}

func (page Page) HasPrevInType() bool {
	return page.PrevInType != nil
}

func (page Page) HasNextInType() bool {
	return page.NextInType != nil
}

func (pages Pages) Filter(filter interface{}) Pages {
	if filter == nil || filter == "" {
		return pages
	}
	matchList := func(value string) func(...string) bool {
		include := make(map[string]bool)
		exclude := make(map[string]bool)
		for _, val := range strings.Split(value, ",") {
			if strings.HasPrefix(val, "-") {
				exclude[val[1:]] = true
			} else {
				include[val] = true
			}
		}
		return func(values ...string) bool {
			for _, val := range values {
				if include[val] {
					return true
				}
				if exclude[val] {
					return false
				}
			}
			if len(include) > 0 {
				return false
			}
			return true
		}
	}

	npages := make(Pages, 0)
	switch fs := filter.(type) {
	case string:
		matcher := matchList(fs)
		for _, page := range pages {
			if !matcher(page.Type) {
				continue
			}
			npages = append(npages, page)
		}
	case map[string]interface{}:
		matchers := make([]func(*Page) bool, 0)
		for k, v := range fs {
			newk := k
			newv := v
			switch k {
			case "type":
				m := matchList(newv.(string))
				matcher := func(page *Page) bool {
					return m(page.Type)
				}
				matchers = append(matchers, matcher)
			case "section":
				m := matchList(newv.(string))
				matcher := func(page *Page) bool {
					return m(page.Section.Name())
				}
				matchers = append(matchers, matcher)
			default:
				matcher := func(page *Page) bool {
					mv, ok := page.Meta[newk]
					if !ok {
						return true
					}
					switch value := mv.(type) {
					case []string:
						m := matchList(newv.(string))
						return m(value...)
					// case []interface{}:
					//	m := matchList(newv.(string))
					//	newvalue := make([]string, len(value))
					//	for i, v := range value {
					//		newvalue[i] = v.(string)
					//	}
					//	return m(newvalue...)
					default:
						return utils.Compare(value, newv) >= 0
					}
				}
				matchers = append(matchers, matcher)
			}
		}
		for _, page := range pages {
			matched := true
			for _, m := range matchers {
				if !m(page) {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
			npages = append(npages, page)
		}
	default:
		npages = pages
	}
	return npages
}

func (pages Pages) OrderBy(key string) Pages {
	if key == "" {
		return pages
	}
	sortfs := make([]func(int, int) int, 0)
	for _, k := range strings.Split(key, ",") {
		var (
			sortf   func(int, int) int
			reverse bool
		)
		if strings.HasSuffix(strings.ToUpper(k), " DESC") {
			k = k[:len(k)-5]
			reverse = true
		}
		switch k {
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
				return utils.Compare(pages[i].Meta[k], pages[j].Meta[k])
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
		var result int
		for _, f := range sortfs {
			result = f(i, j)
			if result != 0 {
				return result > 0
			}
		}
		return result >= 0
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
			name = strings.ToLower(name)
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
	meta.load(filemeta)

	now := time.Now()
	page := &Page{
		File:     file,
		Type:     section.FirstName(),
		Meta:     meta,
		Date:     now,
		Modified: now,
		Section:  section,
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
	if page.Path == "" {
		vars := map[string]string{
			"{date:%Y}":      page.Date.Format("2006"),
			"{date:%m}":      page.Date.Format("01"),
			"{date:%d}":      page.Date.Format("02"),
			"{date:%H}":      page.Date.Format("15"),
			"{filename}":     utils.FileBaseName(file),
			"{slug}":         b.conf.GetSlug(page.Title),
			"{section}":      section.Name(),
			"{section:slug}": section.Slug,
		}
		if slug, ok := meta["slug"]; ok && slug != "" {
			vars["{slug}"] = slug.(string)
		}
		if vars["{filename}"] == "index" {
			vars["{filename}"] = page.Type
		}
		page.Path = utils.StringReplace(meta.GetString("page_path"), vars)
	}
	page.Path = b.conf.GetRelURL(page.Path)
	page.Permalink = b.conf.GetURL(page.Path)
	return b.hooks.AfterPageParse(page)
}
