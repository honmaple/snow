package site

import (
	"context"
	"errors"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Option struct {
		IncludeDrafts bool
	}
	Site struct {
		ctx              *core.Context
		hook             hook.Hook
		option           *Option
		writer           core.Writer
		store            *ContentStore
		contentProcessor *content.Processor
		tplset           template.TemplateSet
	}
	SiteOption func(*Site)
)

func (site *Site) Build(ctx context.Context) error {
	site.store.Reset()

	if err := site.loadContent(); err != nil {
		return err
	}

	if err := site.hook.BeforeBuild(); err != nil {
		return err
	}
	if err := site.buildStatic(ctx); err != nil {
		return err
	}
	if err := site.buildContent(ctx); err != nil {
		return err
	}
	return site.hook.AfterBuild(ctx, site.writer)
}

func WithOption(opt *Option) SiteOption {
	return func(site *Site) {
		site.option = opt
	}
}

func WithHook(h hook.Hook) SiteOption {
	return func(site *Site) {
		site.hook = h
	}
}

func WithWriter(w core.Writer) SiteOption {
	return func(site *Site) {
		site.writer = w
	}
}

func New(ctx *core.Context, opts ...SiteOption) (*Site, error) {
	site := &Site{
		ctx:              ctx,
		store:            NewContentStore(),
		contentProcessor: content.NewProcessor(ctx),
	}
	for _, opt := range opts {
		opt(site)
	}
	if site.option == nil {
		site.option = &Option{
			IncludeDrafts: false,
		}
	}
	if site.writer == nil {
		return nil, errors.New("writer is required")
	}
	if site.tplset == nil {
		tplset, err := template.NewSet(ctx)
		if err != nil {
			return nil, err
		}
		site.tplset = tplset
	}
	if site.hook == nil {
		h, err := hook.New(ctx)
		if err != nil {
			return nil, err
		}
		site.hook = h
	}
	w, err := site.hook.HandleWriter(site.writer)
	if err != nil {
		return nil, err
	}
	site.writer = w

	set, err := site.hook.HandleTemplateSet(site.tplset)
	if err != nil {
		return nil, err
	}
	site.tplset = set
	return site, nil
}
