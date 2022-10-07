package page

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/honmaple/snow/builder/page/markup"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Builder struct {
	conf    *config.Config
	types   map[string]bool
	markup  *markup.Markup
	context map[string]interface{}
}

func (b *Builder) parse(typ string, meta map[string]string) *Page {
	page := &Page{Type: typ}
	for k, v := range meta {
		switch strings.ToLower(k) {
		case "type":
			page.Type = v
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
		case "category", "categories":
			page.Categories = []string{v}
		case "authors":
			page.Authors = strings.Split(v, ",")
		case "status":
			page.Status = v
		case "output":
			page.SaveAs = v
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

func (b *Builder) Read() (Pages, error) {
	pages := make(Pages, 0)

	dir := b.conf.GetString("content_dir")
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		// 忽略部分类型的目录，比如extra
		if b.conf.GetBool(fmt.Sprintf(ignoreTemplate, d.Name())) {
			continue
		}
		root := filepath.Join(dir, d.Name())
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || err != nil {
				return err
			}
			meta, err := b.markup.Read(path)
			if err != nil {
				return err
			}
			pages = append(pages, b.parse(d.Name(), meta))
			b.types[d.Name()] = true
			return nil
		})
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
	conf.SetDefault("theme.path", "themes/simple")
	conf.SetDefault("theme.paginate", 10)

	conf.SetDefault("theme.templates.extra.ignore", true)

	conf.SetDefault("theme.templates.pages.lookup", []string{"page.html", "single.html"})
	conf.SetDefault("theme.templates.pages.output", "pages/{slug}.html")

	conf.SetDefault("theme.templates.posts.lookup", []string{"post.html", "single.html"})
	conf.SetDefault("theme.templates.posts.output", "articles/{date:%Y}/{date:%m}/{slug}.html")

	conf.SetDefault("theme.templates.drafts.lookup", []string{"draft.html", "single.html"})
	conf.SetDefault("theme.templates.drafts.output", "drafts/{date:%Y}/{date:%m}/{slug}.html")

	conf.SetDefault("theme.templates.index.list.lookup", []string{"index.html"})
	conf.SetDefault("theme.templates.index.list.output", "index{number}.html")

	conf.SetDefault("theme.templates.tags.lookup", []string{"tags.html"})
	conf.SetDefault("theme.templates.tags.output", "tags/index.html")
	conf.SetDefault("theme.templates.tags.list.lookup", []string{"tag.html", "index.html"})
	conf.SetDefault("theme.templates.tags.list.output", "tags/{slug}/index{number}.html")

	conf.SetDefault("theme.templates.categories.lookup", []string{"categories.html"})
	conf.SetDefault("theme.templates.categories.output", "categories/index.html")
	conf.SetDefault("theme.templates.categories.list.lookup", []string{"category.html", "index.html"})
	conf.SetDefault("theme.templates.categories.list.output", "categories/{slug}/index{number}.html")

	conf.SetDefault("theme.templates.authors.lookup", []string{"authors.html"})
	conf.SetDefault("theme.templates.authors.output", "authors/index.html")
	conf.SetDefault("theme.templates.authors.list.lookup", []string{"author.html", "index.html"})
	conf.SetDefault("theme.templates.authors.list.output", "authors/{slug}/index{number}.html")

	conf.SetDefault("theme.templates.archives.lookup", []string{"archives.html"})
	conf.SetDefault("theme.templates.archives.output", "archives/index.html")

	conf.SetDefault("theme.templates.year_archives.list.lookup", []string{"period_archives.html"})
	conf.SetDefault("theme.templates.year_archives.list.output", "archives/{slug}/index.html")
	conf.SetDefault("theme.templates.year_archives.list.groupby", "date:2006")
	conf.SetDefault("theme.templates.year_archives.list.paginate", 0)

	conf.SetDefault("theme.templates.month_archives.list.lookup", []string{"period_archives.html"})
	conf.SetDefault("theme.templates.month_archives.list.output", "archives/{slug}/index.html")
	conf.SetDefault("theme.templates.month_archives.list.groupby", "date:2006/01")
	conf.SetDefault("theme.templates.month_archives.list.paginate", 0)
	return &Builder{
		conf:   conf,
		types:  make(map[string]bool),
		markup: markup.New(conf),
		context: map[string]interface{}{
			"site":   conf.GetStringMap("site"),
			"params": conf.GetStringMap("params"),
		},
	}
}
