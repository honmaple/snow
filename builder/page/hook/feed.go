package hook

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Feed struct {
	hook.BaseHook
	conf *config.Config
}

func (f *Feed) Name() string {
	return "feed"
}

func (f *Feed) write(pages page.Pages, output string) {
	limit := f.conf.GetInt("params.feed.limit")
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
	site := f.conf.Site
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
	format := f.conf.GetString("params.feed.format")
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
	_ = content
	output = filepath.Join(f.conf.GetString("output_dir"), output)
	fmt.Println("Writing Feed to", output)
	// writefile := filepath.Join(b.conf.GetString("output_dir"), file)
	// if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
	//	os.MkdirAll(dir, 0755)
	// }
	// return ioutil.WriteFile(writefile, []byte(content), 0755)
}

func (f *Feed) BeforePageList(pages page.Pages) page.Pages {
	out := f.conf.GetStringMapString("params.feed.output")
	for k, v := range out {
		for label, pages := range pages.GroupBy(k) {
			vars := map[string]string{
				"{slug}": label.Name,
			}
			f.write(pages, utils.StringReplace(v, vars))
		}
	}
	return pages
}

func newFeed(conf *config.Config) hook.Hook {
	defaultConfig := map[string]interface{}{
		"params.feed.limit":  10,
		"params.feed.format": "atom",
		"params.feed.output": map[string]interface{}{
			"":         "feeds.xml",
			"tag":      "tags/{slug}/feeds.xml",
			"category": "categories/{slug}/feeds.xml",
			"author":   "authors/{slug}/feeds.xml",
		},
	}
	for k, v := range defaultConfig {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
	return &Feed{conf: conf}
}
