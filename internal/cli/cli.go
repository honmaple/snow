package cli

import (
	"fmt"
	"os"

	"github.com/honmaple/snow/internal/core"
	"github.com/urfave/cli/v2"

	_ "github.com/honmaple/snow/internal/site/content/parser/html"
	_ "github.com/honmaple/snow/internal/site/content/parser/markdown"
	_ "github.com/honmaple/snow/internal/site/content/parser/orgmode"

	_ "github.com/honmaple/snow/internal/site/template/data"
	_ "github.com/honmaple/snow/internal/site/template/i18n"

	_ "github.com/honmaple/snow/internal/site/hook/assets"
	_ "github.com/honmaple/snow/internal/site/hook/encrypt"
	_ "github.com/honmaple/snow/internal/site/hook/filter"
	_ "github.com/honmaple/snow/internal/site/hook/minify"
	_ "github.com/honmaple/snow/internal/site/hook/pelican"
	_ "github.com/honmaple/snow/internal/site/hook/rewrite"
	_ "github.com/honmaple/snow/internal/site/hook/shortcode"
)

const (
	PROCESS     = "snow"
	VERSION     = "0.1.6"
	DESCRIPTION = "snow is a static site generator."
)

func commonAction(clx *cli.Context) (*core.Config, error) {
	conf := core.DefaultConfig()
	if err := conf.LoadFromFile(clx.String("config")); err != nil {
		return nil, err
	}
	if clx.Bool("debug") {
		conf.SetDebug()
	}
	if mode := clx.String("mode"); mode != "" {
		conf.SetMode(mode)
	}
	return conf, nil
}

func Execute() {
	app := &cli.App{
		Name:    PROCESS,
		Usage:   DESCRIPTION,
		Version: VERSION,
		Commands: []*cli.Command{
			initCommand,
			buildCommand,
			serverCommand,
			hookCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
