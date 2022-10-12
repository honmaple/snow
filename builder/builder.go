package builder

import (
	"embed"
	"fmt"
	"sync"

	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/template"
	"github.com/honmaple/snow/config"
)

type Builder interface {
	Build() error
}

var (
	//go:embed themes
	themeFS embed.FS
)

func Build(conf *config.Config) error {
	tmpl, err := template.New(conf, themeFS)
	if err != nil {
		return err
	}
	_ = tmpl

	bs := []Builder{
		page.NewBuilder(conf, tmpl),
		static.NewBuilder(conf),
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
