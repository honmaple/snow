package builder

import (
	"fmt"
	"sync"

	// "github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/static"
	"github.com/honmaple/snow/builder/template"
	"github.com/honmaple/snow/config"
)

type (
	Builder interface {
		Build() error
	}
)

func Build(conf *config.Config) error {
	tmpl := template.New(conf)
	_ = tmpl

	bs := []Builder{
		// page.NewBuilder(conf, tmpl),
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
