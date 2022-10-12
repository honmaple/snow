package pongo2

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2/v4"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Template struct {
	conf    *config.Config
	output  string
	context map[string]interface{}
}

func (t *Template) Lookup(names ...string) string {
	override := t.conf.GetString("theme.override")
	if override != "" {
		for _, name := range names {
			file := filepath.Join(override, name)
			if utils.FileExists(file) {
				return file
			}
		}
	}

	templates := filepath.Join(t.conf.GetString("theme.path"), "templates")
	for _, name := range names {
		file := filepath.Join(templates, name)
		if utils.FileExists(file) {
			return file
		}
	}
	return ""
}

func (t *Template) Write(tmpl string, file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	writefile := filepath.Join(t.output, file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	tpl := pongo2.Must(pongo2.FromFile(tmpl))
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
	fmt.Println("Writing ", writefile)
	return tpl.ExecuteWriter(c, f)
}

func New(conf *config.Config) *Template {
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
	pongo2.RegisterFilter("absURL", t.absURL)
	pongo2.RegisterFilter("relURL", t.relURL)
	pongo2.RegisterFilter("timesince", t.timeSince)
	return t
}
