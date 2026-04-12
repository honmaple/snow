package template

import (
	"io/fs"
	"maps"
	"os"
	"sync"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
)

type (
	Template interface {
		Name() string
		Execute(*core.Context, map[string]any) (string, error)
		ExecuteRaw(*core.Context, map[string]any) (string, error)
	}
	templateImpl struct {
		n   string
		tpl *pongo2.Template
	}
)

func (t *templateImpl) Name() string {
	return t.n
}

func (t *templateImpl) Execute(ctx *core.Context, vars map[string]any) (string, error) {
	nvars := make(map[string]any)
	maps.Copy(nvars, vars)

	for k, v := range TransientVars {
		if _, ok := nvars[k]; !ok {
			nvars[k] = v
		}
	}

	for k, f := range TransientFuncs {
		if _, ok := nvars[k]; !ok {
			nvars[k] = f(ctx, vars)
		}
	}
	return t.tpl.Execute(nvars)
}

func (t *templateImpl) ExecuteRaw(ctx *core.Context, vars map[string]any) (string, error) {
	nvars := make(map[string]any)
	maps.Copy(nvars, vars)

	return t.tpl.Execute(nvars)
}

type (
	TemplateSet interface {
		Lookup(...string) Template
		FromFile(string) (Template, error)
		FromBytes(string, []byte) (Template, error)
		FromString(string, string) (Template, error)
	}
	templateSet struct {
		cache  sync.Map
		tplset *pongo2.TemplateSet
	}
)

func (set *templateSet) Lookup(names ...string) Template {
	for _, name := range names {
		if name == "" {
			continue
		}
		v, ok := set.cache.Load(name)
		if ok {
			return v.(Template)
		}
		// 模版未找到不输出日志, 编译模版有问题才输出
		template, err := set.FromFile(name)
		if err != nil {
			continue
		}
		set.cache.Store(name, template)
		return template
	}
	return nil
}

func (set *templateSet) FromFile(name string) (Template, error) {
	tpl, err := set.tplset.FromFile(name)
	if err != nil {
		return nil, err
	}
	return &templateImpl{n: name, tpl: tpl}, nil
}

func (set *templateSet) FromBytes(name string, b []byte) (Template, error) {
	tpl, err := set.tplset.FromBytes(b)
	if err != nil {
		return nil, err
	}
	return &templateImpl{n: name, tpl: tpl}, nil
}

func (set *templateSet) FromString(name string, b string) (Template, error) {
	tpl, err := set.tplset.FromString(b)
	if err != nil {
		return nil, err
	}
	return &templateImpl{n: name, tpl: tpl}, nil
}

func NewSet(ctx *core.Context) (TemplateSet, error) {
	tplFS, err := fs.Sub(ctx.Theme, "templates")
	if err != nil {
		return nil, err
	}

	internalFS, err := fs.Sub(ctx.Theme, "internal/templates")
	if err != nil {
		return nil, err
	}

	set := pongo2.NewSet("app",
		pongo2.NewFSLoader(os.DirFS("templates")),
		pongo2.NewFSLoader(tplFS),
		pongo2.NewFSLoader(internalFS),
	)

	set.Globals["dict"] = dict
	set.Globals["slice"] = slice
	set.Globals["scratch"] = newScratch
	set.Globals["newScratch"] = newScratchFunc

	set.RegisterFilter("parser", parser)
	set.RegisterFilter("slient", slient)
	set.RegisterFilter("jsonify", jsonify)

	for k, v := range GlobalVars {
		set.Globals[k] = v
	}

	for k, f := range GlobalFuncs {
		set.Globals[k] = f(ctx)
	}

	for k, f := range GlobalTags {
		set.RegisterTag(k, f(ctx))
	}

	for k, f := range GlobalFilters {
		set.RegisterFilter(k, f(ctx))
	}
	return &templateSet{
		tplset: set,
	}, nil
}
