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
		Type       string
		Title      string
		Date       time.Time
		Modified   time.Time
		Status     string
		Categories []string
		Authors    []string
		Tags       []string
		Slug       string
		SaveAs     string

		URL     string
		Content string
		Summary string

		Next     *Page
		Previous *Page
	}
	Pages []*Page

	Label struct {
		URL  string
		Name string
	}
	Section map[Label]Pages
)

func (sec Section) add(name string, page *Page) {
	label := Label{Name: name}
	pages, ok := sec[label]
	if !ok {
		pages = make(Pages, 0)
	}
	sec[label] = append(pages, page)
}

func matchList(list []interface{}, k string) bool {
	for _, t := range list {
		if t == k {
			return true
		}
	}
	return false
}

func (pages Pages) Filter(filter map[string]interface{}) Pages {
	newPages := make(Pages, 0)
	for _, page := range pages {
		matched := true
		for k, v := range filter {
			switch k {
			case "types":
				matched = matchList(v.([]interface{}), page.Type)
			}
			if !matched {
				break
			}
		}
		if !matched {
			continue
		}
		newPages = append(newPages, page)
	}
	return newPages
}

func (pages Pages) OrderBy(key string, reverse bool) Pages {
	switch key {
	case "date":
		sortf := func(i, j int) bool {
			if reverse {
				return pages[i].Date.Before(pages[j].Date)
			}
			return pages[i].Date.After(pages[j].Date)
		}
		sort.SliceStable(pages, sortf)
	case "type":
		sortf := func(i, j int) bool {
			if reverse {
				return pages[i].Type < pages[j].Type
			}
			return pages[i].Type > pages[j].Type
		}
		sort.SliceStable(pages, sortf)
	case "title":
		sortf := func(i, j int) bool {
			if reverse {
				return pages[i].Title < pages[j].Title
			}
			return pages[i].Title > pages[j].Title
		}
		sort.SliceStable(pages, sortf)
	}
	return pages
}

func (pages Pages) GroupBy(key string) Section {
	switch key {
	case "type":
		sec := make(Section)
		for _, page := range pages {
			sec.add(page.Type, page)
		}
		return sec
	case "tag":
		sec := make(Section)
		for _, page := range pages {
			for _, name := range page.Tags {
				sec.add(name, page)
			}
		}
		return sec
	case "author":
		sec := make(Section)
		for _, page := range pages {
			for _, name := range page.Authors {
				sec.add(name, page)
			}
		}
		return sec
	case "category":
		sec := make(Section)
		for _, page := range pages {
			for _, name := range page.Categories {
				sec.add(name, page)
			}
		}
		return sec
	default:
		if strings.HasPrefix(key, "date:") {
			format := key[5:]
			sec := make(Section)
			for _, page := range pages {
				sec.add(page.Date.Format(format), page)
			}
			return sec
		}
		return Section{Label{}: pages}
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
