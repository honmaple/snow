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
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"d"},
				Usage:   "dry run",
			},
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"C"},
				Value:   false,
				Usage:   "Clean output content",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Value:   "output",
				Usage:   "Build output content",
			},
		}, flags...),
		Action: buildAction,
	}
)

func buildAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
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
		w = writer.NewDebugWriter(ctx)
	} else {
		w = writer.NewDiskWriter(ctx)
	}

	siteOpt := &site.Option{
		IncludeDrafts: clx.Bool("include-drafts"),
	}
	return build(ctx, site.WithWriter(w), site.WithOption(siteOpt))
}

func build(ctx *core.Context, opts ...site.SiteOption) error {
	site, err := site.New(ctx, opts...)
	if err != nil {
		return err
	}
	if err := site.Load(); err != nil {
		return err
	}
	return site.Build()
}
