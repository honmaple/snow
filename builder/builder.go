package builder

import (
	"github.com/honmaple/snow/builder/page"
	// "github.com/honmaple/snow/builder/static"
	// "github.com/honmaple/snow/builder/extra"
	"github.com/honmaple/snow/config"
)

type Builder interface {
	Build() error
}

func Build(conf *config.Config) error {
	registerFilter(conf)

	bs := []Builder{
		page.NewBuilder(conf),
		// extra.NewBuilder(conf),
		// static.NewBuilder(conf),
	}
	for _, b := range bs {
		if err := b.Build(); err != nil {
			return err
		}
	}
	return nil
}
