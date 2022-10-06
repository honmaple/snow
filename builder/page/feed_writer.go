package page

import (
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/honmaple/snow/utils"
)

func (b *Builder) writeFeed(pages []*Page, output string) {
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

func (b *Builder) writeSectionFeed(key string, sections Sections) {
	output := b.conf.GetString(key)
	if output == "" {
		return
	}
	for _, section := range sections {
		vars := map[string]string{
			"{slug}": section.Name,
		}
		b.writeFeed(section.Pages, utils.StringReplace(output, vars))
	}
}
