package filter

import (
	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
)

type (
	Option struct {
	}
	FilterHook struct {
		hook.HookImpl
		ctx *core.Context
		tpl *pongo2.Template
	}
)

func (h *FilterHook) HandlePage(page *content.Page) *content.Page {
	if h.tpl == nil {
		return page
	}
	args := page.FrontMatter.AllSettings()

	result, err := h.tpl.Execute(args)
	if err == nil && result == "True" {
		return page
	}
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	var tpl *pongo2.Template
	if expr := ctx.Config.GetString("hooks.filter.option.page_filter"); expr != "" {
		t, err := pongo2.FromString("{{" + expr + "}}")
		if err != nil {
			return nil, err
		}
		tpl = t
	}
	return &FilterHook{
		ctx: ctx,
		tpl: tpl,
	}, nil
}

func init() {
	hook.Register("filter", New)
}
