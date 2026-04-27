package hook

import (
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	BuildHook interface {
		AfterBuild() error
		BeforeBuild() error
		HandleInit(template.TemplateSet) error
	}
	ContentHook interface {
		HandlePage(*content.Page) *content.Page
		HandlePages(content.Pages) content.Pages

		HandleSection(*content.Section) *content.Section
		HandleSections(content.Sections) content.Sections

		HandleTaxonomy(*content.Taxonomy) *content.Taxonomy
		HandleTaxonomies(content.Taxonomies) content.Taxonomies
	}
	Hook interface {
		BuildHook
		ContentHook
	}
)

type HookImpl struct{}

func (HookImpl) AfterBuild() error                                              { return nil }
func (HookImpl) BeforeBuild() error                                             { return nil }
func (HookImpl) HandleInit(template.TemplateSet) error                          { return nil }
func (HookImpl) HandlePage(result *content.Page) *content.Page                  { return result }
func (HookImpl) HandlePages(results content.Pages) content.Pages                { return results }
func (HookImpl) HandleSection(result *content.Section) *content.Section         { return result }
func (HookImpl) HandleSections(results content.Sections) content.Sections       { return results }
func (HookImpl) HandleTaxonomy(result *content.Taxonomy) *content.Taxonomy      { return result }
func (HookImpl) HandleTaxonomies(results content.Taxonomies) content.Taxonomies { return results }
