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
	}
)

func (h *FilterHook) HandlePage(page *content.Page) *content.Page {
	expr := h.ctx.Config.GetString("hooks.filter.option.page_filter")
	if expr == "" {
		return page
	}
	tpl, err := pongo2.FromString("{{" + expr + "}}")
	if err != nil {
		h.ctx.Logger.Warnf("filter expr %s err: %s", expr, err.Error())
		return page
	}
	args := page.FrontMatter.AllSettings()

	result, err := tpl.Execute(args)
	if err == nil && result == "True" {
		return page
	}
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	return &FilterHook{
		ctx: ctx,
	}, nil
}

func init() {
	hook.Register("filter", New)
}
