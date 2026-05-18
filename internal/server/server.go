package server

import (
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
)

type (
	Server struct {
		fs         fs.FS
		ctx        *core.Context
		site       *site.Site
		livereload *Livereload
	}
)

func (m *Server) serveError(w http.ResponseWriter, code int, msg string) {
	if msg == "" {
		msg = http.StatusText(code)
	}
	http.Error(w, msg, code)
}

func (m *Server) serve404(w http.ResponseWriter, r *http.Request) {
	file, err := m.fs.Open("/404.html")
	if err != nil {
		if os.IsNotExist(err) {
			m.serveError(w, http.StatusNotFound, "")
			return
		}
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}
	http.ServeContent(w, r, "404.html", stat.ModTime(), seeker)
}

func (m *Server) serveHTML(w http.ResponseWriter, r *http.Request, file fs.File) {
	if m.livereload != nil {
		m.livereload.ServeHTML(w, file)
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
	return
}

func (m *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		path = stdpath.Join(path, "index.html")
	}

	file, err := m.fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			m.serve404(w, r)
			return
		}
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}
	if stat.IsDir() {
		indexFile, err := m.fs.Open(stdpath.Join(path, "index.html"))
		if err != nil {
			m.serve404(w, r)
			return
		}
		defer indexFile.Close()

		m.serveHTML(w, r, indexFile)
		return
	}
	if strings.HasSuffix(path, ".html") {
		m.serveHTML(w, r, file)
		return
	}

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		m.serveError(w, http.StatusInternalServerError, "")
		return
	}
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), seeker)
	return
}

func (srv *Server) Start(listen string) error {
	if listen == "" {
		listen = srv.ctx.GetBaseURL()
	}
	u, err := url.Parse(listen)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", srv)

	if srv.livereload != nil {
		mux.HandleFunc("/livereload", srv.livereload.HandleWS)
		mux.HandleFunc("/livereload.js", srv.livereload.HandleJS)
	}

	srv.ctx.Logger.Infoln("Listen", u.String(), "...")
	return http.ListenAndServe(u.Host, mux)
}

func New(ctx *core.Context, site *site.Site, fs fs.FS, autoload bool) (*Server, error) {
	srv := &Server{
		fs:   fs,
		ctx:  ctx,
		site: site,
	}
	if autoload {
		srv.livereload = &Livereload{
			clients: make(map[*client]struct{}),
		}
		go srv.watchFiles()
	}
	return srv, nil
}
