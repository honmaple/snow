package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/config"
)

type (
	Hook interface {
		Name() string
		page.Hook
		static.Hook
	}
	Hooks       []Hook
	hookCreator func(*config.Config) Hook
)

var (
	_hooks = make(map[string]hookCreator)
)

func (hooks Hooks) PageHooks() (result []page.Hook) {
	for _, hook := range hooks {
		result = append(result, hook)
	}
	return
}

func (hooks Hooks) StaticHooks() (result []static.Hook) {
	for _, hook := range hooks {
		result = append(result, hook)
	}
	return
}

func New(conf *config.Config) Hooks {
	names := conf.GetStringSlice("hooks")

	hooks := make([]Hook, 0)
	for _, name := range names {
		if hook, ok := _hooks[name]; ok {
			hooks = append(hooks, hook(conf))
		}
	}
	return hooks
}

func Register(name string, creator hookCreator) {
	_hooks[name] = creator
}
