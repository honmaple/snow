package page

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/honmaple/snow/utils"
)

type (
	SectionConfig struct {
		Slug         string      `json:"slug"`
		Path         string      `json:"path"`
		Weight       int         `json:"weight"`
		Template     string      `json:"template"`
		Filter       interface{} `json:"filter"`
		Orderby      string      `json:"orderby"`
		Paginate     int         `json:"paginate"`
		PagePath     string      `json:"page_path"`
		PageTemplate string      `json:"page_template"`
		FeedPath     string      `json:"feed_path"`
		FeedTemplate string      `json:"feed_template"`
	}
	Section struct {
		Path      string
		Permalink string
		Pages     Pages
		Title     string
		Content   string
		Parent    *Section
		Children  Sections
		Config    SectionConfig
	}
	Sections []*Section
)

func (sec Section) set(meta Meta) {
	for k, v := range meta {
		switch strings.ToLower(k) {
		case "title":
			sec.Title = v.(string)
		case "content":
			sec.Content = v.(string)
		case "path":
			sec.Config.Path = v.(string)
		case "template":
			sec.Config.Template = v.(string)
		case "page_path":
			sec.Config.PagePath = v.(string)
		case "page_template":
			sec.Config.PageTemplate = v.(string)
		}
	}
}

func (sec Section) Name() string {
	if sec.Parent == nil {
		return sec.Title
	}
	return fmt.Sprintf("%s/%s", sec.Parent.Name(), sec.Title)
}

func (b *Builder) loadSection(parent *Section, path string) (*Section, error) {
	names, err := utils.FileList(path)
	if err != nil {
		return nil, err
	}
	section := &Section{Title: utils.FileBaseName(path), Parent: parent}

	section.Config = b.newSectionConfig(section.Name(), false)
	for _, name := range names {
		file := filepath.Join(path, name)
		info, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(name, "_index.") {
			meta, err := b.readFile(file)
			if err != nil {
				return nil, err
			}
			section.set(meta)
			continue
		}
		if info.IsDir() {
			sec, err := b.loadSection(section, file)
			if err != nil {
				return nil, err
			}
			section.Children = append(section.Children, sec)
			continue
		}
		if page, err := b.loadPage(section, file); err == nil {
			section.Pages = append(section.Pages, page)
			b.pages = append(b.pages, page)
		}
	}
	conf := section.Config
	section.Path = b.conf.GetRelURL(utils.StringReplace(conf.Path, map[string]string{"{number}": "", "{number:one}": "1"}))
	section.Permalink = b.conf.GetURL(section.Path)
	section.Pages = section.Pages.Filter(conf.Filter).OrderBy(conf.Orderby)
	return section, nil
}

func (b *Builder) loadSections() error {
	for _, dir := range b.Dirs() {
		sec, err := b.loadSection(nil, dir)
		if err != nil {
			return err
		}
		b.sections = append(b.sections, sec)
	}
	return nil
}

func (b *Builder) newSectionConfig(name string, custom bool) SectionConfig {
	meta := b.conf.GetStringMap("sections")
	metaList := make([]string, 0)
	for m := range meta {
		if b.conf.GetBool(fmt.Sprintf("sections.%s.custom", m)) == custom {
			metaList = append(metaList, m)
		}
	}
	// 从最短路径开始匹配, 子目录的配置可以继承父目录
	sort.SliceStable(metaList, func(i, j int) bool {
		return len(metaList[i]) < len(metaList[j])
	})

	c := SectionConfig{
		Path:         b.conf.GetString("sections._default.path"),
		Orderby:      b.conf.GetString("sections._default.orderby"),
		Paginate:     b.conf.GetInt("sections._default.paginate"),
		Template:     b.conf.GetString("sections._default.template"),
		PagePath:     b.conf.GetString("sections._default.page_path"),
		PageTemplate: b.conf.GetString("sections._default.page_template"),
		FeedPath:     b.conf.GetString("sections._default.feed_path"),
		FeedTemplate: b.conf.GetString("sections._default.feed_template"),
	}

	for _, m := range metaList {
		if m != name && !strings.HasPrefix(name, m+"/") {
			continue
		}
		if k := fmt.Sprintf("sections.%s.slug", m); b.conf.IsSet(k) {
			c.Slug = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.weight", m); b.conf.IsSet(k) {
			c.Weight = b.conf.GetInt(k)
		}
		if k := fmt.Sprintf("sections.%s.path", m); b.conf.IsSet(k) {
			c.Path = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.filter", m); b.conf.IsSet(k) {
			c.Filter = b.conf.Get(k)
		}
		if k := fmt.Sprintf("sections.%s.orderby", m); b.conf.IsSet(k) {
			c.Orderby = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.template", m); b.conf.IsSet(k) {
			c.Template = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.page_path", m); b.conf.IsSet(k) {
			c.PagePath = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.page_template", m); b.conf.IsSet(k) {
			c.PageTemplate = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.paginate", m); b.conf.IsSet(k) {
			c.Paginate = b.conf.GetInt(k)
		}
		if k := fmt.Sprintf("sections.%s.feed_path", m); b.conf.IsSet(k) {
			c.FeedPath = b.conf.GetString(k)
		}
		if k := fmt.Sprintf("sections.%s.feed_template", m); b.conf.IsSet(k) {
			c.FeedTemplate = b.conf.GetString(k)
		}
		if m == name {
			break
		}
	}

	if c.Slug != "" {
		name = c.Slug
	}
	vars := map[string]string{"{section}": name}
	c.Path = utils.StringReplace(c.Path, vars)
	c.Template = utils.StringReplace(c.Template, vars)
	c.PagePath = utils.StringReplace(c.PagePath, vars)
	c.PageTemplate = utils.StringReplace(c.PageTemplate, vars)
	c.FeedPath = utils.StringReplace(c.FeedPath, vars)
	c.FeedTemplate = utils.StringReplace(c.FeedTemplate, vars)
	return c
}
