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
	// pongo2模版不支持单个实例注册filter或者tag，所以不支持多语言多主题
	th, err := theme.New(conf)
	if err != nil {
		return err
	}
	hs := hook.New(conf, th)

	bs := make(Builders, 0)
	for _, langc := range conf.Languages {
		bs = append(bs, page.NewBuilder(*langc, th, hs.PageHooks()))
		bs = append(bs, static.NewBuilder(*langc, th, hs.StaticHooks()))
	}
	return bs.Build(context.Background())
}
