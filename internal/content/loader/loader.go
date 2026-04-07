package loader

import (
	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/utils"
)

type (
	DiskLoader struct {
		ctx    *core.Context
		hook   types.Hook
		parser parser.Parser

		pages         *utils.Slice[*types.Page]
		sections      *utils.Slice[*types.Section]
		taxonomies    *utils.Slice[*types.Taxonomy]
		taxonomyTerms *utils.Slice[*types.TaxonomyTerm]
	}
	LoaderOption func(*DiskLoader)
)

func (d *DiskLoader) Load() (types.Store, error) {
	if _, err := d.loadContents(); err != nil {
		return nil, err
	}
	return d, nil
}

func WithHook(h types.Hook) LoaderOption {
	return func(d *DiskLoader) {
		d.hook = h
	}
}

func WithParser(p parser.Parser) LoaderOption {
	return func(d *DiskLoader) {
		d.parser = p
	}
}

func New(ctx *core.Context, opts ...LoaderOption) *DiskLoader {
	d := &DiskLoader{
		ctx:           ctx,
		pages:         utils.NewSlice[*types.Page](),
		sections:      utils.NewSlice[*types.Section](),
		taxonomies:    utils.NewSlice[*types.Taxonomy](),
		taxonomyTerms: utils.NewSlice[*types.TaxonomyTerm](),
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.parser == nil {
		d.parser = parser.New(ctx)
	}
	return d
}
