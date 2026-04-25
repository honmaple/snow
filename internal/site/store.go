package site

import (
	"fmt"
	stdpath "path"

	"github.com/honmaple/snow/internal/site/content/types"
)

type Store struct {
	pages         map[string]*Set[*types.Page]
	sections      map[string]*Set[*types.Section]
	taxonomies    map[string]*Set[*types.Taxonomy]
	taxonomyTerms map[string]*Set[*types.TaxonomyTerm]
}

func (d *Store) Reset() {
	d.pages = make(map[string]*Set[*types.Page])
	d.sections = make(map[string]*Set[*types.Section])
	d.taxonomies = make(map[string]*Set[*types.Taxonomy])
	d.taxonomyTerms = make(map[string]*Set[*types.TaxonomyTerm])
}

func (d *Store) AllSections() map[string]types.Sections {
	results := make(map[string]types.Sections)
	for lang, set := range d.sections {
		results[lang] = set.List()
	}
	return results
}

func (d *Store) Sections(lang string) types.Sections {
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Store) GetSection(path string, lang string) *types.Section {
	set, ok := d.sections[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(path)
	return result
}

func (d *Store) GetSectionURL(path string, lang string) string {
	result := d.GetSection(path, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Store) AllPages() map[string]types.Pages {
	results := make(map[string]types.Pages)
	for lang, set := range d.pages {
		results[lang] = set.List()
	}
	return results
}

func (d *Store) Pages(lang string) types.Pages {
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Store) GetPage(path string, lang string) *types.Page {
	set, ok := d.pages[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(path)
	return result
}

func (d *Store) GetPageURL(path string, lang string) string {
	result := d.GetPage(path, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Store) AllTaxonomies() map[string]types.Taxonomies {
	results := make(map[string]types.Taxonomies)
	for lang, set := range d.taxonomies {
		results[lang] = set.List()
	}
	return results
}

func (d *Store) Taxonomies(lang string) types.Taxonomies {
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	return set.List()
}

func (d *Store) GetTaxonomy(name string, lang string) *types.Taxonomy {
	set, ok := d.taxonomies[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(name)
	return result
}

func (d *Store) GetTaxonomyURL(name string, lang string) string {
	result := d.GetTaxonomy(name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Store) GetTaxonomyTerms(name string, lang string) types.TaxonomyTerms {
	taxonomy := d.GetTaxonomy(name, lang)
	if taxonomy == nil {
		return nil
	}
	return taxonomy.Terms
}

func (d *Store) GetTaxonomyTerm(taxonomyName string, name string, lang string) *types.TaxonomyTerm {
	set, ok := d.taxonomyTerms[lang]
	if !ok {
		return nil
	}
	result, _ := set.Find(fmt.Sprintf("%s:%s", taxonomyName, name))
	return result
}

func (d *Store) GetTaxonomyTermURL(taxonomyName string, name string, lang string) string {
	result := d.GetTaxonomyTerm(taxonomyName, name, lang)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *Store) insertSection(section *types.Section) {
	set, ok := d.sections[section.Lang]
	if !ok {
		set = newSet[*types.Section]()

		d.sections[section.Lang] = set
	}
	if section.File.Dir == "" {
		set.Add("/", section)
	} else {
		set.Add(section.File.Dir, section)
	}
}

func (d *Store) insertPage(page *types.Page) {
	set, ok := d.pages[page.Lang]
	if !ok {
		set = newSet[*types.Page]()
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
			break
		}
		section := d.GetSection(currentDir, page.Lang)
		if section != nil {
			section.Pages = append(section.Pages, page)
		}
		currentDir = stdpath.Dir(currentDir)
	}
	if root := d.GetSection("/", page.Lang); root != nil {
		root.Pages = append(root.Pages, page)
	}
}

func (d *Store) insertTaxonomy(taxonomy *types.Taxonomy) {
	set, ok := d.taxonomies[taxonomy.Lang]
	if !ok {
		set = newSet[*types.Taxonomy]()
		d.taxonomies[taxonomy.Lang] = set
	}
	set.Add(taxonomy.Name, taxonomy)
}

func (d *Store) insertTaxonomyTerm(term *types.TaxonomyTerm) {
	set, ok := d.taxonomyTerms[term.Taxonomy.Lang]
	if !ok {
		set = newSet[*types.TaxonomyTerm]()
		d.taxonomyTerms[term.Taxonomy.Lang] = set
	}
	set.Add(fmt.Sprintf("%s:%s", term.Taxonomy.Name, term.GetFullName()), term)

	for _, child := range term.Children {
		d.insertTaxonomyTerm(child)
	}
}

func NewStore() *Store {
	store := &Store{}
	store.Reset()
	return store
}
