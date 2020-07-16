package page

import (
	"time"
)

type (
	Page struct {
		Title      string
		Date       time.Time
		Modified   time.Time
		Status     string
		Categories []string
		Authors    []string
		Tags       []string
		SaveAs     string

		URL     string
		Content string
		Summary string
		Draft   bool

		Next     *Page
		Previous *Page
	}
	label struct {
		URL   string
		Name  string
		Pages []*Page
	}
	Section map[string]*label
)

func (s *label) add(page *Page) {
	if s.Pages == nil {
		s.Pages = make([]*Page, 0)
	}
	s.Pages = append(s.Pages, page)
}

func (s Section) add(name string, page *Page) {
	if _, ok := s[name]; !ok {
		s[name] = &label{Name: name, Pages: make([]*Page, 0)}
	}
	s[name].add(page)
}

type paginator struct {
	HasNext bool
	HasPrev bool
	Next    *paginator
	Prev    *paginator
	Pages   int
	Page    int
	List    []*Page
}

func Paginator(pages []*Page, number int) []*paginator {
	var maxpage int

	length := len(pages)
	if length%number == 0 {
		maxpage = length / number
	} else {
		maxpage = length/number + 1
	}

	paginators := make([]*paginator, 0)
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
