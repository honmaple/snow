package hook

import (
	"fmt"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/content"
	"github.com/honmaple/snow/internal/core"
)

type Registry struct {
	hooks []Hook
}

func (r *Registry) AfterBuild() error {
	for _, hook := range r.hooks {
		if err := hook.AfterBuild(); err != nil {
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
	names := ctx.Config.GetStringSlice("registered_hooks")
	if len(names) == 0 {
		names = make([]string, 0)
		for name := range ctx.Config.GetStringMap("hooks") {
			names = append(names, name)
		}
		sort.SliceStable(names, func(i, j int) bool {
			wi := ctx.Config.GetInt("hooks." + names[i] + ".weight")
			wj := ctx.Config.GetInt("hooks." + names[j] + ".weight")
			if wi == wj {
				return names[i] > names[j]
			}
			return wi > wj
		})
	}

	hooks := make([]Hook, 0)
	for _, name := range names {
		factory, ok := factories[name]
		if !ok {
			ctx.Logger.Warnf("The hook %s not found", name)
			continue
		}
		h, err := factory(ctx)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return &Registry{hooks: hooks}, nil
}

func NewEmpty() *Registry {
	return &Registry{}
}

func Print() {
	names := make([]string, 0)
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
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
