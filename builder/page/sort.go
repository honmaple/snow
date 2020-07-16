package page

import (
)

// import (
//	"sort"
//	"strings"
// )

// type Sort struct {
//	context.Plugin
//	conf    *snow.Configuration
//	context *context.Context
// }

// func newSort(conf *snow.Configuration, context *context.Context) *Sort {
//	return &Sort{conf: conf, context: context}
// }

// func (s *Sort) byDate(pages []*context.Page, reverse bool) {
//	sortf := func(i, j int) bool {
//		if reverse {
//			return pages[i].Date.Before(pages[j].Date)
//		}
//		return pages[i].Date.After(pages[j].Date)
//	}
//	sort.SliceStable(pages, sortf)
// }

// func (s *Sort) byTitle(pages []*context.Page, reverse bool) {
//	sortf := func(i, j int) bool {
//		if reverse {
//			return pages[i].Title < pages[j].Title
//		}
//		return pages[i].Title > pages[j].Title
//	}
//	sort.SliceStable(pages, sortf)
// }

// func (s *Sort) Register() {
//	var sortf func(pages []*context.Page, reverse bool)

//	for _, orderby := range s.conf.Orderby {
//		reverse := false
//		if strings.HasPrefix("-", orderby) {
//			reverse = true
//			orderby = orderby[1:]
//		}
//		switch orderby {
//		case "title":
//			sortf = s.byTitle
//		default:
//			sortf = s.byDate
//		}
//		sortf(s.context.Pages(), reverse)

//		for k := range s.context.Categories() {
//			sortf(s.context.Categories[k], reverse)
//		}
//		for k := range s.context.Tags() {
//			sortf(s.context.Tags[k], reverse)
//		}
//		for k := range s.context.Authors() {
//			sortf(s.context.Authors[k], reverse)
//		}
//	}
// }

// func sort(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
//	pages := in.Interface().([]*context.Page)
//	key := param.Interface().(string)
//	switch key {
//	case "title":
//	}
//	s := filepath.Join("/static", in.String())
//	return pongo2.AsValue(s), nil
// }
