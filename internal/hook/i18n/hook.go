package i18n

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/hook"
	"github.com/honmaple/snow/internal/theme/template"
)

type i18nHook struct {
	hook.HookImpl

	ctx  *core.Context
	i18n *I18n
}

func New(ctx *core.Context) (hook.Hook, error) {
	i18n := &I18n{}
	if err := i18n.LoadTranslations(ctx); err != nil {
		ctx.Logger.Warnf("load translate err: %s", err.Error())
	}
	template.Register("__i18n__", i18n)

	h := &i18nHook{ctx: ctx, i18n: i18n}
	return h, nil
}

func init() {
	hook.Register("i18n", New)
}
