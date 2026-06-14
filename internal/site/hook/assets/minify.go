package assets

import (
	"fmt"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

const (
	assetFilterCSSMin = "cssmin"
	assetFilterJSMin  = "jsmin"
)

type MinifyFilter struct {
	name string
}

func (f *MinifyFilter) Name() string {
	return f.name
}

func (f *MinifyFilter) cssmin(buf []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Bytes("css", buf)
}

func (f *MinifyFilter) jsmin(buf []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	return m.Bytes("js", buf)
}

func (f *MinifyFilter) Execute(buf []byte) (result []byte, err error) {
	switch f.name {
	case assetFilterCSSMin:
		result, err = f.cssmin(buf)
	case assetFilterJSMin:
		result, err = f.jsmin(buf)
	default:
		err = fmt.Errorf("unknown assets filter %q", f.name)
	}
	return result, err
}
