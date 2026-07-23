package core

import (
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"

	"github.com/honmaple/snow/internal/theme"
	"github.com/honmaple/snow/internal/utils/mergefs"
	"github.com/sirupsen/logrus"
)

type (
	Context struct {
		*LocaleContext
		Logger         Logger
		FS             VirtualFS
		OtherLanguages map[string]*LocaleContext
	}
	ContextOption func(*Context)
)

func (ctx *Context) For(lang string) *LocaleContext {
	if lang == ctx.GetDefaultLanguage() {
		return ctx.LocaleContext
	}
	lctx, ok := ctx.OtherLanguages[lang]
	if ok {
		return lctx
	}
	return ctx.LocaleContext
}

func (ctx *Context) GetAllLanguages() []string {
	langs := []string{
		ctx.GetDefaultLanguage(),
	}
	for lang := range ctx.OtherLanguages {
		langs = append(langs, lang)
	}
	return langs
}

func (ctx *Context) GetDefaultLanguage() string {
	return ctx.Config.GetString("language")
}

func (ctx *Context) VerifyLanguage(lang string) bool {
	if lang == "" {
		return false
	}
	if lang == ctx.GetDefaultLanguage() {
		return true
	}
	_, ok := ctx.OtherLanguages[lang]
	return ok
}

func (ctx *Context) GetFS(name string, includeTheme bool, includeInternal bool) (fs.FS, error) {
	name = normalizeVirtualPath(name)
	fsys := make([]fs.FS, 0, 4)

	if subFS := subDirFSIfExists(ctx.FS, name); subFS != nil {
		fsys = append(fsys, subFS)
	}

	if includeTheme {
		if theme := ctx.GetTheme(); theme != "" {
			if subFS := subDirFSIfExists(ctx.FS, stdpath.Join(MountThemes, theme, name)); subFS != nil {
				fsys = append(fsys, subFS)
			}
		}
	}
	if includeInternal {
		if subFS := subDirFSIfExists(theme.FS, name); subFS != nil {
			fsys = append(fsys, subFS)
			item, err := newMountEntry(subFS, "internal")
			if err == nil {
				fsys = append(fsys, item)
			}
		}
	}
	return mergefs.Merge(fsys...), nil
}

func WithLogger(log Logger) ContextOption {
	return func(ctx *Context) {
		ctx.Logger = log
	}
}

func NewContext(conf *Config, opts ...ContextOption) (*Context, error) {
	ctx := &Context{
		LocaleContext: &LocaleContext{
			Config: conf,
		},
		OtherLanguages: make(map[string]*LocaleContext),
	}
	for _, opt := range opts {
		opt(ctx)
	}
	if ctx.Logger == nil {
		level := logrus.InfoLevel
		if conf.GetBool("debug") {
			level = logrus.DebugLevel
		}
		ctx.Logger = &logrus.Logger{
			Out: os.Stderr,
			Formatter: &logrus.TextFormatter{
				DisableTimestamp: true,
				FullTimestamp:    false,
			},
			Level: level,
		}
	}

	if theme := ctx.GetTheme(); theme != "" {
		if err := conf.MergeFromThemeConfig(filepath.Join(MountThemes, theme)); err != nil {
			return nil, err
		}
	}
	ctx.FS = newVirtualFS(os.DirFS("."))

	defaultLanguage := ctx.GetDefaultLanguage()
	if defaultLanguage == "" {
		defaultLanguage = "en"
	}

	for lang := range conf.GetStringMap("languages") {
		if lang == defaultLanguage {
			continue
		}
		lctx := &LocaleContext{
			Config: NewConfig(),
		}
		for _, key := range conf.AllKeys() {
			lctx.Config.Set(key, conf.Get(key))
		}

		lconf := conf.Sub(fmt.Sprintf("languages.%s", lang))
		for _, key := range lconf.AllKeys() {
			lctx.Config.Set(key, lconf.Get(key))
		}
		lctx.Config.Set("language", lang)

		ctx.OtherLanguages[lang] = lctx
	}
	return ctx, nil
}
