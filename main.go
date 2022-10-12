package main

import (
	"fmt"
	"os"

	"github.com/honmaple/snow/builder"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/server"
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

func before(ctx *cli.Context) error {
	path := ctx.String("conf")
	return conf.Load(path)
}

func newAction(ctx *cli.Context) error {
	return nil
}

func initAction(ctx *cli.Context) error {
	return nil
}

func buildAction(ctx *cli.Context) error {
	return builder.Build(conf)
}

func serveAction(ctx *cli.Context) error {
	return server.Serve(conf)
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
				Name:   "build",
				Usage:  "build and output",
				Action: buildAction,
			},
			{
				Name:   "serve",
				Usage:  "serve host",
				Action: serveAction,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
