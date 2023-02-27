package builder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
	"github.com/urfave/cli/v2"

	_ "github.com/honmaple/snow/builder/page/markup/html"
	_ "github.com/honmaple/snow/builder/page/markup/markdown"
	_ "github.com/honmaple/snow/builder/page/markup/orgmode"

	_ "github.com/honmaple/snow/builder/hook/assets"
	_ "github.com/honmaple/snow/builder/hook/encrypt"
	_ "github.com/honmaple/snow/builder/hook/i18n"
	_ "github.com/honmaple/snow/builder/hook/pelican"
	_ "github.com/honmaple/snow/builder/hook/shortcode"
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
			Name:    "debug",
			Aliases: []string{"D"},
			Value:   false,
			Usage:   "Enable debug mode",
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
		first  bool
		c      = config.DefaultConfig()
	)
	fmt.Printf("Welcome to snow %s.\n", VERSION)
	prompts := Prompts{
		&PromptString{
			Usage:       "> Where do you want to create your new web site? ",
			Value:       name,
			FilePath:    true,
			Destination: &name,
		},
		&PromptString{
			Usage:       "> What will be the title of this web site? ",
			Value:       c.GetString("site.title"),
			Required:    true,
			Destination: &title,
		},
		&PromptString{
			Usage:       "> Who will be the author of this web site? ",
			Value:       c.GetString("site.author"),
			Required:    true,
			Destination: &author,
		},
		&PromptString{
			Usage:       "> What is your URL prefix? (no trailing slash) ",
			Value:       c.GetString("site.url"),
			Required:    true,
			Destination: &url,
		},
		&PromptBool{
			Usage:       "> Do you want to create first page? ",
			Value:       true,
			Destination: &first,
		},
	}

	r := bufio.NewReader(os.Stdin)
	if err := prompts.Excute(r); err != nil {
		return err
	}
	if name != "" && name != "." {
		if err := os.Mkdir(name, 0755); err != nil {
			return err
		}
	}

	if first {
		root := filepath.Join(name, c.GetString("content_dir"), "posts")
		if err := os.MkdirAll(root, 0755); err != nil {
			return err
		}
		file, err := os.Create(filepath.Join(root, "first-page.md"))
		if err != nil {
			return err
		}
		defer file.Close()

		file.WriteString(fmt.Sprintf(`---
title: First Page
date: %s
categories:
 - Linux/Emacs
authors:
 - snow
tags: [linux,emacs,snow]
---

# Hello Snow`, time.Now().Format("2006-01-02 15:04:05")))
	}

	c.SetConfigFile(filepath.Join(name, "config.yaml"))
	c.Set("site.title", title)
	c.Set("site.author", author)
	c.Set("mode.publish", map[string]interface{}{
		"site": map[string]interface{}{"url": url},
	})
	return c.WriteConfig()
}

func commonAction(clx *cli.Context) error {
	if clx.Bool("debug") {
		conf.SetDebug()
	}
	if filter := clx.String("filter"); filter != "" {
		conf.Set("build_filter", filter)
	}
	if err := conf.SetMode(clx.String("mode")); err != nil {
		return err
	}
	conf.SetOutput(clx.String("output"))
	if clx.Bool("clean") {
		conf.Log.Infoln("Removing the contents of", conf.GetOutput())
		return utils.RemoveDir(conf.GetOutput())
	}
	return nil
}

func buildAction(clx *cli.Context) error {
	if clx.Bool("hooks") {
		hook.Print()
		return nil
	}
	if err := commonAction(clx); err != nil {
		return err
	}
	return Build(conf)
}

func serverAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
		return err
	}
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
				Action: initAction,
			},
			{
				Name:  "build",
				Usage: "build and output",
				Flags: append([]cli.Flag{
					&cli.BoolFlag{
						Name:  "hooks",
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
