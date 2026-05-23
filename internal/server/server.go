package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/writer"
)

type (
	Server struct {
		fs         *writer.MemoryWriter
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

func (s *Server) Start(listen string) error {
	if listen == "" {
		listen = s.ctx.GetBaseURL()
	}
	u, err := url.Parse(listen)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", s)

	if s.livereload != nil {
		mux.HandleFunc("/livereload", s.livereload.HandleWS)
		mux.HandleFunc("/livereload.js", s.livereload.HandleJS)
	}

	s.ctx.Logger.Infoln("Listen", u.String(), "...")
	return http.ListenAndServe(u.Host, mux)
}

func Serve(conf *core.Config, listen string, autoload bool, includeDrafts bool) error {
	ctx, err := core.NewContext(conf)
	if err != nil {
		return err
	}

	s := &Server{
		fs:  writer.NewMemoryWriter(),
		ctx: ctx,
	}

	site, err := site.New(ctx, site.IncludeDrafts(includeDrafts))
	if err != nil {
		return err
	}
	if err := site.Build(context.TODO(), s.fs); err != nil {
		return err
	}

	s.site = site
	if autoload {
		s.livereload = &Livereload{
			clients: make(map[*client]struct{}),
		}
		go s.watchFiles()
	}
	return s.Start(listen)
}
