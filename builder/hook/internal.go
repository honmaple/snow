package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type internal struct {
	BaseHook
	conf config.Config
}

func (b *internal) Name() string {
	return "internal"
}

func (b *internal) BeforePagesWrite(pages page.Pages) page.Pages {
	var (
		prev  *page.Page
		terms = pages.GroupBy("type")
	)
	for _, term := range terms {
		var prevInType *page.Page
		for _, page := range term.List {
			page.PrevInType = prevInType
			if prevInType != nil {
				prevInType.NextInType = page
			}
			prevInType = page

			page.Prev = prev
			if prev != nil {
				prev.Next = page
			}
			prev = page
		}
	}
	return pages
}

func init() {
	Register("internal", func(conf config.Config, theme theme.Theme) Hook {
		return &internal{conf: conf}
	})
}