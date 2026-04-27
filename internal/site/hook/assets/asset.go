package assets

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/bep/golibsass/libsass"
	"github.com/honmaple/snow/internal/core"
	"github.com/spf13/cast"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type Asset struct {
	Files       []string
	Filters     []map[string]map[string]any
	Output      string
	ShowVersion bool
}

func getFirstKey[K comparable, V any](m map[K]V) (key K, ok bool) {
	for k := range m {
		return k, true
	}
	return key, false
}

func (n *Asset) filter(name string, w io.Writer, r io.Reader, opt map[string]any) (err error) {
	switch name {
	case "libscss":
		err = n.libscss(w, r, opt)
	case "cssmin":
		err = n.cssmin(w, r, opt)
	case "jsmin":
		err = n.jsmin(w, r, opt)
	}
	return err
}

func (n *Asset) libscss(w io.Writer, r io.Reader, opt map[string]any) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	opts := libsass.Options{}
	if opt != nil {
		opts.IncludePaths = cast.ToStringSlice(opt["paths"])
	}

	transpiler, err := libsass.New(opts)
	if err != nil {
		return err
	}

	result, err := transpiler.Execute(string(bs))
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, result.CSS)
	return err
}

func (n *Asset) cssmin(w io.Writer, r io.Reader, _ map[string]any) error {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Minify("css", w, r)
}

func (n *Asset) jsmin(w io.Writer, r io.Reader, _ map[string]any) error {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	// 多个js文件合并如果没有;会有问题
	defer w.Write([]byte(";"))
	return m.Minify("js", w, r)
}

func (n *Asset) Execute(ctx *core.Context) error {
	var (
		b bytes.Buffer
	)
	for _, file := range n.Files {
		matchedFiles, err := fs.Glob(ctx.Theme, file)
		if err != nil {
			return err
		}

		for _, match := range matchedFiles {
			var (
				buf []byte
				err error
			)

			if strings.HasPrefix(match, "@theme/") {
				f, err := ctx.Theme.Open(match)
				if err != nil {
					return err
				}
				buf, err = io.ReadAll(f)
			} else {
				buf, err = os.ReadFile(match)
			}
			if err != nil {
				return err
			}
			var (
				w = bytes.NewBuffer(nil)
				r = bytes.NewBuffer(buf)
			)
			// filters为空时返回原数据
			w.Write(r.Bytes())
			for _, filter := range n.Filters {
				name, ok := getFirstKey(filter)
				if !ok {
					continue
				}
				w.Reset()
				if err := n.filter(name, w, r, filter[name]); err != nil {
					return err
				}
				r.Reset()
				r.Write(w.Bytes())
			}
			b.Write(w.Bytes())
		}

	}
	// 边读边写
	// if err := ctx.Config.Write(opt.Output, io.TeeReader(&b, h)); err != nil {
	//	return "", err
	// }
	return nil
}
