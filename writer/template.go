/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: template.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-06 22:16:19 (CST)
 Last Update: Saturday 2019-09-07 02:30:12 (CST)
		  By:
 Description:
 *********************************************************************************/
package writer

import (
	"fmt"
	"snow/core"
	"strconv"
	"strings"
)

// TemplateWriter ..
type TemplateWriter struct {
	Writer
}

func NewTemplateWriter() *TemplateWriter {
	return &TemplateWriter{}
}

func (s *TemplateWriter) url(savePath string) string {
	conf := core.G.Conf
	prefix := fmt.Sprintf("%s/", conf.Site.URL)
	if conf.Site.Relative {
		s := strings.Split(savePath, "/")
		if len(s) == 0 {
			prefix = "./"
		} else {
			prefix = strings.Join(make([]string, len(s)), "../")
		}
	}
	return fmt.Sprintf("%s%s", prefix, savePath)
}

// Write ..
func (s *TemplateWriter) Write() {
	template := core.G.Conf.Theme.Template
	s.WritePages(template.Page)
	s.WriteArticles(template.Article)
	s.WriteIndex(template.Index)
	s.WriteTag(template.Tag)
	s.WriteAuthor(template.Author)
	s.WriteCategory(template.Category)

	s.WriteFile(template.Tags.GetSavePath(nil), template.Tags, nil)
	s.WriteFile(template.Authors.GetSavePath(nil), template.Authors, nil)
	s.WriteFile(template.Archives.GetSavePath(nil), template.Archives, nil)
	s.WriteFile(template.Categories.GetSavePath(nil), template.Categories, nil)
}

func (s *TemplateWriter) WriteFile(file string, template *core.Template, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["siteurl"] = s.url(file)
	content, err := template.HTML(context)
	if err != nil {
		core.G.Logger.Error(err.Error())
		return
	}
	s.write(file, content)
}

// WriteList ..
func (s *TemplateWriter) WriteList(articles []*core.Article, template *core.Template, context map[string]string) {
	if context == nil {
		context = make(map[string]string)
	}
	if _, ok := context["{number}"]; !ok {
		context["{number}"] = ""
	}

	items := make([]map[string]interface{}, 0)
	for _, article := range articles {
		items = append(items, article.Context())
	}

	if template.Pagination < 0 || (core.G.Conf.Theme.Pagination <= 0 && template.Pagination <= 0) {
		s.WriteFile(template.GetSavePath(context), template, map[string]interface{}{
			"articles": items,
		})
		return
	}

	number := core.G.Conf.Theme.Pagination
	if template.Pagination > 0 {
		number = template.Pagination
	}
	paginator := Paginator{Items: items, Number: int(number)}
	for index, ins := range paginator.List() {
		if index == 0 {
			context["{number}"] = ""
		} else {
			context["{number}"] = strconv.Itoa(index + 1)
		}
		s.WriteFile(template.GetSavePath(context), template, map[string]interface{}{
			"articles": ins,
		})
	}
}

// WriteIndex ..
func (s *TemplateWriter) WriteIndex(template *core.Template) {
	s.WriteList(core.G.Articles, template, nil)
}

// WriteTag ..
func (s *TemplateWriter) WriteTag(template *core.Template) {
	for slug, articles := range core.G.Tags {
		s.WriteList(articles, template, map[string]string{
			"{slug}": slug,
		})
	}
}

// WriteCategory ..
func (s *TemplateWriter) WriteCategory(template *core.Template) {
	for slug, articles := range core.G.Categories {
		s.WriteList(articles, template, map[string]string{
			"{slug}": slug,
		})
	}
}

// WriteAuthor ..
func (s *TemplateWriter) WriteAuthor(template *core.Template) {
	for slug, articles := range core.G.Authors {
		s.WriteList(articles, template, map[string]string{
			"{slug}": slug,
		})
	}
}

// WriteArticle ..
func (s *TemplateWriter) WriteArticle(article *core.Article, template *core.Template) {
	context := map[string]string{
		"{slug}":      article.Title,
		"{date:2006}": article.Date.Format("2006"),
		"{date:01}":   article.Date.Format("01"),
		"{date:02}":   article.Date.Format("02"),
		"{date:15}":   article.Date.Format("15"),
	}
	savePath := template.GetSavePath(context)
	if article.SavePath != "" {
		savePath = article.SavePath
	}
	s.WriteFile(savePath, template, article.Context())
}

// WritePage ..
func (s *TemplateWriter) WritePage(page *core.Page, template *core.Template) {
	context := map[string]string{
		"{slug}":      page.Title,
		"{date:2006}": page.Date.Format("2006"),
		"{date:01}":   page.Date.Format("01"),
		"{date:02}":   page.Date.Format("02"),
		"{date:15}":   page.Date.Format("15"),
	}
	savePath := template.GetSavePath(context)
	if page.SavePath != "" {
		savePath = page.SavePath
	}
	s.WriteFile(savePath, template, page.Context())
}

// WriteArticle ..
func (s *TemplateWriter) WriteArticles(template *core.Template) {
	for _, article := range core.G.Articles {
		s.WriteArticle(article, template)
	}
}

// WritePage ..
func (s *TemplateWriter) WritePages(template *core.Template) {
	for _, page := range core.G.Pages {
		s.WritePage(page, template)
	}
}
