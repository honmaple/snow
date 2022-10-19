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

func (b *Builder) parse(file string, typ string, meta map[string]string) *Page {
	now := time.Now()
	page := &Page{Type: typ, Meta: meta, Date: now, Modified: now}
	for k, v := range meta {
		if v == "" {
			continue
		}
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
		case "category":
			page.Categories = []string{v}
		case "categories":
			page.Categories = utils.SplitTrim(v, ",")
		case "authors":
			page.Authors = utils.SplitTrim(v, ",")
		case "url":
			page.URL = v
		case "slug":
			page.Slug = v
		case "summary":
			page.Summary = v
		case "content":
			page.Content = v
		}
	}
	if page.URL == "" {
		output := b.conf.GetString(fmt.Sprintf(outputTemplate, page.Type))
		if output == "" {
			output = fmt.Sprintf("%s/{slug}.html", page.Type)
		}
		vars := map[string]string{
			"{date:%Y}":  page.Date.Format("2006"),
			"{date:%m}":  page.Date.Format("01"),
			"{date:%d}":  page.Date.Format("02"),
			"{date:%H}":  page.Date.Format("15"),
			"{filename}": filepath.Base(file),
			"{slug}":     page.Title,
		}
		if page.Slug != "" {
			vars["{slug}"] = page.Slug
		}
		page.URL = utils.StringReplace(output, vars)
	}
	return b.hooks.AfterPageParse(meta, page)
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
			page := b.parse(info.Name(), d.Name(), meta)
			pages = append(pages, page)
			return nil
		})
	}
	return pages, nil
}

func (b *Builder) Build(watcher *fsnotify.Watcher) error {
	now := time.Now()
	pages, err := b.Read(watcher)
	if err != nil {
		return err
	}
	if filter := b.conf.Get("page_filter"); filter != nil {
		pages = pages.Filter(filter)
	}
	if order := b.conf.GetString("page_orderby"); order != "" {
		pages.OrderBy(order)
	}
	pages = b.hooks.BeforePagesWrite(pages)

	defer func() {
		b.conf.Log.Infoln("Done: Processed", len(pages), "pages", "in", time.Now().Sub(now))
	}()
	return b.Write(pages)
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{
		conf:   conf,
		markup: markup.New(conf),
		theme:  theme,
		hooks:  hooks,
	}
}
