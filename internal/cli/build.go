package cli

import (
	"context"

	"github.com/honmaple/snow/internal/content"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/static"
	"github.com/honmaple/snow/internal/utils"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	buildCommand = &cli.Command{
		Name:  "build",
		Usage: "build site",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "hooks",
				Usage: "List all hooks",
			},
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
				Name:  "drafts",
				Usage: "Build with drafts",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry run",
			},
		},
		Action: buildAction,
	}
)

func buildAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
		return err
	}

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	if out := conf.GetString("output_dir"); out != "" && clx.Bool("clean") {
		ctx.Logger.Infoln("Removing the contents of", out)
		return utils.RemoveDir(out)
	}

	var w core.Writer
	if clx.Bool("dry-run") {
		w = writer.NewDebugWriter(ctx)
	} else {
		w = writer.NewDebugWriter(ctx)
	}
	return build(ctx, w)
}

func build(ctx *core.Context, w core.Writer) error {
	staticBuilder, err := static.New(ctx, static.WithWriter(w))
	if err != nil {
		return err
	}

	contentBuilder, err := content.New(ctx, content.WithWriter(w))
	if err != nil {
		return err
	}
	return core.Build(context.TODO(), staticBuilder, contentBuilder)
}
