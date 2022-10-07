package page

import (
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/honmaple/snow/utils"
)

func (b *Builder) writeFeed(pages Pages, output string) {
	limit := b.conf.GetInt("feed.limit")
	if len(pages) == 0 || limit == 0 || output == "" {
		return
	}
	if limit > len(pages) {
		limit = len(pages)
	}

	var (
		err     error
		content string
	)
	site := b.conf.Site
	feed := &feeds.Feed{
		Title:       site.Title,
		Description: site.SubTitle,
		Link:        &feeds.Link{Href: site.URL},
		Author:      &feeds.Author{Name: site.Author, Email: site.Email},
		Created:     time.Now(),
	}
	for _, page := range pages[:limit] {
		feed.Add(&feeds.Item{
			Title:       page.Title,
			Description: page.Summary,
			Link:        &feeds.Link{Href: page.URL},
			Author:      &feeds.Author{Name: strings.Join(page.Authors, ","), Email: ""},
			Created:     page.Date,
		})
	}
	format := b.conf.GetString("feed.format")
	switch format {
	case "rss":
		content, err = feed.ToRss()
	case "atom":
		content, err = feed.ToAtom()
	case "json":
		content, err = feed.ToJSON()
	default:
		return
	}
	if err != nil {
		return
	}

	b.writeFile(output, content)
}

func (b *Builder) writeSectionFeed(key string, sections Section) {
	output := b.conf.GetString(key)
	if output == "" {
		return
	}
	for label, pages := range sections {
		vars := map[string]string{
			"{slug}": label.Name,
		}
		b.writeFeed(pages, utils.StringReplace(output, vars))
	}
}
