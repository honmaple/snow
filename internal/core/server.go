package core

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"net/url"
	filepath "path"
)

type Server struct {
	fs  fs.FS
	ctx *Context
}

func (m *Server) serve404(w http.ResponseWriter, r *http.Request) {
	file, err := m.fs.Open("/404.html")
	if err == nil {
		http.Error(w, "404", 404)
		return
	}
	defer file.Close()

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
	http.ServeContent(w, r, "404.html", stat.ModTime(), seeker)
}

func (m *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}
	file, err := m.fs.Open(path)
	if err != nil {
		m.serve404(w, r)
		return
	}

	stat, err := file.Stat()
	if err != nil {
		m.serve404(w, r)
		return
	}
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		m.serve404(w, r)
		return
	}

	http.ServeContent(w, r, filepath.Base(path), stat.ModTime(), seeker)
}

func Serve(ctx *Context, listen string, fs fs.FS) error {
	u, err := url.Parse(listen)
	if err != nil {
		return err
	}
	srv := &Server{ctx: ctx, fs: fs}

	mux := http.NewServeMux()
	mux.Handle("/", srv)

	ctx.Logger.Infoln("Listen", u.String(), "...")
	return http.ListenAndServe(u.Host, mux)
}
