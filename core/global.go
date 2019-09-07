/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: global.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-30 15:47:08 (CST)
 Last Update: Saturday 2019-09-07 17:02:38 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
)

// Global ..
type Global struct {
	Authors    map[string][]*Article
	Categories map[string][]*Article
	Tags       map[string][]*Article
	Articles   []*Article
	Pages      []*Page

	Vars    map[string]interface{}
	Context map[string]interface{}

	Logger *logrus.Logger
	Conf   *Configuration

	Develop bool
	Server  bool
	Version string
}

var G = &Global{
	Tags:       make(map[string][]*Article),
	Authors:    make(map[string][]*Article),
	Categories: make(map[string][]*Article),
	Pages:      make([]*Page, 0),
	Articles:   make([]*Article, 0),
	Vars:       make(map[string]interface{}),
	Context:    make(map[string]interface{}),
}

// AddVar ..
func (s *Global) AddVar(key string, value interface{}) {
	s.Vars[key] = value
}

// AddTag ..
func (s *Global) AddTag(article *Article) {
	articles := []*Article{article}
	for _, name := range article.Tags {
		if v, ok := s.Tags[name]; ok {
			s.Tags[name] = append(v, article)
		} else {
			s.Tags[name] = articles

		}
	}
}

// AddCategory ..
func (s *Global) AddCategory(article *Article) {
	if v, ok := s.Categories[article.Category]; ok {
		s.Categories[article.Category] = append(v, article)
	} else {
		s.Categories[article.Category] = []*Article{article}
	}
}

// AddAuthor ..
func (s *Global) AddAuthor(article *Article) {
	articles := []*Article{article}
	authors := append(article.Authors, article.Author)
	for _, name := range authors {
		if v, ok := s.Authors[name]; ok {
			s.Authors[name] = append(v, article)
		} else {
			s.Authors[name] = articles
		}
	}
}

// addPage ..
func (s *Global) AddPage(page *Page) {
	s.Pages = append(s.Pages, page)
}

// addArticle ..
func (s *Global) AddArticle(article *Article) {
	s.AddTag(article)
	s.AddCategory(article)
	s.AddAuthor(article)
	s.Articles = append(s.Articles, article)
}

// Context ..
func (s *Global) site() map[string]interface{} {
	return map[string]interface{}{
		"url":      s.Conf.Site.URL,
		"title":    s.Conf.Site.Title,
		"subtitle": s.Conf.Site.SubTitle,
	}
}

// Context ..
func (s *Global) articles() []map[string]interface{} {
	articles := make([]map[string]interface{}, 0)
	for _, article := range s.Articles {
		articles = append(articles, article.Context())
	}
	return articles
}

// Context ..
func (s *Global) pages() []map[string]interface{} {
	pages := make([]map[string]interface{}, 0)
	for _, page := range s.Pages {
		pages = append(pages, page.Context())
	}
	return pages
}

// Context ..
func (s *Global) tags() []map[string]interface{} {
	tags := make([]map[string]interface{}, 0)
	for tag, articles := range s.Tags {
		ins := make([]map[string]interface{}, 0)
		for _, article := range articles {
			ins = append(ins, article.Context())
		}
		tags = append(tags, map[string]interface{}{
			"name":     tag,
			"url":      tag,
			"articles": ins,
		})
	}
	return tags
}

// Context ..
func (s *Global) categories() []map[string]interface{} {
	categories := make([]map[string]interface{}, 0)
	for cate, articles := range s.Categories {
		ins := make([]map[string]interface{}, 0)
		for _, article := range articles {
			ins = append(ins, article.Context())
		}
		categories = append(categories, map[string]interface{}{
			"name":     cate,
			"url":      cate,
			"articles": ins,
		})
	}
	return categories
}

// Context ..
func (s *Global) authors() []map[string]interface{} {
	authors := make([]map[string]interface{}, 0)
	for author, articles := range s.Tags {
		ins := make([]map[string]interface{}, 0)
		for _, article := range articles {
			ins = append(ins, article.Context())
		}
		authors = append(authors, map[string]interface{}{
			"name":     author,
			"url":      author,
			"articles": ins,
		})
	}
	return authors
}

// Context ..
func (s *Global) context() map[string]interface{} {
	return map[string]interface{}{
		"site":       s.site(),
		"articles":   s.articles(),
		"pages":      s.pages(),
		"categories": s.categories(),
		"tags":       s.tags(),
		"authors":    s.authors(),
	}
}

// GetPages ..
func (s *Global) GetPages() []string {
	fs := make([]string, 0)
	for _, dir := range s.Conf.PageDir {
		if files, err := ListFiles(filepath.Join(s.Conf.Dir, dir)); err != nil {
			s.Logger.Warn(err.Error())
			continue
		} else {
			fs = append(fs, files...)
		}
	}
	return fs
}

// GetArticles ..
func (s *Global) GetArticles() []string {
	fs := make([]string, 0)
	for _, dir := range s.Conf.ArticleDir {
		if files, err := ListFiles(filepath.Join(s.Conf.Dir, dir)); err != nil {
			s.Logger.Warn(err.Error())
			continue
		} else {
			fs = append(fs, files...)
		}
	}
	return fs
}

// Finish ..
func (s *Global) GetContext() map[string]interface{} {
	if s.Context != nil || len(s.Context) != 0 {
		s.Context = s.context()
	}
	vars := make(map[string]interface{}, len(s.Vars))
	for k, v := range s.Vars {
		vars[k] = v
	}
	for k, v := range s.Context {
		vars[k] = v
	}
	return vars
}
