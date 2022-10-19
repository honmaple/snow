package pongo2

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type Template struct {
	conf    config.Config
	output  string
	context map[string]interface{}
	tplset  *pongo2.TemplateSet
	loader  *loader
	cache   sync.Map
}

func (t *Template) lookup(names ...string) *pongo2.Template {
	for _, name := range names {
		v, ok := t.cache.Load(name)
		if ok {
			return v.(*pongo2.Template)
		}
		buf, err := t.loader.GetBytes(name)
		if err != nil {
			continue
		}
		// 模版未找到不输出日志, 编译模版有问题才输出
		tpl, err := t.tplset.FromBytes(buf)
		if err != nil {
			t.conf.Log.Errorln("Parse template:", err.Error())
			return nil
		}
		t.cache.Store(name, tpl)
		return tpl
	}
	return nil
}

func (t *Template) Write(names []string, file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	tpl := t.lookup(names...)
	if tpl == nil {
		return nil
	}

	writefile := filepath.Join(t.output, file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := make(map[string]interface{})
	for k, v := range t.context {
		c[k] = v
	}
	for k, v := range context {
		c[k] = v
	}
	t.conf.Log.Debugln("Writing", writefile)
	// _ = tpl
	// return nil
	return tpl.ExecuteWriter(c, f)
}

func New(conf config.Config, theme fs.FS) *Template {
	t := &Template{
		conf:   conf,
		output: conf.GetOutput(),
		context: map[string]interface{}{
			"site":   conf.GetStringMap("site"),
			"params": conf.GetStringMap("params"),
		},
	}
	t.loader = newLoader(theme, conf.GetString("theme.override"))
	t.tplset = pongo2.NewSet("app", t.loader)

	pongo2.RegisterFilter("absURL", t.absURL)
	pongo2.RegisterFilter("relURL", t.relURL)
	pongo2.RegisterFilter("timesince", t.timeSince)
	return t
}
