package hook

import (
	"github.com/honmaple/snow/builder/content"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	internalHook struct {
		BaseHook
		conf   config.Config
		filter func(*content.Page) bool
	}
)

func (self *internalHook) Name() string {
	return "internal"
}

func (self *internalHook) Page(p *content.Page) *content.Page {
	if self.filter != nil && !self.filter(p) {
		return nil
	}
	return p
}

func newInternalHook(conf config.Config, theme theme.Theme) Hook {
	return &internalHook{
		conf:   conf,
		filter: content.FilterExpr(conf.GetString("hooks.internal.filter")),
	}
}

func init() {
	Register("internal", newInternalHook)
}
