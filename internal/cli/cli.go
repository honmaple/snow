package cli

import (
	"fmt"
	"os"

	"github.com/honmaple/snow/internal/core"
	"github.com/urfave/cli/v2"

	_ "github.com/honmaple/snow/internal/content/parser/html"
	_ "github.com/honmaple/snow/internal/content/parser/markdown"
	_ "github.com/honmaple/snow/internal/content/parser/orgmode"

	_ "github.com/honmaple/snow/internal/hook/assets"
	_ "github.com/honmaple/snow/internal/hook/encrypt"
	_ "github.com/honmaple/snow/internal/hook/i18n"
	_ "github.com/honmaple/snow/internal/hook/pelican"
	_ "github.com/honmaple/snow/internal/hook/shortcode"
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
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "output",
			Usage:   "Build output content",
		},
		&cli.BoolFlag{
			Name:    "clean",
			Aliases: []string{"C"},
			Value:   false,
			Usage:   "Clean output content",
		},
		&cli.StringFlag{
			Name:    "filter",
			Aliases: []string{"F"},
			Value:   "",
			Usage:   "Filter when build",
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
		conf.Set("debug", true)
	}
	if filter := clx.String("filter"); filter != "" {
		conf.Set("hooks.internal.filter", filter)
	}
	if output := clx.String("output"); output != "" {
		conf.Set("output_dir", output)
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
