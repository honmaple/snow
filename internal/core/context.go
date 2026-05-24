package core

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/honmaple/snow/internal/theme"
	"github.com/honmaple/snow/internal/utils/mergefs"
	"github.com/sirupsen/logrus"
)

type (
	Context struct {
		*LocaleContext
		Logger         Logger
		ThemeFS        fs.FS
		OtherLanguages map[string]*LocaleContext
	}
	ContextOption func(*Context)
)

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

func (ctx *Context) GetFS(path string, internal bool) (fs.FS, error) {
	fsys := make([]fs.FS, 0)

	if _, err := os.Stat(path); err == nil {
		fsys = append(fsys, os.DirFS(path))
	}

	if subFS, err := fs.Sub(ctx.ThemeFS, path); err == nil {
		fsys = append(fsys, subFS)
	}
	if internal {
		if subFS, err := fs.Sub(ctx.ThemeFS, "internal/"+path); err == nil {
			fsys = append(fsys, subFS)
		}
	}
	return mergefs.Merge(fsys...), nil
}

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

	themeFS, err := theme.New(ctx.GetThemeDir(), ctx.GetTheme())
	if err != nil {
		return nil, err
	}
	ctx.ThemeFS = themeFS

	if err := conf.MergeFromThemeConfig(ctx.ThemeFS); err != nil {
		return nil, err
	}

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
