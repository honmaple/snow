package static

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	Static struct {
		URL     string
		File    string
		IsTheme bool
	}
	Statics []*Static
	Builder struct {
		conf  config.Config
		theme theme.Theme
		hooks Hooks
	}
)

func (s *Static) src() string {
	if s.IsTheme {
		return fmt.Sprintf("theme/%s", s.File)
	}
	return s.File
}

func (b *Builder) Dirs() []string {
	return b.conf.GetStringSlice("static_dirs")
}

func (b *Builder) parser() func(string, bool) *Static {
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

	return func(file string, isTheme bool) *Static {
		if extsIsSet && !exts[filepath.Ext(file)] {
			return nil
		}
		rawFile := file
		if isTheme {
			file = fmt.Sprintf("theme/%s", file)
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
			return &Static{URL: output, File: rawFile, IsTheme: isTheme}
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
			return &Static{URL: output, File: rawFile, IsTheme: isTheme}
		}
		return &Static{URL: file, File: rawFile, IsTheme: isTheme}
	}

}

func (b *Builder) Read(dirs []string) ([]*Static, error) {
	parse := b.parser()

	files := make([]*Static, 0)
	// 默认添加主题的静态文件
	root := b.theme.Root()
	fs.WalkDir(root, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if file := parse(path, true); file != nil {
			files = append(files, file)
		}
		return nil
	})

	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if file := parse(path, false); file != nil {
				files = append(files, file)
			}
			return nil
		})
	}
	return files, nil
}

func (b *Builder) Build() error {
	dirs := b.Dirs()
	if len(dirs) == 0 {
		return nil
	}

	now := time.Now()
	files, err := b.Read(dirs)
	if err != nil {
		return err
	}
	files = b.hooks.BeforeStaticsWrite(files)
	defer func() {
		b.conf.Log.Infoln("Done: Processed", len(files), "static files", "in", time.Now().Sub(now))
	}()
	return b.Write(files)
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{conf: conf, theme: theme, hooks: hooks}
}
