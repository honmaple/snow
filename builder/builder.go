package builder

import (
	"context"
	"fmt"
	"sync"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	Builder interface {
		Dirs() []string
		Build(context.Context) error
	}
	Builders []Builder
)

func (bs Builders) Dirs() []string {
	dirs := make([]string, 0)
	for _, b := range bs {
		dirs = append(dirs, b.Dirs()...)
	}
	return dirs
}

func (bs Builders) Build(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, b := range bs {
		wg.Add(1)
		go func(builder Builder) {
			defer wg.Done()
			if err := builder.Build(ctx); err != nil {
				fmt.Println(err.Error())
			}
		}(b)
	}
	wg.Wait()
	return nil
}

func Build(conf config.Config) error {
	bs, err := newBuilder(conf)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return bs.Build(ctx)
}

func newBuilder(conf config.Config) (Builder, error) {
	t, err := theme.New(conf)
	if err != nil {
		return nil, err
	}

	hooks := hook.New(conf, t)
	return Builders{
		page.NewBuilder(conf, t, hooks.PageHooks()),
		static.NewBuilder(conf, t, hooks.StaticHooks()),
	}, nil
}
