package template

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

var (
	Globals        = make(map[string]interface{})
	GlobalFuncs    = make(map[string]func(map[string]interface{}) interface{})
	RegisterTag    = pongo2.RegisterTag
	RegisterFilter = pongo2.RegisterFilter
)

func Register(k string, v interface{}) {
	Globals[k] = v
}

func RegisterFunc(k string, v func(map[string]interface{}) interface{}) {
	GlobalFuncs[k] = v
}

type (
	Writer interface {
		Name() string
		Write(string, map[string]interface{}) error
	}
	Interface interface {
		Lookup(string) (Writer, error)
	}
)

type (
	writer struct {
		n string
		t *template
		w *pongo2.Template
	}
	template struct {
		conf    config.Config
		output  string
		context map[string]interface{}
		loader  *loader
		tplset  *pongo2.TemplateSet
	}
)

func (t *writer) Name() string {
	return t.n
}

func (t *writer) Write(file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	writefile := filepath.Join(t.t.output, file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	vars := make(map[string]interface{})
	for k, v := range t.t.context {
		vars[k] = v
	}
	for k, v := range context {
		vars[k] = v
	}
	for k, v := range Globals {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}
	for k, v := range GlobalFuncs {
		if _, ok := vars[k]; !ok {
			vars[k] = v(vars)
		}
	}
	t.t.conf.Log.Debugln("Writing", writefile)
	// _ = tpl
	// return nil
	return t.w.ExecuteWriter(vars, f)
}

func (t *template) Lookup(name string) (Writer, error) {
	buf, err := t.loader.GetBytes(name)
	if err != nil {
		return nil, err
	}
	tpl, err := t.tplset.FromBytes(buf)
	if err != nil {
		t.conf.Log.Errorf("%s: %s", name, err.Error())
		return nil, err
	}
	return &writer{n: name, t: t, w: tpl}, nil
}

func New(conf config.Config, theme fs.FS) Interface {
	t := &template{
		conf:   conf,
		output: conf.GetOutput(),
		context: map[string]interface{}{
			"site":   conf.GetStringMap("site"),
			"params": conf.GetStringMap("params"),
			"config": conf.AllSettings(),
		},
		loader: newLoader(theme, conf.GetString("theme.override")),
	}
	t.tplset = pongo2.NewSet("app", t.loader)

	pongo2.RegisterFilter("absURL", t.absURL)
	pongo2.RegisterFilter("relURL", t.relURL)
	pongo2.RegisterFilter("timesince", t.timeSince)
	return t
}
