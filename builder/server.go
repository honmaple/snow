package builder

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/config"
)

func watchBuilder(conf config.Config, b Builder, ctx context.Context) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					conf.Log.Infoln("The", event.Name, "has been modified. Rebuilding...")
					if err := b.Build(ctx); err != nil {
						conf.Log.Errorln("Build error", err.Error())
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				conf.Log.Fatalln("Watch error", err.Error())
			}
		}
	}()
	dirs := b.Dirs()
	for _, dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			return nil, err
		}
	}
	if err := b.Build(ctx); err != nil {
		return nil, err
	}
	conf.Log.Infoln("Watching", strings.Join(dirs, ", "))
	return watcher, nil
}

type fileServer struct {
	output string
	server http.Handler
}

func (s *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// 默认index.html会重定向到到./
	if strings.HasSuffix(path, "/index.html") {
		file, err := os.Open(filepath.Join(s.output, filepath.Clean(path)))
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), file)
		return
	}
	s.server.ServeHTTP(w, r)
}

func Server(conf config.Config, listen string, autoload bool) error {
	if listen == "" {
		listen = conf.GetString("site.url")
	}
	if strings.HasPrefix(listen, "http://") {
		listen = listen[7:]
	} else if strings.HasPrefix(listen, "https://") {
		listen = listen[8:]
	}
	b, err := newBuilder(conf)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if autoload {
		watcher, err := watchBuilder(conf, b, ctx)
		if err != nil {
			return err
		}
		defer watcher.Close()
	} else if err := b.Build(ctx); err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", &fileServer{
		output: conf.GetOutput(),
		server: http.FileServer(http.Dir(conf.GetOutput())),
	})

	conf.Log.Infoln("Listen", listen, "...")
	return http.ListenAndServe(listen, mux)
}
