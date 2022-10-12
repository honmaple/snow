package template

import (
	"github.com/honmaple/snow/builder/template/html"
	"github.com/honmaple/snow/builder/template/pongo2"
	"github.com/honmaple/snow/config"
)

type Template interface {
	Write(string, string, map[string]interface{}) error
	Lookup(...string) string
}

func New(conf *config.Config) Template {
	var tmpl Template

	engine := conf.GetString("theme.engine")
	switch engine {
	case "pongo2":
		tmpl = pongo2.New(conf)
	default:
		tmpl = html.New(conf)
	}
	return tmpl
}
