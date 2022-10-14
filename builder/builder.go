package builder

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"

	_ "github.com/honmaple/snow/builder/page/hook"
)

type (
	Builder interface {
		Build(*fsnotify.Watcher) error
	}
	Builders []Builder
)

func (bs Builders) Build(watcher *fsnotify.Watcher) error {
	var wg sync.WaitGroup
	for _, b := range bs {
		wg.Add(1)
		go func(builder Builder) {
			if err := builder.Build(watcher); err != nil {
				fmt.Println(err.Error())
			}
			wg.Done()
		}(b)
	}
	wg.Wait()
	return nil
}

func Build(conf config.Config) error {
	b, err := newBuilder(conf)
	if err != nil {
		return err
	}
	return b.Build(nil)
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
