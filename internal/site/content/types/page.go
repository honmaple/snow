package types

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Page struct {
		*Node

		IsBundle  bool
		Draft     bool
		Hidden    bool
		WordCount int64

		Date     time.Time
		Modified time.Time

		Path      string
		Permalink string

		Assets  Assets
		Formats Formats

		Prev *Page
		Next *Page
	}
	Pages []*Page
)

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

func (pages Pages) Related() *RelatedPages {
	return &RelatedPages{list: pages}
}

func (pages Pages) FilterBy(filter string) Pages {
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

func (pages Pages) SortBy(key string) {
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
			return utils.Compare(pages[i].FrontMatter.Get(k), pages[j].FrontMatter.Get(k))
		}
	}))
}

func (pages Pages) OrderBy(key string) Pages {
	newPs := make(Pages, len(pages))
	copy(newPs, pages)

	newPs.SortBy(key)
	return newPs
}

func (pages Pages) GroupBy(key string) TaxonomyTerms {
	var groupf func(*Page) []string

	if strings.HasPrefix(key, "date:") {
		format := key[5:]
		groupf = func(page *Page) []string {
			return []string{page.Date.Format(format)}
		}
	} else {
		groupf = func(page *Page) []string {
			value := page.FrontMatter.Get(key)
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

	results := make(TaxonomyTerms, 0)
	resultMap := make(map[string]*TaxonomyTerm)
	for _, page := range pages {
		for _, name := range groupf(page) {
			var (
				currentTerm *TaxonomyTerm
				currentName string
			)
			for part := range strings.SplitSeq(strings.Trim(name, "/"), "/") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if currentName == "" {
					currentName = part
				} else {
					currentName += "/" + part
				}

				term, ok := resultMap[currentName]
				if !ok {
					term = &TaxonomyTerm{
						Name:     part,
						Pages:    make(Pages, 0),
						Parent:   currentTerm,
						Children: make(TaxonomyTerms, 0),
					}
					resultMap[currentName] = term

					if currentTerm == nil {
						results = append(results, term)
					} else {
						currentTerm.Children = append(currentTerm.Children, term)
					}
				}
				term.Pages = append(term.Pages, page)

				currentTerm = term
			}
		}
	}
	return results
}

func (pages Pages) Paginate(number int, path string, paginatePath string) []*Paginator[*Page] {
	return Paginate(pages, number, path, paginatePath)
}

func FilterExpr(filter string) func(*Page) bool {
	tpl, err := pongo2.FromString("{{" + filter + "}}")
	if err != nil {
		return func(page *Page) bool {
			return true
		}
	}
	return func(page *Page) bool {
		args := page.FrontMatter.AllSettings()

		result, err := tpl.Execute(args)
		return err == nil && result == "True"
	}
}

type RelatedPages struct {
	mu    sync.RWMutex
	once  sync.Once
	list  Pages
	index map[*Page]int
}

func (r *RelatedPages) init() {
	r.once.Do(func() {
		r.index = make(map[*Page]int)
		for i, page := range r.list {
			r.index[page] = i
		}
	})
}

func (r *RelatedPages) Prev(page *Page) *Page {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()
	idx, ok := r.index[page]
	if !ok || idx == 0 {
		return nil
	}
	return r.list[idx-1]
}

func (r *RelatedPages) Next(page *Page) *Page {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()
	idx, ok := r.index[page]
	if !ok || idx == len(r.list)-1 {
		return nil
	}
	return r.list[idx+1]
}
