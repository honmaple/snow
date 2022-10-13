package pongo2

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Template struct {
	conf    *config.Config
	theme   *pongo2.TemplateSet
	output  string
	context map[string]interface{}
}

func (t *Template) lookup(names ...string) *pongo2.Template {
	override := t.conf.GetString("theme.override")
	if override != "" {
		for _, name := range names {
			file := filepath.Join(override, name)
			if !utils.FileExists(file) {
				continue
			}
			tpl, err := pongo2.FromFile(file)
			if err == nil {
				return tpl
			}
		}
	}
	for _, name := range names {
		tpl, err := t.theme.FromCache(name)
		if err == nil {
			return tpl
		}
	}
	return nil
}

func (t *Template) Write(names []string, file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	tpl := t.lookup(names...)
	if tpl == nil {
		return nil
	}

	writefile := filepath.Join(t.output, file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := make(map[string]interface{})
	for k, v := range t.context {
		c[k] = v
	}
	for k, v := range context {
		c[k] = v
	}
	fmt.Println("Writing", writefile)
	_ = tpl
	return nil
	// return tpl.ExecuteWriter(c, f)
}

func New(conf *config.Config, theme fs.FS) *Template {
	output := conf.GetString("output_dir")
	if output == "" {
		output = "output"
	}
	t := &Template{
		conf:   conf,
		output: output,
		context: map[string]interface{}{
			"site":   conf.GetStringMap("site"),
			"params": conf.GetStringMap("params"),
		},
	}
	t.theme = pongo2.NewSet("app", pongo2.NewFSLoader(theme))

	// https://github.com/flosch/pongo2/issues/313
	// fsLoader := pongo2.NewFSLoader(theme)
	// override := t.conf.GetString("theme.override")
	// if override != "" && utils.FileExists(override) {
	//	loader := pongo2.MustNewLocalFileSystemLoader(override)
	//	set := pongo2.NewSet("app", loader, fsLoader)
	//	fmt.Println(set.FromCache("post.html"))
	// }

	pongo2.RegisterFilter("absURL", t.absURL)
	pongo2.RegisterFilter("relURL", t.relURL)
	pongo2.RegisterFilter("timesince", t.timeSince)
	return t
}
