package core

import (
	"fmt"
	stdpath "path"
	"strings"

	"github.com/gosimple/slug"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/viper"
)

type LocaleContext struct {
	Config *Config
}

func (ctx *LocaleContext) GetDefaultLanguage() string {
	return ctx.Config.GetString("site.language")
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

func (ctx *LocaleContext) GetHighlightStyle() string {
	return ctx.Config.GetString("content_highlight_style")
}

func (ctx *LocaleContext) GetSummary(content string) string {
	length := ctx.Config.GetInt("content_truncate_len")
	ellipsis := ctx.Config.GetString("content_truncate_ellipsis")
	return utils.TruncateHTML(content, length, ellipsis)
}

func (ctx *LocaleContext) GetSlug(name string) string {
	if ctx.Config.GetBool("slugify") {
		return slug.Make(name)
	}
	return name
}

func (ctx *LocaleContext) GetPathSlug(path string) string {
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

func (ctx *LocaleContext) GetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return ctx.Config.GetString("site.url") + path
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
