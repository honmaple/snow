package core

import (
	"io/fs"
	"os"

	"github.com/honmaple/snow/internal/theme"
	"github.com/sirupsen/logrus"
)

type (
	Context struct {
		*LocaleContext
		Theme          fs.FS
		Logger         Logger
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

func WithTheme(theme fs.FS) ContextOption {
	return func(ctx *Context) {
		ctx.Theme = theme
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
	if ctx.Theme == nil {
		t, err := theme.New(conf.GetString("theme"))
		if err != nil {
			return nil, err
		}
		ctx.Theme = t
	}

	if err := conf.MergeFromThemeConfig(ctx.Theme); err != nil {
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
		lctx.Config.MergeConfigMap(conf.AllSettings())
		lctx.Config.MergeConfigMap(conf.GetStringMap("languages." + lang))
		lctx.Config.Set("language", lang)

		ctx.OtherLanguages[lang] = lctx
	}
	return ctx, nil
}
