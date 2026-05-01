package cli

import (
	"io/fs"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	serverCommand = &cli.Command{
		Name:  "server",
		Usage: "server site",
		Flags: append([]cli.Flag{
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
		}, flags...),
		Action: serverAction,
	}
)

func serverAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
		return err
	}

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	w := writer.NewMemoryWriter(ctx)

	opts := []site.SiteOption{
		site.WithWriter(w),
		site.WithOption(&site.Option{
			IncludeDrafts: clx.Bool("include-drafts"),
		}),
	}
	if err := build(ctx, opts...); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if clx.Bool("autoload") {
		paths := []string{
			ctx.GetContentDir(),
		}
		for _, path := range paths {
			err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					return watcher.Add(p)
				}
				return nil
			})
			if err != nil {
				ctx.Logger.Errorf("Error watching %s: %v", path, err)
			}
		}
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op == fsnotify.Write {
						ctx.Logger.Infoln("The", event.Name, "has been modified. Rebuilding...")
						w.Reset()
						if err := build(ctx, opts...); err != nil {
							ctx.Logger.Errorln("Build error", err.Error())
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					ctx.Logger.Errorln("Watch error", err.Error())
				}
			}
		}()
	}
	return core.Serve(ctx, clx.String("listen"), w)
}
