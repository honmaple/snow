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

type Result struct {
	value any
}

func (r Result) Int() int {
	return cast.ToInt(r.value)
}

func (r Result) Int32() int32 {
	return cast.ToInt32(r.value)
}

func (r Result) Int64() int64 {
	return cast.ToInt64(r.value)
}

func (r Result) Bool() bool {
	return cast.ToBool(r.value)
}

func (r Result) String() string {
	return cast.ToString(r.value)
}

func (r Result) StringMap() map[string]any {
	return cast.ToStringMap(r.value)
}

func (r Result) StringSlice() []string {
	return cast.ToStringSlice(r.value)
}

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
	key := fmt.Sprintf("modes.%s", mode)
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

func (conf *Config) Reset(m map[string]any, force bool) {
	for k, v := range m {
		if !force && conf.IsSet(k) {
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

	conf.MergeFromDefaultConfig(false)
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

func (conf *Config) MergeFromDefaultConfig(force bool) {
	conf.Reset(siteConfig, force)
	conf.Reset(pageConfig, force)
	conf.Reset(sectionConfig, force)
	conf.Reset(taxonomyConfig, force)
	conf.Reset(hookConfig, force)
}

var (
	siteConfig = map[string]any{
		"base_url":                  "http://127.0.0.1:8000",
		"title":                     "snow",
		"description":               "snow is a static site generator.",
		"author":                    "honmaple",
		"language":                  "en",
		"theme_dir":                 "themes",
		"static_dir":                "static",
		"output_dir":                "output",
		"content_dir":               "content",
		"slugify":                   true,
		"content_truncate_len":      49,
		"content_truncate_ellipsis": "...",
		"content_highlight_style":   "monokai",
		"formats.rss.template":      "partials/rss.xml",
		"formats.atom.template":     "partials/atom.xml",
	}
	pageConfig = map[string]any{
		"pages._default.path":     "{path:slug}/{slug}/",
		"pages._default.template": "page.html",
	}
	sectionConfig = map[string]any{
		"sections._default.path":          "{path:slug}/",
		"sections._default.template":      "section.html",
		"sections._default.sort_by":       "date desc",
		"sections._default.paginate":      10,
		"sections._default.paginate_path": "{name}{number}{extension}",
	}
	taxonomyConfig = map[string]any{
		"taxonomies._default.path":               "{taxonomy}/",
		"taxonomies._default.sort_by":            "name",
		"taxonomies._default.term.path":          "{taxonomy}/{term:slug}/",
		"taxonomies._default.term.sort_by":       "date desc",
		"taxonomies._default.term.paginate_path": "{name}{number}{extension}",
	}
	hookConfig = map[string]any{
		"hooks.assets.enabled":    true,
		"hooks.encrypt.enabled":   true,
		"hooks.shortcode.enabled": true,

		// 先encrypt再shortcode
		"hooks.encrypt.weight":   2,
		"hooks.shortcode.weight": 1,
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
	conf.MergeFromDefaultConfig(true)
	return conf
}
