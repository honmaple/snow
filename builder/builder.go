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
		Build(context.Context) error
	}
	Builders []Builder
)

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
	return bs.Build(context.Background())
}

func newBuilder(conf config.Config) (Builder, error) {
	th, err := theme.New(conf)
	if err != nil {
		return nil, err
	}
	hs := hook.New(conf, th)
	return Builders{
		page.NewBuilder(conf, th, hs.PageHooks()),
		static.NewBuilder(conf, th, hs.StaticHooks()),
	}, nil
}
