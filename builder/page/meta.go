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

		Prev       *Page
		Next       *Page
		PrevInType *Page
		NextInType *Page
	}
	Pages []*Page

	Label struct {
		Name     string
		List     Pages
		Children Labels
	}
	Labels []*Label
)

func (ls Labels) Has(name string) bool {
	for _, l := range ls {
		if l.Name == name {
			return true
		}
	}
	return false
}

func (ls Labels) Orderby(key string) Labels {
	var (
		reverse = false
		sortf   func(int, int) bool
	)
	if strings.HasSuffix(strings.ToLower(key), " desc") {
		key = key[:len(key)-5]
		reverse = true
	}
	switch key {
	case "name":
		sortf = func(i, j int) bool {
			return ls[i].Name < ls[j].Name
		}
	case "count":
		sortf = func(i, j int) bool {
			return len(ls[i].List) < len(ls[j].List)
		}
	}
	if sortf == nil {
		return ls
	}
	if reverse {
		sort.SliceStable(ls, func(i, j int) bool {
			return !sortf(i, j)
		})
	} else {
		sort.SliceStable(ls, sortf)
	}
	return ls
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

func (pages Pages) filter(filter interface{}) chan *Page {
	ch := make(chan *Page)
	go func() {
		defer close(ch)

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
				if exclude[typ] {
					return -1
				}
				if include[typ] {
					return 1
				}
				return 0
			}
		}

		switch fs := filter.(type) {
		case string:
			matcher := matchList(fs)
			for _, page := range pages {
				if matcher(page.Type) < 0 {
					continue
				}
				ch <- page
			}
		case map[string]interface{}:
			matcherf := func(value string, names func(*Page) []string) func(*Page) bool {
				m := matchList(value)
				return func(page *Page) bool {
					matched := false
					for _, name := range names(page) {
						if m := m(name); m < 0 {
							return false
						} else if m >= 0 {
							matched = true
						}
					}
					return matched
				}
			}
			matchers := make([]func(*Page) bool, 0)
			for k, v := range fs {
				switch k {
				case "type":
					matcher := matcherf(v.(string), func(page *Page) []string {
						return []string{page.Type}
					})
					matchers = append(matchers, matcher)
				case "tag":
					matcher := matcherf(v.(string), func(page *Page) []string {
						return page.Tags
					})
					matchers = append(matchers, matcher)
				case "category":
					matcher := matcherf(v.(string), func(page *Page) []string {
						return page.Categories
					})
					matchers = append(matchers, matcher)
				case "author":
					matcher := matcherf(v.(string), func(page *Page) []string {
						return page.Authors
					})
					matchers = append(matchers, matcher)
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
			for _, page := range pages {
				if !matchAll(page) {
					continue
				}
				ch <- page
			}
		default:
			for _, page := range pages {
				ch <- page
			}
		}
	}()
	return ch
}

func (pages Pages) Filter(key interface{}) Pages {
	ps := make(Pages, 0)
	for page := range pages.filter(key) {
		ps = append(ps, page)
	}
	return ps
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
		if strings.HasSuffix(strings.ToLower(k), " desc") {
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

func (pages Pages) GroupBy(key string) Labels {
	var groupf func(*Page) []string

	labels := make(Labels, 0)
	labelm := make(map[string]*Label)
	switch key {
	case "type":
		groupf = func(page *Page) []string {
			return []string{page.Type}
		}
	case "category":
		groupf = func(page *Page) []string {
			return page.Categories
		}
	case "tag":
		groupf = func(page *Page) []string {
			return page.Tags
		}
	case "author":
		groupf = func(page *Page) []string {
			return page.Authors
		}
	default:
		if strings.HasPrefix(key, "date:") {
			format := key[5:]
			groupf = func(page *Page) []string {
				return []string{page.Date.Format(format)}
			}
		}
	}
	if groupf == nil {
		labels = append(labels, &Label{List: pages})
		return labels
	}
	for _, page := range pages {
		for _, name := range groupf(page) {
			var parent *Label

			for _, subname := range utils.SplitPrefix(name, "/") {
				label, ok := labelm[subname]
				if !ok {
					label = &Label{Name: subname}
					if parent == nil {
						labels = append(labels, label)
					}
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
		}
	}
	return labels
}
