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
	ctx   *Context
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

func (b *Builder) Build(ctx context.Context) error {
	now := time.Now()
	defer func() {
		if count := len(b.ctx.Statics()); count > 0 {
			lang := ""
			if !b.conf.IsDefaultLanguage(b.conf.Site.Language) {
				lang = "[" + b.conf.Site.Language + "]"
			}
			b.conf.Log.Infoln("Done:", lang+"Static Processed", count, "static files", "in", time.Now().Sub(now))
		}
	}()

	// 因为viper不能识别文件名中的".", 所以这里通过获取".path"的前缀来获取文件名
	names := make([]string, 0)
	for _, name := range b.conf.Sub("statics").AllKeys() {
		if strings.HasSuffix(name, ".path") {
			names = append(names, name[:len(name)-5])
		}
	}
	for _, name := range names {
		output := b.conf.GetString("statics." + name + ".path")
		if output == "" {
			continue
		}
		isTheme := strings.HasPrefix(name, "@theme/")
		isIgnore := b.ignoreFile(name)

		walkFunc := func(file string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			staticFile := &Static{Name: file}
			if isTheme {
				staticFile.Root = b.theme
				staticFile.Name = filepath.Join("@theme", staticFile.Name)
			} else {
				staticFile.Root = os.DirFS(".")
			}

			staticName := strings.TrimPrefix(staticFile.Name, name)
			if isIgnore(staticName) {
				return nil
			}

			if strings.HasSuffix(output, "/") {
				staticFile.Path = filepath.Join(output, filepath.Base(name), staticName)
			} else {
				staticFile.Path = filepath.Join(output, staticName)
			}
			staticFile.Path = b.conf.GetRelURL(staticFile.Path)
			staticFile.Permalink = b.conf.GetURL(staticFile.Path)

			staticFile = b.hooks.Static(staticFile)
			if staticFile == nil {
				return nil
			}
			b.ctx.insertStatic(staticFile)
			return nil
		}

		if isTheme {
			fs.WalkDir(b.theme, name[7:], func(file string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				info, err := d.Info()
				return walkFunc(file, info, err)
			})
			continue
		}
		filepath.Walk(name, walkFunc)
	}
	return b.Write()
}

func (b *Builder) Write() error {
	for _, static := range b.hooks.Statics(b.ctx.Statics()) {
		// src := static.File.Name()
		// dst := filepath.Join(b.conf.OutputDir, static.Path)
		// b.conf.Log.Debugln("Copying", src, "to", dst)
		file, err := static.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		if err := b.conf.Write(static.Path, file); err != nil {
			return err
		}
	}
	return nil
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	return &Builder{
		conf:  conf,
		theme: theme,
		hooks: hooks,
		ctx:   newContext(conf),
	}
}
