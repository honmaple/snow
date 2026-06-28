package core

import (
	"fmt"
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/utils"
	"github.com/honmaple/snow/internal/utils/slugify"
	"github.com/spf13/viper"
)

type LocaleContext struct {
	Config *Config
}

func (ctx *LocaleContext) GetLanguage() string {
	return ctx.Config.GetString("language")
}

func (ctx *LocaleContext) GetTheme() string {
	return ctx.Config.GetString("theme")
}

func (ctx *LocaleContext) GetThemeDir() string {
	return ctx.Config.GetString("theme_dir")
}

func (ctx *LocaleContext) GetStaticDir() string {
	return ctx.Config.GetString("static_dir")
}

func (ctx *LocaleContext) GetContentDir() string {
	return ctx.Config.GetString("content_dir")
}

func (ctx *LocaleContext) GetOutputDir() string {
	return ctx.Config.GetString("output_dir")
}

func (ctx *LocaleContext) GetSummary(content string) string {
	length := ctx.Config.GetInt("content_truncate_len")
	ellipsis := ctx.Config.GetString("content_truncate_ellipsis")
	return utils.TruncateHTML(content, length, ellipsis)
}

func (ctx *LocaleContext) GetSlug(name string) string {
	opts := make([]slugify.Option, 0)
	if k := "slugify.lowercase"; ctx.Config.IsSet(k) {
		opts = append(opts, slugify.WithLowercase(ctx.Config.GetBool(k)))
	}
	if k := "slugify.preserve_chars"; ctx.Config.IsSet(k) {
		opts = append(opts, slugify.WithPreserveChars(ctx.Config.GetString(k)))
	}
	if k := "slugify.preserve_unicode"; ctx.Config.IsSet(k) {
		opts = append(opts, slugify.WithPreserveUnicode(ctx.Config.GetBool(k)))
	}
	return slugify.Make(name, opts...)
}

func (ctx *LocaleContext) GetPathSlug(path string) string {
	names := strings.Split(path, "/")
	slugs := make([]string, len(names))
	for i, name := range names {
		slugs[i] = ctx.GetSlug(name)
	}
	return strings.Join(slugs, "/")
}

func (ctx *LocaleContext) applyPathStyle(path string, opts ...slugify.Option) string {
	ext := stdpath.Ext(path)
	if ext != "" {
		path = strings.TrimSuffix(path, ext)
	}

	names := strings.Split(path, "/")
	slugs := make([]string, len(names))
	for i, name := range names {
		if len(opts) > 0 {
			slugs[i] = slugify.Make(name, opts...)
		} else {
			slugs[i] = ctx.GetSlug(name)
		}
	}
	return strings.Join(slugs, "/") + strings.ToLower(ext)
}

func (ctx *LocaleContext) ApplyPathStyle(path string, style string) string {
	for s := range strings.SplitSeq(style, ",") {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "lower":
			path = strings.ToLower(path)
		case "slug":
			path = ctx.applyPathStyle(path)
		case "slug_unicode":
			path = ctx.applyPathStyle(path, slugify.WithPreserveUnicode(true))
		}
	}
	return path
}

func (ctx *LocaleContext) GetBaseURL() string {
	return ctx.Config.GetString("base_url")
}

func (ctx *LocaleContext) GetRelURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (ctx *LocaleContext) GetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return ctx.GetBaseURL() + path
}

func (ctx *LocaleContext) GetPageConfig(dir string) map[string]any {
	config := viper.New()

	currentDir := dir
	for {
		if currentDir == "" || currentDir == "." {
			break
		}
		for k, v := range ctx.Config.GetStringMap("pages." + currentDir) {
			if config.IsSet(k) {
				continue
			}
			config.Set(k, v)
		}
		currentDir = stdpath.Dir(currentDir)
	}
	for k, v := range ctx.Config.GetStringMap("pages._default") {
		if config.IsSet(k) {
			continue
		}
		config.Set(k, v)
	}
	return config.AllSettings()
}

func (ctx *LocaleContext) GetSectionConfig(dir string) map[string]any {
	config := viper.New()

	currentDir := dir
	for {
		if currentDir == "" || currentDir == "." {
			break
		}
		for k, v := range ctx.Config.GetStringMap("sections." + currentDir) {
			if config.IsSet(k) {
				continue
			}
			config.Set(k, v)
		}
		currentDir = stdpath.Dir(currentDir)
	}
	for k, v := range ctx.Config.GetStringMap("sections._default") {
		if config.IsSet(k) {
			continue
		}
		config.Set(k, v)
	}
	return config.AllSettings()
}

func (ctx *LocaleContext) GetTaxonomyConfig(name string, keyName string) Result {
	var val any

	if key := fmt.Sprintf("taxonomies.%s.%s", name, keyName); ctx.Config.IsSet(key) {
		val = ctx.Config.Get(key)
	}
	if val == nil || val == "" {
		val = ctx.Config.Get("taxonomies._default." + keyName)
	}
	return Result{value: val}
}

func (ctx *LocaleContext) GetFormatConfig(name string, keyName string) Result {
	var val any

	if key := fmt.Sprintf("formats.%s.%s", name, keyName); ctx.Config.IsSet(key) {
		val = ctx.Config.Get(key)
	}
	if val == nil || val == "" {
		val = ctx.Config.Get("formats._default." + keyName)
	}
	return Result{value: val}
}

func (ctx *LocaleContext) GetMarkupConfig(name string, keyName string) Result {
	var val any

	if key := fmt.Sprintf("markups.%s.%s", name, keyName); ctx.Config.IsSet(key) {
		val = ctx.Config.Get(key)
	}
	if val == nil || val == "" {
		val = ctx.Config.Get("markups._default." + keyName)
	}
	return Result{value: val}
}
