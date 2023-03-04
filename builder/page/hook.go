package page

type (
	Hook interface {
		AfterPageParse(*Page) *Page
		BeforePageWrite(*Page) *Page
		BeforePagesWrite(Pages) Pages
		BeforeSectionsWrite(Sections) Sections
		BeforeTaxonomiesWrite(Taxonomies) Taxonomies
		BeforeTemplateWrite(map[string]interface{}) map[string]interface{}
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

func (hooks Hooks) BeforeSectionsWrite(sections Sections) Sections {
	for _, hook := range hooks {
		sections = hook.BeforeSectionsWrite(sections)
	}
	return sections
}

func (hooks Hooks) BeforeTaxonomiesWrite(taxonomies Taxonomies) Taxonomies {
	for _, hook := range hooks {
		taxonomies = hook.BeforeTaxonomiesWrite(taxonomies)
	}
	return taxonomies
}

func (hooks Hooks) BeforeTemplateWrite(vars map[string]interface{}) map[string]interface{} {
	for _, hook := range hooks {
		vars = hook.BeforeTemplateWrite(vars)
	}
	return vars
}
