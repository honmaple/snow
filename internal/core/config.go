package core

import (
	"fmt"
	"io/fs"
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
	conf.Set("debug", true)
}

func (conf *Config) SetOutput(output string) {
	conf.Set("output_dir", output)
}

func (conf *Config) SetMode(mode string) {
	key := fmt.Sprintf("mode.%s", mode)
	if !conf.IsSet(key) {
		fmt.Printf("The mode %s not found", mode)
		return
	}
	var modeConfig *Config
	if file := conf.GetString(fmt.Sprintf("%s.include", key)); file != "" {
		modeConfig = &Config{
			Viper: viper.New(),
		}
		if err := modeConfig.LoadFromFile(file); err != nil {
			fmt.Printf("The mode %s not found", mode)
			return
		}
	} else {
		modeConfig = &Config{
			Viper: conf.Sub(key),
		}
	}
	keys := conf.AllKeys()
	for _, k := range modeConfig.AllKeys() {
		conf.Set(k, modeConfig.Get(k))
	}
	for _, k := range keys {
		if !modeConfig.IsSet(k) {
			conf.Set(k, conf.Get(k))
		}
	}
}

func (conf *Config) Reset(m map[string]any) {
	for k, v := range m {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
}

func (conf *Config) LoadFromFile(file string) error {
	if file != "" && utils.FileExists(file) {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		v := viper.New()
		v.SetConfigFile(file)

		if err := v.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
			return err
		}
		for _, k := range v.AllKeys() {
			conf.Set(k, v.Get(k))
		}
	}

	conf.Reset(siteConfig)
	conf.Reset(otherConfig)
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return nil
}

func (conf *Config) MergeFromThemeConfig(theme fs.FS) error {
	var themeFile string

	for _, file := range []string{"theme.yaml", "theme.toml", "theme.json"} {
		if _, err := fs.Stat(theme, file); err == nil {
			themeFile = file
			break
		}
	}
	if themeFile == "" {
		return nil
	}

	content, err := fs.ReadFile(theme, themeFile)
	if err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigFile(themeFile)

	if err := v.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
		return err
	}
	for _, k := range v.AllKeys() {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v.Get(k))
	}
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
	// 默认不需要修改的配置
	otherConfig = map[string]any{
		"content_truncate_len":      49,
		"content_truncate_ellipsis": "...",
		"content_highlight_style":   "monokai",
		"slugify":                   true,
		"formats.rss.template":      "partials/rss.xml",
		"formats.atom.template":     "partials/atom.xml",
	}
	// 默认需要修改的配置
	siteConfig = map[string]any{
		"site.url":      "http://127.0.0.1:8000",
		"site.title":    "snow",
		"site.subtitle": "snow is a static site generator.",
		"site.author":   "honmaple",
		"site.language": "en",
		"theme_dir":     "themes",
		"static_dir":    "static",
		"output_dir":    "output",
		"content_dir":   "content",
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
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return conf
}
