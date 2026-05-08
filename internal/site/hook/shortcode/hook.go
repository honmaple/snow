package shortcode

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type ShortcodeHook struct {
	hook.HookImpl
	ctx *core.Context
	set *ShortcodeSet
}

func (h *ShortcodeHook) HandlePage(page *content.Page) *content.Page {
	if h.set == nil {
		return page
	}
	page.Summary = h.set.Render(page, page.Summary)
	page.Content = h.set.Render(page, page.Content)
	return page
}

func (h *ShortcodeHook) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	sc, err := NewShortcodeSet(h.ctx, set)
	if err != nil {
		return nil, err
	}
	h.set = sc
	return set, nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	h := &ShortcodeHook{
		ctx: ctx,
	}
	return h, nil
}

func init() {
	hook.Register("shortcode", New)
}
