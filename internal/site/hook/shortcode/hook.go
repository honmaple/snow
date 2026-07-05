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
	vars := map[string]any{
		"current_lang": page.Lang,
		"page":         page,
	}
	page.Summary = h.set.Render(page.File.Path, page.Summary, vars)
	page.Content = h.set.Render(page.File.Path, page.Content, vars)
	return page
}

func (h *ShortcodeHook) HandleSection(section *content.Section) *content.Section {
	if h.set == nil {
		return section
	}
	vars := map[string]any{
		"current_lang": section.Lang,
		"section":      section,
	}
	section.Summary = h.set.Render(section.File.Path, section.Summary, vars)
	section.Content = h.set.Render(section.File.Path, section.Content, vars)
	return section
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
