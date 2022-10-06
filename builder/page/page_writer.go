package page

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flosch/pongo2/v4"
	"github.com/honmaple/snow/utils"
	"io/ioutil"
)

const (
	lookupTemplate   = "theme.templates.%s.lookup"
	outputTemplate   = "theme.templates.%s.save_as"
	paginateTemplate = "theme.templates.%s.paginate"
)

func (b *Builder) Write(pages []*Page) error {
	templates := b.conf.GetStringMap("theme.templates")
	for name := range templates {
		switch name {
		case "page":
			b.writePage(pages)
		case "index":
			b.writeIndex(pages)
		case "tags":
			b.writeTags(pages)
		case "authors":
			b.writeAuthors(pages)
		case "categories":
			b.writeCategories(pages)
		case "archives":
			b.writeArchives(pages)
		default:
			b.writeOther(name, pages)
		}
	}
	return nil
}

func (b *Builder) lookup(key string) (string, string) {
	output := b.conf.GetString(fmt.Sprintf(outputTemplate, key))
	if output == "" {
		return "", ""
	}

	names := b.conf.GetStringSlice(fmt.Sprintf(lookupTemplate, key))
	if len(names) == 0 {
		return "", ""
	}

	layouts := b.conf.GetStringSlice("layout_dirs")
	for _, layout := range layouts {
		for _, name := range names {
			file := filepath.Join(layout, name)
			if utils.FileExists(file) {
				return file, output
			}
		}
	}

	layout := filepath.Join(b.conf.GetString("theme.path"), "templates")
	for _, name := range names {
		file := filepath.Join(layout, name)
		if utils.FileExists(file) {
			return file, output
		}
	}
	return "", ""
}

func (b *Builder) writePage(pages []*Page) {
	layout, output := b.lookup("page")
	if layout == "" || output == "" {
		return
	}
	for _, page := range pages {
		vars := map[string]string{
			"{date:%Y}": page.Date.Format("2006"),
			"{date:%m}": page.Date.Format("01"),
			"{date:%d}": page.Date.Format("02"),
			"{date:%H}": page.Date.Format("15"),
			"{slug}":    page.Title,
		}

		dest := utils.StringReplace(output, vars)
		if page.SaveAs != "" {
			dest = page.SaveAs
		}
		b.writeTemplate(layout, dest, map[string]interface{}{
			"article": page,
		})
	}
	b.writeFeed(pages, b.conf.GetString("feed.all"))
}

func (b *Builder) writeOther(key string, pages []*Page) {
	layout, output := b.lookup(key)
	if layout == "" || output == "" {
		return
	}
	b.writeTemplate(layout, output, map[string]interface{}{
		"pages": pages,
	})
}

func (b *Builder) writeSingle(key string, sections Sections) {
	layout, output := b.lookup(key)
	if layout == "" || output == "" {
		return
	}
	b.writeTemplate(layout, output, map[string]interface{}{
		"sections": sections,
	})
}

func (b *Builder) writeSection(key string, sections Sections) {
	layout, output := b.lookup(key)
	if layout == "" || output == "" {
		return
	}

	paginate := b.conf.GetInt("theme.paginate")
	if k := fmt.Sprintf(paginateTemplate, key); b.conf.IsSet(k) {
		paginate = b.conf.GetInt(k)
	}
	for _, section := range sections {
		for index, por := range Paginator(section.Pages, paginate) {
			vars := map[string]string{
				"{slug}":   section.Name,
				"{number}": strconv.Itoa(index + 1),
			}
			if index == 0 {
				vars["{number}"] = ""
			}
			dest := utils.StringReplace(output, vars)
			b.writeTemplate(layout, dest, map[string]interface{}{
				"paginator": por,
			})
		}
	}
}

func (b *Builder) writeIndex(pages []*Page) {
	indexs := Sections{{Pages: pages}}
	b.writeSection("index", indexs)
}

func (b *Builder) writeCategories(pages []*Page) {
	categories := make(Sections, 0)
	for _, page := range pages {
		for _, name := range page.Categories {
			categories.add(name, page)
		}
	}
	b.writeSingle("categories", categories)
	b.writeSection("category", categories)
	b.writeSectionFeed("feed.categories", categories)
}

func (b *Builder) writeTags(pages []*Page) {
	tags := make(Sections, 0)
	for _, page := range pages {
		for _, name := range page.Tags {
			tags.add(name, page)
		}
	}
	b.writeSingle("tags", tags)
	b.writeSection("tag", tags)
	b.writeSectionFeed("feed.tags", tags)
}

func (b *Builder) writeAuthors(pages []*Page) {
	authors := make(Sections, 0)
	for _, page := range pages {
		for _, name := range page.Authors {
			authors.add(name, page)
		}
	}
	b.writeSingle("authors", authors)
	b.writeSection("author", authors)
	b.writeSectionFeed("feed.authors", authors)
}

func (b *Builder) writeArchives(pages []*Page) {
	year, month := make(Sections, 0), make(Sections, 0)
	for _, page := range pages {
		year.add(page.Date.Format("2006"), page)
		month.add(page.Date.Format("200601"), page)
	}
	b.writeSingle("archives", year)

	// period_archives.html
	b.writeSection("year_archive", year)
	b.writeSection("month_archive", month)
}

func (b *Builder) writeFile(file, content string) error {
	writefile := filepath.Join(b.conf.GetString("output_dir"), file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}
	return ioutil.WriteFile(writefile, []byte(content), 0755)
}

func (b *Builder) writeTemplate(tmpl string, file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	writefile := filepath.Join(b.conf.GetString("output_dir"), file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	tpl := pongo2.Must(pongo2.FromFile(tmpl))
	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := make(map[string]interface{})
	for k, v := range b.context {
		c[k] = v
	}
	for k, v := range context {
		c[k] = v
	}
	return tpl.ExecuteWriter(c, f)
}
