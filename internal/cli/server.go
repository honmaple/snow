package cli

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/internal/content"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/static"
	"github.com/honmaple/snow/internal/writer"
	"github.com/urfave/cli/v2"
)

var (
	serverCommand = &cli.Command{
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
	}
)

type Server struct {
	ctx *core.Context
	w   *writer.MemoryWriter
}

func (m *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}
	file, ok := m.w.Find(path)
	if !ok {
		file, ok = m.w.Find("/404.html")
		if !ok {
			http.Error(w, "404", 404)
			return
		}
	}

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "404", 404)
		return
	}
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "404", 404)
		return
	}

	http.ServeContent(w, r, filepath.Base(path), stat.ModTime(), seeker)
}

func serverAction(clx *cli.Context) error {
	if err := commonAction(clx); err != nil {
		return err
	}

	listen := clx.String("listen")

	u, err := url.Parse(listen)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	w := writer.NewMemoryWriter(ctx)

	staticBuilder, err := static.New(ctx, static.WithWriter(w))
	if err != nil {
		return err
	}

	contentBuilder, err := content.New(ctx, content.WithWriter(w))
	if err != nil {
		return err
	}
	if err := core.Build(context.TODO(), staticBuilder, contentBuilder); err != nil {
		return err
	}

	srv := &Server{
		ctx: ctx,
		w:   w,
	}

	mux := http.NewServeMux()
	mux.Handle("/", srv)

	ctx.Logger.Infoln("Listen", listen, "...")
	return http.ListenAndServe(u.Host, mux)
}
