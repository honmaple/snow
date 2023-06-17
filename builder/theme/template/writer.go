package template

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"

	"path/filepath"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/config"
)

var (
	Globals        = make(map[string]interface{})
	GlobalFuncs    = make(map[string]func(map[string]interface{}) interface{})
	ConfigFuncs    = make(map[string]func(config.Config) pongo2.FilterFunction)
	RegisterTag    = pongo2.RegisterTag
	RegisterFilter = pongo2.RegisterFilter
)

func Register(k string, v interface{}) {
	Globals[k] = v
}

func RegisterFunc(k string, v func(map[string]interface{}) interface{}) {
	GlobalFuncs[k] = v
}

func RegisterConfigFilter(k string, v func(config.Config) pongo2.FilterFunction) {
	ConfigFuncs[k] = v
}

func RegisterStringFilter(k string, f func(string) interface{}) {
	pongo2.RegisterFilter(k, func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		v, ok := in.Interface().(string)
		if !ok {
			return nil, newError(k, errors.New("filter input argument must be of type 'string'"))
		}
		return pongo2.AsValue(f(v)), nil
	})
}

func Expr(expr string) (*pongo2.Template, error) {
	tpl, err := pongo2.FromString("{{" + expr + "}}")
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

type (
	Writer interface {
		Name() string
		Write(string, map[string]interface{}) error
		Execute(map[string]interface{}) (string, error)
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
		conf   config.Config
		loader *loader
		tplset *pongo2.TemplateSet
		langs  map[string]interface{}
	}
)

func (t *writer) Name() string {
	return t.n
}

func (t *writer) Write(file string, ctx map[string]interface{}) error {
	if file == "" {
		return nil
	}
	if filepath.Clean(file) != file {
		return fmt.Errorf("The path '%s' is not valid", file)
	}
	vars := make(map[string]interface{})
	for k, v := range ctx {
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

	var w bytes.Buffer
	if err := t.w.ExecuteWriter(vars, &w); err != nil {
		return err
	}
	// t.t.conf.Log.Debugln("Writing", writefile)
	return t.t.conf.Write(file, &w)
}

func (t *writer) Execute(ctx map[string]interface{}) (string, error) {
	vars := make(map[string]interface{})
	for k, v := range ctx {
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
	return t.w.Execute(vars)
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

func (t *template) newConfig(ctx map[string]interface{}) interface{} {
	lang := ctx["current_lang"]
	if lang != nil && lang != t.conf.Site.Language {
		if v, ok := t.langs[lang.(string)]; ok {
			return v
		}
	}
	return t.langs[t.conf.Site.Language]
}

func New(conf config.Config, theme fs.FS) Interface {
	t := &template{
		conf:   conf,
		loader: newLoader(theme, conf.GetString("theme.override")),
	}
	t.tplset = pongo2.NewSet("app", t.loader)

	t.langs = make(map[string]interface{})
	for lang := range conf.Languages {
		t.langs[lang] = t.conf.With(lang).AllSettings()
	}
	t.langs[conf.Site.Language] = conf.AllSettings()

	Register("dict", dict)
	Register("slice", slice)

	RegisterFunc("config", t.newConfig)
	RegisterFunc("scratch", newScratch)
	RegisterFunc("newScratch", newScratchFunc)

	RegisterFilter("absURL", t.absURL)
	RegisterFilter("relURL", t.relURL)
	RegisterFilter("slient", t.slient)
	RegisterFilter("jsonify", t.jsonify)

	for k, f := range ConfigFuncs {
		RegisterFilter(k, f(conf))
	}
	return t
}
