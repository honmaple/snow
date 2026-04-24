package i18n

import (
	"io/fs"
	"os"
	stdpath "path"
	"reflect"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"gopkg.in/yaml.v3"
)

type (
	I18n struct {
		lang         string
		translations map[string]map[string]*Translation
	}
	Translation struct {
		Id         string `yaml:"id"          json:"id"`
		Tr         string `yaml:"tr"          json:"tr"`
		IgnoreCase bool   `yaml:"ignore_case" json:"ignore_case"`
	}
)

func (i *I18n) Translate(id string, lang string) string {
	if lang == "" {
		lang = i.lang
	}
	tr, ok := i.translations[lang]
	if !ok {
		return id
	}
	t, ok := tr[id]
	if !ok {
		return id
	}
	return t.Tr
}

func (i *I18n) LoadTranslations(ctx *core.Context) error {
	trans := make(map[string]map[string]*Translation)

	// 加载主题下以及当前目录下的的翻译文件
	for _, dir := range []fs.FS{ctx.Theme, os.DirFS(".")} {
		if _, err := fs.Stat(dir, "i18n"); err != nil {
			continue
		}
		entries, err := fs.ReadDir(dir, "i18n")
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := stdpath.Base(entry.Name())
			lang := name[:len(name)-len(stdpath.Ext(name))]
			if _, ok := trans[lang]; !ok {
				trans[lang] = make(map[string]*Translation)
			}

			buf, err := fs.ReadFile(dir, stdpath.Join("i18n", entry.Name()))
			if err != nil {
				return err
			}
			ts := make([]*Translation, 0)
			if err := yaml.Unmarshal(buf, &ts); err != nil {
				return err
			}
			for _, t := range ts {
				trans[lang][t.Id] = t

				if t.IgnoreCase {
					trans[lang][strings.ToLower(t.Id)] = t
				}
			}
		}
	}

	// languages.en:
	//   translations: "en.yaml"
	//   translations:
	//     - id: "tag"
	//       tr: "tag"
	//     - id: "author"
	//       tr: "author"
	languages := ctx.Config.GetStringMap("languages")
	for lang := range languages {
		k := "languages." + lang + ".translations"

		ts := make([]*Translation, 0)
		switch reflect.ValueOf(ctx.Config.Get(k)).Kind() {
		case reflect.String:
			file := ctx.Config.GetString(k)
			if file == "" {
				continue
			}
			buf, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			if err := yaml.Unmarshal(buf, &ts); err != nil {
				return err
			}
		case reflect.Slice:
			for _, c := range ctx.Config.GetSubSlice(k) {
				ts = append(ts, &Translation{
					Id:         c.GetString("id"),
					Tr:         c.GetString("tr"),
					IgnoreCase: c.GetBool("ignore_case"),
				})
			}
		}
		if _, ok := trans[lang]; !ok {
			trans[lang] = make(map[string]*Translation)
		}
		for _, t := range ts {
			trans[lang][t.Id] = t

			if t.IgnoreCase {
				trans[lang][strings.ToLower(t.Id)] = t
			}
		}
	}
	i.translations = trans
	return nil
}

func New(ctx *core.Context) (*I18n, error) {
	i18n := &I18n{}
	if err := i18n.LoadTranslations(ctx); err != nil {
		return nil, &core.Error{
			Op:   "load i18n translations",
			Path: "i18n",
			Err:  err,
		}
	}
	return i18n, nil
}
