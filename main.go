package main

import (
	"fmt"
	"os"

	"github.com/honmaple/snow/builder"
	"github.com/honmaple/snow/config"
	"github.com/urfave/cli/v2"
)

const (
	PROCESS     = "snow"
	VERSION     = "0.1.0"
	DESCRIPTION = "snow is a static site generator."
)

var (
	conf = config.DefaultConfig()
)

func before(clx *cli.Context) error {
	path := clx.String("conf")
	return conf.Load(path)
}

func newAction(clx *cli.Context) error {
	return nil
}

func initAction(clx *cli.Context) error {
	return nil
}

func buildAction(clx *cli.Context) error {
	if err := conf.SetMode(clx.String("mode")); err != nil {
		return err
	}
	conf.SetOutput(clx.String("output"))
	return builder.Build(conf)
}

func serveAction(clx *cli.Context) error {
	conf.SetOutput(clx.String("output"))
	return builder.Serve(conf, clx.String("listen"), clx.Bool("autoload"))
}

func main() {
	app := &cli.App{
		Name:    PROCESS,
		Usage:   DESCRIPTION,
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "Load configuration from `FILE`",
			},
		},
		Before: before,
		Commands: []*cli.Command{
			{
				Name:   "new",
				Usage:  "create new page",
				Action: newAction,
			},
			{
				Name:   "init",
				Usage:  "first init",
				Action: initAction,
			},
			{
				Name:  "build",
				Usage: "build and output",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Value:   "",
						Usage:   "Build site with mode",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "output",
						Usage:   "Build output content",
					},
				},
				Action: buildAction,
			},
			{
				Name:  "serve",
				Usage: "serve host",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Value:   "",
						Usage:   "Build site with mode",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "output",
						Usage:   "Build output content",
					},
					&cli.StringFlag{
						Name:    "listen",
						Aliases: []string{"l"},
						Value:   "",
						Usage:   "Listen address",
					},
					&cli.BoolFlag{
						Name:    "autoload",
						Aliases: []string{"r"},
						Usage:   "Autoload when file change",
					},
				},
				Action: serveAction,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
