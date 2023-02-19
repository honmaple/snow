package assets

import (
	"fmt"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	assets struct {
		hook.BaseHook
		conf  config.Config
		opts  map[string]option
		theme theme.Theme
	}
	option struct {
		files      []string
		output     string
		filters    []string
		filterOpts []filterOption
	}
)

const (
	filesTemplate   = "params.assets.%s.files"
	outputTemplate  = "params.assets.%s.output"
	filtersTemplate = "params.assets.%s.filters"
)

func (self *assets) Name() string {
	return "assets"
}

func (self *assets) BeforeStaticsWrite(statics static.Statics) static.Statics {
	for _, opt := range self.opts {
		if err := self.execute(opt); err != nil {
			self.conf.Log.Errorln("hook assets:", err.Error())
		}
	}
	return statics
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	opts := make(map[string]option)
	meta := conf.GetStringMap("params.assets")
	for name := range meta {
		opt := option{
			files:  conf.GetStringSlice(fmt.Sprintf(filesTemplate, name)),
			output: conf.GetString(fmt.Sprintf(outputTemplate, name)),
		}
		if len(opt.files) == 0 || opt.output == "" {
			continue
		}
		opt.filters, opt.filterOpts = filterOptions(conf.Get(fmt.Sprintf(filtersTemplate, name)))
		opts[name] = opt
	}
	h := &assets{conf: conf, opts: opts, theme: theme}
	pongo2.RegisterTag("assets", pongo2.TagParser(h.assetParser))
	return h
}

func init() {
	hook.Register("assets", New)
}
