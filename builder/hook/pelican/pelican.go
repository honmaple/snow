package pelican

import (
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
)

type (
	pelican struct {
		hook.BaseHook
		conf config.Config
	}
)

func (self *pelican) Name() string {
	return "pelican"
}

func (self *pelican) AfterPageParse(p *page.Page) *page.Page {
	if v, ok := p.Meta["tag"]; ok {
		p.Meta["tags"] = utils.SplitTrim(v.(string), ",")
	}
	if v, ok := p.Meta["author"]; ok {
		p.Meta["authors"] = utils.SplitTrim(v.(string), ",")
	}
	if v, ok := p.Meta["category"]; ok {
		p.Meta["categories"] = utils.SplitTrim(v.(string), ",")
	}
	return p
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	return &pelican{conf: conf}
}

func init() {
	hook.Register("pelican", New)
}
