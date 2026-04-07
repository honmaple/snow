package static

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/static/loader"
	"github.com/honmaple/snow/internal/static/types"
)

type (
	Builder struct {
		ctx    *core.Context
		once   sync.Once
		hook   types.Hook
		store  types.Store
		loader types.Loader
		writer core.Writer
	}
	BuilderOption func(*Builder)
)

func (b *Builder) writeStatic(ctx context.Context, static *types.Static) error {
	var (
		src fs.File
		err error
	)

	if strings.HasPrefix(static.File, "@theme/") {
		src, err = b.ctx.Theme.Open(static.File)
	} else {
		src, err = os.Open(static.File)
	}
	if err != nil {
		return err
	}
	defer src.Close()
	return b.writer.Write(ctx, static.Path, src)
}

func (b *Builder) Build(ctx context.Context) error {
	store, err := b.loader.Load()
	if err != nil {
		return err
	}

	b.ctx.Logger.Infof("%d statics", len(store.Statics()))

	b.once.Do(func() {
		b.store = store
	})

	statics := b.hook.HandleStatics(store.Statics())
	for _, static := range statics {
		if err := b.writeStatic(ctx, static); err != nil {
			return err
		}
	}
	return nil
}

func WithLoader(l types.Loader) BuilderOption {
	return func(b *Builder) {
		b.loader = l
	}
}

func WithWriter(w core.Writer) BuilderOption {
	return func(b *Builder) {
		b.writer = w
	}
}

func New(ctx *core.Context, opts ...BuilderOption) (*Builder, error) {
	b := &Builder{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(b)
	}
	if b.writer == nil {
		return nil, errors.New("static writer is required")
	}
	if b.loader == nil {
		b.loader = loader.New(ctx)
	}
	return b, nil
}

func Build(ctx *core.Context, opts ...BuilderOption) error {
	b, err := New(ctx, opts...)
	if err != nil {
		return err
	}
	return b.Build(context.TODO())
}
