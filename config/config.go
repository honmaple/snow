package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gosimple/slug"
	"github.com/honmaple/snow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Site struct {
	URL      string
	Title    string
	SubTitle string
	Language string
}

type Config struct {
	*viper.Viper
	Log   *logrus.Logger
	Cache *sync.Map

	writer Writer

	Site            Site
	OutputDir       string
	ContentDir      string
	DefaultLanguage string

	Languages map[string]*Config
}

type Writer interface {
	Write(string, io.Reader) error
	Watch(string) error
}

func (conf *Config) With(lang string) Config {
	langc, ok := conf.Languages[lang]
	if !ok {
		return *conf
	}
	return *langc
}

func (conf *Config) IsValidLanguage(lang string) bool {
	return conf.DefaultLanguage == lang || conf.IsSet("languages."+lang)
}

func (conf *Config) IsDefaultLanguage(lang string) bool {
	return conf.DefaultLanguage == lang
}

func (conf *Config) SetDebug() {
	conf.Log.SetLevel(logrus.DebugLevel)
}

func (conf *Config) SetFilter(filter string) {
	conf.Set("hooks.internal.filter", filter)
}

func (conf *Config) SetOutput(output string) {
	conf.Set("output_dir", output)
}

func (conf *Config) SetMode(mode string) {
	key := fmt.Sprintf("mode.%s", mode)
	if !conf.IsSet(key) {
		conf.Log.Fatalf("The mode %s not found", mode)
	}
	var c *Config
	if file := conf.GetString(fmt.Sprintf("%s.include", key)); file != "" {
		c = &Config{
			Viper: viper.New(),
		}
		if err := c.Load(file); err != nil {
			conf.Log.Fatal(err.Error())
		}
	} else {
		c = &Config{
			Viper: conf.Sub(key),
		}
	}
	keys := conf.AllKeys()
	for _, k := range c.AllKeys() {
		conf.Set(k, c.Get(k))
	}
	for _, k := range keys {
		if !c.IsSet(k) {
			conf.Set(k, conf.Get(k))
		}
	}
}

func (conf *Config) WithWriter(w Writer) Config {
	conf.writer = w
	for _, langc := range conf.Languages {
		langc.writer = w
	}
	return *conf
}

func (conf *Config) Watch(file string) error {
	if conf.writer == nil {
		return nil
	}
	return conf.writer.Watch(file)
}

func (conf *Config) Write(file string, r io.Reader) error {
	if file == "" {
		return nil
	}
	output := filepath.Join(conf.OutputDir, file)

	conf.Log.Debugln("Writing", output)
	if conf.writer != nil {
		return conf.writer.Write(file, r)
	}

	if dir := filepath.Dir(output); !utils.FileExists(output) {
		os.MkdirAll(dir, 0755)
	}
	dstFile, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, r)
	return err
}

func (conf *Config) GetSummary(text string) string {
	length := conf.GetInt("content_truncate_len")
	ellipsis := conf.GetString("content_truncate_ellipsis")
	return utils.TruncateHTML(text, length, ellipsis)
}

func (conf *Config) GetHighlightStyle() string {
	return conf.GetString("content_highlight_style")
}

func (conf *Config) GetSlug(name string) string {
	if conf.GetBool("slugify") {
		return slug.Make(name)
	}
	return name
}

func (conf *Config) GetRelURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if conf.IsDefaultLanguage(conf.Site.Language) {
		return path
	}
	return "/" + conf.Site.Language + path
}

func (conf *Config) GetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return conf.Site.URL + path
}

func (conf *Config) Reset(m map[string]interface{}) {
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

func (conf *Config) ResetByFile(file string, r io.Reader) {
	v := viper.New()
	v.SetConfigFile(file)

	if err := v.ReadConfig(r); err == nil {
		for _, k := range v.AllKeys() {
			if conf.IsSet(k) {
				continue
			}
			conf.Set(k, v.Get(k))
		}
	}
}

func (conf *Config) Unmarshal(key string, val interface{}) error {
	if conf.IsSet(key) {
		return conf.UnmarshalKey(key, val, func(decoderConfig *mapstructure.DecoderConfig) {
			decoderConfig.TagName = "json"
		})
	}
	return nil
}

func (conf *Config) Load(path string) error {
	if path != "" && utils.FileExists(path) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		conf.SetConfigFile(path)
		if err := conf.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
			return err
		}
	}
	conf.Reset(siteConfig)
	conf.Reset(otherConfig)
	conf.Reset(staticConfig)
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return nil
}

