package rewrite

import (
	"fmt"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/utils"
)

type (
	Option struct {
		Src  string `json:"src"`
		Dst  string `json:"dst"`
		Type string `json:"type"`
	}
	RewriteHook struct {
		hook.HookImpl
		ctx  *core.Context
		opts []*Option
	}
)

func (h *RewriteHook) HandlePage(page *content.Page) *content.Page {
	for _, opt := range h.opts {
		if !page.FrontMatter.IsSet(opt.Src) {
			continue
		}
		if v := page.FrontMatter.Get(opt.Src); v != nil {
			page.FrontMatter.Set(opt.Src, nil)

			switch opt.Type {
			case "list":
				page.FrontMatter.Set(opt.Dst, utils.SplitTrim(v.(string), ","))
			default:
				page.FrontMatter.Set(opt.Dst, v)
			}
		}
	}
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	var opts []*Option

	if err := hook.Unmarshal(ctx.Config.Get("hooks.rewrite.option"), &opts); err != nil {
		return nil, err
	}
	for i, opt := range opts {
		if opt.Src == "" {
			return nil, fmt.Errorf("hooks.rewrite.option[%d].src is required", i)
		}
		if opt.Dst == "" {
			return nil, fmt.Errorf("hooks.rewrite.option[%d].dst is required", i)
		}
		switch opt.Type {
		case "", "list":
		default:
			return nil, fmt.Errorf("hooks.rewrite.option[%d].type %q is invalid", i, opt.Type)
		}
	}
	return &RewriteHook{ctx: ctx, opts: opts}, nil
}

func init() {
	hook.Register("rewrite", New)
}
