package hook

import (
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/config"
)

type Minify struct {
	hook.BaseHook
	conf config.Config
}

func (Minify) Name() string {
	return "minify"
}

func newMinify(conf config.Config) hook.Hook {
	defaultConfig := map[string]interface{}{
		"params.minify.types": []string{"css", "js", "html"},
	}
	for k, v := range defaultConfig {
		if conf.IsSet(k) {
			continue
		}
		conf.Set(k, v)
	}
	return &Minify{conf: conf}
}
