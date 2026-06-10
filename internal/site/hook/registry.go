package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
)

type Registry struct {
	hooks []Hook
	names []string
}

func (r *Registry) AfterBuild(ctx context.Context, writer core.Writer) error {
	for _, hook := range r.hooks {
		if err := hook.AfterBuild(ctx, writer); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) BeforeBuild() error {
	for _, hook := range r.hooks {
		if err := hook.BeforeBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) HandleWriter(writer core.Writer) (core.Writer, error) {
	for _, hook := range r.hooks {
		result, err := hook.HandleWriter(writer)
		if err != nil {
			return nil, err
		}
		writer = result
	}
	return writer, nil
}

func (r *Registry) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	for _, hook := range r.hooks {
		result, err := hook.HandleTemplateSet(set)
		if err != nil {
			return nil, err
		}
		set = result
	}
	return set, nil
}

func (r *Registry) HandlePage(result *content.Page) *content.Page {
	for _, hook := range r.hooks {
		result = hook.HandlePage(result)
		if result == nil {
			return nil
		}
	}
	return result
}

func (r *Registry) HandlePages(results content.Pages) content.Pages {
	for _, hook := range r.hooks {
		results = hook.HandlePages(results)
		if len(results) == 0 {
			return nil
		}
	}
	return results
}

func (r *Registry) HandleSection(result *content.Section) *content.Section {
	for _, hook := range r.hooks {
		result = hook.HandleSection(result)
		if result == nil {
			return nil
		}
	}
	return result
}

func (r *Registry) HandleSections(results content.Sections) content.Sections {
	for _, hook := range r.hooks {
		results = hook.HandleSections(results)
		if len(results) == 0 {
			return nil
		}
	}
	return results
}

func (r *Registry) HandleTaxonomy(result *content.Taxonomy) *content.Taxonomy {
	for _, hook := range r.hooks {
		result = hook.HandleTaxonomy(result)
		if result == nil {
			return nil
		}
	}
	return result
}

func (r *Registry) HandleTaxonomies(results content.Taxonomies) content.Taxonomies {
	for _, hook := range r.hooks {
		results = hook.HandleTaxonomies(results)
		if len(results) == 0 {
			return nil
		}
	}
	return results
}

func New(ctx *core.Context) (*Registry, error) {
	names := make([]string, 0)
	for name := range ctx.Config.GetStringMap("hooks") {
		if !ctx.Config.GetBool(fmt.Sprintf("hooks.%s.enabled", name)) {
			continue
		}
		names = append(names, name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		wi := ctx.Config.GetInt("hooks." + names[i] + ".weight")
		wj := ctx.Config.GetInt("hooks." + names[j] + ".weight")
		if wi == wj {
			return names[i] < names[j]
		}
		return wi < wj
	})

	hooks := make([]Hook, 0)
	enabled := make([]string, 0, len(names))
	for _, name := range names {
		factory, ok := factories[name]
		if !ok {
			return nil, fmt.Errorf("hook %q is enabled but not registered", name)
		}

		h, err := factory(ctx)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
		enabled = append(enabled, fmt.Sprintf("%s(%d)", name, ctx.Config.GetInt("hooks."+name+".weight")))
	}
	if len(enabled) > 0 {
		ctx.Logger.Debugf("Enabled hooks: %s", strings.Join(enabled, ", "))
	}
	return &Registry{hooks: hooks, names: names}, nil
}

func Unmarshal(data any, value any) error {
	if data == nil {
		return nil
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, value)
}

func Print(ctx *core.Context) {
	names := make([]string, 0)
	for name := range factories {
		names = append(names, name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		wi := ctx.Config.GetInt("hooks." + names[i] + ".weight")
		wj := ctx.Config.GetInt("hooks." + names[j] + ".weight")
		if wi == wj {
			return names[i] < names[j]
		}
		return wi < wj
	})
	for i, name := range names {
		if ctx.Config.GetBool(fmt.Sprintf("hooks.%s.enabled", name)) {
			names[i] = name + "(enabled)"
		}
	}
	fmt.Println(strings.Join(names, ", "))
}

func Register(name string, c Factory) {
	if _, ok := factories[name]; ok {
		panic(fmt.Sprintf("The hook %s has been registered", name))
	}
	factories[name] = c
}

type Factory func(*core.Context) (Hook, error)

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}
