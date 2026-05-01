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
	_ "github.com/honmaple/snow/internal/site/hook/pelican"
	_ "github.com/honmaple/snow/internal/site/hook/rewrite"
	_ "github.com/honmaple/snow/internal/site/hook/shortcode"
)

const (
	PROCESS     = "snow"
	VERSION     = "0.1.4"
	DESCRIPTION = "snow is a static site generator."
)

var (
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "mode",
			Aliases: []string{"m"},
			Value:   "",
			Usage:   "Build site with special mode",
		},
		&cli.BoolFlag{
			Name:  "include-drafts",
			Usage: "Build site with drafts",
			Value: true,
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"D"},
			Value:   false,
			Usage:   "Enable debug mode",
		},
	}
	conf = core.DefaultConfig()
)

func beforeAction(clx *cli.Context) error {
	return conf.LoadFromFile(clx.String("config"))
}

func commonAction(clx *cli.Context) error {
	if clx.Bool("debug") {
		conf.SetDebug()
	}
	if mode := clx.String("mode"); mode != "" {
		conf.SetMode(mode)
	}
	return nil
}

func Execute() {
	app := &cli.App{
		Name:    PROCESS,
		Usage:   DESCRIPTION,
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "load configuration from `FILE`",
			},
		},
		Before: beforeAction,
		Commands: []*cli.Command{
			initCommand,
			buildCommand,
			serverCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
