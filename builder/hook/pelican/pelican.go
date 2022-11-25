package pelican

import (
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type pelican struct {
	hook.BaseHook
	conf config.Config
}

func (s *pelican) Name() string {
	return "pelican"
}

func (e *pelican) AfterPageParse(page *page.Page) *page.Page {
	if v, ok := page.Meta["save_as"]; ok {
		page.URL = v.(string)
	}
	return page
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	return &pelican{conf: conf}
}

func init() {
	hook.Register("pelican", New)
}
