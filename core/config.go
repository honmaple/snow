/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: config.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 14:40:41 (CST)
 Last Update: Saturday 2019-09-07 17:07:06 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"github.com/BurntSushi/toml"
)

// VERSION ..
const VERSION = "0.1.0"

type Configuration struct {
	Log              *Log      `toml:"log"`
	Site             *Site     `toml:"site"`
	Feed             *Feed     `toml:"feed"`
	Theme            *Theme    `toml:"theme"`
	Plugin           *Plugin   `toml:"plugin"`
	Statics          []*Static `toml:"static"`
	Output           string    `toml:"output"`
	Lang             string    `toml:"lang"`
	Locale           string    `toml:"locale"`
	Orderby          []string  `toml:"orderby"`
	Dir              string    `toml:dir`
	PageDir          []string  `toml:"pagedir"`
	ArticleDir       []string  `toml:"articledir"`
	UseDirAsCategory bool      `toml:"usedirascategory"`
}

// Site ..
type Site struct {
	URL      string `toml:"url"`
	Title    string `toml:"title"`
	SubTitle string `toml:"subtitle"`
	Relative bool   `toml:"relative"`
	Author   string `toml:"author"`
}

// Static ..
type Static struct {
	File     string `toml:"file"`
	SavePath string `toml:"save_path"`
}

// Theme ..
type Theme struct {
	Include    includeThemeFile `toml:"include"`
	Path       string           `toml:"path"`
	Pagination int64            `toml:"pagination"`
	Template   *Template        `toml:"template"`
}

// Plugin ..
type Plugin struct {
	Include includePluginFile `toml:"include"`
	Path    string            `toml:"path"`
}

// includeThemeFile ..
type includeThemeFile string

// includePluginFile ..
type includePluginFile string

func (d *includeThemeFile) UnmarshalText(text []byte) error {
	var themeconf Theme
	if _, err := toml.DecodeFile(string(text), &themeconf); err != nil {
		return err
	}
	conf.Theme = &themeconf
	conf.Theme.Include = includeThemeFile(text)
	return nil
}

func (d *includePluginFile) UnmarshalText(text []byte) error {
	var pluginconf Plugin
	if _, err := toml.DecodeFile(string(text), &pluginconf); err != nil {
		return err
	}
	conf.Plugin = &pluginconf
	conf.Plugin.Include = includePluginFile(text)
	return nil
}

var conf = &Configuration{
	Lang:             "en",
	Locale:           "en_US.UTF-8",
	Dir:              "content",
	Output:           "output",
	Orderby:          []string{"date"},
	ArticleDir:       []string{"markdown"},
	PageDir:          []string{"page"},
	UseDirAsCategory: false,

	Log: &Log{
		Level:     "debug",
		Timestamp: false,
	},
	Site: &Site{
		URL:      "https://honmaple.me",
		Title:    "honmaple's blog",
		SubTitle: "",
	},
	Feed: &Feed{
		Limit:    10,
		Title:    "{title}",
		Format:   "atom",
		SavePath: "feeds/atom.xml",
		Category: &Feed{
			Limit:    10,
			Title:    "{slug} - {title}",
			Format:   "atom",
			SavePath: "feeds/{slug}.atom.xml",
		},
		Tag: &Feed{
			Limit:    10,
			Title:    "{slug} - {title}",
			Format:   "atom",
			SavePath: "",
		},
		Author: &Feed{
			Limit:    10,
			Title:    "{slug} - {title}",
			Format:   "atom",
			SavePath: "",
		},
	},
	Theme: &Theme{
		Path:       "theme/simple",
		Pagination: 10,
	},
	Plugin: &Plugin{
		Path: "plugin",
	},
}

// ParseConf ..
func Parse(confpath string) *Configuration {
	if _, err := toml.DecodeFile(confpath, conf); err != nil {
		Exit(err.Error(), 1)
	}
	keys := []string{"Index", "Article", "Page", "Tag", "Author", "Category", "Tags", "Authors", "Categories", "Archives"}
	for _, key := range keys {
		if value, err := GetField(template, key); err != nil {
			Exit(err.Error(), 1)
		} else {
			SetField(conf.Theme.Template, key, value)
		}
	}
	return conf
}
