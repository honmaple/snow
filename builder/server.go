package builder

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/config"
)

func watchBuilder(b Builder) (*fsnotify.Watcher, error) {
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
				fmt.Println("Autoload", event.Name)
				if event.Op == fsnotify.Write {
					if err := b.Build(nil); err != nil {
						fmt.Println("Build error", err.Error())
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println(err)
			}
		}
	}()
	if err := b.Build(watcher); err != nil {
		return nil, err
	}
	// fmt.Println("Watching", strings.Join(watcher.WatchList(), ", "))
	return watcher, nil
}

func Serve(conf config.Config, listen string, autoload bool) error {
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
	if autoload {
		watcher, err := watchBuilder(b)
		if err != nil {
			return err
		}
		defer watcher.Close()
	} else if err := b.Build(nil); err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(conf.GetOutput())))

	fmt.Println("Listen", listen, "...")
	return http.ListenAndServe(listen, mux)
}
