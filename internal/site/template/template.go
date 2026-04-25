package template

import (
	"fmt"
	"io"
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
		Execute(map[string]any) (string, error)
		ExecuteRaw(map[string]any) (string, error)
	}
	templateImpl struct {
		n      string
		tpl    *pongo2.Template
		tplset *templateSet
	}
)

func (t *templateImpl) Name() string {
	return t.n
}

func (t *templateImpl) Execute(vars map[string]any) (string, error) {
	nvars := make(map[string]any)
	maps.Copy(nvars, vars)

	for k, fn := range t.tplset.funcs {
		if _, ok := nvars[k]; !ok {
			nvars[k] = fn(vars)
		}
	}

	for k, fn := range t.tplset.filters {
		t.tplset.ReplaceFilter(k, fn(vars))
	}
	return t.tpl.Execute(nvars)
}

func (t *templateImpl) ExecuteRaw(vars map[string]any) (string, error) {
	nvars := make(map[string]any)
	maps.Copy(nvars, vars)

	return t.tpl.Execute(nvars)
}

type (
	TemplateSet interface {
		Lookup(...string) Template
		FromFile(string) (Template, error)
		FromBytes([]byte) (Template, error)
		FromString(string) (Template, error)

		Register(string, any) error
		RegisterTag(string, pongo2.TagParser) error
		RegisterFilter(string, pongo2.FilterFunction) error
		RegisterTransient(string, TransientFunction) error
		RegisterTransientFilter(string, TransientFilterFunction) error
	}
	templateSet struct {
		ctx     *core.Context
		cache   sync.Map
		tplset  *pongo2.TemplateSet
		loaders []pongo2.TemplateLoader
		// 依赖于ctx上下文的变量
		funcs   map[string]TransientFunction
		filters map[string]TransientFilterFunction
	}
	TransientFunction       func(map[string]any) any
	TransientFilterFunction func(map[string]any) pongo2.FilterFunction
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
		template, err := set.fromFile(name)
		if err != nil {
			set.ctx.Logger.Warnf("parse template %s err: %s", name, err.Error())
			continue
		}
		if template == nil {
			continue
		}
		set.cache.Store(name, template)
		return template
	}
	return nil
}

func (set *templateSet) fromFile(name string) (Template, error) {
	for _, loader := range set.loaders {
		r, err := loader.Get(name)
		if err != nil {
			continue
		}
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		tpl, err := set.tplset.FromBytes(buf)
		if err != nil {
			return nil, err
		}
		return &templateImpl{n: name, tpl: tpl, tplset: set}, nil
	}
	return nil, nil
}

func (set *templateSet) FromFile(name string) (Template, error) {
	tpl, err := set.tplset.FromFile(name)
	if err != nil {
		return nil, err
	}
	return &templateImpl{n: name, tpl: tpl, tplset: set}, nil
}

func (set *templateSet) FromBytes(b []byte) (Template, error) {
	tpl, err := set.tplset.FromBytes(b)
	if err != nil {
		return nil, err
	}
	return &templateImpl{tpl: tpl, tplset: set}, nil
}

func (set *templateSet) FromString(b string) (Template, error) {
	tpl, err := set.tplset.FromString(b)
	if err != nil {
		return nil, err
	}
	return &templateImpl{tpl: tpl, tplset: set}, nil
}

func (set *templateSet) Register(name string, val any) error {
	set.tplset.Globals[name] = val
	return nil
}

func (set *templateSet) RegisterTag(name string, fn pongo2.TagParser) error {
	return set.tplset.RegisterTag(name, fn)
}

func (set *templateSet) RegisterFilter(name string, fn pongo2.FilterFunction) error {
	return set.tplset.RegisterFilter(name, fn)
}

func (set *templateSet) ReplaceFilter(name string, fn pongo2.FilterFunction) error {
	if set.tplset.FilterExists(name) {
		return set.tplset.ReplaceFilter(name, fn)
	}
	return set.tplset.RegisterFilter(name, fn)
}

func (set *templateSet) RegisterTransient(name string, fn TransientFunction) error {
	if _, ok := set.funcs[name]; !ok {
		set.funcs[name] = fn
	}
	return nil
}

func (set *templateSet) RegisterTransientFilter(name string, fn TransientFilterFunction) error {
	if _, ok := set.filters[name]; !ok {
		set.filters[name] = fn
	}
	return nil
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

	loaders := []pongo2.TemplateLoader{
		newLoader(os.DirFS("templates")),
		newLoader(tplFS),
		newLoader(internalFS),
	}
	tplset := pongo2.NewSet("app", loaders...)

	set := &templateSet{
		ctx:     ctx,
		tplset:  tplset,
		loaders: loaders,
		funcs:   make(map[string]TransientFunction),
		filters: make(map[string]TransientFilterFunction),
	}
	for _, fc := range factories {
		if err := fc(ctx, set); err != nil {
			return nil, err
		}
	}
	return set, nil
}

type Factory func(*core.Context, TemplateSet) error

func Register(name string, fc Factory) {
	factoriesOnce.Do(func() {
		factories = make(map[string]Factory)
	})

	if _, ok := factories[name]; ok {
		panic(fmt.Errorf("factory with name '%s' is already registered", name))
	}
	factories[name] = fc
}

var (
	factories     map[string]Factory
	factoriesOnce sync.Once
)
