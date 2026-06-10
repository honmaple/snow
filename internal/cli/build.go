package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	buildCommand = &cli.Command{
		Name:  "build",
		Usage: "Build site",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "load configuration from `FILE`",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"D"},
				Value:   false,
				Usage:   "enable debug mode",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry run",
			},
			&cli.StringFlag{
				Name:    "root-dir",
				Aliases: []string{"r"},
				Value:   ".",
				Usage:   "directory to use as root of project",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Value:   "output",
				Usage:   "build output content",
			},
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"C"},
				Value:   false,
				Usage:   "clean output content",
			},
			&cli.StringFlag{
				Name:    "mode",
				Aliases: []string{"m"},
				Value:   "",
				Usage:   "build site with special mode",
			},
			&cli.BoolFlag{
				Name:  "include-drafts",
				Usage: "build site with content marked as draft",
				Value: false,
			},
		},
		Action: buildAction,
	}
)

func buildAction(clx *cli.Context) error {
	return runInRootDir(clx.String("root-dir"), func() error {
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

			files, err := os.ReadDir(out)
			if err != nil {
				return err
			}
			for _, file := range files {
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}
				if err := os.RemoveAll(filepath.Join(out, file.Name())); err != nil {
					return err
				}
			}
		}

		s, err := site.New(ctx, site.IncludeDrafts(clx.Bool("include-drafts")))
		if err != nil {
			return err
		}

		var w core.Writer
		if clx.Bool("dry-run") {
			w = writer.NewNullWriter()
		} else {
			w = writer.NewDiskWriter(conf.GetString("output_dir"))
		}
		return s.Build(context.TODO(), w)
	})
}

func runInRootDir(root string, fn func() error) error {
	if root == "" || root == "." {
		return fn()
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return err
	}
	defer os.Chdir(wd)

	return fn()
}