func (conf *Config) Init() {
	conf.Site = Site{
		URL:      conf.GetString("site.url"),
		Title:    conf.GetString("site.title"),
		SubTitle: conf.GetString("site.subtitle"),
		Language: conf.GetString("site.language"),
	}
	conf.OutputDir = conf.GetString("output_dir")
	conf.ContentDir = conf.GetString("content_dir")
	conf.DefaultLanguage = conf.GetString("site.language")

	conf.Languages = make(map[string]*Config)
	for lang := range conf.GetStringMap("languages") {
		if lang == conf.DefaultLanguage {
			continue
		}
		langc := &Config{
			Viper:  viper.New(),
			Log:    conf.Log,
			Cache:  conf.Cache,
			writer: conf.writer,
		}
		langc.MergeConfigMap(conf.AllSettings())
		for _, ignore := range conf.GetStringSlice("languages." + lang + ".ignores") {
			langc.Set(ignore, make(map[string]interface{}))
		}
		langc.MergeConfigMap(conf.GetStringMap("languages." + lang))

		langc.Site = Site{
			URL:      langc.GetString("site.url"),
			Title:    langc.GetString("site.title"),
			SubTitle: langc.GetString("site.subtitle"),
			Language: lang,
		}
		langc.OutputDir = langc.GetString("output_dir")
		langc.ContentDir = langc.GetString("content_dir")

		langc.Languages = conf.Languages
		langc.DefaultLanguage = conf.DefaultLanguage

		conf.Languages[lang] = langc
	}
	conf.Languages[conf.DefaultLanguage] = conf
}

var (
	sectionConfig = map[string]interface{}{
		"sections._default.path":          "{section:slug}/index.html",
		"sections._default.orderby":       "weight",
		"sections._default.template":      "section.html",
		"sections._default.paginate":      10,
		"sections._default.paginate_path": "{name}{number}{extension}",
		"sections._default.page_path":     "{section:slug}/{slug}/index.html",
		"sections._default.page_orderby":  "date desc",
		"sections._default.page_template": "page.html",
	}
	taxonomyConfig = map[string]interface{}{
		"taxonomies._default.path":               "{taxonomy}/index.html",
		"taxonomies._default.orderby":            "name",
		"taxonomies._default.term_path":          "{taxonomy}/{term:slug}/index.html",
		"taxonomies._default.term_paginate_path": "{name}{number}{extension}",
		"taxonomies._default.term_orderby":       "date desc",

		"taxonomies.categories.weight": 1,
		"taxonomies.tags.weight":       2,
		"taxonomies.authors.weight":    3,
	}
	staticConfig = map[string]interface{}{
		"statics.@theme/_internal/static.path": "static",
		"statics.@theme/static.path":           "static",
		"statics.static.path":                  "static",
	}
	// 默认不需要修改的配置
	otherConfig = map[string]interface{}{
		"content_truncate_len":      49,
		"content_truncate_ellipsis": "...",
		"content_highlight_style":   "monokai",
		"slugify":                   true,
		"formats.rss.template":      "_internal/partials/rss.xml",
		"formats.atom.template":     "_internal/partials/atom.xml",
	}
	// 默认需要修改的配置
	siteConfig = map[string]interface{}{
		"site.url":       "http://127.0.0.1:8000",
		"site.title":     "snow",
		"site.subtitle":  "snow is a static site generator.",
		"site.author":    "snow",
		"site.language":  "en",
		"theme.config":   "theme.yaml",
		"theme.override": "layouts",
		"output_dir":     "output",
		"content_dir":    "content",
	}
)

func DefaultConfig() Config {
	c := Config{
		Log: &logrus.Logger{
			Out: os.Stderr,
			Formatter: &logrus.TextFormatter{
				DisableTimestamp: true,
				FullTimestamp:    false,
			},
			Level: logrus.InfoLevel,
		},
		Viper: viper.New(),
		Cache: new(sync.Map),
	}
	for k, v := range siteConfig {
		c.SetDefault(k, v)
	}
	return c
}
