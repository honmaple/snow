package page

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type (
	Builder struct {
		conf    config.Config
		theme   theme.Theme
		hooks   Hooks
		readers map[string]Reader
	}
	Reader interface {
		Read(io.Reader) (Meta, error)
	}
)

func (b *Builder) parse(file string, typ string, meta Meta) *Page {
	page := meta.Page(file, typ)
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
	return b.hooks.AfterPageParse(page)
}

func (b *Builder) read(dirs []string) (Pages, error) {
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
			reader, ok := b.readers[filepath.Ext(path)]
			if !ok {
				return fmt.Errorf("no reader for %s", path)
			}
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			meta, err := reader.Read(bytes.NewBuffer(buf))
			if err != nil {
				return err
			}
			if meta != nil {
				pages = append(pages, b.parse(info.Name(), dinfo.Name(), meta))
			}
			return nil
		})
		if err != nil {
			b.conf.Log.Errorln(err.Error())
		}
	}
	return pages, nil
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

func (b *Builder) Build(ctx context.Context) error {
	dirs := b.Dirs()
	if len(dirs) == 0 {
		return nil
	}
	now := time.Now()
	pages, err := b.read(dirs)
	if err != nil {
		return err
	}
	pages = b.hooks.BeforePagesWrite(pages)

	labels := pages.GroupBy("type")
	defer func() {
		ls := make([]string, len(labels))
		for i, label := range labels {
			ls[i] = fmt.Sprintf("%d %s", len(label.List), label.Name)
		}
		b.conf.Log.Infoln("Done: Processed", strings.Join(ls, ", "), "in", time.Now().Sub(now))
	}()
	return b.Write(pages, labels)
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	readers := make(map[string]Reader)
	for ext, c := range _readers {
		readers[ext] = c(conf)
	}
	return &Builder{
		conf:    conf,
		theme:   theme,
		hooks:   hooks,
		readers: readers,
	}
}

type creator func(config.Config) Reader

var _readers = make(map[string]creator)

func Register(ext string, c creator) {
	_readers[ext] = c
}
