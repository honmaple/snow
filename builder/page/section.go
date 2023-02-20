package page

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/honmaple/snow/utils"
)

type sectionChan struct {
	errs     chan error
	pages    chan *Page
	sections chan *Section
	assets   chan string
}

func (ch *sectionChan) sendErr(err error) {
	ch.errs <- err
}

func (ch *sectionChan) sendAsset(files ...string) {
	for _, file := range files {
		ch.assets <- file
	}
}

func (ch *sectionChan) sendPage(pages ...*Page) {
	for _, page := range pages {
		ch.pages <- page
	}
}

func (ch *sectionChan) sendSection(sections ...*Section) {
	for _, section := range sections {
		ch.sections <- section
	}
}

type (
	Section struct {
		// slug:
		// weight:
		// aliases:
		// transparent:
		// filter:
		// orderby:
		// paginate:
		// paginate_path: {name}{number}{extension}
		// path:
		// template:
		// page_path:
		// page_template:
		// feed_path:
		// feed_template:
		File      string
		Meta      Meta
		Path      string
		Permalink string
		Slug      string
		Title     string
		Content   string
		Pages     Pages
		Assets    []string
		Parent    *Section
		Children  Sections
	}
	Sections []*Section
)

func (sec *Section) vars() map[string]string {
	return map[string]string{"{section}": sec.Name(), "{section:slug}": sec.Slug}
}

func (sec *Section) allPages() Pages {
	pages := make(Pages, 0)
	if !sec.Meta.GetBool("transparent") {
		pages = append(pages, sec.Pages...)
	}
	for _, child := range sec.Children {
		if child.isHidden() {
			continue
		}
		pages = append(pages, child.allPages()...)
	}
	return pages
}

func (sec *Section) allSections() Sections {
	sections := make(Sections, 0)
	for _, child := range sec.Children {
		sections = append(sections, child)
		sections = append(sections, child.allSections()...)
	}
	return sections
}

func (sec *Section) isRoot() bool {
	return sec.Parent == nil
}

func (sec *Section) isHidden() bool {
	return sec.Meta.GetBool("hidden")
}

func (sec *Section) Paginator() []*paginator {
	return sec.Pages.Paginator(
		sec.Meta.GetInt("paginate"),
		sec.Path,
		sec.Meta.GetString("paginate_path"),
	)
}

func (sec *Section) Root() *Section {
	if sec.Parent == nil {
		return sec
	}
	return sec.Parent.Root()
}

func (sec *Section) Name() string {
	if sec.Parent == nil || sec.Parent.Parent == nil {
		return sec.Title
	}
	return fmt.Sprintf("%s/%s", sec.Parent.Name(), sec.Title)
}

func (sec *Section) FirstName() string {
	if sec.Parent == nil || sec.Parent.Title == "" {
		return sec.Title
	}
	return sec.Parent.FirstName()
}

func (b *Builder) loadSectionDone(section *Section) {
	pages := section.Pages
	if section.isHidden() {
		pages = section.Parent.allPages()
	}
	section.Pages = pages.Filter(section.Meta.Get("filter")).OrderBy(section.Meta.GetString("orderby"))

	sort.SliceStable(section.Children, func(i, j int) bool {
		ti := section.Children[i]
		tj := section.Children[j]
		if wi, wj := ti.Meta.GetInt("weight"), tj.Meta.GetInt("weight"); wi == wj {
			return ti.Title > tj.Title
		} else {
			return wi > wj
		}
	})
}

