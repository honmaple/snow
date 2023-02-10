package assets

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type (
	assets struct {
		hook.BaseHook
		conf config.Config
		opts map[string]option
	}
	option struct {
		retain     bool
		files      []string
		output     string
		filters    []string
		filterOpts []filterOption
	}
	filterOption = *viper.Viper
)

func filterOptions(conf config.Config, key string) ([]string, []filterOption) {
	data := conf.Get(key)
	if data == nil {
		return nil, nil
	}
	if reflect.TypeOf(data).Kind() == reflect.Slice {
		names := make([]string, 0)
		opts := make([]filterOption, 0)
		for _, item := range data.([]interface{}) {
			for k, v := range cast.ToStringMap(item) {
				subv := viper.New()
				subv.Set(k, v)
				opts = append(opts, subv)
				names = append(names, k)
				break
			}
		}
		return names, opts
	}
	return nil, nil
}

func (ws *assets) Name() string {
	return "assets"
}

func (ws *assets) execute(opt option, statics static.Statics, exclude map[string]bool) error {
	files := make(map[string]bool)
	for _, file := range opt.files {
		files[file] = true
	}

	var b bytes.Buffer
	for _, static := range statics {
		if !files[static.File.Name()] {
			continue
		}
		// 原始文件不继续写入到output目录
		if !opt.retain {
			exclude[static.File.Name()] = true
		}
		buf, err := ioutil.ReadAll(static.File)
		if err != nil {
			return err
		}
		var (
			w = bytes.NewBuffer(nil)
			r = bytes.NewBuffer(buf)
		)
		for i, filter := range opt.filters {
			w.Reset()
			if err := ws.filter(filter, w, r, opt.filterOpts[i]); err != nil {
				return err
			}
			r = w
		}
		b.Write(w.Bytes())
	}
	return ws.conf.WriteOutput(opt.output, &b)
}

func (ws *assets) BeforeStaticsWrite(statics static.Statics) static.Statics {
	excludeFiles := make(map[string]bool)
	for _, opt := range ws.opts {
		if err := ws.execute(opt, statics, excludeFiles); err != nil {
			ws.conf.Log.Errorln("assets err", err.Error())
		}
	}
	newstatics := make(static.Statics, 0)
	for _, static := range statics {
		if excludeFiles[static.File.Name()] {
			continue
		}
		newstatics = append(newstatics, static)
	}
	return newstatics
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	const (
		filesTemplate   = "params.assets.%s.files"
		outputTemplate  = "params.assets.%s.output"
		retainTemplate  = "params.assets.%s.retain"
		filtersTemplate = "params.assets.%s.filters"
	)

	opts := make(map[string]option)
	meta := conf.GetStringMap("params.assets")
	for name := range meta {
		opt := option{
			files:  conf.GetStringSlice(fmt.Sprintf(filesTemplate, name)),
			output: conf.GetString(fmt.Sprintf(outputTemplate, name)),
			retain: conf.GetBool(fmt.Sprintf(retainTemplate, name)),
		}
		if len(opt.files) == 0 || opt.output == "" {
			continue
		}
		opt.filters, opt.filterOpts = filterOptions(conf, fmt.Sprintf(filtersTemplate, name))
		opts[name] = opt
	}
	return &assets{conf: conf, opts: opts}
}

func init() {
	hook.Register("assets", New)
}
