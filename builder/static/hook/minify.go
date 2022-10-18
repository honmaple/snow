package hook

import (
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type Minify struct {
	hook.BaseHook
	conf config.Config
}

func (Minify) Name() string {
	return "minify"
}

func (Minify) BeforeStaticsWrite(statics static.Statics) static.Statics {
	return statics
}

func newMinify(conf config.Config, theme theme.Theme) hook.Hook {
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
