package assets

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bep/golibsass/libsass"
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

func (self *assets) execute(opt option) (string, error) {
	var (
		b bytes.Buffer
		h = md5.New()
	)
	for _, file := range opt.files {
		var matches []string

		if strings.HasPrefix(file, "@theme/_internal/") {
			matches = []string{file}
		} else {
			files, err := filepath.Glob(self.theme.Path(file))
			if err != nil {
				return "", err
			}
			matches = files
		}

		for _, match := range matches {
			var (
				buf []byte
				err error
			)

			if strings.HasPrefix(match, "@theme/") {
				f, err := self.theme.Open(match)
				if err != nil {
					return "", err
				}
				buf, err = ioutil.ReadAll(f)
			} else {
				buf, err = ioutil.ReadFile(match)
			}
			if err != nil {
				return "", err
			}
			var (
				w = bytes.NewBuffer(nil)
				r = bytes.NewBuffer(buf)
			)
			// filters为空时返回原数据
			w.Write(r.Bytes())
			for i, filter := range opt.filters {
				w.Reset()
				if err := self.filter(filter, w, r, opt.filterOpts[i]); err != nil {
					return "", err
				}
				r.Reset()
				r.Write(w.Bytes())
			}
			b.Write(w.Bytes())
		}

	}
	// 边读边写
	if err := self.conf.Write(opt.output, io.TeeReader(&b, h)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (self *assets) filter(name string, w io.Writer, r io.Reader, opt filterOption) (err error) {
	switch name {
	case "libscss":
		err = self.libscss(w, r, opt)
	case "cssmin":
		err = self.cssmin(w, r, opt)
	case "jsmin":
		err = self.jsmin(w, r, opt)
	}
	return err
}

func (self *assets) libscss(w io.Writer, r io.Reader, opt filterOption) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	opts := libsass.Options{}
	if opt != nil {
		paths := make([]string, 0)
		for _, path := range cast.ToStringSlice(opt["path"]) {
			paths = append(paths, self.theme.Path(path))
		}
		opts.IncludePaths = paths
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

func (self *assets) cssmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Minify("css", w, r)
}

func (self *assets) jsmin(w io.Writer, r io.Reader, opt filterOption) error {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	// 多个js文件合并如果没有;会有问题
	defer w.Write([]byte(";"))
	return m.Minify("js", w, r)
}
