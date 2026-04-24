package shortcode

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
)

type shortcodeHook struct {
	hook.HookImpl
	ctx       *core.Context
	shortcode *shortcode
}

func (h *shortcodeHook) HandlePage(page *content.Page) *content.Page {
	page.Summary = h.shortcode.Render(page, page.Summary)
	page.Content = h.shortcode.Render(page, page.Content)
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	sc, err := NewShortcode(ctx)
	if err != nil {
		return nil, err
	}
	h := &shortcodeHook{
		ctx:       ctx,
		shortcode: sc,
	}
	return h, nil
}

func init() {
	hook.Register("shortcode", New)
}
