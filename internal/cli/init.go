package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/honmaple/snow/internal/core"
	"github.com/urfave/cli/v2"
)

var (
	initCommand = &cli.Command{
		Name:   "init",
		Usage:  "init a new site",
		Action: initAction,
	}
)

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
		c      = core.DefaultConfig()
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
	c.Set("mode.publish", map[string]any{
		"site": map[string]any{"url": url},
	})
	return c.WriteConfig()
}
