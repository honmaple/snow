package core

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/gosimple/slug"
	"github.com/honmaple/snow/internal/theme"
	"github.com/honmaple/snow/internal/utils"
	"github.com/sirupsen/logrus"
)

type (
	Logger interface {
		Debug(...any)
		Debugf(string, ...any)
		Debugln(...any)
		Info(...any)
		Infof(string, ...any)
		Infoln(...any)
		Warn(...any)
		Warnf(string, ...any)
		Warnln(...any)
		Error(...any)
		Errorf(string, ...any)
		Errorln(...any)
	}
	Context struct {
		Theme   fs.FS
		Logger  Logger
		Config  *Config
		Locales map[string]*Context
	}
	ContextOption func(*Context)
)

func (ctx *Context) For(lang string) *Context {
	if lang == ctx.GetDefaultLanguage() {
		return ctx
	}
	c, ok := ctx.Locales[lang]
	if ok {
		return c
	}
	return ctx
}

func (ctx *Context) IsValidLanguage(lang string) bool {
	return ctx.GetDefaultLanguage() == lang || ctx.Config.IsSet("languages."+lang)
}

func (ctx *Context) IsHome(path string) bool {
	return ctx.Config.GetString("content_dir") == path
}

func (ctx *Context) GetConfigMap(lang string) map[string]any {
	if lang == "" {
		return ctx.Config.AllSettings()
	}
	locale, ok := ctx.Locales[lang]
	if !ok {
		return ctx.Config.AllSettings()
	}
	return locale.Config.AllSettings()
}

func (ctx *Context) GetSummary(text string) string {
	length := ctx.Config.GetInt("content_truncate_len")
	ellipsis := ctx.Config.GetString("content_truncate_ellipsis")
	return utils.TruncateHTML(text, length, ellipsis)
}

func (ctx *Context) GetSlug(name string) string {
	if ctx.Config.GetBool("slugify") {
		return slug.Make(name)
	}
	return name
}

func (ctx *Context) GetPathSlug(path string) string {
	if ctx.Config.GetBool("slugify") {
		names := strings.Split(path, "/")
		slugs := make([]string, len(names))
		for i, name := range names {
			slugs[i] = slug.Make(name)
		}
		return strings.Join(slugs, "/")
	}
	return path
}

func (ctx *Context) GetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return ctx.Config.GetString("site.url") + path
}

func (ctx *Context) GetRelURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (ctx *Context) GetDefaultLanguage() string {
	return ctx.Config.GetString("site.language")
}

func (ctx *Context) GetStaticDir() string {
	return ctx.Config.GetString("static_dir")
}

func (ctx *Context) GetContentDir() string {
	return ctx.Config.GetString("content_dir")
}

func (ctx *Context) GetStaticConfig(file string, key string) string {
	strs := strings.Split(file, "/")
	for i := len(strs); i > 0; i-- {
		value := ctx.Config.GetString(fmt.Sprintf("static.%s.%s", strings.Join(strs[:i], "/"), key))
		if value != "" {
			return value
		}
	}
	return ctx.Config.GetString(fmt.Sprintf("static._default.%s", key))
}

func (ctx *Context) GetSectionConfig(dir string, key string) string {
	if dir == "." {
		return ctx.Config.GetString(fmt.Sprintf("sections._default.%s", key))
	}
	// dir := filepath.Dir(path)
	// 获取配置 content/posts/linux/emacs/page-01.md
	// 查找顺序: posts/linux/emacs -> posts/linux -> posts -> _default
	strs := strings.Split(dir, "/")
	for i := len(strs); i > 0; i-- {
		value := ctx.Config.GetString(fmt.Sprintf("sections.%s.%s", strings.Join(strs[:i], "/"), key))
		if value != "" {
			return value
		}
	}
	return ctx.Config.GetString(fmt.Sprintf("sections._default.%s", key))
}

func (ctx *Context) GetHighlightStyle() string {
	return ctx.Config.GetString("content_highlight_style")
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
		Config:  conf,
		Locales: make(map[string]*Context),
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

	defaultLanguage := conf.GetString("site.language")
	if defaultLanguage == "" {
		defaultLanguage = "en"
	}

	for lang := range conf.GetStringMap("languages") {
		if lang == defaultLanguage {
			continue
		}
		lctx := &Context{
			Logger: ctx.Logger,
			Config: NewConfig(),
		}
		lctx.Config.MergeConfigMap(conf.AllSettings())
		lctx.Config.MergeConfigMap(conf.GetStringMap("languages." + lang))
		lctx.Config.Set("site.language", lang)

		t, err := theme.New(lctx.Config.GetString("theme"))
		if err != nil {
			return nil, err
		}
		lctx.Theme = t

		ctx.Locales[lang] = lctx
	}
	return ctx, nil
}
