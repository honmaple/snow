package cli

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
	"github.com/urfave/cli/v2"
)

var (
	//go:embed init.md
	initFile    embed.FS
	initCommand = &cli.Command{
		Name:   "init",
		Usage:  "Init a new site",
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
			Value:       "Snow",
			Required:    true,
			Destination: &title,
		},
		&PromptString{
			Usage:       "> Who will be the author of this web site? ",
			Value:       "honmaple",
			Required:    true,
			Destination: &author,
		},
		&PromptString{
			Usage:       "> What is your URL prefix? (no trailing slash) ",
			Value:       "http://example.com",
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
		b, err := initFile.ReadFile("init.md")
		if err != nil {
			return err
		}
		root := filepath.Join(name, "content", "posts")
		if err := os.MkdirAll(root, 0755); err != nil {
			return err
		}
		os.WriteFile(filepath.Join(root, "hello-snow.md"), b, 0644)
	}

	conf := core.NewConfig()
	conf.SetConfigFile(filepath.Join(name, "config.yaml"))
	conf.Set("base_url", "http://127.0.0.1:8000")
	conf.Set("title", title)
	conf.Set("description", "")
	conf.Set("author", author)
	conf.Set("taxonomies.tags", map[string]any{
		"path": "/{taxonomy}/",
	})
	conf.Set("modes.publish", map[string]any{
		"base_url": url,
	})
	return conf.WriteConfig()
}
