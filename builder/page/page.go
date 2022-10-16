package page

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/builder/page/markup"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Builder struct {
	conf   config.Config
	theme  theme.Theme
	markup *markup.Markup
	hooks  Hooks
}

func (b *Builder) getDirs(root string) ([]os.FileInfo, error) {
	pageDirs := b.conf.GetStringSlice("page_dirs")
	if len(pageDirs) > 0 {
		dirs := make([]os.FileInfo, 0)
		for _, dir := range pageDirs {
			info, err := os.Stat(filepath.Join(root, dir))
			if err != nil {
				return nil, err
			}
			dirs = append(dirs, info)
		}
		return dirs, nil
	}
	return ioutil.ReadDir(root)
}

func (b *Builder) parse(typ string, meta map[string]string) *Page {
	now := time.Now()
	page := &Page{Type: typ, Meta: meta, Date: now, Modified: now}
	for k, v := range meta {
		switch strings.ToLower(k) {
		case "type":
			page.Type = v
		case "title":
			page.Title = strings.TrimSpace(v)
		case "date":
			if t, err := utils.ParseTime(v); err == nil {
				page.Date = t
			}
		case "modified":
			if t, err := utils.ParseTime(v); err == nil {
				page.Modified = t
			}
		case "tags":
			page.Tags = utils.SplitTrim(v, ",")
		case "category", "categories":
			page.Categories = []string{v}
		case "authors":
			page.Authors = utils.SplitTrim(v, ",")
		case "output":
			page.SaveAs = v
		case "summary":
			page.Summary = v
		case "content":
			page.Content = v
		}
	}
	if page.SaveAs != "" {
		page.URL = page.SaveAs
	} else if output := b.conf.GetString(fmt.Sprintf(outputTemplate, page.Type)); output == "" {
		page.URL = fmt.Sprintf("pages/%s.html", page.Title)
	} else {
		vars := map[string]string{
			"{date:%Y}": page.Date.Format("2006"),
			"{date:%m}": page.Date.Format("01"),
			"{date:%d}": page.Date.Format("02"),
			"{date:%H}": page.Date.Format("15"),
			"{slug}":    page.Title,
		}
		page.URL = utils.StringReplace(output, vars)
	}
	return page
}

func (b *Builder) Read(watcher *fsnotify.Watcher) (Pages, error) {
	root := b.conf.GetString("content_dir")
	dirs, err := b.getDirs(root)
	if err != nil {
		return nil, err
	}

	pages := make(Pages, 0)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		abs := filepath.Join(root, d.Name())
		if watcher != nil {
			if err := watcher.Add(abs); err != nil {
				return nil, err
			}
		}
		filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			meta, err := b.markup.Read(path)
			if err != nil {
				return err
			}
			pages = append(pages, b.parse(d.Name(), meta))
			return nil
		})
	}
	return pages, nil
}

func (b *Builder) Build(watcher *fsnotify.Watcher) error {
	pages, err := b.Read(watcher)
	if err != nil {
		return err
	}

	pages.OrderBy(b.conf.GetString("page_order"))
	pages = b.hooks.BeforePageList(pages)
	return b.Write(pages)
}

var defaultConfig = map[string]interface{}{
	"page_order":                             "date desc",
	"page_paginate":                          10,
	"page_meta.pages.lookup":                 []string{"page.html", "single.html"},
	"page_meta.pages.output":                 "pages/{slug}.html",
	"page_meta.posts.lookup":                 []string{"post.html", "single.html"},
	"page_meta.posts.output":                 "posts/{date:%Y}/{date:%m}/{slug}.html",
	"page_meta.drafts.lookup":                []string{"draft.html", "single.html"},
	"page_meta.drafts.output":                "drafts/{date:%Y}/{date:%m}/{slug}.html",
	"page_meta.index.list.lookup":            []string{"index.html", "list.html"},
	"page_meta.index.list.filter":            "-pages",
	"page_meta.index.list.output":            "index{number}.html",
	"page_meta.tags.lookup":                  []string{"tags.html", "section.html"},
	"page_meta.tags.output":                  "tags/index.html",
	"page_meta.tags.list.lookup":             []string{"tag.html", "list.html"},
	"page_meta.tags.list.groupby":            "tag",
	"page_meta.tags.list.filter":             "-pages",
	"page_meta.tags.list.output":             "tags/{slug}/index{number}.html",
	"page_meta.categories.lookup":            []string{"categories.html", "section.html"},
	"page_meta.categories.output":            "categories/index.html",
	"page_meta.categories.list.lookup":       []string{"category.html", "list.html"},
	"page_meta.categories.list.filter":       "-pages",
	"page_meta.categories.list.groupby":      "category",
	"page_meta.categories.list.output":       "categories/{slug}/index{number}.html",
	"page_meta.authors.lookup":               []string{"authors.html", "section.html"},
	"page_meta.authors.output":               "authors/index.html",
	"page_meta.authors.list.lookup":          []string{"author.html", "list.html"},
	"page_meta.authors.list.filter":          "-pages",
	"page_meta.authors.list.groupby":         "author",
	"page_meta.authors.list.output":          "authors/{slug}/index{number}.html",
	"page_meta.archives.lookup":              []string{"archives.html"},
	"page_meta.archives.output":              "archives/index.html",
	"page_meta.year_archives.list.lookup":    []string{"period_archives.html"},
	"page_meta.year_archives.list.output":    "archives/{slug}/index.html",
	"page_meta.year_archives.list.filter":    "-pages",
	"page_meta.year_archives.list.groupby":   "date:2006",
	"page_meta.year_archives.list.paginate":  0,
	"page_meta.month_archives.list.lookup":   []string{"period_archives.html"},
	"page_meta.month_archives.list.output":   "archives/{slug}/index.html",
	"page_meta.month_archives.list.filter":   "-pages",
	"page_meta.month_archives.list.groupby":  "date:2006/01",
	"page_meta.month_archives.list.paginate": 0,
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	keys := conf.AllKeys()
	for k, v := range defaultConfig {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
	for _, k := range keys {
		conf.Set(k, conf.Get(k))
	}
	return &Builder{
		conf:   conf,
		markup: markup.New(conf),
		theme:  theme,
		hooks:  hooks,
	}
}
