package loader

import (
	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/content/types"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/utils"
)

type (
	DiskLoader struct {
		ctx        *core.Context
		hook       types.Hook
		parser     parser.Parser
		parserExts map[string]bool

		pages      *utils.Slice[*types.Page]
		sections   *utils.Slice[*types.Section]
		taxonomies *utils.Slice[*types.Taxonomy]

		taxonomyTermMap map[string]*types.TaxonomyTerm
	}
	LoaderOption func(*DiskLoader)
)

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
		ctx:             ctx,
		pages:           utils.NewSlice[*types.Page](),
		sections:        utils.NewSlice[*types.Section](),
		taxonomies:      utils.NewSlice[*types.Taxonomy](),
		taxonomyTermMap: make(map[string]*types.TaxonomyTerm),
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.parser == nil {
		d.parser = parser.New(ctx)
	}

	d.parserExts = make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		d.parserExts[ext] = true
	}
	return d
}
