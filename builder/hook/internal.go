package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	internalHook struct {
		BaseHook
		conf   config.Config
		filter func(*page.Page) bool
	}
)

func (self *internalHook) Name() string {
	return "internal"
}

func (self *internalHook) Page(p *page.Page) *page.Page {
	if self.filter != nil && !self.filter(p) {
		return nil
	}
	return p
}

func newInternalHook(conf config.Config, theme theme.Theme) Hook {
	return &internalHook{
		conf:   conf,
		filter: page.FilterExpr(conf.GetString("build_filter")),
	}
}

func init() {
	RegisterInternal("internal", newInternalHook)
}
