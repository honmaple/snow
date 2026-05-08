package minify

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
)

type (
	MinifyHook struct {
		hook.HookImpl
		ctx *core.Context
	}
)

func (h *MinifyHook) HandleWriter(w core.Writer) (core.Writer, error) {
	return NewWriter(h.ctx, w), nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	return &MinifyHook{
		ctx: ctx,
	}, nil
}

func init() {
	hook.Register("minify", New)
}
