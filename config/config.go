package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/mitchellh/mapstructure"
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

func (conf Config) Unmarshal(key string, val interface{}) error {
	if conf.IsSet(key) {
		return conf.UnmarshalKey(key, val, func(decoderConfig *mapstructure.DecoderConfig) {
			decoderConfig.TagName = "json"
		})
	}
	return nil
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
	conf.Reset(siteConfig)
	conf.Reset(sectionConfig)
	conf.Reset(taxonomyConfig)
	return nil
}

func (conf Config) GetRelURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		return utils.StringConcat("/", path)
	}
	return path
}

func (conf Config) GetURL(path string) string {
	return utils.StringConcat(conf.GetString("site.url"), conf.GetRelURL(path))
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

var (
	sectionConfig = map[string]interface{}{
		"sections._default.path":          "{section}/index{number}.html",
		"sections._default.orderby":       "date desc",
		"sections._default.paginate":      10,
		"sections._default.template":      "posts.html",
		"sections._default.page_path":     "{section}/{slug}/index.html",
		"sections._default.page_template": "page.html",

		"sections._index.path":     "index{number}.html",
		"sections._index.paginate": 10,
		"sections._index.template": "index.html",
		"sections._index.custom":   true,
	}
	taxonomyConfig = map[string]interface{}{
		"taxonomies._default.path":          "{taxonomy}/index.html",
		"taxonomies._default.template":      "{taxonomy}/list.html",
		"taxonomies._default.term_path":     "{taxonomy}/{slug}/index.html",
		"taxonomies._default.term_template": "{taxonomy}/single.html",

		"taxonomies.categories.weight": 1,
		"taxonomies.tags.weight":       2,
		"taxonomies.authors.weight":    3,
	}
	siteConfig = map[string]interface{}{
		"site.url":       "http://127.0.0.1:8000",
		"site.title":     "snow",
		"site.subtitle":  "snow is a static site generator.",
		"theme.override": "layouts",
		"output_dir":     "output",
		"content_dir":    "content",
	}
)

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
	for k, v := range siteConfig {
		c.SetDefault(k, v)
	}
	return c
}
