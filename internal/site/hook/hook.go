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
		// HandleContent runs after all content for a language has been parsed and
		// inserted into the store, before rendering starts. It is intended for
		// body transformations that need to resolve references across content.
		HandleContent(ContentStore, string)

		HandleSection(*content.Section) *content.Section
	}
	ContentStore interface {
		Pages(string) content.Pages
		HiddenPages(string) content.Pages
		Sections(string) content.Sections
	}
	Hook interface {
		BuildHook
		ContentHook
	}
)

type HookImpl struct{}

func (HookImpl) AfterBuild(context.Context, core.Writer) error          { return nil }
func (HookImpl) BeforeBuild() error                                     { return nil }
func (HookImpl) HandlePage(result *content.Page) *content.Page          { return result }
func (HookImpl) HandleContent(ContentStore, string)                     {}
func (HookImpl) HandleSection(result *content.Section) *content.Section { return result }
func (HookImpl) HandleWriter(writer core.Writer) (core.Writer, error) {
	return writer, nil
}
func (HookImpl) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	return set, nil
}
