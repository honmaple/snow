/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: feed.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 15:02:22 (CST)
 Last Update: Saturday 2019-09-07 17:11:20 (CST)
		  By:
 Description:
 *********************************************************************************/
package writer

import (
	"github.com/gorilla/feeds"
	"snow/core"
	"time"
)

type FeedWriter struct {
	Writer
}

func NewFeedWriter() *FeedWriter {
	return &FeedWriter{}
}

func (s *FeedWriter) feed(title string, articles []*core.Article, feedconf *core.Feed) *feeds.Feed {
	site := core.G.Conf.Site
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: site.URL},
		Description: site.SubTitle,
		Author:      &feeds.Author{Name: site.Author, Email: ""},
		Created:     time.Now(),
	}
	limit := feedconf.Limit
	if limit > 0 {
		if len(articles) < limit {
			limit = len(articles)
		}
		articles = articles[:limit]
	}
	for _, article := range articles {
		feed.Add(&feeds.Item{
			Title:       article.Title,
			Link:        &feeds.Link{Href: article.URL},
			Description: article.Summary,
			Author:      &feeds.Author{Name: article.Author, Email: ""},
			Created:     article.Date,
		})
	}
	return feed
}

func (s *FeedWriter) Write() {
	feedconf := core.G.Conf.Feed
	s.WriteIndex(feedconf)
	s.WriteTag(feedconf.Tag)
	s.WriteAuthor(feedconf.Author)
	s.WriteCategory(feedconf.Category)
}

func (s *FeedWriter) WriteList(articles []*core.Article, feedconf *core.Feed, context map[string]string) {
	var (
		content string
		err     error
	)
	if context == nil {
		context = make(map[string]string)
	}
	if _, ok := context["{slug}"]; !ok {
		context["{slug}"] = ""
	}
	if _, ok := context["{title}"]; !ok {
		context["{title}"] = core.G.Conf.Site.Title
	}
	if _, ok := context["{lang}"]; !ok {
		context["{lang}"] = core.G.Conf.Lang
	}
	feed := s.feed(core.StringReplace(feedconf.Title, context), articles, feedconf)
	switch feedconf.Format {
	case "rss":
		content, err = feed.ToRss()
	case "atom":
		content, err = feed.ToAtom()
	case "json":
		content, err = feed.ToJSON()
	default:
		core.G.Logger.Errorf("%s is error feed type, should be rss,atom or json", feedconf.Format)
		return
	}
	if err != nil {
		core.G.Logger.Error(err.Error)
		return
	}
	s.Writer.write(core.StringReplace(feedconf.SavePath, context), content)
}

func (s *FeedWriter) WriteIndex(feedconf *core.Feed) {
	s.WriteList(core.G.Articles, feedconf, nil)
}

func (s *FeedWriter) WriteTag(feedconf *core.Feed) {
	for slug, articles := range core.G.Tags {
		vars := map[string]string{
			"{slug}": slug,
		}
		s.WriteList(articles, feedconf, vars)
	}
}

func (s *FeedWriter) WriteCategory(feedconf *core.Feed) {
	for slug, articles := range core.G.Categories {
		vars := map[string]string{
			"{slug}": slug,
		}
		s.WriteList(articles, feedconf, vars)
	}
}

func (s *FeedWriter) WriteAuthor(feedconf *core.Feed) {
	for slug, articles := range core.G.Authors {
		vars := map[string]string{
			"{slug}": slug,
		}
		s.WriteList(articles, feedconf, vars)
	}
}
