package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
)

type BaseHook struct {
}

func (BaseHook) BeforePage(page *page.Page) *page.Page {
	return page
}

func (BaseHook) BeforePageList(pages page.Pages) page.Pages {
	return pages
}

func (BaseHook) BeforePageSection(section page.Section) page.Section {
	return section
}

func (BaseHook) BeforeStatic(static *static.Static) *static.Static {
	return static
}

func (BaseHook) BeforeStaticList(statics static.Statics) static.Statics {
	return statics
}
