package page

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flosch/pongo2/v4"
	"github.com/honmaple/snow/utils"
)

func (b *Builder) Write(pages []*Page) error {
	b.writeIndex(pages)
	b.writePage(pages)
	b.writeArchives(pages)
	b.writeTags(pages)
	b.writeAuthors(pages)
	b.writeCategories(pages)
	return nil
}

func (b *Builder) lookup(names ...string) string {
	// layouts := b.conf1.LayoutDirs
	layouts := b.conf.GetStringSlice("theme.layouts")
	for _, name := range names {
		for _, layout := range layouts {
			file := filepath.Join(layout, name)
			if utils.FileExists(file) {
				return file
			}
		}
	}
	return ""
}

func (b *Builder) writeIndex(pages []*Page) {
	layout := b.lookup("index.html")
	if layout == "" {
		return
	}

	// layoutDest := "index{number}.html"
	layoutDest := b.conf.GetString("theme.templates.index.save_as")

	paginate := b.conf.GetInt("theme.paginate")
	if k := "theme.templates.index.paginate"; b.conf.IsSet(k) {
		paginate = b.conf.GetInt(k)
	}

	for index, por := range Paginator(pages, paginate) {
		vars := map[string]string{
			"{number}": strconv.Itoa(index + 1),
		}
		if index == 0 {
			vars["{number}"] = ""
		}
		dest := utils.StringReplace(layoutDest, vars)
		b.writeTemplate(layout, dest, map[string]interface{}{
			"paginator": por,
		})
	}
}

func (b *Builder) writePage(pages []*Page) {
	layout := b.lookup("article.html", "page.html")
	if layout == "" {
		return
	}
	// layoutDest := "articles/{date:%Y}/{date:%m}/{slug}.html"
	layoutDest := b.conf.GetString("theme.templates.page.save_as")

	for _, page := range pages {
		vars := map[string]string{
			"{date:%Y}": page.Date.Format("2006"),
			"{date:%m}": page.Date.Format("01"),
			"{date:%d}": page.Date.Format("02"),
			"{slug}":    page.Title,
		}

		dest := utils.StringReplace(layoutDest, vars)
		if page.SaveAs != "" {
			dest = page.SaveAs
		}
		b.writeTemplate(layout, dest, map[string]interface{}{
			"article": page,
		})
	}
	b.writeFeed(pages, b.conf.GetString("feed.all"))
}

func (b *Builder) writeSingle(layout, layoutDest string, section Section) {
	if layout == "" || layoutDest == "" {
		return
	}
	b.writeTemplate(layout, layoutDest, nil)
}

func (b *Builder) writeSection(layout, key string, section Section) {
	paginate := b.conf.GetInt("theme.paginate")
	if k := fmt.Sprintf("%s.paginate", key); b.conf.IsSet(k) {
		paginate = b.conf.GetInt(k)
	}
	layoutDest := b.conf.GetString(fmt.Sprintf("%s.save_as", key))
	for name, label := range section {
		for index, por := range Paginator(label.Pages, paginate) {
			vars := map[string]string{
				"{slug}":   name,
				"{number}": strconv.Itoa(index + 1),
			}
			if index == 0 {
				vars["{number}"] = ""
			}
			dest := utils.StringReplace(layoutDest, vars)
			b.writeTemplate(layout, dest, map[string]interface{}{
				"paginator": por,
			})
		}
	}
}

func (b *Builder) writeCategories(pages []*Page) {
	categories := make(Section)
	for _, page := range pages {
		for _, name := range page.Categories {
			categories.add(name, page)
		}
	}
	b.writeSingle(
		b.lookup("categories.html"),
		b.conf.GetString("theme.templates.categories.save_as"), categories)
	b.writeSection(
		b.lookup("category.html", "index.html"),
		"theme.templates.category",
		categories)
	b.writeSectionFeed("feed.categories", categories)
}

func (b *Builder) writeTags(pages []*Page) {
	tags := make(Section)
	for _, page := range pages {
		for _, name := range page.Tags {
			tags.add(name, page)
		}
	}
	b.writeSingle(
		b.lookup("tags.html"),
		b.conf.GetString("theme.templates.tags.save_as"), tags)
	b.writeSection(
		b.lookup("tag.html", "index.html"),
		"theme.templates.tag", tags)
	b.writeSectionFeed("feed.tags", tags)
}

func (b *Builder) writeAuthors(pages []*Page) {
	authors := make(Section)
	for _, page := range pages {
		for _, name := range page.Authors {
			authors.add(name, page)
		}
	}
	b.writeSingle(
		b.lookup("authors.html"),
		b.conf.GetString("theme.templates.authors.save_as"), authors)
	b.writeSection(
		b.lookup("author.html", "index.html"),
		"theme.templates.author", authors)
	b.writeSectionFeed("feed.authors", authors)
}

func (b *Builder) writeArchives(pages []*Page) {
	year, month := make(Section), make(Section)
	for _, page := range pages {
		year.add(page.Date.Format("2006"), page)
		month.add(page.Date.Format("200601"), page)
	}

	// layoutDest := "archives/{date:%Y}/{date:%m}/index{number}.html"
	b.writeSection(
		b.lookup("year_archive.html", "archive.html", "index.html"),
		"theme.templates.year_archive", year)
	b.writeSection(
		b.lookup("month_archive.html", "archive.html", "index.html"),
		"theme.templates.month_archive", year)
}

func (b *Builder) writeTemplate(file string, dest string, context map[string]interface{}) error {
	writefile := filepath.Join(b.conf.GetString("output"), dest)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	tpl := pongo2.Must(pongo2.FromFile(file))
	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := make(map[string]interface{})
	for k, v := range context {
		c[k] = v
	}
	// c["tags"] = ctx.Tags()
	// c["archives"] = ctx.Archives()
	// c["authors"] = ctx.Authors()
	// c["categories"] = ctx.Categories()
	return tpl.ExecuteWriter(c, f)
}
