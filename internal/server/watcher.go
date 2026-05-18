package server

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func (srv *Server) watchDir(watcher *fsnotify.Watcher, path string) error {
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return watcher.Add(p)
		}
		return nil
	})
}

func (srv *Server) watch(paths []string, fn func(string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil && !os.IsExist(err) {
			continue
		}
		if err := srv.watchDir(watcher, path); err != nil {
			srv.ctx.Logger.Errorf("Error watching %s: %v", path, err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) {
				srv.ctx.Logger.Infoln("The", event.Name, "has been modified. Rebuilding...")

				fn(event.Name)
			} else if event.Has(fsnotify.Create) {
				srv.ctx.Logger.Infoln("The", event.Name, "has been created. Rebuilding...")

				fn(event.Name)
			} else if event.Has(fsnotify.Remove) {
				srv.ctx.Logger.Infoln("The", event.Name, "has been removed. Rebuilding...")

				fn(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("监听错误: %v", err)
		}
	}
}

func (srv *Server) watchFiles() {
	contentDir := srv.ctx.GetContentDir()

	paths := []string{
		contentDir, "static", "templates",
	}
	if name := srv.ctx.GetTheme(); name != "" {
		paths = append(paths, filepath.Join("themes", name, "static"))
		paths = append(paths, filepath.Join("themes", name, "templates"))
	}
	if err := srv.watch(paths, func(file string) {
		ctx := context.Background()

		if err := srv.site.Build(ctx); err != nil {
			srv.ctx.Logger.Errorf("Rebuilding err:", err.Error())
			return
		}
		if srv.livereload != nil {
			srv.livereload.Notify("*.html")
		}
	}); err != nil {
		srv.ctx.Logger.Errorf("Watch files err: %s", err.Error())
	}
}
