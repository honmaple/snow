package template

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/builder/template/html"
	"github.com/honmaple/snow/builder/template/pongo2"
	"github.com/honmaple/snow/config"
)

type Template interface {
	Write([]string, string, map[string]interface{}) error
}

func New(conf *config.Config, themes fs.FS) (Template, error) {
	var (
		err   error
		tmpl  Template
		theme fs.FS
	)

	name := conf.GetString("theme.path")
	switch name {
	case "simple":
		theme, err = fs.Sub(themes, filepath.Join("themes", name, "templates"))
	default:
		theme, err = fs.Sub(os.DirFS(name), "templates")
	}
	if err != nil {
		return nil, err
	}

	engine := conf.GetString("theme.engine")
	switch engine {
	case "pongo2":
		tmpl = pongo2.New(conf, theme)
	default:
		tmpl = html.New(conf)
	}
	return tmpl, nil
}
