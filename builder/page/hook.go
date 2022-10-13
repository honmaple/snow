package page

type (
	Hook interface {
		BeforePage(*Page) *Page
		BeforePageSection(Section) Section
		BeforePageList(Pages) Pages
	}
	Hooks []Hook
)

func (hooks Hooks) BeforePage(page *Page) *Page {
	for _, hook := range hooks {
		page = hook.BeforePage(page)
	}
	return page
}

func (hooks Hooks) BeforePageList(pages Pages) Pages {
	for _, hook := range hooks {
		pages = hook.BeforePageList(pages)
	}
	return pages
}

func (hooks Hooks) BeforePageSection(section Section) Section {
	for _, hook := range hooks {
		section = hook.BeforePageSection(section)
	}
	return section
}
