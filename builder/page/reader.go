package page

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	Builder struct {
		conf    config.Config
		theme   theme.Theme
		hooks   Hooks
		readers map[string]Reader

		ctx         *Context
		ignoreFiles []*regexp.Regexp
	}
	Reader interface {
		Read(io.Reader) (Meta, error)
	}
)

func (b *Builder) langRange(f func(lang string, isdefault bool)) {
	for lang := range b.conf.Languages {
		f(lang, false)
	}
	f(b.conf.Site.Language, true)
}

func (b *Builder) findLang(path string, filemeta Meta) string {
	if filemeta != nil {
		if v := filemeta.GetString("lang"); v != "" && b.conf.Languages[v] {
			return v
		}
	}
	ext := filepath.Ext(path)
	if ext != "" {
		lang := filepath.Ext(path[:len(path)-len(ext)])
		if lang != "" && b.conf.Languages[lang[1:]] {
			return lang[1:]
		}
	}
	return b.conf.Site.Language
}

func (b *Builder) ignoreFile(file string) bool {
	for _, re := range b.ignoreFiles {
		if re.MatchString(file) {
			return true
		}
	}
	return false
}

func (b *Builder) readFile(file string) (Meta, error) {
	reader, ok := b.readers[filepath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no reader for %s", file)
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta, err := reader.Read(f)
	if err != nil {
		return nil, fmt.Errorf("Read file %s: %s", file, err.Error())
	}
	if len(meta) == 0 {
		return nil, fmt.Errorf("Read file %s: no meta", file)
	}
	return meta, nil
}

func (b *Builder) findFiles(path string, pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(path, pattern))
	if len(matches) == 0 {
		return nil
	}

	files := make([]string, 0)
	for _, m := range matches {
		if _, ok := b.readers[filepath.Ext(m)]; !ok {
			continue
		}
		files = append(files, m)
	}
	return files
}

func (b *Builder) Build(ctx context.Context) error {
	now := time.Now()
	defer func() {
		ps := make([]string, 0)
		ls := make([]string, 0)
		ts := make([]string, 0)
		b.langRange(func(lang string, isdefault bool) {
			showLang := " " + lang
			if isdefault {
				showLang = ""
			}
			if count := len(b.ctx.Pages(lang)); count > 0 {
				ps = append(ps, fmt.Sprintf("%d%s normal pages", count, showLang))
			}
			if count := len(b.ctx.HiddenPages(lang)); count > 0 {
				ps = append(ps, fmt.Sprintf("%d%s hidden pages", count, showLang))
			}
			if count := len(b.ctx.SectionPages(lang)); count > 0 {
				ps = append(ps, fmt.Sprintf("%d%s section pages", count, showLang))
			}

			for _, section := range b.ctx.Sections(lang) {
				if section.isRoot() {
					continue
				}
				if count := len(section.Pages) + len(section.HiddenPages) + len(section.SectionPages); count > 0 {
					ls = append(ls, fmt.Sprintf("%d%s %s", count, showLang, section.RealName()))
				}
			}

			for _, taxonomy := range b.ctx.Taxonomies(lang) {
				if count := len(taxonomy.Terms); count > 0 {
					ts = append(ts, fmt.Sprintf("%d%s %s", count, showLang, taxonomy.Name))
				}
			}
		})
		if len(ps) > 0 {
			b.conf.Log.Infoln("Done: Page Processed", strings.Join(ps, ", "), "in", time.Now().Sub(now))
		}
		if len(ls) > 0 {
			b.conf.Log.Infoln("Done: Section Processed", strings.Join(ls, ", "), "in", time.Now().Sub(now))
		}
		if len(ts) > 0 {
			b.conf.Log.Infoln("Done: Taxonomy Processed", strings.Join(ts, ", "), "in", time.Now().Sub(now))
		}
	}()

	var wg sync.WaitGroup

	tasks := newTaskPool(&wg, 100, func(i interface{}) {
		defer wg.Done()
		b.insertPage(i.(string))
	})
	defer tasks.Release()

	rootDir := b.conf.GetString("content_dir")
	if rootDir == "" {
		rootDir = "content"
	}
	b.conf.Watch(rootDir)

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if b.ignoreFile(path) {
			return nil
		}
		if info.IsDir() {
			// is not root
			if path != rootDir {
				files := b.findFiles(path, "index.*")
				for _, file := range files {
					tasks.Invoke(file)
				}
				if len(files) > 0 {
					return fs.SkipDir
				}
			}
			b.insertSection(path)
			return nil
		}
		if _, ok := b.readers[filepath.Ext(path)]; !ok {
			b.insertAsset(path)
			return nil
		}
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_index.") || strings.HasPrefix(basename, ".") {
			return nil
		}
		tasks.Invoke(path)
		return nil
	}
	if err := filepath.WalkDir(rootDir, walkDir); err != nil {
		return err
	}
	tasks.Wait()
	return b.Write()
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
		ctx:     newContext(conf),
	}
}

type creator func(config.Config) Reader

var _readers = make(map[string]creator)

func Register(ext string, c creator) {
	_readers[ext] = c
}
