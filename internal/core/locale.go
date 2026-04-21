package core

import (
	stdpath "path"
	"strings"

	"github.com/gosimple/slug"
	"github.com/honmaple/snow/internal/utils"
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

func (ctx *LocaleContext) GetSectionConfig(dir string, keyName string) string {
	currentDir := dir
	for {
		if currentDir == "" || currentDir == "." {
			break
		}
		sectionKey := "sections." + currentDir
		if ctx.Config.IsSet(sectionKey) {
			section := ctx.Config.Sub(sectionKey)
			if section != nil && section.IsSet(keyName) {
				if val := section.GetString(keyName); val != "" {
					return val
				}
			}
		}
		currentDir = stdpath.Dir(currentDir)
	}
	return ctx.Config.GetString("sections._default." + keyName)
}
