package cli

import (
	"github.com/honmaple/snow/internal/server"
	"github.com/urfave/cli/v2"
)

var (
	serverCommand = &cli.Command{
		Name:  "server",
		Usage: "Server site",
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
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   "",
				Usage:   "listen address",
			},
			&cli.BoolFlag{
				Name:    "autoload",
				Aliases: []string{"R"},
				Usage:   "autoload when file change",
			},
			&cli.StringFlag{
				Name:    "mode",
				Aliases: []string{"m"},
				Value:   "",
				Usage:   "build site with special mode",
			},
			&cli.BoolFlag{
				Name:  "include-drafts",
				Usage: "include content marked as draft",
				Value: false,
			},
		},
		Action: serverAction,
	}
)

func serverAction(clx *cli.Context) error {
	conf, err := commonAction(clx)
	if err != nil {
		return err
	}
	return server.Serve(conf, clx.String("listen"), clx.Bool("autoload"), clx.Bool("include-drafts"))
}
