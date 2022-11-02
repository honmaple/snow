package builder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/config"
	"github.com/urfave/cli/v2"

	_ "github.com/honmaple/snow/builder/page/hook"
	_ "github.com/honmaple/snow/builder/static/hook"
)

const (
	PROCESS     = "snow"
	VERSION     = "0.1.0"
	DESCRIPTION = "snow is a static site generator."
)

var (
	conf  = config.DefaultConfig()
	flags = []cli.Flag{
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
			Name:    "debug",
			Aliases: []string{"D"},
			Value:   false,
			Usage:   "debug mode",
		},
	}
)

func before(clx *cli.Context) error {
	path := clx.String("conf")
	return conf.Load(path)
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
	)

	r := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to snow %s.\n", VERSION)
	for {
		fmt.Printf("> Where do you want to create your new web site? [%s] ", name)
		input, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input != "" {
			name = input
		}
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
		fmt.Printf("> What is your URL prefix? (no trailing slash) ")
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

	if clx.Bool("first-page") || clx.Args().Get(1) == "--first-page" {
		root := filepath.Join(name, c.GetString("content_dir"), "posts")
		os.MkdirAll(root, 0755)
		f, err := os.Create(filepath.Join(root, "first-page.md"))
		if err != nil {
			return err
		}
		defer f.Close()

		content := fmt.Sprintf(`title: First Page
author: snow
date: %s
category: Linux
tags: linux,emacs

# Hello World`, time.Now().Format("2006-01-02T15:04:05"))
		f.WriteString(content)
	}
	c.SetConfigFile(filepath.Join(name, "config.yaml"))
	c.Set("site.title", title)
	c.Set("site.author", author)
	c.Set("mode.publish", map[string]interface{}{
		"site": map[string]interface{}{"url": url},
	})
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

func serverAction(clx *cli.Context) error {
	if clx.Bool("debug") {
		conf.SetDebug()
	}
	if err := conf.SetMode(clx.String("mode")); err != nil {
		return err
	}
	conf.SetOutput(clx.String("output"))
	return Server(conf, clx.String("listen"), clx.Bool("autoload"))
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
				Usage:   "load configuration from `FILE`",
			},
		},
		Before: before,
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "init a new site",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "first-page",
						Usage: "create first example page",
					},
				},
				Action: initAction,
			},
			{
				Name:  "build",
				Usage: "build and output",
				Flags: append([]cli.Flag{
					&cli.BoolFlag{
						Name:  "listhooks",
						Usage: "List all hooks",
					},
				}, flags...),
				Action: buildAction,
			},
			{
				Name:  "server",
				Usage: "server local files",
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
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
