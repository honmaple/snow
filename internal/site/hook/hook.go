package hook

import (
	"context"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	BuildHook interface {
		AfterBuild(context.Context, core.Writer) error
		BeforeBuild() error
		HandleWriter(core.Writer) (core.Writer, error)
		HandleTemplateSet(template.TemplateSet) (template.TemplateSet, error)
	}
	ContentHook interface {
		HandlePage(*content.Page) *content.Page
		// Collection hooks are reserved for future use. The current site build
		// flow calls the single-item hooks below, but does not call collection
		// hooks while loading or rendering content.
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

func (HookImpl) AfterBuild(context.Context, core.Writer) error                  { return nil }
func (HookImpl) BeforeBuild() error                                             { return nil }
func (HookImpl) HandlePage(result *content.Page) *content.Page                  { return result }
func (HookImpl) HandlePages(results content.Pages) content.Pages                { return results }
func (HookImpl) HandleSection(result *content.Section) *content.Section         { return result }
func (HookImpl) HandleSections(results content.Sections) content.Sections       { return results }
func (HookImpl) HandleTaxonomy(result *content.Taxonomy) *content.Taxonomy      { return result }
func (HookImpl) HandleTaxonomies(results content.Taxonomies) content.Taxonomies { return results }
func (HookImpl) HandleWriter(writer core.Writer) (core.Writer, error) {
	return writer, nil
}
func (HookImpl) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	return set, nil
}
