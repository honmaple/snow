package site

import (
	"fmt"
	"iter"
	stdpath "path"
	"sync"

	"github.com/honmaple/snow/internal/site/content"
)

type Set[T any] struct {
	list  []T
	index map[string]T
}

func (s *Set[T]) List() []T {
	return s.list
}

func (s *Set[T]) Iter() iter.Seq2[int, T] {
	return func(yield func(key int, value T) bool) {
		for k, v := range s.list {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *Set[T]) Add(key string, val T) {
	if _, ok := s.index[key]; !ok {
		s.list = append(s.list, val)
		s.index[key] = val
	}
}

func (s *Set[T]) Find(key string) (T, bool) {
	value, ok := s.index[key]
	return value, ok
}

func newSet[T any]() *Set[T] {
	return &Set[T]{
		list:  make([]T, 0),
		index: make(map[string]T),
	}
}

type ContentStore struct {
	mu            sync.RWMutex
	pages         map[string]*Set[*content.Page]
	hiddenPages   map[string]*Set[*content.Page]
	sections      map[string]*Set[*content.Section]
	taxonomies    map[string]*Set[*content.Taxonomy]
	taxonomyTerms map[string]*Set[*content.TaxonomyTerm]
}

func (d *ContentStore) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pages = make(map[string]*Set[*content.Page])
	d.hiddenPages = make(map[string]*Set[*content.Page])
	d.sections = make(map[string]*Set[*content.Section])
	d.taxonomies = make(map[string]*Set[*content.Taxonomy])
	d.taxonomyTerms = make(map[string]*Set[*content.TaxonomyTerm])
}

func (d *ContentStore) AllSections() map[string]content.Sections {
	d.mu.RLock()
	defer d.mu.RUnlock()
	results := make(map[string]content.Sections)
	for lang, set := range d.sections {
		results[lang] = set.List()
	}
	return results
}

func (d *ContentStore) Sections(lang string) content.Sections {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *ContentStore) getSection(path string, lang string) *content.Section {
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(path)
	return result
}

func (d *ContentStore) GetSection(path string, lang string) *content.Section {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getSection(path, lang)
}

func (d *ContentStore) GetSectionURL(path string, lang string) string {
	result := d.GetSection(path, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *ContentStore) AllPages() map[string]content.Pages {
	d.mu.RLock()
	defer d.mu.RUnlock()
	results := make(map[string]content.Pages)
	for lang, set := range d.pages {
		results[lang] = set.List()
	}
	return results
}

func (d *ContentStore) HiddenPages(lang string) content.Pages {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.hiddenPages[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *ContentStore) Pages(lang string) content.Pages {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *ContentStore) GetPage(path string, lang string) *content.Page {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(path)
	return result
}

func (d *ContentStore) GetPageURL(path string, lang string) string {
	result := d.GetPage(path, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *ContentStore) AllTaxonomies() map[string]content.Taxonomies {
	d.mu.RLock()
	defer d.mu.RUnlock()
	results := make(map[string]content.Taxonomies)
	for lang, set := range d.taxonomies {
		results[lang] = set.List()
	}
	return results
}

func (d *ContentStore) Taxonomies(lang string) content.Taxonomies {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *ContentStore) GetTaxonomy(name string, lang string) *content.Taxonomy {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(name)
	return result
}

func (d *ContentStore) GetTaxonomyURL(name string, lang string) string {
	result := d.GetTaxonomy(name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *ContentStore) GetTaxonomyTerms(name string, lang string) content.TaxonomyTerms {
	taxonomy := d.GetTaxonomy(name, lang)
	if taxonomy == nil {
		return nil
	}
	return taxonomy.Terms
}

func (d *ContentStore) GetTaxonomyTerm(taxonomyName string, name string, lang string) *content.TaxonomyTerm {
	d.mu.RLock()
	defer d.mu.RUnlock()
	set, ok := d.taxonomyTerms[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(fmt.Sprintf("%s:%s", taxonomyName, name))
	return result
}

func (d *ContentStore) GetTaxonomyTermURL(taxonomyName string, name string, lang string) string {
	result := d.GetTaxonomyTerm(taxonomyName, name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *ContentStore) insertSection(section *content.Section) {
	d.mu.Lock()
	defer d.mu.Unlock()

	set, ok := d.sections[section.Lang]
	if !ok {
		set = newSet[*content.Section]()
		d.sections[section.Lang] = set
	}
	if section.IsHome() {
		set.Add("/", section)
	} else {
		set.Add(section.File.Dir, section)

		currentDir := stdpath.Dir(section.File.Dir)
		for {
			if currentDir == "" || currentDir == "." {
				currentDir = "/"
			}
			if parent := d.getSection(currentDir, section.Lang); parent != nil {
				section.Parent = parent
				parent.Children = append(parent.Children, section)
				break
			}
			if currentDir == "/" {
				break
			}
			currentDir = stdpath.Dir(currentDir)
		}
	}
}

func (d *ContentStore) insertPage(page *content.Page) {
	if page.Hidden {
		d.insertHiddenPage(page)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	set, ok := d.pages[page.Lang]
	if !ok {
		set = newSet[*content.Page]()
		d.pages[page.Lang] = set
	}
	set.Add(page.File.Path, page)

	sectionPath := page.File.Dir
	if page.IsBundle {
		sectionPath = stdpath.Dir(sectionPath)
	}

	currentDir := sectionPath
	for {
		if currentDir == "" || currentDir == "." {
			currentDir = "/"
		}
		if section := d.getSection(currentDir, page.Lang); section != nil {
			page.Section = section
			section.Pages = append(section.Pages, page)
			break
		}
		if currentDir == "/" {
			break
		}
		currentDir = stdpath.Dir(currentDir)
	}
}

func (d *ContentStore) insertHiddenPage(page *content.Page) {
	d.mu.Lock()
	defer d.mu.Unlock()

	set, ok := d.hiddenPages[page.Lang]
	if !ok {
		set = newSet[*content.Page]()
		d.hiddenPages[page.Lang] = set
	}
	set.Add(page.File.Path, page)

	sectionPath := page.File.Dir
	if page.IsBundle {
		sectionPath = stdpath.Dir(sectionPath)
	}

	currentDir := sectionPath
	for {
		if currentDir == "" || currentDir == "." {
			currentDir = "/"
		}
		if section := d.getSection(currentDir, page.Lang); section != nil {
			page.Section = section
			section.HiddenPages = append(section.HiddenPages, page)
			break
		}
		if currentDir == "/" {
			break
		}
		currentDir = stdpath.Dir(currentDir)
	}
}

func (d *ContentStore) insertTaxonomy(taxonomy *content.Taxonomy) {
	d.mu.Lock()
	defer d.mu.Unlock()

	set, ok := d.taxonomies[taxonomy.Lang]
	if !ok {
		set = newSet[*content.Taxonomy]()
		d.taxonomies[taxonomy.Lang] = set
	}
	set.Add(taxonomy.Name, taxonomy)
}

func (d *ContentStore) insertTaxonomyTerm(term *content.TaxonomyTerm) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.addTaxonomyTerm(term)
}

func (d *ContentStore) addTaxonomyTerm(term *content.TaxonomyTerm) {
	set, ok := d.taxonomyTerms[term.Taxonomy.Lang]
	if !ok {
		set = newSet[*content.TaxonomyTerm]()
		d.taxonomyTerms[term.Taxonomy.Lang] = set
	}
	set.Add(fmt.Sprintf("%s:%s", term.Taxonomy.Name, term.GetFullName()), term)

	for _, child := range term.Children {
		d.addTaxonomyTerm(child)
	}
}

func NewContentStore() *ContentStore {
	store := &ContentStore{}
	store.Reset()
	return store
}
