package builder

import (
	"fmt"
	"sync"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"

	_ "github.com/honmaple/snow/builder/page/hook"
)

type Builder interface {
	Build() error
}

func Build(conf *config.Config) error {
	t, err := theme.New(conf)
	if err != nil {
		return err
	}

	hooks := hook.New(conf)

	bs := []Builder{
		page.NewBuilder(conf, t, hooks.PageHooks()),
		static.NewBuilder(conf, t, hooks.StaticHooks()),
	}
	var wg sync.WaitGroup
	for _, b := range bs {
		wg.Add(1)
		go func(b1 Builder) {
			if err := b1.Build(); err != nil {
				fmt.Println(err.Error())
			}
			wg.Done()
		}(b)
	}
	wg.Wait()
	return nil
}