func (b *Builder) loadSection(parent *Section, path string) (*Section, error) {
	names, err := utils.FileList(path)
	if err != nil {
		return nil, err
	}

	var (
		ch = &sectionChan{
			errs:     make(chan error),
			pages:    make(chan *Page),
			assets:   make(chan string),
			sections: make(chan *Section),
		}
		wg      = sync.WaitGroup{}
		section = b.newSection(parent, path)
	)

	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			file := filepath.Join(path, name)
			info, err := os.Stat(file)
			if err != nil {
				ch.sendErr(err)
				return
			}
			if info.IsDir() {
				// 如果包括index.md, index.org等文件，整个目录为一个page，而不是section
				matches, _ := filepath.Glob(filepath.Join(file, "index.*"))
				matched := -1
				for i, m := range matches {
					if _, ok := b.readers[filepath.Ext(m)]; ok {
						matched = i
						break
					}
				}
				if matched == -1 {
					sec, err := b.loadSection(section, file)
					if err != nil {
						ch.sendErr(err)
						return
					}
					if sec.Meta.GetBool("transparent") {
						ch.sendPage(sec.Pages...)
						// ch.sendAsset(sec.Assets...)
						// ch.sendSection(sec.Children...)
					}
					ch.sendSection(sec)
					return
				}
				file = matches[matched]
			} else if _, ok := b.readers[filepath.Ext(file)]; !ok {
				ch.sendAsset(file)
				return
			}

			if strings.HasPrefix(name, "_index.") {
				return
			}

			filemeta, err := b.readFile(file)
			if err != nil {
				ch.sendErr(err)
				return
			}
			// 自定义页面, 该页面的page列表与父级section一致
			if template, ok := filemeta["template"]; ok && template != "" {
				meta := make(Meta)
				meta.load(filemeta)
				meta["hidden"] = true
				child := &Section{
					Meta:    meta,
					Title:   meta.GetString("title"),
					Content: meta.GetString("content"),
					Parent:  section,
				}
				child.Path = b.conf.GetRelURL(meta.GetString("path"))
				child.Permalink = b.conf.GetURL(child.Path)
				ch.sendSection(child)
				return
			}
			ch.sendPage(b.newPage(section, file, filemeta))
		}(name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		wg.Wait()
		cancel()
	}()
LOOP:
	for {
		select {
		case page := <-ch.pages:
			if !page.Meta.GetBool("hidden") {
				section.Pages = append(section.Pages, page)
			}
		case file := <-ch.assets:
			section.Assets = append(section.Assets, file)
		case child := <-ch.sections:
			section.Children = append(section.Children, child)
			defer b.loadSectionDone(child)
		case err := <-ch.errs:
			return nil, err
		case <-ctx.Done():
			break LOOP
		}
	}
	b.loadSectionDone(section)
	return section, nil
}

func (b *Builder) loadSections() error {
	root, err := b.loadSection(nil, b.Dirs()[0])
	if err != nil {
		return err
	}
	b.pages = root.allPages()
	b.sections = root.allSections()
	return nil
}

func (b *Builder) newSection(parent *Section, path string) *Section {
	section := &Section{
		Parent: parent,
	}
	// 根目录
	if parent == nil {
		section.Meta = make(Meta)
		section.Meta.load(b.conf.GetStringMap("sections._default"))
	} else {
		section.Meta = parent.Meta.copy()
		section.Title = utils.FileBaseName(path)
	}
	// _index.md包括配置信息
	matches, _ := filepath.Glob(filepath.Join(path, "_index.*"))
	matched := -1
	for i, m := range matches {
		if _, ok := b.readers[filepath.Ext(m)]; ok {
			matched = i
			break
		}
	}
	if matched > -1 {
		filemeta, _ := b.readFile(matches[matched])
		section.Meta.load(filemeta)
	}

	name := section.Name()
	if name != "" {
		section.Meta.load(b.conf.GetStringMap("sections." + name))
	}

	for k, v := range section.Meta {
		switch strings.ToLower(k) {
		case "title":
			section.Title = v.(string)
		case "content":
			section.Content = v.(string)
		}
	}

	slug := section.Meta.GetString("slug")
	if slug == "" {
		names := strings.Split(name, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = b.conf.GetSlug(name)
		}
		slug = strings.Join(slugs, "/")
	}
	section.Slug = slug
	section.Path = b.conf.GetRelURL(utils.StringReplace(section.Meta.GetString("path"), section.vars()))
	section.Permalink = b.conf.GetURL(section.Path)
	return section
}
