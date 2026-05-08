package minify

import (
	"context"
	"io"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

type MinifyWriter struct {
	core.Writer
	ctx    *core.Context
	minify *minify.M
}

func (w *MinifyWriter) check(ext string) bool {
	if !w.ctx.Config.IsSet("hooks.minify.option" + ext) {
		return true
	}
	return w.ctx.Config.GetBool("hooks.minify.option" + ext)
}

func (w *MinifyWriter) Write(ctx context.Context, file string, r io.Reader) error {
	ext := filepath.Ext(file)
	switch ext {
	case ".html":
		if w.check(ext) {
			r = w.minify.Reader("html", r)
		}
	case ".css":
		if w.check(ext) {
			r = w.minify.Reader("css", r)
		}
	case ".js":
		if w.check(ext) {
			r = w.minify.Reader("js", r)
		}
	}
	return w.Writer.Write(ctx, file, r)
}

func NewWriter(ctx *core.Context, w core.Writer) *MinifyWriter {
	m := minify.New()
	m.AddFunc("html", html.Minify)
	m.AddFunc("css", css.Minify)
	m.AddFunc("js", js.Minify)
	return &MinifyWriter{Writer: w, minify: m, ctx: ctx}
}
