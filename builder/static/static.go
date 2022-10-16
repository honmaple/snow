package static

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	Static struct {
		URL  string
		File string
	}
	Statics []*Static
	Builder struct {
		conf  config.Config
		theme theme.Theme
		hooks Hooks
	}
)

func (b *Builder) GetDirs() []string {
	return b.conf.GetStringSlice("static_dirs")
}

func (b *Builder) parser() func(string) *Static {
	exts := make(map[string]bool)
	extsIsSet := b.conf.IsSet("static_exts")
	if extsIsSet {
		for _, ext := range b.conf.GetStringSlice("static_exts") {
			exts[ext] = true
		}
	}

	meta := b.conf.GetStringMapString("static_meta")
	if _, ok := meta["theme/static"]; !ok {
		meta["theme/static"] = "static/"
	}
	metaList := make([]string, 0)
	for m := range meta {
		metaList = append(metaList, m)
	}
	sort.SliceStable(metaList, func(i, j int) bool {
		return len(metaList[i]) > len(metaList[j])
	})

	return func(file string) *Static {
		if extsIsSet && !exts[filepath.Ext(file)] {
			return nil
		}
		// viper 无法处理带.的key, 无法直接使用meta[file]
		if k := fmt.Sprintf("static_meta.%s", file); b.conf.IsSet(k) {
			output := b.conf.GetString(k)
			if output == "" {
				return nil
			}
			if strings.HasSuffix(output, "/") {
				output = filepath.Join(output, filepath.Base(file))
			}
			return &Static{URL: output, File: file}
		}
		for _, m := range metaList {
			if !strings.HasPrefix(file, m) {
				continue
			}
			output := meta[m]
			if output == "" {
				return nil
			}
			if strings.HasSuffix(output, "/") {
				output = filepath.Join(output, strings.TrimPrefix(file, m))
			}
			return &Static{URL: output, File: file}
		}
		return &Static{URL: file, File: file}
	}

}

func (b *Builder) Read(watcher *fsnotify.Watcher) ([]*Static, error) {
	parse := b.parser()

	files := make([]*Static, 0)
	// 默认添加主题的静态文件
	root := b.theme.Root()
	fs.WalkDir(root, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if file := parse(filepath.Join("theme", path)); file != nil {
			files = append(files, file)
		}
		return nil
	})

	dirs := b.conf.GetStringSlice("static_dirs")
	for _, dir := range dirs {
		if watcher != nil {
			if err := watcher.Add(dir); err != nil {
				return nil, err
			}
		}
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if file := parse(path); file != nil {
				files = append(files, file)
			}
			return nil
		})
	}
	return files, nil
}

func (b *Builder) Build(watcher *fsnotify.Watcher) error {
	files, err := b.Read(watcher)
	if err != nil {
		return err
	}
	files = b.hooks.BeforeStaticList(files)
	return b.Write(files)
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{conf: conf, theme: theme, hooks: hooks}
}
