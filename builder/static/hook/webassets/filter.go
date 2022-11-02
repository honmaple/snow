package webassets

import (
	"io"

	libsass "github.com/wellington/go-libsass"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

func (ws *webassets) filter(name string, w io.Writer, r io.Reader, opt filterOption) (err error) {
	switch name {
	case "libscss":
		err = ws.libscss(w, r, opt)
	case "cssmin":
		err = ws.cssmin(w, r, opt)
	case "jsmin":
		err = ws.jsmin(w, r, opt)
	}
	return err
}

func (ws *webassets) libscss(w io.Writer, r io.Reader, opt filterOption) error {
	paths := make([]string, 0)
	if opt != nil {
		paths = opt.GetStringSlice("path")
	}

	comp, err := libsass.New(w, r, libsass.IncludePaths(paths))
	if err != nil {
		return err
	}
	return comp.Run()
}

func (ws *webassets) cssmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Minify("css", w, r)
}

func (ws *webassets) jsmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	return m.Minify("js", w, r)
}
