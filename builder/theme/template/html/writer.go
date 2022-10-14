package html

import (
	"github.com/honmaple/snow/config"
)

type Template struct {
	context map[string]interface{}
}

func (t *Template) Write(names []string, file string, context map[string]interface{}) error {
	return nil
}

func New(conf config.Config) *Template {
	return &Template{
		context: map[string]interface{}{
			"site":   conf.GetStringMap("site"),
			"params": conf.GetStringMap("params"),
		},
	}
}
