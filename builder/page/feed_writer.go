package page

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/honmaple/snow/utils"
)

func (b *Builder) writeFile(file, content string) error {
	writefile := filepath.Join(b.conf.GetString("output"), file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}
	return ioutil.WriteFile(writefile, []byte(content), 0755)
}

func (b *Builder) writeFeed(pages []*Page, dest string) {
	limit := b.conf.GetInt("feed.limit")
	if len(pages) == 0 || limit == 0 || dest == "" {
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

	b.writeFile(dest, content)
}

func (b *Builder) writeSectionFeed(key string, section Section) {
	dest := b.conf.GetString(key)
	if dest == "" {
		return
	}
	for slug, label := range section {
		vars := map[string]string{
			"{slug}": slug,
		}
		b.writeFeed(label.Pages, utils.StringReplace(dest, vars))
	}
}
