package site

import (
	"context"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Site struct {
		ctx              *core.Context
		hook             hook.Hook
		contentProcessor *content.Processor
		tplset           template.TemplateSet
		includeDrafts    bool
	}
	SiteOption func(*Site)
)

func (site *Site) newTemplateSet() (template.TemplateSet, error) {
	tplset, err := template.NewSet(site.ctx)
	if err != nil {
		return nil, err
	}
	return site.hook.HandleTemplateSet(tplset)
}

func (site *Site) Build(ctx context.Context, w core.Writer) error {
	writer, err := site.hook.HandleWriter(w)
	if err != nil {
		return err
	}

	if err := site.hook.BeforeBuild(); err != nil {
		return err
	}
	if err := site.BuildStatic(ctx, writer); err != nil {
		return err
	}
	if err := site.BuildContent(ctx, writer); err != nil {
		return err
	}
	return site.hook.AfterBuild(ctx, writer)
}

func IncludeDrafts(b bool) SiteOption {
	return func(site *Site) {
		site.includeDrafts = b
	}
}

func New(ctx *core.Context, opts ...SiteOption) (*Site, error) {
	site := &Site{
		ctx:              ctx,
		contentProcessor: content.NewProcessor(ctx),
	}
	for _, opt := range opts {
		opt(site)
	}

	h, err := hook.New(ctx)
	if err != nil {
		return nil, err
	}
	site.hook = h

	tplset, err := site.newTemplateSet()
	if err != nil {
		return nil, err
	}
	site.tplset = tplset
	return site, nil
}
