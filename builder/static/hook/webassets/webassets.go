package webassets

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type (
	webassets struct {
		hook.BaseHook
		conf config.Config
		opts map[string]option
	}
	option struct {
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

func (ws *webassets) Name() string {
	return "webassets"
}

func (ws *webassets) run(opt option, statics static.Statics) error {
	var b bytes.Buffer

	for _, s := range statics {
		buf, err := s.File.Bytes()
		if err != nil {
			return err
		}
		var (
			w bytes.Buffer
			r = bytes.NewBuffer(buf)
		)
		for i, filter := range opt.filters {
			w.Reset()
			if err := ws.filter(filter, &w, r, opt.filterOpts[i]); err != nil {
				return err
			}
			r = &w
		}
		b.Write(w.Bytes())
	}

	output := filepath.Join(ws.conf.GetOutput(), opt.output)

	dstFile, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, &b)
	return err
}

func (ws *webassets) BeforeStaticsWrite(statics static.Statics) static.Statics {
	for _, opt := range ws.opts {
		// 根据files找出静态文件
		files := make(static.Statics, 0)
		for _, s := range statics {
			for _, f := range opt.files {
				if s.File.Name() == f {
					files = append(files, s)
				}
			}
		}
		if err := ws.run(opt, files); err != nil {
			ws.conf.Log.Errorln("webassets err", err.Error())
		}
	}
	return statics
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	const (
		filesTemplate   = "params.webassets.%s.files"
		outputTemplate  = "params.webassets.%s.output"
		filtersTemplate = "params.webassets.%s.filters"
	)

	opts := make(map[string]option)
	meta := conf.GetStringMap("params.webassets")
	for name := range meta {
		opt := option{
			files:  conf.GetStringSlice(fmt.Sprintf(filesTemplate, name)),
			output: conf.GetString(fmt.Sprintf(outputTemplate, name)),
		}
		if len(opt.files) == 0 || opt.output == "" {
			continue
		}
		opt.filters, opt.filterOpts = filterOptions(conf, fmt.Sprintf(filtersTemplate, name))
		opts[name] = opt
	}
	return &webassets{conf: conf, opts: opts}
}
