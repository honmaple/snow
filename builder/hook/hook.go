package hook

import (
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	pageHooks   = page.Hooks
	staticHooks = static.Hooks
	BaseHook    struct {
		pageHooks
		staticHooks
	}

	Hook interface {
		Name() string
		page.Hook
		static.Hook
	}
	Hooks       []Hook
	hookCreator func(config.Config, theme.Theme) Hook
)

var (
	_hooks map[string]hookCreator
)

func (hooks Hooks) PageHooks() (result page.Hooks) {
	for _, hook := range hooks {
		result = append(result, hook)
	}
	return
}

func (hooks Hooks) StaticHooks() (result static.Hooks) {
	for _, hook := range hooks {
		result = append(result, hook)
	}
	return
}

func New(conf config.Config, theme theme.Theme) Hooks {
	names := conf.GetStringSlice("registered_hooks")
	if len(names) == 0 {
		names = make([]string, 0)
		for name := range conf.GetStringMap("hooks") {
			names = append(names, name)
		}
		sort.SliceStable(names, func(i, j int) bool {
			wi := conf.GetInt("hooks." + names[i] + ".weight")
			wj := conf.GetInt("hooks." + names[j] + ".weight")
			if wi == wj {
				return names[i] > names[j]
			}
			return wi > wj
		})
	}

	hooks := make([]Hook, 0)
	for _, name := range names {
		if creator, ok := _hooks[name]; ok {
			hooks = append(hooks, creator(conf, theme))
		} else {
			conf.Log.Warnf("The hook %s not found", name)
		}
	}
	return hooks
}

func Print() {
	names := make([]string, 0)
	for name := range _hooks {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println(strings.Join(names, ", "))
}

func Register(name string, creator hookCreator) {
	if _, ok := _hooks[name]; ok {
		panic(fmt.Sprintf("The hook %s has been registered", name))
	}
	_hooks[name] = creator
}

func init() {
	_hooks = make(map[string]hookCreator)
}
