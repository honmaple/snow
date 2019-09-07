/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: feed.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-06 23:21:45 (CST)
 Last Update: Friday 2019-09-06 23:30:11 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"fmt"
	"github.com/gorilla/feeds"
)

// Feed ..
type Feed struct {
	Limit    int    `toml:"limit"`
	Title    string `toml:"title"`
	Format   string `toml:"format"`
	SavePath string `toml:"save_path"`
	Tag      *Feed  `toml:"tag"`
	Author   *Feed  `toml:"author"`
	Category *Feed  `toml:"category"`
}

func (s *Feed) GetTitle(slug string) string {
	return StringReplace(s.Title, map[string]string{
		"{slug}":  slug,
		"{title}": G.Conf.Site.Title,
	})
}

func (s *Feed) GetSavePath(slug string) string {
	return StringReplace(s.SavePath, map[string]string{
		"{slug}": slug,
	})
}

// Content ..
func (s *Feed) Content(feed *feeds.Feed) (string, error) {
	feed.Title = s.GetTitle(feed.Id)
	feed.Id = ""
	switch s.Format {
	case "rss":
		return feed.ToRss()
	case "atom":
		return feed.ToAtom()
	case "json":
		return feed.ToJSON()
	default:
		return "", fmt.Errorf("%s is error feed type, should be rss,atom or json", s.Format)
	}
}
