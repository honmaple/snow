package renderer

import (
	"github.com/honmaple/snow/internal/site/content/types"
)

type LocaleStore struct {
	lang  string
	store Store
}

func (s *LocaleStore) Sections(args ...string) types.Sections {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.Sections(lang)
}

func (s *LocaleStore) GetSection(path string, args ...string) *types.Section {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetSection(path, lang)
}

func (s *LocaleStore) GetSectionURL(path string, args ...string) string {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetSectionURL(path, lang)
}

func (s *LocaleStore) Pages(args ...string) types.Pages {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.Pages(lang)
}

func (s *LocaleStore) HiddenPages(args ...string) types.Pages {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.HiddenPages(lang)
}

func (s *LocaleStore) GetPage(path string, args ...string) *types.Page {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetPage(path, lang)
}

func (s *LocaleStore) GetPageURL(path string, args ...string) string {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetPageURL(path, lang)
}

func (s *LocaleStore) Taxonomies(args ...string) types.Taxonomies {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.Taxonomies(lang)
}

func (s *LocaleStore) GetTaxonomy(name string, args ...string) *types.Taxonomy {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetTaxonomy(name, lang)
}

func (s *LocaleStore) GetTaxonomyURL(name string, args ...string) string {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetTaxonomyURL(name, lang)
}

func (s *LocaleStore) GetTaxonomyTerm(taxonomyName string, name string, args ...string) *types.TaxonomyTerm {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetTaxonomyTerm(taxonomyName, name, lang)
}

func (s *LocaleStore) GetTaxonomyTermURL(taxonomyName string, name string, args ...string) string {
	lang := s.lang
	if len(args) > 0 {
		lang = args[0]
	}
	return s.store.GetTaxonomyTermURL(taxonomyName, name, lang)
}
