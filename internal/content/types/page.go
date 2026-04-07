package types

import (
	// "strings"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Page struct {
		File        string
		FrontMatter *FrontMatter

		Lang     string
		Date     time.Time
		Modified time.Time

		Title       string
		Description string
		Summary     string
		Content     string

		Slug         string
		Path         string
		Permalink    string
		RelPermalink string

		Draft      bool
		Assets     []*Asset
		WordCount  int64
		Taxonomies Taxonomies

		// Params        map[string]any
		Formats Formats
	}
	Pages []*Page
)

func (page *Page) IsBundle() bool {
	return utils.FileBaseName(page.File) == "index"
}

func (pages Pages) Paginate(number int, path string, paginatePath string) []*Paginator[*Page] {
	return Paginate(pages, number, path, paginatePath)
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

// func (pages Pages) OrderBy(key string) Pages {
//	newPs := make(Pages, len(pages))
//	copy(newPs, pages)

//	newPs.setSort(key)
//	return newPs
// }

// func (pages Pages) GroupBy(key string) TaxonomyTerms {
//	var groupf func(*Page) []string

//	terms := make(TaxonomyTerms, 0)
//	termm := make(map[string]*TaxonomyTerm)

//	if strings.HasPrefix(key, "date:") {
//		format := key[5:]
//		groupf = func(page *Page) []string {
//			return []string{page.Date.Format(format)}
//		}
//	} else {
//		groupf = func(page *Page) []string {
//			value := page.Meta[key]
//			switch v := value.(type) {
//			case string:
//				return []string{v}
//			case []string:
//				return v
//			default:
//				return nil
//			}
//		}
//	}
//	for _, page := range pages {
//		for _, name := range groupf(page) {
//			// 不要忽略大小写
//			var parent *TaxonomyTerm

//			for _, subname := range utils.SplitPrefix(name, "/") {
//				term, ok := termm[subname]
//				if !ok {
//					term = &TaxonomyTerm{Name: subname[strings.LastIndex(subname, "/")+1:], Parent: parent}
//					if parent == nil {
//						terms = append(terms, term)
//					}
//				}
//				term.List = append(term.List, page)
//				termm[subname] = term
//				if parent != nil {
//					if !parent.Children.Has(subname) {
//						parent.Children = append(parent.Children, term)
//					}
//				}
//				parent = term
//			}
//		}
//	}
//	return terms
// }

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
