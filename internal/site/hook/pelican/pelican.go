package pelican

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/utils"
)

type (
	PelicanHook struct {
		hook.HookImpl
		ctx *core.Context
	}
)

func (h *PelicanHook) HandlePage(page *content.Page) *content.Page {
	if v := page.FrontMatter.Get("tag"); v != nil {
		page.FrontMatter.Set("tag", nil)
		page.FrontMatter.Set("tags", utils.SplitTrim(v.(string), ","))
	}
	if v := page.FrontMatter.Get("tags"); v != nil {
		if vv, ok := v.(string); ok {
			page.FrontMatter.Set("tags", utils.SplitTrim(vv, ","))
		}
	}
	if v := page.FrontMatter.Get("author"); v != nil {
		page.FrontMatter.Set("author", nil)
		page.FrontMatter.Set("authors", utils.SplitTrim(v.(string), ","))
	}
	if v := page.FrontMatter.Get("category"); v != nil {
		page.FrontMatter.Set("category", nil)
		page.FrontMatter.Set("categories", utils.SplitTrim(v.(string), ","))
	}
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	return &PelicanHook{ctx: ctx}, nil
}

func init() {
	hook.Register("pelican", New)
}
