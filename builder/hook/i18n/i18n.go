package i18n

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"gopkg.in/yaml.v3"
)

type (
	i18n struct {
		hook.BaseHook
		conf  config.Config
		lang  string
		trans map[string]map[string]tran
	}
	tran struct {
		Id    string `yaml:"id"    json:"id"`
		Tr    string `yaml:"tr"    json:"tr"`
		Lower bool   `yaml:"lower" json:"lower"`
	}
)

func (self *i18n) Name() string {
	return "i18n"
}

func (self *i18n) Tran(id string, lang string) string {
	if lang == "" {
		lang = self.lang
	}
	tr, ok := self.trans[lang]
	if !ok {
		return id
	}
	t, ok := tr[id]
	if !ok {
		return id
	}
	return t.Tr
}

func getTrans(conf config.Config, theme theme.Theme) map[string]map[string]tran {
	trans := make(map[string]map[string]tran)

	// 先加载主题目录下的翻译文件
	files, _ := fs.Glob(theme, "i18n/*")
	for _, file := range files {
		stat, err := fs.Stat(theme, file)
		if err != nil || stat.IsDir() {
			continue
		}
		buf, err := fs.ReadFile(theme, file)
		if err != nil {
			continue
		}
		ts := make([]tran, 0)
		if err := yaml.Unmarshal(buf, &ts); err != nil {
			continue
		}
		tr := make(map[string]tran)
		for _, t := range ts {
			tr[t.Id] = t

			if t.Lower {
				tr[strings.ToLower(t.Id)] = t
			}
		}

		name := filepath.Base(stat.Name())
		lang := name[:len(name)-len(filepath.Ext(name))]
		trans[lang] = tr
	}

	languages := conf.GetStringMap("languages")
	for lang := range languages {
		k := "languages." + lang + ".translations"

		ts := make([]tran, 0)
		switch reflect.ValueOf(conf.Get(k)).Kind() {
		case reflect.String:
			file := conf.GetString(k)
			if file == "" {
				continue
			}
			buf, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			if err := yaml.Unmarshal(buf, &ts); err != nil {
				continue
			}
		case reflect.Slice:
			if err := conf.Unmarshal(k, &ts); err != nil {
				continue
			}
		}
		tr, ok := trans[lang]
		if !ok {
			tr = make(map[string]tran)
		}
		for _, t := range ts {
			tr[t.Id] = t

			if t.Lower {
				tr[strings.ToLower(t.Id)] = t
			}
		}
		trans[lang] = tr
	}
	return trans
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	e := &i18n{
		conf:  conf,
		lang:  conf.DefaultLanguage,
		trans: getTrans(conf, theme),
	}
	e.register()
	return e
}

func init() {
	hook.Register("i18n", New)
}
