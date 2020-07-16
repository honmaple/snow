package page

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/page/markup"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type (
	Builder struct {
		conf   *config.Config
		markup *markup.Markup
	}
)

func (b *Builder) parse(meta map[string]string) *Page {
	page := new(Page)
	for k, v := range meta {
		switch strings.ToLower(k) {
		case "title":
			page.Title = v
		case "date":
			if t, err := utils.ParseTime(v); err != nil {
				page.Date = time.Now()
			} else {
				page.Date = t
			}
		case "modified":
			if t, err := utils.ParseTime(v); err != nil {
				page.Modified = time.Now()
			} else {
				page.Modified = t
			}
		case "tags":
			page.Tags = strings.Split(v, ",")
		case "category":
			page.Categories = []string{v}
		case "authors":
			page.Authors = strings.Split(v, ",")
		case "status":
			page.Status = v
		case "save_as":
			page.SaveAs = v
		}
	}
	return page
}

func (b *Builder) Read() ([]*Page, error) {
	dirs := b.conf.GetStringSlice("page_dirs")
	pages := make([]*Page, 0)

	for _, dir := range dirs {
		matched, err := filepath.Glob(dir)
		if err != nil {
			continue
		}
		for _, match := range matched {
			file := filepath.Join(dir, match)
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}
			meta, err := b.markup.Read(file)
			if err != nil {
				continue
			}
			pages = append(pages, b.parse(meta))
		}
	}
	return pages, nil
}

func (b *Builder) Build() error {
	pages, err := b.Read()
	if err != nil {
		return err
	}
	return b.Write(pages)
}

func NewBuilder(conf *config.Config) *Builder {
	conf.SetDefault("feed.limit", 10)
	conf.SetDefault("feed.format", "atom")
	conf.SetDefault("feed.all", "feeds.xml")
	conf.SetDefault("feed.tags", "tags/{slug}/feeds.xml")
	conf.SetDefault("feed.authors", "authors/{slug}/feeds.xml")
	conf.SetDefault("feed.categories", "categories/{slug}/feeds.xml")
	conf.SetDefault("theme.paginate", 10)
	conf.SetDefault("theme.templates.index.save_as", "index{number}.html")
	conf.SetDefault("theme.templates.tag.save_as", "tags/{slug}/index{number}.html")
	conf.SetDefault("theme.templates.tags.save_as", "tags/index.html")
	conf.SetDefault("theme.templates.category.save_as", "categories/{slug}/index{number}.html")
	conf.SetDefault("theme.templates.categories.save_as", "categories/index.html")
	conf.SetDefault("theme.templates.author.save_as", "authors/{slug}/index{number}.html")
	conf.SetDefault("theme.templates.authors.save_as", "authors/index.html")
	conf.SetDefault("theme.templates.archives.save_as", "archives/index.html")
	return &Builder{
		conf: conf,
	}
}
