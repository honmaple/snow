/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: article.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-04 15:12:59 (CST)
 Last Update: Thursday 2019-09-05 02:29:32 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"time"
)

type ArticleType interface {
	Context() map[string]interface{}
}

// Article ..
type Article struct {
	Title   string
	URL     string
	Date    time.Time
	Summary string
	Content string

	SavePath string

	Tags     []string
	Category string
	Author   string
	Authors  []string

	Next     *Article
	Previous *Article
}

type Page struct {
	Title   string
	URL     string
	Date    time.Time
	Summary string
	Content string

	SavePath string
}

// Context ..
func (s *Article) Context() map[string]interface{} {
	return map[string]interface{}{
		"title":    s.Title,
		"date":     s.Date,
		"content":  s.Content,
		"summary":  s.Summary,
		"tags":     s.Tags,
		"category": s.Category,
		"author":   s.Author,
		"authors":  s.Authors,
	}
}

// Context ..
func (s *Page) Context() map[string]interface{} {
	return map[string]interface{}{
		"title":   s.Title,
		"date":    s.Date,
		"content": s.Content,
		"summary": s.Summary,
	}
}
