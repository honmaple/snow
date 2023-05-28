package static

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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

func (b *Builder) ignoreFile(path string) func(file string) bool {
	exts := make(map[string]bool)
	for _, ext := range b.conf.GetStringSlice("statics." + path + ".exts") {
		exts[ext] = true
	}
	ignores := make([]*regexp.Regexp, 0)
	for _, pattern := range b.conf.GetStringSlice("statics." + path + ".ignore_files") {
		re, err := regexp.Compile(pattern)
		if err != nil {
			b.conf.Log.Errorln(err.Error())
			continue
		}
		ignores = append(ignores, re)
	}
	return func(file string) bool {
		if len(exts) > 0 && !exts[filepath.Ext(file)] {
			return true
		}
		if strings.HasPrefix(file, "/") {
			file = file[1:]
		}
		for _, re := range ignores {
			if re.MatchString(file) {
				return true
			}
		}
		return false
	}
}

func (b *Builder) loadStatics() Statics {
	staticFiles := make([]*Static, 0)

	for name := range b.conf.GetStringMap("statics") {
		output := b.conf.GetString("statics." + name + ".path")
		if output == "" {
			continue
		}

		isTheme := strings.HasPrefix(name, "@theme/")
		isIgnored := b.ignoreFile(name)

		walkFunc := func(file string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			var root fs.FS

			if isTheme {
				root = b.theme
			} else {
				root = os.DirFS(".")
			}

			staticFile := &Static{
				File: &localFile{file: file, root: root, isTheme: isTheme},
			}
			staticName := strings.TrimPrefix(staticFile.File.Name(), name)

			if isIgnored(staticName) {
				return nil
			}

			if strings.HasSuffix(output, "/") {
				staticFile.Path = filepath.Join(output, filepath.Base(name), staticName)
			} else {
				staticFile.Path = filepath.Join(output, staticName)
			}
			staticFiles = append(staticFiles, staticFile)
			return nil
		}

		if isTheme {
			fs.WalkDir(b.theme, name[7:], func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				info, err := d.Info()
				return walkFunc(path, info, err)
			})
			continue
		}
		filepath.Walk(name, walkFunc)
	}
	return staticFiles
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
