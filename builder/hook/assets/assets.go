package assets

import (
	"fmt"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
)

type (
	assets struct {
		hook.BaseHook
		conf  config.Config
		opts  map[string]option
		hash  map[string]string
		theme theme.Theme
	}
	option struct {
		files      []string
		output     string
		version    bool
		filters    []string
		filterOpts []filterOption
	}
)

const (
	filesTemplate   = "hooks.assets.%s.files"
	outputTemplate  = "hooks.assets.%s.output"
	filtersTemplate = "hooks.assets.%s.filters"
	versionTemplate = "hooks.assets.%s.version"
)

func (self *assets) Name() string {
	return "assets"
}

func (self *assets) Statics(statics static.Statics) static.Statics {
	for name, opt := range self.opts {
		h, err := self.execute(opt)
		if err != nil {
			self.conf.Log.Errorln("hook assets:", err.Error())
		} else {
			self.hash[name] = h
		}
	}
	return statics
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	opts := make(map[string]option)
	meta := conf.GetStringMap("hooks.assets")
	for name := range meta {
		opt := option{
			files:   conf.GetStringSlice(fmt.Sprintf(filesTemplate, name)),
			output:  conf.GetString(fmt.Sprintf(outputTemplate, name)),
			version: conf.GetBool(fmt.Sprintf(versionTemplate, name)),
		}
		if len(opt.files) == 0 || opt.output == "" {
			continue
		}
		opt.filters, opt.filterOpts = filterOptions(conf.Get(fmt.Sprintf(filtersTemplate, name)))
		opts[name] = opt
	}
	h := &assets{conf: conf, opts: opts, hash: make(map[string]string), theme: theme}
	template.RegisterTag("assets", h.assetParser)
	return h
}

func init() {
	hook.Register("assets", New)
}
