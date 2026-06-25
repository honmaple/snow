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
	return slug.Make(name)
}

func (ctx *LocaleContext) GetPathSlug(path string) string {
	names := strings.Split(path, "/")
	slugs := make([]string, len(names))
	for i, name := range names {
		slugs[i] = ctx.GetSlug(name)
	}
	return strings.Join(slugs, "/")
}

func (ctx *LocaleContext) GetBaseURL() string {
	return ctx.Config.GetString("base_url")
}

func (ctx *LocaleContext) GetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return ctx.Config.GetString("base_url") + path
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

// func (ctx *LocaleContext) CleanOutputDir() error {
//	if out := ctx.GetOutputDir(); out != "" {
//		ctx.Logger.Infoln("Removing the contents of", out)

//		files, err := os.ReadDir(out)
//		if err != nil {
//			return err
//		}
//		for _, file := range files {
//			if strings.HasPrefix(file.Name(), ".") {
//				continue
//			}
//			if err := os.RemoveAll(filepath.Join(out, file.Name())); err != nil {
//				return err
//			}
//		}
//		return nil
//	}
//	return nil
// }
