package builder

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/honmaple/snow/config"
)

type (
	memoryFile struct {
		reader  io.ReadSeeker
		modTime time.Time
	}
	memoryServer struct {
		mu         sync.RWMutex
		conf       config.Config
		cache      sync.Map
		watcher    *fsnotify.Watcher
		autoload   bool
		watchFiles []string
		watchExist map[string]bool
	}
)

func (m *memoryServer) Close() error {
	if m.watcher == nil {
		return nil
	}
	return m.watcher.Close()
}

func (m *memoryServer) Watch(file string) error {
	if m.watcher == nil || !m.autoload {
		return nil
	}
	m.mu.RLock()
	exist := m.watchExist[file]
	m.mu.RUnlock()

	if !exist {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.watchFiles = append(m.watchFiles, file)
		return m.watcher.Add(file)
	}
	return nil
}

func (m *memoryServer) Write(file string, r io.Reader) error {
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	// TODO: large file handle
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	m.cache.Store(file, &memoryFile{bytes.NewReader(buf), time.Now()})
	return nil
}

func (m *memoryServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}
	v, ok := m.cache.Load(path)
	if !ok {
		v, ok = m.cache.Load("/404.html")
		if !ok {
			http.Error(w, "404", 404)
			return
		}
	}
	file := v.(*memoryFile)

	http.ServeContent(w, r, filepath.Base(path), file.modTime, file.reader)
}

func (m *memoryServer) Build(ctx context.Context) error {
	b, err := newBuilder(m.conf)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					m.conf.Log.Infoln("The", event.Name, "has been modified. Rebuilding...")
					if err := b.Build(ctx); err != nil {
						m.conf.Log.Errorln("Build error", err.Error())
					}
				}
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				m.conf.Log.Fatalln("Watch error", err.Error())
			}
		}
	}()
	if err := b.Build(ctx); err != nil {
		return err
	}
	if len(m.watchFiles) > 0 {
		m.conf.Log.Infoln("Watching", strings.Join(m.watchFiles, ", "))
	}
	return nil
}

func Server(conf config.Config, listen string, autoload bool) error {
	if listen == "" {
		listen = conf.Site.URL
	}
	u, err := url.Parse(listen)
	if err != nil {
		return err
	}

	mem := newServer(conf, autoload)
	defer mem.Close()

	if err := mem.Build(context.Background()); err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", mem)

	conf.Log.Infoln("Listen", listen, "...")
	return http.ListenAndServe(u.Host, mux)
}

func newServer(conf config.Config, autoload bool) *memoryServer {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		conf.Log.Error(err.Error())
	}
	m := &memoryServer{
		watcher:  watcher,
		autoload: autoload,
	}
	m.conf = conf.WithWriter(m)
	return m
}
