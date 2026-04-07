package types

type (
	Hook interface {
		HandlePage(*Page) *Page
		HandlePages(Pages) Pages

		HandleSection(*Section) *Section
		HandleSections(Sections) Sections

		HandleTaxonomy(*Taxonomy) *Taxonomy
		HandleTaxonomies(Taxonomies) Taxonomies
	}
	EmptyHook struct {
	}
)

func (EmptyHook) HandlePage(result *Page) *Page                  { return result }
func (EmptyHook) HandlePages(results Pages) Pages                { return results }
func (EmptyHook) HandleSection(result *Section) *Section         { return result }
func (EmptyHook) HandleSections(results Sections) Sections       { return results }
func (EmptyHook) HandleTaxonomy(result *Taxonomy) *Taxonomy      { return result }
func (EmptyHook) HandleTaxonomies(results Taxonomies) Taxonomies { return results }
