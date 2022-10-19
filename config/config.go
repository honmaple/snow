package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
	Log *logrus.Logger
}

func (conf Config) SetDebug() {
	conf.Log.SetLevel(logrus.DebugLevel)
}

func (conf Config) Reset(m map[string]interface{}) {
	keys := conf.AllKeys()
	for k, v := range m {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
	for _, k := range keys {
		conf.Set(k, conf.Get(k))
	}
}

func (conf Config) Load(path string) error {
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
	conf.Reset(defaultConfig)
	return nil
}

func (conf Config) GetOutput() string {
	return conf.GetString("output_dir")
}

func (conf Config) SetOutput(output string) {
	if output != "" {
		conf.Set("output", output)
	}
}

func (conf Config) SetMode(mode string) error {
	if mode == "" {
		return nil
	}
	key := fmt.Sprintf("mode.%s", mode)
	if !conf.IsSet(key) {
		return fmt.Errorf("mode %s not found", key)
	}
	var c *Config
	if file := conf.GetString(fmt.Sprintf("%s.include", key)); file != "" {
		c = &Config{
			Viper: viper.New(),
		}
		if err := c.Load(file); err != nil {
			return err
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
	return nil
}

var defaultConfig = map[string]interface{}{
	"site.url":                                "http://127.0.0.1:8080",
	"site.title":                              "snow",
	"site.subtitle":                           "snow is a static site generator.",
	"theme.path":                              "simple",
	"theme.engine":                            "pongo2",
	"theme.override":                          "layouts",
	"output_dir":                              "output",
	"content_dir":                             "content",
	"static_dirs":                             []string{"static"},
	"page_order":                              "date desc",
	"page_paginate":                           10,
	"page_meta.pages.lookup":                  []string{"page.html", "single.html"},
	"page_meta.pages.output":                  "pages/{slug}.html",
	"page_meta.posts.lookup":                  []string{"post.html", "single.html"},
	"page_meta.posts.output":                  "posts/{date:%Y}/{date:%m}/{slug}.html",
	"page_meta.drafts.lookup":                 []string{"draft.html", "single.html"},
	"page_meta.drafts.output":                 "drafts/{date:%Y}/{date:%m}/{slug}.html",
	"page_meta.index.list.lookup":             []string{"index.html", "list.html"},
	"page_meta.index.list.filter":             "-pages",
	"page_meta.index.list.output":             "index{number}.html",
	"page_meta.tags.lookup":                   []string{"tags.html"},
	"page_meta.tags.filter":                   "-pages",
	"page_meta.tags.groupby":                  "tag",
	"page_meta.tags.output":                   "tags/index.html",
	"page_meta.tags.list.lookup":              []string{"tag.html", "list.html"},
	"page_meta.tags.list.groupby":             "tag",
	"page_meta.tags.list.filter":              "-pages",
	"page_meta.tags.list.output":              "tags/{slug}/index{number}.html",
	"page_meta.categories.lookup":             []string{"categories.html"},
	"page_meta.categories.filter":             "-pages",
	"page_meta.categories.groupby":            "category",
	"page_meta.categories.output":             "categories/index.html",
	"page_meta.categories.list.lookup":        []string{"category.html", "list.html"},
	"page_meta.categories.list.filter":        "-pages",
	"page_meta.categories.list.groupby":       "category",
	"page_meta.categories.list.output":        "categories/{slug}/index{number}.html",
	"page_meta.authors.lookup":                []string{"authors.html"},
	"page_meta.authors.filter":                "-pages",
	"page_meta.authors.groupby":               "author",
	"page_meta.authors.output":                "authors/index.html",
	"page_meta.authors.list.lookup":           []string{"author.html", "list.html"},
	"page_meta.authors.list.filter":           "-pages",
	"page_meta.authors.list.groupby":          "author",
	"page_meta.authors.list.output":           "authors/{slug}/index{number}.html",
	"page_meta.archives.lookup":               []string{"archives.html"},
	"page_meta.archives.output":               "archives/index.html",
	"page_meta.period_archives.list.lookup":   []string{"period_archives.html"},
	"page_meta.period_archives.list.output":   "archives/{slug}/index.html",
	"page_meta.period_archives.list.filter":   "-pages",
	"page_meta.period_archives.list.groupby":  "date:2006/01",
	"page_meta.period_archives.list.paginate": 0,
}

func DefaultConfig() Config {
	c := Config{
		Viper: viper.New(),
		Log: &logrus.Logger{
			Out: os.Stderr,
			Formatter: &logrus.TextFormatter{
				DisableTimestamp: true,
				FullTimestamp:    false,
			},
			Level: logrus.InfoLevel,
		},
	}
	for k, v := range defaultConfig {
		c.SetDefault(k, v)
	}
	return c
}
