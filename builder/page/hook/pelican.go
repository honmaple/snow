package hook

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

func (e *pelican) AfterPageParse(meta map[string]string, page *page.Page) *page.Page {
	if v, ok := meta["save_as"]; ok {
		page.URL = v
	}
	return page
}

func newPelican(conf config.Config, theme theme.Theme) hook.Hook {
	return &pelican{conf: conf}
}
