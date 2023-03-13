package sofile

import (
	"path/filepath"
	"plugin"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	hookFunc    = func(config.Config, theme.Theme) hook.Hook
	pageHooks   = page.Hooks
	staticHooks = static.Hooks
	sofile      struct {
		pageHooks
		staticHooks
		conf config.Config
	}
)

func (self *sofile) Name() string {
	return "sofile"
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	hooks := make(hook.Hooks, 0)
	for _, file := range conf.GetStringSlice("params.sofiles") {
		if filepath.Ext(file) != ".so" {
			continue
		}
		p, err := plugin.Open(file)
		if err != nil {
			conf.Log.Fatalln(file, err.Error())
		}
		v, err := p.Lookup("NewHook")
		if err != nil {
			conf.Log.Fatalln(file, err.Error())
		}
		hooks = append(hooks, v.(hookFunc)(conf, theme))
	}
	return &sofile{
		conf:        conf,
		pageHooks:   hooks.PageHooks(),
		staticHooks: hooks.StaticHooks(),
	}
}

func init() {
	hook.Register("sofile", New)
}
