package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
)

type BaseHook struct {
}

func (BaseHook) AfterPageParse(page *page.Page) *page.Page {
	return page
}

func (BaseHook) BeforePageWrite(page *page.Page) *page.Page {
	return page
}

func (BaseHook) BeforePagesWrite(pages page.Pages) page.Pages {
	return pages
}

func (BaseHook) BeforeLabelsWrite(labels page.Labels) page.Labels {
	return labels
}

func (BaseHook) BeforeStaticWrite(static *static.Static) *static.Static {
	return static
}

func (BaseHook) BeforeStaticsWrite(statics static.Statics) static.Statics {
	return statics
}
