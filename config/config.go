package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
	Site Site
}

type Site struct {
	URL         string `yaml:"url"`
	Title       string `yaml:"title"`
	SubTitle    string `yaml:"subtitle"`
	Keywords    string `yaml:"keywords"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Email       string `yaml:"email"`
	Relative    bool   `yaml:"relative"`
	Language    string `yaml:"language"`
}

type Dir struct {
	Root       string   `yaml:"root"`
	Output     string   `yaml:"output"`
	PageDirs   []string `yaml:"page_dirs"`
	ExtraDirs  []string `yaml:"extra_dirs"`
	StaticDirs []string `yaml:"static_dirs"`
}

var defaultConfig = map[string]interface{}{
	"site.baseURL":  "http://127.0.0.1:8080",
	"site.title":    "snow",
	"site.subtitle": "snow is a static site generator.",
	"output":        "output",
	"page_dirs":     []string{"content"},
	"extra_dirs":    []string{"extra"},
	"static_dirs":   []string{"static"},
}

func (c *Config) Load(path string) error {
	if utils.FileExists(path) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		c.SetConfigFile(path)
		if err := c.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
			return err
		}
	}
	return c.Viper.Unmarshal(c)
}

func DefaultConfig() *Config {
	c := &Config{
		Viper: viper.New(),
	}
	for k, v := range defaultConfig {
		c.SetDefault(k, v)
	}
	return c
}
