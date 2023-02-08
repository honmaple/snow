package static

import (
	"context"
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

type Builder struct {
	conf  config.Config
	theme theme.Theme
	hooks Hooks
}

func (b *Builder) genFile(file string, output string, isTheme bool) *Static {
	if output == "" {
		return nil
	}
	if strings.HasSuffix(output, "/") {
		output = filepath.Join(output, filepath.Base(file))
	}
	if isTheme {
		return &Static{URL: output, File: file, Root: b.theme, IsTheme: isTheme}
	}
	return &Static{URL: output, File: file, Root: os.DirFS("."), IsTheme: isTheme}
}

func (b *Builder) loader() func(string, bool) *Static {
	exts := make(map[string]bool)
	extsIsSet := b.conf.IsSet("static_exts")
	if extsIsSet {
		for _, ext := range b.conf.GetStringSlice("static_exts") {
			exts[ext] = true
		}
	}

	meta := b.conf.GetStringMapString("static_paths")
	if _, ok := meta["@theme/static"]; !ok {
		meta["@theme/static"] = "static/"
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
			file = fmt.Sprintf("@theme/%s", file)
		}
		// viper 无法处理带.的key, 无法直接使用meta[file]
		if k := fmt.Sprintf("static_paths.%s", file); b.conf.IsSet(k) {
			return b.genFile(rawFile, b.conf.GetString(k), isTheme)
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
			return b.genFile(rawFile, output, isTheme)
		}
		return b.genFile(rawFile, file, isTheme)
	}
}

func (b *Builder) loadStatics() Statics {
	loader := b.loader()

	files := make([]*Static, 0)
	// 默认添加主题的静态文件
	fs.WalkDir(b.theme, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if file := loader(path, true); file != nil {
			files = append(files, file)
		}
		return nil
	})

	for _, dir := range b.Dirs() {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if file := loader(path, false); file != nil {
				files = append(files, file)
			}
			return nil
		})
	}
	return files
}

func (b *Builder) Dirs() []string {
	return b.conf.GetStringSlice("static_dirs")
}

func (b *Builder) Build(ctx context.Context) error {
	now := time.Now()

	files := b.loadStatics()
	defer func() {
		b.conf.Log.Infoln("Done: Static Processed", len(files), "static files", "in", time.Now().Sub(now))
	}()
	files = b.hooks.BeforeStaticsWrite(files)
	return b.Write(files)
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{
		conf:  conf,
		theme: theme,
		hooks: hooks,
	}
}
