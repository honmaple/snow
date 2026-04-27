package shortcode

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type ShortcodeHook struct {
	hook.HookImpl
	sc  *Shortcode
	ctx *core.Context
}

func (h *ShortcodeHook) HandlePage(page *content.Page) *content.Page {
	page.Summary = h.sc.Render(page, page.Summary)
	page.Content = h.sc.Render(page, page.Content)
	return page
}

func (h *ShortcodeHook) HandleInit(set template.TemplateSet) error {
	sc, err := NewShortcode(h.ctx, set)
	if err != nil {
		return err
	}
	h.sc = sc
	return nil
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
