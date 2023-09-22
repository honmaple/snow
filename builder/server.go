package builder

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
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
		files      sync.Map
		watcher    *fsnotify.Watcher
		autoload   bool
		watchFiles sync.Map
	}
)

func (m *memoryServer) reset() {
	m.files.Range(func(k, v interface{}) bool {
		m.files.Delete(k)
		return true
	})
}

func (m *memoryServer) Watch(file string) error {
	if m.watcher == nil || !m.autoload {
		return nil
	}

	_, exist := m.watchFiles.LoadOrStore(file, true)
	if !exist {
		return filepath.WalkDir(file, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if file == path || info.IsDir() {
				return m.watcher.Add(path)
			}
			return nil
		})
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
	m.files.Store(file, &memoryFile{bytes.NewReader(buf), time.Now()})
	return nil
}

func (m *memoryServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}
	v, ok := m.files.Load(path)
	if !ok {
		v, ok = m.files.Load("/404.html")
		if !ok {
			http.Error(w, "404", 404)
			return
		}
	}
	file := v.(*memoryFile)

	http.ServeContent(w, r, filepath.Base(path), file.modTime, file.reader)
}

func (m *memoryServer) Build(ctx context.Context) error {
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					m.conf.Log.Infoln("The", event.Name, "has been modified. Rebuilding...")
					m.conf.Cache.Delete(event.Name)
					m.reset()
					if err := Build(m.conf); err != nil {
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
	if err := Build(m.conf); err != nil {
		return err
	}

	watchFiles := make([]string, 0)
	m.watchFiles.Range(func(k, v interface{}) bool {
		watchFiles = append(watchFiles, k.(string))
		return true
	})
	sort.Strings(watchFiles)
	if len(watchFiles) > 0 {
		m.conf.Log.Infoln("Watching", strings.Join(watchFiles, ", "))
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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	m := &memoryServer{
		watcher:  watcher,
		autoload: autoload,
	}
	m.conf = conf.WithWriter(m)

	if err := m.Build(context.Background()); err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", m)

	conf.Log.Infoln("Listen", listen, "...")
	return http.ListenAndServe(u.Host, mux)
}
