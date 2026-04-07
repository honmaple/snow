package core

import (
	"os"
	"reflect"
	"strings"

	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func (conf *Config) GetSubSlice(key string) []*viper.Viper {
	data := conf.Get(key)
	if data == nil {
		return nil
	}

	if reflect.TypeOf(data).Kind() == reflect.Slice {
		var vList []*viper.Viper
		for _, item := range data.([]any) {
			subv := viper.New()
			for k, v := range cast.ToStringMap(item) {
				subv.Set(k, v)
			}
			vList = append(vList, subv)
		}
		return vList
	}
	return nil
}

func (conf *Config) SetDebug() {
	// conf.Log.SetLevel(logrus.DebugLevel)
}

func (conf *Config) SetFilter(filter string) {
	conf.Set("hooks.internal.filter", filter)
}

func (conf *Config) SetOutput(output string) {
	conf.Set("output_dir", output)
}

func (conf *Config) SetMode(mode string) {
	// key := fmt.Sprintf("mode.%s", mode)
	// if !conf.IsSet(key) {
	//	conf.Log.Fatalf("The mode %s not found", mode)
	// }
	// var c *Config
	// if file := conf.GetString(fmt.Sprintf("%s.include", key)); file != "" {
	//	c = &Config{
	//		Viper: viper.New(),
	//	}
	//	if err := c.Load(file); err != nil {
	//		conf.Log.Fatal(err.Error())
	//	}
	// } else {
	//	c = &Config{
	//		Viper: conf.Sub(key),
	//	}
	// }
	// keys := conf.AllKeys()
	// for _, k := range c.AllKeys() {
	//	conf.Set(k, c.Get(k))
	// }
	// for _, k := range keys {
	//	if !c.IsSet(k) {
	//		conf.Set(k, conf.Get(k))
	//	}
	// }
}

func (conf *Config) Reset(m map[string]any) {
	for k, v := range m {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
	for _, k := range conf.AllKeys() {
		conf.Set(k, conf.Get(k))
	}
}

func (conf *Config) LoadFromFile(path string) error {
	if path != "" && utils.FileExists(path) {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		conf.SetConfigFile(path)
		if err := conf.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
			return err
		}
	}

	// if n := conf.GetString("theme"); n != "" {
	//	path := filepath.Join("themes", n, "theme.yaml")
	//	if utils.FileExists(path) {
	//		content, err := os.ReadFile(path)
	//		if err != nil {
	//			return err
	//		}
	//		conf.SetConfigFile(path)
	//		if err := conf.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
	//			return err
	//		}
	//	}
	// }

	conf.Reset(siteConfig)
	conf.Reset(otherConfig)
	conf.Reset(staticConfig)
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return nil
}

var (
	sectionConfig = map[string]any{
		"sections._default.path":          "{path:slug}/index.html",
		"sections._default.orderby":       "weight",
		"sections._default.template":      "section.html",
		"sections._default.paginate":      10,
		"sections._default.paginate_path": "{name}{number}{extension}",
		"sections._default.page_path":     "{path:slug}/{slug}/index.html",
		"sections._default.page_orderby":  "date desc",
		"sections._default.page_template": "page.html",
	}
	taxonomyConfig = map[string]any{
		"taxonomies._default.path":               "{taxonomy}/index.html",
		"taxonomies._default.orderby":            "name",
		"taxonomies._default.term_path":          "{taxonomy}/{term:slug}/index.html",
		"taxonomies._default.term_paginate_path": "{name}{number}{extension}",
		"taxonomies._default.term_orderby":       "date desc",

		"taxonomies.categories.weight": 1,
		"taxonomies.tags.weight":       2,
		"taxonomies.authors.weight":    3,
	}
	staticConfig = map[string]any{
		"statics.@theme/_internal/static.path": "static",
		"statics.@theme/static.path":           "static",
		"statics.static.path":                  "static",
	}
	// 默认不需要修改的配置
	otherConfig = map[string]any{
		"content_truncate_len":      49,
		"content_truncate_ellipsis": "...",
		"content_highlight_style":   "monokai",
		"slugify":                   true,
		"formats.rss.template":      "_internal/partials/rss.xml",
		"formats.atom.template":     "_internal/partials/atom.xml",
	}
	// 默认需要修改的配置
	siteConfig = map[string]any{
		"site.url":      "http://127.0.0.1:8000",
		"site.title":    "snow",
		"site.subtitle": "snow is a static site generator.",
		"site.author":   "snow",
		"site.language": "en",
		// "theme.config":   "theme.yaml",
		// "theme.override": "templates",
		"static_dir":  "static",
		"output_dir":  "output",
		"content_dir": "content",
		"themes_dir":  "themes",
	}
)

func NewConfig() *Config {
	return &Config{
		Viper: viper.New(),
	}
}

func DefaultConfig() *Config {
	conf := &Config{
		Viper: viper.New(),
	}
	conf.Reset(siteConfig)
	conf.Reset(otherConfig)
	conf.Reset(staticConfig)
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return conf
}
