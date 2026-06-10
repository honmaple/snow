package snakecase

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	SnakeHook struct {
		hook.HookImpl
		ctx *core.Context
	}
)

func (h *SnakeHook) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	return &TemplateSet{set}, nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	h := &SnakeHook{
		ctx: ctx,
	}
	return h, nil
}

func init() {
	hook.Register("snakecase", New)
}
