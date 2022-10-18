package page

import (
	"sort"
	"strings"
	"time"

	"github.com/honmaple/snow/utils"
)

type (
	Page struct {
		Meta       map[string]string
		Type       string
		Title      string
		Date       time.Time
		Modified   time.Time
		Categories []string
		Authors    []string
		Tags       []string
		Slug       string

		URL     string
		Content string
		Summary string

		Prev *Page
		Next *Page
	}
	Pages []*Page

	Label struct {
		Name     string
		List     Pages
		Children Labels
	}
	Labels  []*Label
)

func (ls Labels) Has(name string) bool {
	for _, l := range ls {
		if l.Name == name {
			return true
		}
	}
	return false
}

func (labels Labels) add(name string, page *Page, labelm map[string]*Label, tree bool) Labels {
	if !tree {
		label, ok := labelm[name]
		if !ok {
			label = &Label{Name: name}
			labels = append(labels, label)
		}
		label.List = append(label.List, page)
		labelm[name] = label
		return labels
	}
	var parent *Label
	for _, subname := range utils.SplitPrefix(name, "/") {
		label, ok := labelm[subname]
		if !ok {
			label = &Label{Name: subname}
			labels = append(labels, label)
		}
		label.List = append(label.List, page)
		labelm[subname] = label
		if parent != nil {
			if !parent.Children.Has(subname) {
				parent.Children = append(parent.Children, label)
			}
		}
		parent = label
	}
	return labels
}

func (page Page) HasPrev() bool {
	return page.Prev != nil
}

func (page Page) HasNext() bool {
	return page.Next != nil
}

func (pages Pages) Filter(filter interface{}) Pages {
	matchList := func(value string) func(string) int {
		include := make(map[string]bool)
		exclude := make(map[string]bool)
		for _, typ := range strings.Split(value, ",") {
			if strings.HasPrefix(typ, "-") {
				exclude[typ[1:]] = true
			} else {
				include[typ] = true
			}
		}
		return func(typ string) int {
			if include[typ] {
				return 1
			}
			if exclude[typ] {
				return -1
			}
			return 0
		}
	}
	switch fs := filter.(type) {
	case string:
		matcher := matchList(fs)
		newPages := make(Pages, 0)
		for _, page := range pages {
			if matcher(page.Type) >= 0 {
				newPages = append(newPages, page)
			}
		}
		return newPages
	case map[string]interface{}:
		matchers := make([]func(*Page) bool, 0)
		for k, v := range fs {
			switch k {
			case "type":
				matcher := matchList(v.(string))
				matchers = append(matchers, func(page *Page) bool {
					return matcher(page.Type) >= 0
				})
			case "tag":
				matcher := matchList(v.(string))
				matchers = append(matchers, func(page *Page) bool {
					matched := false
					for _, name := range page.Tags {
						if m := matcher(name); m < 0 {
							return false
						} else if m > 0 {
							matched = true
						}
					}
					return matched
				})
			case "category":
				matcher := matchList(v.(string))
				matchers = append(matchers, func(page *Page) bool {
					matched := false
					for _, name := range page.Categories {
						if m := matcher(name); m < 0 {
							return false
						} else if m > 0 {
							matched = true
						}
					}
					return matched
				})
			case "author":
				matcher := matchList(v.(string))
				matchers = append(matchers, func(page *Page) bool {
					matched := false
					for _, name := range page.Authors {
						if m := matcher(name); m < 0 {
							return false
						} else if m > 0 {
							matched = true
						}
					}
					return matched
				})
			}
		}
		matchAll := func(page *Page) bool {
			for _, m := range matchers {
				if !m(page) {
					return false
				}
			}
			return true
		}
		newPages := make(Pages, 0)
		for _, page := range pages {
			if matchAll(page) {
				newPages = append(newPages, page)
			}
		}
		return newPages
	default:
		return pages
	}
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
		if strings.HasSuffix(k, " desc") || strings.HasSuffix(k, " DESC") {
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
				d1 := pages[i].Date
				d2 := pages[j].Date
				if d1 == d2 {
					return 0
				}
				if d1.Before(d2) {
					return 1
				}
				return -1
			}
		case "modified":
			sortf = func(i, j int) int {
				d1 := pages[i].Modified
				d2 := pages[j].Modified
				if d1 == d2 {
					return 0
				}
				if d1.Before(d2) {
					return 1
				}
				return -1
			}
		case "type":
			sortf = func(i, j int) int {
				return strings.Compare(pages[i].Type, pages[j].Type)
			}
		default:
			sortf = func(i, j int) int {
				return strings.Compare(pages[i].Meta[k], pages[j].Meta[k])
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

func (pages Pages) groupBy(key string) (map[string]*Label, Labels) {
	labels := make(Labels, 0)
	labelm := make(map[string]*Label)
	switch key {
	case "type":
		for _, page := range pages {
			labels.add(page.Type, page, labelm, false)
		}
	case "category":
		for _, page := range pages {
			// 增加子分类功能 Linux/Python/Flask
			for _, name := range page.Categories {
				labels = labels.add(name, page, labelm, true)
			}
		}
	case "tag":
		for _, page := range pages {
			for _, name := range page.Tags {
				labels = labels.add(name, page, labelm, false)
			}
		}
	case "author":
		for _, page := range pages {
			for _, name := range page.Authors {
				labels = labels.add(name, page, labelm, false)
			}
		}
	default:
		if strings.HasPrefix(key, "date:") {
			format := key[5:]
			for _, page := range pages {
				labels = labels.add(page.Date.Format(format), page, labelm, false)
			}
		} else {
			labels = append(labels, &Label{List: pages})
		}
	}
	return labelm, labels
}

func (pages Pages) GroupBy(key string) Labels {
	_, labels := pages.groupBy(key)
	return labels
}
