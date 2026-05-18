package cli

import (
	"context"
	"slices"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/utils"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	buildCommand = &cli.Command{
		Name:  "build",
		Usage: "Build site",
		Flags: slices.Concat(flags, []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry run",
			},
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"C"},
				Value:   false,
				Usage:   "clean output content",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Value:   "output",
				Usage:   "build output content",
			},
		}),
		Action: buildAction,
	}
)

func buildAction(clx *cli.Context) error {
	conf, err := commonAction(clx)
	if err != nil {
		return err
	}

	if out := clx.String("output-dir"); out != "" {
		conf.Set("output_dir", out)
	}

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	if out := ctx.GetOutputDir(); out != "" && clx.Bool("clean") {
		ctx.Logger.Infoln("Removing the contents of", out)
		if err := utils.RemoveDir(out); err != nil {
			return err
		}
	}

	var w core.Writer
	if clx.Bool("dry-run") {
		w = writer.NewNullWriter(ctx)
	} else {
		w = writer.NewDiskWriter(ctx)
	}

	site, err := site.New(ctx, site.WithWriter(w), site.WithOption(&site.Option{
		IncludeDrafts: clx.Bool("include-drafts"),
	}))
	if err != nil {
		return err
	}
	return site.Build(context.TODO())
}
