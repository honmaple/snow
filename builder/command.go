package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/config"
	"github.com/urfave/cli/v2"

	"bufio"

	"strings"

	_ "github.com/honmaple/snow/builder/page/hook"
	_ "github.com/honmaple/snow/builder/static/hook"
)

const (
	PROCESS     = "snow"
	VERSION     = "0.1.0"
	DESCRIPTION = "snow is a static site generator."
)

var (
	conf = config.DefaultConfig()
)

func before(clx *cli.Context) error {
	path := clx.String("conf")
	return conf.Load(path)
}

func newAction(clx *cli.Context) error {
	return nil
}

func initAction(clx *cli.Context) error {
	name := clx.Args().First()
	if name == "" {
		name = "."
	}

	var (
		url    string
		title  string
		author string
		err    error
	)

	r := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to snow %s.\n", VERSION)
	for {
		fmt.Printf("> Where do you want to create your new web site? [%s]", name)
		name, err = r.ReadString('\n')
		if err != nil {
			return err
		}
		name = strings.TrimSpace(name)
		if name != "" {
			if n := filepath.Clean(name); n != name {
				return fmt.Errorf("site path is not right")
			}
		}
	TITLE:
		fmt.Printf("> What will be the title of this web site? ")
		title, err = r.ReadString('\n')
		if err != nil {
			return err
		}
		title = strings.TrimSpace(title)
		if title == "" {
			fmt.Println("title is required")
			goto TITLE
		}
	AUTHOR:
		fmt.Printf("> Who will be the author of this web site? ")
		author, err = r.ReadString('\n')
		if err != nil {
			return err
		}
		author = strings.TrimSpace(author)
		if author == "" {
			fmt.Println("author is required")
			goto AUTHOR
		}
	URL:
		fmt.Printf("> What is your URL prefix? (see above example; no trailing slash) ")
		url, err = r.ReadString('\n')
		if err != nil {
			return err
		}
		url = strings.TrimSpace(url)
		if url == "" {
			fmt.Println("url is required")
			goto URL
		}
		break
	}
	if name != "" && name != "." {
		if err := os.Mkdir(name, 0755); err != nil {
			return err
		}
	}
	c := config.DefaultConfig()
	c.SetConfigFile(filepath.Join(name, "config.yaml"))
	c.Set("site.url", url)
	c.Set("site.title", title)
	c.Set("site.author", author)
	return c.WriteConfig()
}

func buildAction(clx *cli.Context) error {
	if clx.Bool("listhooks") {
		hook.Print()
		return nil
	}
	if clx.Bool("debug") {
		conf.SetDebug()
	}
	if err := conf.SetMode(clx.String("mode")); err != nil {
		return err
	}
	conf.SetOutput(clx.String("output"))
	return Build(conf)
}

func serveAction(clx *cli.Context) error {
	if err := conf.SetMode(clx.String("mode")); err != nil {
		return err
	}
	conf.SetOutput(clx.String("output"))
	return Serve(conf, clx.String("listen"), clx.Bool("autoload"))
}

func Excute() {
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
				Name:  "build",
				Usage: "build and output",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Value:   "",
						Usage:   "Build site with mode",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "output",
						Usage:   "Build output content",
					},
					&cli.BoolFlag{
						Name:  "listhooks",
						Value: false,
						Usage: "List all hooks",
					},
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"D"},
						Value:   false,
						Usage:   "debug mode",
					},
				},
				Action: buildAction,
			},
			{
				Name:  "serve",
				Usage: "serve host",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Value:   "",
						Usage:   "Build site with mode",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "output",
						Usage:   "Build output content",
					},
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
				},
				Action: serveAction,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
