package page

type (
	Hook interface {
		AfterPageParse(*Page) *Page
		BeforePageWrite(*Page) *Page
		BeforePagesWrite(Pages) Pages
		BeforeLabelsWrite(Labels) Labels
	}
	Hooks []Hook
)

func (hooks Hooks) AfterPageParse(page *Page) *Page {
	for _, hook := range hooks {
		page = hook.AfterPageParse(page)
	}
	return page
}

func (hooks Hooks) BeforePageWrite(page *Page) *Page {
	for _, hook := range hooks {
		page = hook.BeforePageWrite(page)
	}
	return page
}

func (hooks Hooks) BeforePagesWrite(pages Pages) Pages {
	for _, hook := range hooks {
		pages = hook.BeforePagesWrite(pages)
	}
	return pages
}

func (hooks Hooks) BeforeLabelsWrite(labels Labels) Labels {
	for _, hook := range hooks {
		labels = hook.BeforeLabelsWrite(labels)
	}
	return labels
}
