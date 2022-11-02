package page

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func (b *Builder) Dirs() []string {
	root := b.conf.GetString("content_dir")

	pageDirs := b.conf.GetStringSlice("page_dirs")
	if len(pageDirs) > 0 {
		dirs := make([]string, len(pageDirs))
		for i, dir := range pageDirs {
			dirs[i] = filepath.Join(root, dir)
		}
		return dirs
	}
	subDirs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil
	}
	dirs := make([]string, len(subDirs))
	for i, dir := range subDirs {
		dirs[i] = filepath.Join(root, dir.Name())
	}
	return dirs
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
			"{filename}": utils.FileBaseName(file),
			"{slug}":     page.Title,
		}
		if page.Slug != "" {
			vars["{slug}"] = page.Slug
		}
		page.URL = utils.StringReplace(output, vars)
	}
	return b.hooks.AfterPageParse(meta, page)
}

func (b *Builder) Read(dirs []string) (Pages, error) {
	pages := make(Pages, 0)
	for _, d := range dirs {
		dinfo, err := os.Stat(d)
		if err != nil {
			return nil, err
		}
		err = filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			meta, err := b.markup.Read(path)
			if err != nil {
				return err
			}
			page := b.parse(info.Name(), dinfo.Name(), meta)
			pages = append(pages, page)
			return nil
		})
		if err != nil {
			b.conf.Log.Errorln(err.Error())
		}
	}
	return pages, nil
}

func (b *Builder) Build(ctx context.Context) error {
	dirs := b.Dirs()
	if len(dirs) == 0 {
		return nil
	}
	now := time.Now()
	pages, err := b.Read(dirs)
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
