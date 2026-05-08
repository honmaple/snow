package cli

import (
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/urfave/cli/v2"
)

var (
	hookCommand = &cli.Command{
		Name:   "hooks",
		Usage:  "list hooks",
		Action: hookAction,
	}
)

func hookAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
		return err
	}

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}
	hook.Print(ctx)
	return nil
}
