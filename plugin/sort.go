/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: sort.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-05 23:05:27 (CST)
 Last Update: Friday 2019-09-06 00:54:44 (CST)
		  By:
 Description:
 *********************************************************************************/
package plugin

import (
	"snow/core"
	"sort"
	"strings"
)

type SortPlugin struct {
	Plugin
}

func NewSortPlugin() *SortPlugin {
	return &SortPlugin{}
}

func (s *SortPlugin) ByDate(articles []*core.Article, reverse bool) {
	sortf := func(i, j int) bool {
		if reverse {
			return articles[i].Date.Before(articles[j].Date)
		}
		return articles[i].Date.After(articles[j].Date)
	}
	sort.SliceStable(articles, sortf)
}

func (s *SortPlugin) ByTitle(articles []*core.Article, reverse bool) {
	sortf := func(i, j int) bool {
		if reverse {
			return articles[i].Title < articles[j].Title
		}
		return articles[i].Title > articles[j].Title
	}
	sort.SliceStable(articles, sortf)
}

func (s *SortPlugin) Register() {
	var sortf func(artiles []*core.Article, reverse bool)

	for _, orderby := range core.G.Conf.Orderby {
		reverse := false
		if strings.HasPrefix("-", orderby) {
			reverse = true
			orderby = orderby[1:]
		}
		switch orderby {
		case "title":
			sortf = s.ByTitle
		default:
			sortf = s.ByDate
		}
		sortf(core.G.Articles, reverse)

		for k := range core.G.Categories {
			sortf(core.G.Categories[k], reverse)
		}
		for k := range core.G.Tags {
			sortf(core.G.Tags[k], reverse)
		}
		for k := range core.G.Authors {
			sortf(core.G.Authors[k], reverse)
		}
	}
}
