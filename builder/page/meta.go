package page

import (
	"sort"
	"strings"
	"time"
)

type (
	paginator struct {
		HasNext bool
		HasPrev bool
		Next    *paginator
		Prev    *paginator
		Pages   int
		Page    int
		List    Pages
	}
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
		SaveAs     string

		URL     string
		Content string
		Summary string

		Prev *Page
		Next *Page
	}
	Pages []*Page

	Section map[string]Pages
)

func (sec Section) add(name string, page *Page) {
	pages, ok := sec[name]
	if !ok {
		pages = make(Pages, 0)
	}
	sec[name] = append(pages, page)
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

func (pages Pages) GroupBy(key string) Section {
	sec := make(Section)
	switch key {
	case "type":
		for _, page := range pages {
			sec.add(page.Type, page)
		}
		return sec
	case "tag":
		for _, page := range pages {
			for _, name := range page.Tags {
				sec.add(name, page)
			}
		}
		return sec
	case "author":
		for _, page := range pages {
			for _, name := range page.Authors {
				sec.add(name, page)
			}
		}
		return sec
	case "category":
		for _, page := range pages {
			for _, name := range page.Categories {
				sec.add(name, page)
			}
		}
		return sec
	default:
		if strings.HasPrefix(key, "date:") {
			format := key[5:]
			for _, page := range pages {
				sec.add(page.Date.Format(format), page)
			}
			return sec
		}
		sec[""] = pages
		return sec
	}
}

func (pages Pages) Paginator(number int) []*paginator {
	paginators := make([]*paginator, 0)
	if number <= 0 {
		paginators = append(paginators, &paginator{
			Page:  1,
			Pages: 1,
			List:  pages,
		})
		return paginators
	}
	var maxpage int

	length := len(pages)
	if length%number == 0 {
		maxpage = length / number
	} else {
		maxpage = length/number + 1
	}

	for i := 0; i*number < length; i++ {
		por := &paginator{
			Page:  i + 1,
			Pages: maxpage,
		}
		end := (i + 1) * number
		if end > length {
			end = length
		}
		por.List = pages[i*number : end]
		paginators = append(paginators, por)
	}
	return paginators
}
