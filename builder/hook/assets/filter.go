package assets

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	libsass "github.com/wellington/go-libsass"
	// "github.com/bep/golibsass/libsass"

	"github.com/spf13/cast"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type filterOption map[string]interface{}

func filterOptions(data interface{}) (names []string, opts []filterOption) {
	if data == nil {
		return
	}
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		// - libsass:
		//     path: ""
		// - cssmin:
		for _, item := range data.([]interface{}) {
			for k, v := range cast.ToStringMap(item) {
				opts = append(opts, cast.ToStringMap(v))
				names = append(names, k)
				break
			}
		}
		return
	case reflect.String:
		// libcass,css
		names = strings.Split(data.(string), ",")
		opts = make([]filterOption, len(names))
		return
	default:
		return
	}
}

func (ws *assets) execute(opt option) error {
	var b bytes.Buffer
	for _, file := range opt.files {
		var (
			buf []byte
			err error
		)
		if strings.HasPrefix(file, "@theme/") {
			f, err := ws.theme.Open(file[7:])
			if err != nil {
				return err
			}
			buf, err = ioutil.ReadAll(f)
		} else {
			buf, err = ioutil.ReadFile(file)
		}
		if err != nil {
			return err
		}
		var (
			w = bytes.NewBuffer(nil)
			r = bytes.NewBuffer(buf)
		)
		for i, filter := range opt.filters {
			w.Reset()
			if err := ws.filter(filter, w, r, opt.filterOpts[i]); err != nil {
				return err
			}
			r = w
		}
		b.Write(w.Bytes())
	}
	return ws.conf.WriteOutput(opt.output, &b)
}

func (ws *assets) filter(name string, w io.Writer, r io.Reader, opt filterOption) (err error) {
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

func (ws *assets) libscss(w io.Writer, r io.Reader, opt filterOption) error {
	paths := make([]string, 0)
	if opt != nil {
		paths = cast.ToStringSlice(opt["path"])
	}

	comp, err := libsass.New(w, r, libsass.IncludePaths(paths))
	if err != nil {
		return err
	}
	return comp.Run()
}

func (ws *assets) cssmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Minify("css", w, r)
}

func (ws *assets) jsmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	return m.Minify("js", w, r)
}
