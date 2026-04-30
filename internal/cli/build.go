package cli

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/utils"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	buildCommand = &cli.Command{
		Name:  "build",
		Usage: "build site",
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "hooks",
				Usage: "List all hooks",
			},
			// &cli.StringFlag{
			//	Name:    "mode",
			//	Aliases: []string{"m"},
			//	Value:   "",
			//	Usage:   "Build site with special mode",
			// },
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
			// &cli.StringFlag{
			//	Name:    "filter",
			//	Aliases: []string{"F"},
			//	Value:   "",
			//	Usage:   "Filter when build",
			// },
			&cli.BoolFlag{
				Name:  "drafts",
				Usage: "Build with drafts",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry run",
			},
		}, flags...),
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

	if out := ctx.GetOutputDir(); out != "" && clx.Bool("clean") {
		ctx.Logger.Infoln("Removing the contents of", out)
		if err := utils.RemoveDir(out); err != nil {
			return err
		}
	}

	var w core.Writer
	if clx.Bool("dry-run") {
		w = writer.NewDebugWriter(ctx)
	} else {
		w = writer.NewDiskWriter(ctx)
	}
	return build(ctx, w)
}

func build(ctx *core.Context, w core.Writer) error {
	site, err := site.New(ctx, site.WithWriter(w))
	if err != nil {
		return err
	}
	if err := site.Load(); err != nil {
		return err
	}
	return site.Build()
}
