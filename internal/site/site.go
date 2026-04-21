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
	Site struct {
		ctx           *core.Context
		hook          hook.Hook
		store         *Store
		writer        core.Writer
		tplset        template.TemplateSet
		contentParser *content.ContentParser
	}
	SiteOption func(*Site)
)

func (site *Site) Build() error {
	ctx := context.TODO()
	// if err := site.buildStatic(ctx); err != nil {
	//	return err
	// }
	if err := site.buildContent(ctx); err != nil {
		return err
	}
	return nil
}

func (site *Site) Load() error {
	return site.loadContent()
}

func (site *Site) Reload() error {
	site.store.Reset()
	return site.Load()
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
		ctx: ctx,
	}
	site.store = NewStore()
	site.contentParser = content.NewContentParser(ctx)

	for _, opt := range opts {
		opt(site)
	}
	if site.hook == nil {
		site.hook = hook.HookImpl{}
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
	return site, nil
}
