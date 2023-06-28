package page

type (
	Hook interface {
		Page(*Page) *Page
		Section(*Section) *Section

		Pages(Pages) Pages
		Sections(Sections) Sections
		Taxonomies(Taxonomies) Taxonomies
		TaxonomyTerms(TaxonomyTerms) TaxonomyTerms
	}
	Hooks []Hook
)

func (hooks Hooks) Page(page *Page) *Page {
	for _, hook := range hooks {
		page = hook.Page(page)
		if page == nil {
			return nil
		}
	}
	return page
}

func (hooks Hooks) Section(section *Section) *Section {
	for _, hook := range hooks {
		section = hook.Section(section)
		if section == nil {
			return nil
		}
	}
	return section
}

func (hooks Hooks) Pages(pages Pages) Pages {
	for _, hook := range hooks {
		pages = hook.Pages(pages)
		if len(pages) == 0 {
			return nil
		}
	}
	return pages
}

func (hooks Hooks) Sections(sections Sections) Sections {
	for _, hook := range hooks {
		sections = hook.Sections(sections)
		if len(sections) == 0 {
			return nil
		}
	}
	return sections
}

func (hooks Hooks) Taxonomies(taxonomies Taxonomies) Taxonomies {
	for _, hook := range hooks {
		taxonomies = hook.Taxonomies(taxonomies)
		if len(taxonomies) == 0 {
			return nil
		}
	}
	return taxonomies
}

func (hooks Hooks) TaxonomyTerms(terms TaxonomyTerms) TaxonomyTerms {
	for _, hook := range hooks {
		terms = hook.TaxonomyTerms(terms)
		if len(terms) == 0 {
			return nil
		}
	}
	return terms
}
