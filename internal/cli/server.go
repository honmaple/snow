package cli

import (
	"context"
	"slices"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/server"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	serverCommand = &cli.Command{
		Name:  "server",
		Usage: "Server site",
		Flags: slices.Concat(flags, []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   "",
				Usage:   "listen address",
			},
			&cli.BoolFlag{
				Name:    "autoload",
				Aliases: []string{"r"},
				Usage:   "autoload when file change",
			},
		}),
		Action: serverAction,
	}
)

func serverAction(clx *cli.Context) error {
	conf, err := commonAction(clx)
	if err != nil {
		return err
	}

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	memFS := writer.NewMemoryWriter(ctx)

	site, err := site.New(ctx, site.WithWriter(memFS), site.WithOption(&site.Option{
		IncludeDrafts: clx.Bool("include-drafts"),
	}))
	if err != nil {
		return err
	}
	if err := site.Build(context.TODO()); err != nil {
		return err
	}

	srv, err := server.New(ctx, site, memFS, clx.Bool("autoload"))
	if err != nil {
		return err
	}
	return srv.Start(clx.String("listen"))
}
