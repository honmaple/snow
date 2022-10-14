package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/honmaple/snow/utils"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
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
	"site.url":       "http://127.0.0.1:8080",
	"site.title":     "snow",
	"site.subtitle":  "snow is a static site generator.",
	"theme.path":     "simple",
	"theme.engine":   "pongo2",
	"theme.override": "layouts",
	"output_dir":     "output",
	"content_dir":    "content",
}

func DefaultConfig() Config {
	c := Config{
		Viper: viper.New(),
	}
	for k, v := range defaultConfig {
		c.SetDefault(k, v)
	}
	return c
}
